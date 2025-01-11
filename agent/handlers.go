package agent

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/pelletier/go-toml"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	authentication "Dana/agent/Auth"
	"Dana/agent/model"
	"Dana/agent/notification"
	"Dana/config"
	"Dana/internal/snmp"
)

func (a *Server) HealthCheck(ctx echo.Context) error {
	ctx.Logger().Info("HealthCheck endpoint called")
	return ctx.JSON(200, "OK")
}

func (a *Server) Register(ctx echo.Context) error {
	ctx.Logger().Info("Register endpoint called")
	user := &model.User{}
	if err := ctx.Bind(user); err != nil {
		ctx.Logger().Error("Error binding request: ", err)
		return ctx.JSON(400, errors.New("invalid request"))
	}
	if err := a.UserRepo.AddUser(ctx.Request().Context(), user); err != nil {
		ctx.Logger().Error("Error adding user: ", err)
		return ctx.JSON(500, "internal server error")
	}
	ctx.Logger().Info("User registered successfully")
	return ctx.JSON(200, "OK")
}

func (a *Server) Login(ctx echo.Context) error {
	ctx.Logger().Info("Login endpoint called")
	user := &model.User{}
	if err := ctx.Bind(user); err != nil {
		ctx.Logger().Error("Error binding request: ", err)
		return ctx.JSON(400, errors.New("invalid request"))
	}
	if err := a.UserRepo.UserAuth(ctx.Request().Context(), user.Username, user.Password); err != nil {
		ctx.Logger().Error("Authentication failed: ", err)
		return ctx.JSON(401, "unauthorized")
	}
	token, err := authentication.GenerateJWT(user.Username)
	if err != nil {
		ctx.Logger().Error("Error generating token: ", err)
		return ctx.JSON(500, "internal server error")
	}
	ctx.Logger().Info("User logged in successfully")
	return ctx.JSON(200, token)
}

func (a *Server) Query(ctx echo.Context) error {
	ctx.Logger().Info("Query endpoint called")
	status, header, body := a.proxyRequest(ctx, "/query", "GET")
	ctx.Logger().Infof("Query completed with status: %d", status)
	return ctx.Blob(status, header, body)
}

func (a *Server) PostInput(ctx echo.Context) error {
	ctx.Logger().Info("PostInput endpoint called")
	inputData := &model.HandlerInput{}
	if err := ctx.Bind(inputData); err != nil {
		ctx.Logger().Error("Error binding input data: ", err)
		return ctx.JSON(400, errors.New("invalid request"))
	}
	inputData.Type = ctx.Param("type")
	if err := a.InputRepo.AddServerInput(ctx.Request().Context(), inputData); err != nil {
		ctx.Logger().Error("Error adding server input: ", err)
		return ctx.JSON(500, "internal server error")
	}
	tomll, err := ConvertMapToTOML(inputData.Data, inputData.Type)
	if err != nil {
		ctx.Logger().Error("Error converting data to TOML: ", err)
		return ctx.JSON(500, "internal server error")
	}
	err = appendToFile("~/.Dana2/Dana2.conf", string(tomll))
	if err != nil {
		ctx.Logger().Error("Error appending data to file: ", err)
		return err
	}
	newConfig := config.NewConfig()
	if err := newConfig.LoadConfigData(tomll); err != nil {
		ctx.Logger().Error("Error loading config data: ", err)
		return ctx.JSON(500, errors.New("internal server error"))
	}
	ctx.Logger().Info("Config data loaded successfully")
	for _, input := range newConfig.Inputs {
		// Share the snmp translator setting with plugins that need it.
		if tp, ok := input.Input.(snmp.TranslatorPlugin); ok {
			tp.SetTranslator(newConfig.Agent.SnmpTranslator)
		}
		err := input.Init()
		if err != nil {
			return fmt.Errorf("could not initialize input %s: %w", input.LogName(), err)
		}
	}
	iu := &inputUnit{
		dst:    a.InputDstChan,
		inputs: newConfig.Inputs,
	}
	a.runInputs(ctx.Request().Context(), a.StartTime, iu)
	ctx.Logger().Info("Inputs processed successfully")
	return ctx.JSON(200, "OK")
}

