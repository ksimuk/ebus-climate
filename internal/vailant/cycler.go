// Cycled is responsible to calculate start and stop time to fire boiler
package vailant

import (
	"time"

	"github.com/rs/zerolog/log"
)

func (c *eBusClimate) startCycler() {
	c.calculateLoss() // initial calculation
	// launch cycler goroutine every minute
	go func() {
		ticker := time.NewTicker(time.Minute)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				c.calculateLoss()
				c.pingHeating() // keep connection with boiler active
			case <-c.stopChan:
				return
			}
		}
	}()
}

func (c *eBusClimate) getMinuteLoss() float64 {
	// TODO adjust based on inside target temp
	current_weather := c.state.OutsideTemp

	loss3 := c.loss3
	loss7 := c.loss7

	currentLossW := float64(loss3-loss7)/10*(7-current_weather) + float64(loss7)
	if currentLossW < 0 {
		currentLossW = 0
	}
	return float64(currentLossW) / 60 // per minute
}

func (c *eBusClimate) calculateLoss() {
	currentLoss := c.getMinuteLoss()
	c.heatLoss = c.heatLoss - currentLoss
	log.Debug().Msgf("heat loss balance %f, at temps %f, with loss %f", c.heatLoss, c.state.OutsideTemp, currentLoss)

	if c.heatLoss < 0 {
		added := c.runCycle()
		c.heatLoss = c.heatLoss + added
		c.state.ConsumptionHeating += added / 1000 // convert to kWh
		log.Info().Msgf("Starting new heating cycle to cover heat loss, new balance %f", c.heatLoss)
	}
}

func (c *eBusClimate) runCycle() float64 {
	// TODO calculate cycle length increase duration for lower temps
	cycleLength := 10 // default to 10 minutes
	c.runFor(cycleLength)
	return float64(c.power*cycleLength) / 60
}

func (c *eBusClimate) runFor(minutes int) {
	log.Info().Msgf("Start heating cycle for %d minutes", minutes)
	go func() {
		c.StartHeating()
		time.AfterFunc(time.Duration(minutes)*time.Minute, func() {
			c.StopHeating()
		})
	}()
	c.state.LastActive = time.Now().Format(time.RFC3339) // record last launch
}
