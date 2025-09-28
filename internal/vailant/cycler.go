// Cycled is responsible to calculate start and stop time to fire boiler
package vailant

import (
	"time"

	"github.com/rs/zerolog/log"
)

const CYCLE_CHECK_INTERVAL = 1

func (c *eBusClimate) startCycler() {
	c.calculateLoss() // initial calculation
	// launch cycler goroutine every minute
	go func() {
		ticker := time.NewTicker(time.Minute * CYCLE_CHECK_INTERVAL)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				c.calculateConsumption()
				c.calculateLoss()
				c.pingHeating() // keep connection with boiler active

			case <-c.stopChan:
				return
			}
		}
	}()
}

func (c *eBusClimate) calculateConsumption() {
	if c.IsGasActive() {
		c.state.ConsumptionHeating += float64(c.power) * CYCLE_CHECK_INTERVAL / 60 / 1000 // per minute kWh
	}
	c.stateStore.Save(c.state) // save state with new consumption and heat loss

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
	c.state.HeatLoss = c.state.HeatLoss - currentLoss
	log.Debug().Msgf("heat loss balance %f, at temps %f, with loss %f", c.state.HeatLoss, c.state.OutsideTemp, currentLoss)

	if c.state.HeatLoss < 0 {
		c.state.HeatLoss = c.state.HeatLoss + c.runCycle()
		log.Info().Msgf("Starting new heating cycle to cover heat loss, new balance %f", c.state.HeatLoss)
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
}
