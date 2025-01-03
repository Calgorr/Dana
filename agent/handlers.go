package agent

import (
	"errors"

	"github.com/labstack/echo/v4"

	"Dana/agent/model"
)

func (a *Server) HealthCheck(ctx echo.Context) error {
	return ctx.JSON(200, "OK")
}

func (a *Server) PostInput(ctx echo.Context) error {
	input := &model.HandlerInput{}
	if err := ctx.Bind(input); err != nil {
		return ctx.JSON(400, errors.New("invalid request"))
	}
	if err := a.InputRepo.AddServerInput(ctx.Request().Context(), input); err != nil {
		return ctx.JSON(500, "internal server error")
	}
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
