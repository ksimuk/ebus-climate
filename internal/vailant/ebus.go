package vailant

import (
	"strconv"
	"strings"
)

func (c *eBusClimate) onChange(newValues map[string]string) {
	// handles loading updates from boiler
	for key, value := range newValues {
		switch key {
		case "ReturnTemp":
			// parse float
			value := getFloat(value)
			c.returnTemp = value
		case "FlowTemp":
			value := getFloat(value)
			c.flowTemp = value
		case "ModulationTempDesired":
			// parse int
			c.modulationTemp = int(getFloat(value))
		}
	}
	c.onReturnTemperatureChange()
}

func getFloat(values string) float64 {
	// 32.94;65008;ok
	parts := strings.Split(values, ";")
	if len(parts) > 0 {
		if value, err := strconv.ParseFloat(parts[0], 64); err == nil {
			return value
		}
	}
	return 0
}
