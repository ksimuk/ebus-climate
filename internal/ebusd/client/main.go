package client

import (
	"bufio"
	"errors"
	"fmt"
	"net"
	"strings"

	"github.com/rs/zerolog/log"
)

type Circuit string

type Client struct {
	host string
	port int
}

func New(host string, port int) *Client {
	return &Client{
		host: host,
		port: port,
	}
}

func (c Client) request(request string) ([]string, error) {
	connection, err := net.Dial("tcp", fmt.Sprintf("%s:%d", c.host, c.port))
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
		return nil, errors.New("Empty reponse")
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
	request := fmt.Sprintf("read %s %s\n", parameter, args)

	reply, err := c.request(request)

	if err != nil {
		return nil, err
	}
	log.Debug().Msgf("Received reply: %s", reply)
	return reply, nil

}

func (c Client) write(parameter string, value string) error {
	request := fmt.Sprintf("write %s %s\n", parameter, value)
	_, err := c.request(request)
	if err != nil {
		return err
	}
	return nil
}

func (c Client) Get(parameter string) ([]string, error) {
	return c.read(parameter, false)
}

func checkEbusError(reply string) error {
	if strings.HasPrefix(reply, "ERR:") {
		return errors.New(reply)
	}
	return nil
}
