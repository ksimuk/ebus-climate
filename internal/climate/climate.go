package climate

import (
	"errors"
	"strconv"
	"strings"

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
	client *client.Client
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

func New(ebusClient *client.Client) *Climate {
	return &Climate{
		client:      ebusClient,
		heatingFire: false,
	}
}

func (c *Climate) SetDesiredFlowTemperature(temperature int) {
	c.desiredFlowTemperature = temperature
}

func (c *Climate) DesiredFlowTemperature() int {
	return c.desiredFlowTemperature
}

func parseTemperature(response []string) (float64, error) {
	fields := strings.Split(response[0], ";")
	if len(fields) < 1 {
		return 0, errors.New("Invalid response")
	}
	temp, err := strconv.ParseFloat(fields[0], 64)
	if err != nil {
		return 0, err
	}
	return temp, nil
}

// func (c *Climate) FlowTemperature() (float64, error) {
// 	response, err := c.client.Get("FlowTemp")
// 	if err != nil {
// 		return 0, err
// 	}
// 	return parseTemperature(response)
// }

// func (c *Climate) ReturnTemp() (float64, error) {
// 	response, err := c.client.Get("ReturnTemp")
// 	if err != nil {
// 		return 0, err
// 	}
// 	return parseTemperature(response)
// }

func (c *Climate) TargetTemp() float64 {
	return 0
}

func (c *Climate) SetTargetTemperature(temp float64) error {
	return nil
}

func (c *Climate) SetMode(hotWater bool, heating bool) error {
	return nil
}

func (c *Climate) Mode() (Mode, error) {
	return Mode{}, nil
}

func (c *Climate) SetRoomTemperature(sensor Sensor) {
	c.roomSensor = sensor
}

func (c *Climate) RoomTemperature() Sensor {
	return c.roomSensor
}

func (c *Climate) MinTemperature() float64 {
	return c.minTemperature
}

func (c *Climate) CurrentTemperature() float64 {
	return c.RoomTemperature().temperature
}

func (c *Climate) MaxTemperature() float64 {
	return c.maxTemperature
}

func (c *Climate) fireHeating() error {
	c.heatingFire = true
	return nil
}

func (c *Climate) stopHeating() error {
	c.heatingFire = false
	return nil
}

func (c *Climate) IsHeating() bool {
	return c.heatingFire
}

func (c *Climate) FlowTemperature() float64 {
	return c.flowTemperature
}

func (c *Climate) ReturnTemperature() float64 {
	return c.returnTemperature
}

func (c *Climate) HotWaterTemperature() int {
	return c.hotWaterTemperature
}

func (c *Climate) SetHotWaterTemperature(temp int) {
	c.hotWaterTemperature = temp
}