func appendToFile(filePath string, textToAppend string) error {
	filePath = expandHomeDir(filePath)
	// Open the file in append mode, create it if it doesn't exist
	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			return
		}
	}(file)

	// Write the text to the file followed by a newline
	_, err = file.WriteString(textToAppend + "\n")
	if err != nil {
		return err
	}

	return nil
}

func expandHomeDir(path string) string {
	if path[:2] == "~/" {
		usr, _ := user.Current()
		homeDir := usr.HomeDir
		return filepath.Join(homeDir, path[2:])
	}
	return path
}

func (a *Server) GetInput(ctx echo.Context) error {
	ctx.Logger().Info("GetInput endpoint called")
	inputs, err := a.InputRepo.GetServers(ctx.Request().Context())
	if err != nil {
		ctx.Logger().Error("Error retrieving inputs: ", err)
		return ctx.JSON(500, "internal server error")
	}
	ctx.Logger().Info("Inputs retrieved successfully")
	return ctx.JSON(200, inputs)
}

func (a *Server) GetInputByType(ctx echo.Context) error {
	ctx.Logger().Info("GetInputByType endpoint called")
	inputType := ctx.Param("type")
	inputs, err := a.InputRepo.GetServersByType(ctx.Request().Context(), inputType)
	if err != nil {
		ctx.Logger().Error("Error retrieving inputs by type: ", err)
		return ctx.JSON(500, "internal server error")
	}
	ctx.Logger().Info("Inputs by type retrieved successfully")
	return ctx.JSON(200, inputs)
}

func (a *Server) CreateDashboard(ctx echo.Context) error {
	ctx.Logger().Info("CreateDashboard endpoint called")
	dashboard := &model.Dashboard{}
	if err := ctx.Bind(dashboard); err != nil {
		ctx.Logger().Error("Error binding dashboard data: ", err)
		return ctx.JSON(400, errors.New("invalid request"))
	}
	id, err := a.DashboardRepo.CreateDashboard(ctx.Request().Context(), dashboard)
	if err != nil {
		ctx.Logger().Error("Error creating dashboard: ", err)
		return ctx.JSON(500, "internal server error")
	}
	ctx.Logger().Info("Dashboard created successfully")
	return ctx.JSON(201, map[string]interface{}{"id": id})
}

func (a *Server) GetDashboard(ctx echo.Context) error {
	ctx.Logger().Info("GetDashboard endpoint called")
	id := ctx.Param("id")
	dashboard, err := a.DashboardRepo.GetDashboard(ctx.Request().Context(), id)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			ctx.Logger().Warn("Dashboard not found")
			return ctx.JSON(404, "dashboard not found")
		}
		ctx.Logger().Error("Error retrieving dashboard: ", err)
		return ctx.JSON(500, "internal server error")
	}
	ctx.Logger().Info("Dashboard retrieved successfully")
	return ctx.JSON(200, dashboard)
}

func (a *Server) UpdateDashboard(ctx echo.Context) error {
	ctx.Logger().Info("UpdateDashboard endpoint called")
	dashboard := &model.Dashboard{}
	if err := ctx.Bind(dashboard); err != nil {
		ctx.Logger().Error("Error binding dashboard data: ", err)
		return ctx.JSON(400, errors.New("invalid request"))
	}
	id, _ := primitive.ObjectIDFromHex(ctx.Param("id"))
	if err := a.DashboardRepo.UpdateDashboard(ctx.Request().Context(), dashboard, id); err != nil {
		ctx.Logger().Error("Error updating dashboard: ", err)
		return ctx.JSON(500, "internal server error")
	}
	ctx.Logger().Info("Dashboard updated successfully")
	return ctx.JSON(200, "OK")
}

func (a *Server) DeleteDashboard(ctx echo.Context) error {
	ctx.Logger().Info("DeleteDashboard endpoint called")
	id := ctx.Param("id")
	if err := a.DashboardRepo.DeleteDashboard(ctx.Request().Context(), id); err != nil {
		ctx.Logger().Error("Error deleting dashboard: ", err)
		return ctx.JSON(500, "internal server error")
	}
	ctx.Logger().Info("Dashboard deleted successfully")
	return ctx.JSON(200, "OK")
}

