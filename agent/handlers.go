package agent

import (
	"errors"

	"github.com/influxdata/toml/ast"
	"github.com/labstack/echo/v4"

	"Dana/agent/model"
	"Dana/models"
)

func (a *Server) HealthCheck(ctx echo.Context) error {
	return ctx.JSON(200, "OK")
}

func (a *Server) PostInput(ctx echo.Context) error {
	inputData := &model.HandlerInput{}
	if err := ctx.Bind(inputData); err != nil {
		return ctx.JSON(400, errors.New("invalid request"))
	}
	if err := a.InputRepo.AddServerInput(ctx.Request().Context(), inputData); err != nil {
		return ctx.JSON(500, "internal server error")
	}
	table := MapToASTTable(inputData.Data)
	input, err := a.Config.InputMaker(inputData.Type, table)
	if err != nil {
		return ctx.JSON(500, errors.New("internal server error"))
	}

	iu := &inputUnit{
		dst: a.InputDstChan,
		inputs: []*models.RunningInput{
			input,
		},
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

// MapToASTTable converts a map[string]interface{} to an ast.Table
func MapToASTTable(configMap map[string]interface{}) *ast.Table {
	table := &ast.Table{
		Fields: make(map[string]interface{}),
		Type:   ast.TableTypeNormal,
	}

	for key, value := range configMap {
		switch v := value.(type) {
		case map[string]interface{}:
			// Recursively convert nested maps to tables
			table.Fields[key] = MapToASTTable(v)
		case []interface{}:
			// Handle array types
			tables := make([]*ast.Table, 0)
			for _, item := range v {
				if mapItem, ok := item.(map[string]interface{}); ok {
					tables = append(tables, MapToASTTable(mapItem))
				}
			}
			if len(tables) > 0 {
				table.Fields[key] = tables
			} else {
				// If it's not an array of tables, store as is
				table.Fields[key] = value
			}
		default:
			// Store primitive values as is
			table.Fields[key] = value
		}
	}

	return table
}
