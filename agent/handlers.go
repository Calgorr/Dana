package agent

import (
	"errors"

	"github.com/labstack/echo/v4"

	"Dana/agent/model"
)

func (a *Server) HealthCheck(ctx echo.Context) error {
	return ctx.JSON(200, "OK")
}

func (a *Server) PostInputWmi(ctx echo.Context) error {
	wmiModel := &model.Wmi{}
	if err := ctx.Bind(wmiModel); err != nil {
		return ctx.JSON(400, errors.New("invalid request"))
	}
	if err := a.WmiRepo.AddWmiInput(ctx.Request().Context(), wmiModel); err != nil {
		return ctx.JSON(500, "internal server error")
	}
	return ctx.JSON(200, "OK")
}

func (a *Server) GetInputsWmi(ctx echo.Context) error {
	wmiModels, err := a.WmiRepo.GetWmis(ctx.Request().Context())
	if err != nil {
		return ctx.JSON(500, errors.New("internal server error"))
	}
	return ctx.JSON(200, wmiModels)
}

func (a *Server) DeleteInputWmi(ctx echo.Context) error {
	id := ctx.Param("id")
	if err := a.WmiRepo.DeleteWmi(ctx.Request().Context(), id); err != nil {
		if err.Error() == "no document found with the given ID" {
			return ctx.JSON(404, "no input found with the given ID")
		}
		return ctx.JSON(500, err)
	}
	return ctx.JSON(200, "OK")
}

func (a *Server) PostInputSnmp(ctx echo.Context) error {
	snmpModel := &model.Snmp{}
	if err := ctx.Bind(snmpModel); err != nil {
		return ctx.JSON(400, errors.New("invalid request"))
	}
	if err := a.SnmpRepo.AddSnmpInput(ctx.Request().Context(), snmpModel); err != nil {
		return ctx.JSON(500, "internal server error")
	}
	return ctx.JSON(200, "OK")
}

func (a *Server) GetInputsSnmp(ctx echo.Context) error {
	snmpModels, err := a.SnmpRepo.GetSnmps(ctx.Request().Context())
	if err != nil {
		return ctx.JSON(500, errors.New("internal server error"))
	}
	return ctx.JSON(200, snmpModels)
}

func (a *Server) DeleteInputSnmp(ctx echo.Context) error {
	id := ctx.Param("id")
	if err := a.SnmpRepo.DeleteSnmp(ctx.Request().Context(), id); err != nil {
		if err.Error() == "no document found with the given ID" {
			return ctx.JSON(404, "no input found with the given ID")
		}
		return ctx.JSON(500, "internal server error")
	}
	return ctx.JSON(200, "OK")
}

func (a *Server) PostInputPrometheus(ctx echo.Context) error {
	prometheusModel := &model.Prometheus{}
	if err := ctx.Bind(prometheusModel); err != nil {
		return ctx.JSON(400, errors.New("invalid request"))
	}
	if err := a.PrometheusRepo.AddServerInput(ctx.Request().Context(), prometheusModel); err != nil {
		return ctx.JSON(500, "internal server error")
	}
	return ctx.JSON(200, "OK")
}

func (a *Server) GetInputsPrometheus(ctx echo.Context) error {
	prometheusModels, err := a.PrometheusRepo.GetServers(ctx.Request().Context())
	if err != nil {
		return ctx.JSON(500, errors.New("internal server error"))
	}
	return ctx.JSON(200, prometheusModels)
}

func (a *Server) DeleteInputPrometheus(ctx echo.Context) error {
	id := ctx.Param("id")
	if err := a.PrometheusRepo.DeleteServer(ctx.Request().Context(), id); err != nil {
		if err.Error() == "no document found with the given ID" {
			return ctx.JSON(404, "no input found with the given ID")
		}
		return ctx.JSON(500, "internal server error")
	}
	return ctx.JSON(200, "OK")
}

func (a *Server) PostInputInflux(ctx echo.Context) error {
	influxModel := &model.InfluxDbV2{}
	if err := ctx.Bind(influxModel); err != nil {
		return ctx.JSON(400, errors.New("invalid request"))
	}
	if err := a.InfluxRepo.AddServerInput(ctx.Request().Context(), influxModel); err != nil {
		return ctx.JSON(500, "internal server error")
	}
	return ctx.JSON(200, "OK")
}

func (a *Server) GetInputsInflux(ctx echo.Context) error {
	influxModels, err := a.InfluxRepo.GetServers(ctx.Request().Context())
	if err != nil {
		return ctx.JSON(500, errors.New("internal server error"))
	}
	return ctx.JSON(200, influxModels)
}

func (a *Server) DeleteInputInflux(ctx echo.Context) error {
	id := ctx.Param("id")
	if err := a.InfluxRepo.DeleteServer(ctx.Request().Context(), id); err != nil {
		if err.Error() == "no document found with the given ID" {
			return ctx.JSON(404, "no input found with the given ID")
		}
		return ctx.JSON(500, "internal server error")
	}
	return ctx.JSON(200, "OK")
}