func (a *Server) GetDashboards(ctx echo.Context) error {
	ctx.Logger().Info("GetDashboards endpoint called")
	dashboards, err := a.DashboardRepo.GetDashboards(ctx.Request().Context())
	if err != nil {
		ctx.Logger().Error("Error retrieving dashboards: ", err)
		return ctx.JSON(500, "internal server error")
	}
	ctx.Logger().Info("Dashboards retrieved successfully")
	return ctx.JSON(200, dashboards)
}

func (a *Server) CreateFolder(ctx echo.Context) error {
	folder := &model.Folder{}
	if err := ctx.Bind(folder); err != nil {
		ctx.Logger().Error("CreateFolder: Invalid request", "error", err)
		return ctx.JSON(400, errors.New("invalid request"))
	}
	id, err := a.FolderRepo.CreateFolder(ctx.Request().Context(), folder)
	if err != nil {
		ctx.Logger().Error("CreateFolder: Failed to create folder", "error", err)
		return ctx.JSON(500, "internal server error")
	}
	ctx.Logger().Info("CreateFolder: Folder created", "id", id)
	return ctx.JSON(201, map[string]interface{}{"id": id})
}

func (a *Server) GetFolder(ctx echo.Context) error {
	id := ctx.Param("id")
	folder, err := a.FolderRepo.GetFolder(ctx.Request().Context(), id)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			ctx.Logger().Warn("GetFolder: Folder not found", "id", id)
			return ctx.JSON(404, "folder not found")
		}
		ctx.Logger().Error("GetFolder: Internal server error", "error", err)
		return ctx.JSON(500, "internal server error")
	}
	ctx.Logger().Info("GetFolder: Folder retrieved", "id", id)
	return ctx.JSON(200, folder)
}

func (a *Server) UpdateDashboardInFolder(ctx echo.Context) error {
	folderID := ctx.Param("folderID")
	dashboardID := ctx.Param("dashboardID")
	dashboard := &model.Dashboard{}
	if err := ctx.Bind(dashboard); err != nil {
		ctx.Logger().Error("UpdateDashboardInFolder: Invalid request", "error", err)
		return ctx.JSON(400, errors.New("invalid request"))
	}
	if err := a.FolderRepo.UpdateDashboardInFolder(ctx.Request().Context(), folderID, dashboardID, dashboard); err != nil {
		ctx.Logger().Error("UpdateDashboardInFolder: Failed to update dashboard", "error", err)
		return ctx.JSON(500, "internal server error")
	}
	ctx.Logger().Info("UpdateDashboardInFolder: Dashboard updated", "folderID", folderID, "dashboardID", dashboardID)
	return ctx.JSON(200, "OK")
}

func (a *Server) DeleteFolder(ctx echo.Context) error {
	id := ctx.Param("id")
	if err := a.FolderRepo.DeleteFolder(ctx.Request().Context(), id); err != nil {
		ctx.Logger().Error("DeleteFolder: Failed to delete folder", "error", err)
		return ctx.JSON(500, "internal server error")
	}
	ctx.Logger().Info("DeleteFolder: Folder deleted", "id", id)
	return ctx.JSON(200, "OK")
}

func (a *Server) GetFolders(ctx echo.Context) error {
	folders, err := a.FolderRepo.GetFolders(ctx.Request().Context())
	if err != nil {
		ctx.Logger().Error("GetFolders: Failed to retrieve folders", "error", err)
		return ctx.JSON(500, "internal server error")
	}
	ctx.Logger().Info("GetFolders: Folders retrieved")
	return ctx.JSON(200, folders)
}

func (a *Server) AddNotification(ctx echo.Context) error {
	n := &model.Notification{}
	if err := ctx.Bind(n); err != nil {
		ctx.Logger().Error("AddNotification: Invalid request", "error", err)
		return ctx.JSON(400, errors.New("invalid request"))
	}
	if err := a.NotificationRepo.CreateNotification(ctx.Request().Context(), n); err != nil {
		ctx.Logger().Error("AddNotification: Failed to add notification", "error", err)
		return ctx.JSON(500, "internal server error")
	}
	ctx.Logger().Info("AddNotification: Notification added")
	return ctx.JSON(200, "OK")
}

