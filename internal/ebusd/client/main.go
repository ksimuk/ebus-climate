package client

import (
	"bufio"
	"errors"
	"fmt"
	"net"
	"strings"

	"github.com/ksimuk/ebus-climate/internal/config"
	"github.com/rs/zerolog/log"
)

type Circuit string

type Client struct {
	config     *config.Config
	parameters []string
}

func New(config *config.Config, readParameters []string) *Client {
	return &Client{
		config:     config,
		parameters: readParameters,
	}
}

func (c Client) request(request string) ([]string, error) {
	log.Debug().Msgf("Connecting to ebusd at %s:%d", c.config.Ebus.Host, c.config.Ebus.Port)
	connection, err := net.Dial("tcp", fmt.Sprintf("%s:%d", c.config.Ebus.Host, c.config.Ebus.Port))
	if err != nil {
		return nil, err
	}
	defer connection.Close()

	log.Debug().Msgf("Sending request: %s", request)
	_, err = connection.Write([]byte(request))
	if err != nil {
		return nil, err
	}
	reader := bufio.NewReader(connection)
	result := []string{}
	for {
		line, _, err := reader.ReadLine()
		if err != nil {
			return nil, err
		}
		if (len(line)) == 0 {
			break
		}

		result = append(result, string(line))
	}
	if (len(result)) == 0 {
		return nil, errors.New("empty response")
	}
	for _, line := range result {
		err := checkEbusError(line)
		if err != nil {
			return nil, err
		}
	}
	log.Debug().Interface("result", result).Msg("Received reply")
	return result, nil
}

func (c Client) read(parameter string, force bool) ([]string, error) {

	args := ""
	if force {
		args = " -f"
	}
	request := fmt.Sprintf("read -c %s %s%s\n", c.config.Ebus.Circuit, parameter, args)

	reply, err := c.request(request)

	if err != nil {
		return nil, err
	}
	log.Debug().Msgf("Received reply: %s", reply)
	return reply, nil

}

func (c Client) Info() ([]string, error) {
	request := "info\n"
	reply, err := c.request(request)
	if err != nil {
		return nil, err
	}
	log.Debug().Msgf("Received reply: %s", reply)
	return reply, nil
}

func (c Client) write(parameter string, value string) error {
	request := fmt.Sprintf("write -c %s %s %s\n", c.config.Ebus.Circuit, parameter, value)
	res, err := c.request(request)
	if err != nil {
		return err
	}
	log.Debug().Interface("result", res).Msgf("%s result %s", request, res)
	return nil
}

func (c Client) Get(parameter string) ([]string, error) {
	return c.read(parameter, false)
}

func (c Client) Set(parameter string, value string) error {
	return c.write(parameter, value)
}

func (c Client) ReadAll() map[string]string {
	result := make(map[string]string)
	for _, param := range c.parameters {
		res, err := c.read(param, false)
		if err != nil {
			log.Error().Err(err).Msgf("Failed to read parameter %s", param)
			continue
		}
		result[param] = res[0]
	}
	return result
}

func (c Client) State() string {
	request := "state\n"
	reply, err := c.request(request)
	if err != nil {
		return "unknown"
	}
	if len(reply) > 0 {
		return strings.TrimSpace(reply[0])
	}
	return "unknown"
}

func checkEbusError(reply string) error {
	if strings.HasPrefix(reply, "ERR:") {
		return errors.New(reply)
	}
	return nil
}
