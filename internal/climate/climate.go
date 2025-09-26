package climate

import (
	"github.com/ksimuk/ebus-climate/internal/config"
	"github.com/ksimuk/ebus-climate/internal/ebusd/client"
)

type Mode struct {
	heating, hotWater bool
}

type Sensor struct {
	temperature float64
	humidity    float64
}

type Climate struct {
	ebusClient *client.Client
	// Config
	minTemperature         float64
	maxTemperature         float64
	hotWaterTemperature    int
	desiredFlowTemperature int
	isHotWater             bool
	isHeating              bool

	mode Mode

	roomSensor        Sensor
	heatingFire       bool
	flowTemperature   float64
	returnTemperature float64
}

func New(config *config.Config) *Climate {
	ebusClient := client.New(config)
	return &Climate{
		ebusClient: ebusClient,
	}
}

func (c *Climate) Info() ([]string, error) {
	return c.ebusClient.Info()
}