func (a *Server) GetNotification(ctx echo.Context) error {
	channelName := ctx.Param("channelName")
	n, err := a.NotificationRepo.GetNotification(ctx.Request().Context(), channelName)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			ctx.Logger().Warn("GetNotification: Notification not found", "channelName", channelName)
			return ctx.JSON(404, "notification not found")
		}
		ctx.Logger().Error("GetNotification: Internal server error", "error", err)
		return ctx.JSON(500, "internal server error")
	}
	ctx.Logger().Info("GetNotification: Notification retrieved", "channelName", channelName)
	return ctx.JSON(200, n)
}

func (a *Server) DeleteNotification(ctx echo.Context) error {
	channelName := ctx.Param("channelName")
	if err := a.NotificationRepo.DeleteNotification(ctx.Request().Context(), channelName); err != nil {
		ctx.Logger().Error("DeleteNotification: Failed to delete notification", "error", err)
		return ctx.JSON(500, "internal server error")
	}
	ctx.Logger().Info("DeleteNotification: Notification deleted", "channelName", channelName)
	return ctx.JSON(200, "OK")
}

func (a *Server) SendNotification(ctx echo.Context) error {
	notif := new(model.Notification)

	if err := ctx.Bind(notif); err != nil {
		ctx.Logger().Error("SendNotification: Invalid input", "error", err)
		return ctx.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid input",
		})
	}

	notif.ChannelName = ctx.QueryParam("channelName")

	n, err := a.NotificationRepo.GetNotification(ctx.Request().Context(), ctx.QueryParam("channelName"))
	if err != nil {
		ctx.Logger().Error("SendNotification: Failed to retrieve notification", "error", err)
		return ctx.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to retrieve notification",
		})
	}

	if strings.Contains(notif.ChannelName, "telegram") {
		notification.SendNotification("https://api.telegram.org", a.Config.ServerConfig.TelegramToken, "checkname: "+notif.CheckName+"\n"+"level: "+notif.Level+"\n"+"message: "+notif.Message, int64(n.ChatID))
		ctx.Logger().Info("SendNotification: Notification sent to Telegram", "chatID", n.ChatID)
	} else if strings.Contains(notif.ChannelName, "bale") {
		notification.SendNotification("https://tapi.bale.ai", a.Config.ServerConfig.BaleToken, "checkname: "+notif.CheckName+"\n"+"level: "+notif.Level+"\n"+"message: "+notif.Message, int64(n.ChatID))
		ctx.Logger().Info("SendNotification: Notification sent to Bale", "chatID", n.ChatID)
	} else {
		ctx.Logger().Warn("SendNotification: Invalid channel name", "channelName", notif.ChannelName)
		return ctx.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid channel name",
		})
	}

	return ctx.JSON(http.StatusOK, notif)
}

func (a *Server) NotificationEndpointsGet(ctx echo.Context) error {
	status, header, body := a.proxyRequest(ctx, "/api/v2/notificationEndpoints", "GET")
	ctx.Logger().Info("NotificationEndpoints: Proxy request completed", "status", status)
	return ctx.Blob(status, header, body)
}

func (a *Server) NotificationRulesGet(ctx echo.Context) error {
	status, header, body := a.proxyRequest(ctx, "/api/v2/notificationRules", "GET")
	ctx.Logger().Info("NotificationRules: Proxy request completed", "status", status)
	return ctx.Blob(status, header, body)
}

func (a *Server) ChecksGet(ctx echo.Context) error {
	status, header, body := a.proxyRequest(ctx, "/api/v2/checks", "GET")
	ctx.Logger().Info("Checks: Proxy request completed", "status", status)
	return ctx.Blob(status, header, body)
}

func (a *Server) NotificationEndpointsPost(ctx echo.Context) error {
	status, header, body := a.proxyRequest(ctx, "/api/v2/notificationEndpoints", "POST")
	ctx.Logger().Info("NotificationEndpoints: Proxy request completed", "status", status)
	return ctx.Blob(status, header, body)
}

func (a *Server) NotificationRulesPost(ctx echo.Context) error {
	status, header, body := a.proxyRequest(ctx, "/api/v2/notificationRules", "POST")
	ctx.Logger().Info("NotificationRules: Proxy request completed", "status", status)
	return ctx.Blob(status, header, body)
}

