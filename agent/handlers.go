package agent

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/pelletier/go-toml"
	"go.mongodb.org/mongo-driver/mongo"

	authentication "Dana/agent/Auth"
	"Dana/agent/model"
	"Dana/agent/notification"
	"Dana/config"
)

func (a *Server) HealthCheck(ctx echo.Context) error {
	return ctx.JSON(200, "OK")
}

func (a *Server) Register(ctx echo.Context) error {
	user := &model.User{}
	if err := ctx.Bind(user); err != nil {
		return ctx.JSON(400, errors.New("invalid request"))
	}
	if err := a.UserRepo.AddUser(ctx.Request().Context(), user); err != nil {
		return ctx.JSON(500, "internal server error")
	}
	return ctx.JSON(200, "OK")
}

func (a *Server) Login(ctx echo.Context) error {
	user := &model.User{}
	if err := ctx.Bind(user); err != nil {
		return ctx.JSON(400, errors.New("invalid request"))
	}
	if err := a.UserRepo.UserAuth(ctx.Request().Context(), user.Username, user.Password); err != nil {
		return ctx.JSON(401, "unauthorized")
	}
	token, err := authentication.GenerateJWT(user.Username)
	if err != nil {
		return ctx.JSON(500, "internal server error")
	}
	return ctx.JSON(200, token)
}

func (a *Server) Query(ctx echo.Context) error {
	status, header, body := a.proxyRequest(ctx, "/query")
	return ctx.Blob(status, header, body)
}

func (a *Server) PostInput(ctx echo.Context) error {
	inputData := &model.HandlerInput{}
	if err := ctx.Bind(inputData); err != nil {
		return ctx.JSON(400, errors.New("invalid request"))
	}
	if err := a.InputRepo.AddServerInput(ctx.Request().Context(), inputData); err != nil {
		return ctx.JSON(500, "internal server error")
	}
	toml, err := ConvertMapToTOML(inputData.Data, inputData.Type)
	if err != nil {
		return ctx.JSON(500, "internal server error")
	}
	newConfig := config.NewConfig()
	if err := newConfig.LoadConfigData(toml); err != nil {
		return ctx.JSON(500, errors.New("internal server error"))
	}
	iu := &inputUnit{
		dst:    a.InputDstChan,
		inputs: newConfig.Inputs,
	}
	a.runInputs(ctx.Request().Context(), a.StartTime, iu)
	return ctx.JSON(200, "OK")
}

func (a *Server) GetInput(ctx echo.Context) error {
	inputs, err := a.InputRepo.GetServers(ctx.Request().Context())
	if err != nil {
		return ctx.JSON(500, errors.New("internal server error"))
	}
	return ctx.JSON(200, inputs)
}

func (a *Server) GetInputByType(ctx echo.Context) error {
	inputType := ctx.Param("type")
	inputs, err := a.InputRepo.GetServersByType(ctx.Request().Context(), inputType)
	if err != nil {
		return ctx.JSON(500, errors.New("internal server error"))
	}
	return ctx.JSON(200, inputs)
}

func (a *Server) CreateDashboard(ctx echo.Context) error {
	dashboard := &model.Dashboard{}
	if err := ctx.Bind(dashboard); err != nil {
		return ctx.JSON(400, errors.New("invalid request"))
	}
	id, err := a.DashboardRepo.CreateDashboard(ctx.Request().Context(), dashboard)
	if err != nil {
		return ctx.JSON(500, "internal server error")
	}
	return ctx.JSON(201, map[string]interface{}{"id": id})
}

func (a *Server) GetDashboard(ctx echo.Context) error {
	id := ctx.Param("id")
	dashboard, err := a.DashboardRepo.GetDashboard(ctx.Request().Context(), id)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return ctx.JSON(404, "dashboard not found")
		}
		return ctx.JSON(500, "internal server error")
	}
	return ctx.JSON(200, dashboard)
}

func (a *Server) UpdateDashboard(ctx echo.Context) error {
	dashboard := &model.Dashboard{}
	if err := ctx.Bind(dashboard); err != nil {
		return ctx.JSON(400, errors.New("invalid request"))
	}
	if err := a.DashboardRepo.UpdateDashboard(ctx.Request().Context(), dashboard); err != nil {
		return ctx.JSON(500, "internal server error")
	}
	return ctx.JSON(200, "OK")
}

func (a *Server) DeleteDashboard(ctx echo.Context) error {
	id := ctx.Param("id")
	if err := a.DashboardRepo.DeleteDashboard(ctx.Request().Context(), id); err != nil {
		return ctx.JSON(500, "internal server error")
	}
	return ctx.JSON(200, "OK")
}

func (a *Server) GetDashboards(ctx echo.Context) error {
	dashboards, err := a.DashboardRepo.GetDashboards(ctx.Request().Context())
	if err != nil {
		return ctx.JSON(500, "internal server error")
	}
	return ctx.JSON(200, dashboards)
}

func (a *Server) CreateFolder(ctx echo.Context) error {
	folder := &model.Folder{}
	if err := ctx.Bind(folder); err != nil {
		return ctx.JSON(400, errors.New("invalid request"))
	}
	id, err := a.FolderRepo.CreateFolder(ctx.Request().Context(), folder)
	if err != nil {
		return ctx.JSON(500, "internal server error")
	}
	return ctx.JSON(201, map[string]interface{}{"id": id})
}

