package climate

import (
	"errors"
	"strconv"
	"strings"

	"github.com/ksimuk/ebus-climate/internal/ebusd/client"
)

type Climate struct {
	client *client.Client
}

func New(ebusClient *client.Client) *Climate {
	return &Climate{
		client: ebusClient,
	}
}

func (c *Climate) SetFlowTemperature(temperature int) error {

	return nil
}

func (c *Climate) FlowTemperature() (int, error) {
	return 0, nil
}

func (c *Climate) ReturnTemp() (float64, error) {
	response, err := c.client.Get("ReturnTemp")
	if err != nil {
		return 0, err
	}
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