func (a *Server) ChecksPost(ctx echo.Context) error {
	status, header, body := a.proxyRequest(ctx, "/api/v2/checks", "POST")
	ctx.Logger().Info("Checks: Proxy request completed", "status", status)
	return ctx.Blob(status, header, body)
}

func (a *Server) NotificationEndpointsDelete(ctx echo.Context) error {
	status, header, body := a.proxyRequest(ctx, "/api/v2/notificationEndpoints", "DELETE")
	ctx.Logger().Info("NotificationEndpoints: Proxy request completed", "status", status)
	return ctx.Blob(status, header, body)
}

func (a *Server) NotificationRulesDelete(ctx echo.Context) error {
	status, header, body := a.proxyRequest(ctx, "/api/v2/notificationRules", "DELETE")
	ctx.Logger().Info("NotificationRules: Proxy request completed", "status", status)
	return ctx.Blob(status, header, body)
}

func (a *Server) ChecksDelete(ctx echo.Context) error {
	status, header, body := a.proxyRequest(ctx, "/api/v2/checks", "DELETE")
	ctx.Logger().Info("Checks: Proxy request completed", "status", status)
	return ctx.Blob(status, header, body)
}

func (a *Server) proxyRequest(ctx echo.Context, path string, method string) (int, string, []byte) {
	baseURL := a.Config.ServerConfig.InfluxHost + ":" + a.Config.ServerConfig.InfluxPort

	targetURL, err := url.Parse(baseURL)
	if err != nil {
		ctx.Logger().Error("proxyRequest: Invalid target URL", "error", err)
		return http.StatusInternalServerError, "Invalid target URL", nil
	}

	targetURL.Path = path
	targetURL.RawQuery = ctx.QueryString()

	req := ctx.Request()
	client := &http.Client{}

	var bodyReader io.Reader = nil
	if req.Body != nil {
		bodyReader = req.Body
	}
	req.Method = method

	targetReq, err := http.NewRequest(req.Method, targetURL.String(), bodyReader)
	if err != nil {
		ctx.Logger().Error("proxyRequest: Failed to create proxy request", "error", err)
		return http.StatusInternalServerError, "Failed to create proxy request", nil
	}

	for name, values := range req.Header {
		for _, value := range values {
			targetReq.Header.Add(name, value)
		}
	}

	resp, err := client.Do(targetReq)
	if err != nil {
		ctx.Logger().Error("proxyRequest: Failed to contact target server", "error", err)
		return http.StatusBadGateway, "Failed to contact target server", nil
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			ctx.Logger().Warn("proxyRequest: Failed to close response body", "error", err)
		}
	}(resp.Body)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		ctx.Logger().Error("proxyRequest: Failed to read response body", "error", err)
		return http.StatusInternalServerError, "Failed to read response body", nil
	}

	for name, values := range resp.Header {
		for _, value := range values {
			ctx.Response().Header().Add(name, value)
		}
	}
	return resp.StatusCode, resp.Header.Get("Content-Type"), body
}

// ConvertMapToTOML takes a map[string]interface{} and converts it to a TOML-formatted file
func ConvertMapToTOML(data map[string]interface{}, t string) ([]byte, error) {
	tomlTree, err := toml.TreeFromMap(data)
	if err != nil {
		return nil, fmt.Errorf("failed to create TOML tree: %w", err)
	}

	tomlBytes, err := tomlTree.Marshal()
	if err != nil {
		return nil, fmt.Errorf("failed to marshal TOML: %w", err)
	}

	updatedToml := processConfig2(string(tomlBytes), t)

	return []byte(updatedToml), nil
}

func processConfig2(input string, t string) string {
	var result strings.Builder
	scanner := bufio.NewScanner(strings.NewReader(input))

	for scanner.Scan() {
		line := scanner.Text()
		var processedLine string
		if strings.Contains(line, `inputs.`+t) {
			spp := strings.Split(line, `"`)
			for _, v := range spp {
				processedLine += v
			}
			line = "[" + strings.TrimSpace(processedLine) + "]"
		}
		result.WriteString(line + "\n")
	}

	if err := scanner.Err(); err != nil {
		fmt.Println("Error reading input:", err)
	}

	return result.String()
}