func (a *Server) GetFolder(ctx echo.Context) error {
	id := ctx.Param("id")
	folder, err := a.FolderRepo.GetFolder(ctx.Request().Context(), id)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return ctx.JSON(404, "folder not found")
		}
		return ctx.JSON(500, "internal server error")
	}
	return ctx.JSON(200, folder)
}

func (a *Server) UpdateDashboardInFolder(ctx echo.Context) error {
	folderID := ctx.Param("folderID")
	dashboardID := ctx.Param("dashboardID")
	dashboard := &model.Dashboard{}
	if err := ctx.Bind(dashboard); err != nil {
		return ctx.JSON(400, errors.New("invalid request"))
	}
	if err := a.FolderRepo.UpdateDashboardInFolder(ctx.Request().Context(), folderID, dashboardID, dashboard); err != nil {
		return ctx.JSON(500, "internal server error")
	}
	return ctx.JSON(200, "OK")
}

func (a *Server) DeleteFolder(ctx echo.Context) error {
	id := ctx.Param("id")
	if err := a.FolderRepo.DeleteFolder(ctx.Request().Context(), id); err != nil {
		return ctx.JSON(500, "internal server error")
	}
	return ctx.JSON(200, "OK")
}

func (a *Server) GetFolders(ctx echo.Context) error {
	folders, err := a.FolderRepo.GetFolders(ctx.Request().Context())
	if err != nil {
		return ctx.JSON(500, "internal server error")
	}
	return ctx.JSON(200, folders)
}

func (a *Server) AddNotification(ctx echo.Context) error {
	n := &model.Notification{}
	if err := ctx.Bind(n); err != nil {
		return ctx.JSON(400, errors.New("invalid request"))
	}
	if err := a.NotificationRepo.CreateNotification(ctx.Request().Context(), n); err != nil {
		return ctx.JSON(500, "internal server error")
	}
	return ctx.JSON(200, "OK")
}

func (a *Server) GetNotification(ctx echo.Context) error {
	channelName := ctx.Param("channelName")
	n, err := a.NotificationRepo.GetNotification(ctx.Request().Context(), channelName)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return ctx.JSON(404, "notification not found")
		}
		return ctx.JSON(500, "internal server error")
	}
	return ctx.JSON(200, n)
}

func (a *Server) DeleteNotification(ctx echo.Context) error {
	channelName := ctx.Param("channelName")
	if err := a.NotificationRepo.DeleteNotification(ctx.Request().Context(), channelName); err != nil {
		return ctx.JSON(500, "internal server error")
	}
	return ctx.JSON(200, "OK")
}

func (a *Server) SendNotification(ctx echo.Context) error {
	notif := new(model.Notification)

	if err := ctx.Bind(notif); err != nil {
		return ctx.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid input",
		})
	}

	n, err := a.NotificationRepo.GetNotification(ctx.Request().Context(), notif.ChannelName)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to retrieve notification",
		})
	}

	if notif.ChannelName == "telegram" {
		notification.SendNotification("https://api.telegram.org", a.Config.ServerConfig.TelegramToken, notif.Alert, int64(n.ChatID))
	} else if notif.ChannelName == "bale" {
		notification.SendNotification("https://tapi.bale.ai", a.Config.ServerConfig.BaleToken, notif.Alert, int64(n.ChatID))
	} else {
		return ctx.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid channel name",
		})
	}

	return ctx.JSON(http.StatusOK, notif)
}

func (a *Server) NotificationEndpoints(ctx echo.Context) error {
	status, header, body := a.proxyRequest(ctx, "/api/v2/notificationEndpoints")
	return ctx.Blob(status, header, body)
}

func (a *Server) NotificationRules(ctx echo.Context) error {
	status, header, body := a.proxyRequest(ctx, "/api/v2/notificationRules")
	return ctx.Blob(status, header, body)
}

func (a *Server) Checks(ctx echo.Context) error {
	status, header, body := a.proxyRequest(ctx, "/api/v2/checks")
	return ctx.Blob(status, header, body)
}

func (a *Server) proxyRequest(ctx echo.Context, path string) (int, string, []byte) {
	baseURL := a.Config.ServerConfig.InfluxHost + ":" + a.Config.ServerConfig.InfluxPort

	targetURL, err := url.Parse(baseURL)
	if err != nil {
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

	targetReq, err := http.NewRequest(req.Method, targetURL.String(), bodyReader)
	if err != nil {
		return http.StatusInternalServerError, "Failed to create proxy request", nil
	}

	for name, values := range req.Header {
		for _, value := range values {
			targetReq.Header.Add(name, value)
		}
	}

	resp, err := client.Do(targetReq)
	if err != nil {
		return http.StatusBadGateway, "Failed to contact target server", nil
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			return
		}
	}(resp.Body)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return http.StatusInternalServerError, "Failed to read response body", nil
	}

	for name, values := range resp.Header {
		for _, value := range values {
			ctx.Response().Header().Add(name, value)
		}
	}
	return resp.StatusCode, resp.Header.Get("Content-Type"), body
}

// ConvertMapToTOML takes a map[string]interface{} and converts it to a TOML-formatted []byte
func ConvertMapToTOML(data map[string]interface{}, t string) ([]byte, error) {
	tomlTree, err := toml.TreeFromMap(data)
	if err != nil {
		return nil, fmt.Errorf("failed to create TOML tree: %w", err)
	}

	tomlBytes, err := tomlTree.Marshal()
	if err != nil {
		return nil, fmt.Errorf("failed to marshal TOML: %w", err)
	}

	updatedToml := strings.ReplaceAll(string(tomlBytes), "\"inputs."+t+"\"", "input."+t)

	return []byte(updatedToml), nil
}
