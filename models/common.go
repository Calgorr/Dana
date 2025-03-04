package models

import (
	"reflect"

	"Dana"
)

// logName returns the log-friendly name/type.
func logName(pluginType, name, alias string) string {
	if alias == "" {
		return pluginType + "." + name
	}
	return pluginType + "." + name + "::" + alias
}

func SetLoggerOnPlugin(i interface{}, logger Dana.Logger) {
	valI := reflect.ValueOf(i)

	if valI.Type().Kind() != reflect.Ptr {
		valI = reflect.New(reflect.TypeOf(i))
	}

	field := valI.Elem().FieldByName("Log")
	if !field.IsValid() {
		return
	}

	switch field.Type().String() {
	case "Dana2.Logger":
		if field.CanSet() {
			field.Set(reflect.ValueOf(logger))
		}
	default:
		logger.Debugf("Plugin %q defines a 'Log' field on its struct of an unexpected type %q. Expected Dana2.Logger",
			valI.Type().Name(), field.Type().String())
	}
}
