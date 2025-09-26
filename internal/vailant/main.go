package vailant

import (
	"time"

	"github.com/ksimuk/ebus-climate/internal/config"
	"github.com/ksimuk/ebus-climate/internal/ebusd/client"
	"github.com/rs/zerolog/log"
)

type eBusClimate struct {
	ebusClient *client.Client
	stopChan   chan struct{}
}

func New(config *config.Config) *eBusClimate {
	log.Debug().Msg("Creating new eBusClimate instance")
	ebusClient := client.New(config)

	climate := eBusClimate{
		ebusClient: ebusClient,
		stopChan:   make(chan struct{}),
	}

	climate.StartPolling(time.Second*10, func(client *client.Client) {
		res, err := ebusClient.Get("FlowTempDesired")
		if err != nil {
			log.Error().Err(err).Msg("Failed to get FlowTempDesired")
			return
		}
		log.Debug().Interface("result", res).Msg("Received FlowTempDesired")
	})

	return &climate
}

func (c *eBusClimate) Info() ([]string, error) {
	return c.ebusClient.Info()
}

// StartPolling starts a timer to read data from ebusClient at the given interval.
func (c *eBusClimate) StartPolling(interval time.Duration, readFunc func(*client.Client)) {
	log.Debug().Msg("Start ebus pulling")

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				readFunc(c.ebusClient)
			case <-c.stopChan:
				return
			}
		}
	}()
}

// StopPolling stops the polling timer.
func (c *eBusClimate) StopPolling() {
	close(c.stopChan)
}
