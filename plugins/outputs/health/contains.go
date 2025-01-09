package health

import "Dana"

type Contains struct {
	Field string `toml:"field"`
}

func (c *Contains) Check(metrics []Dana.Metric) bool {
	success := false
	for _, m := range metrics {
		ok := m.HasField(c.Field)
		if ok {
			success = true
		}
	}

	return success
}
