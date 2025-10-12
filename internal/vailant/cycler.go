// Cycled is responsible to calculate start and stop time to fire boiler
package vailant

import (
	"math"
	"time"

	"github.com/rs/zerolog/log"
)

const CYCLE_CHECK_INTERVAL = 1
const BASE_TEMP = 20.0
const ADJUSTMENT_THRESHOLD = 0.5 // only adjust if we are more than this far from target

const MIN_RUNTIME = 10 // minimum runtime in minutes 10C
const MAX_RUNTIME = 25 // maximum runtime in minutes -3C

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

// adjust temprature if we below or above the target
// returns number of degrees to adjust current weather for heat loss calculation
// negative means we are above target, positive means we are below target
func (c *eBusClimate) adjustTemp() float64 {
	insideTemp := c.state.InsideTemp
	targetTemp := c.state.TargetTemperature

	adjustment := targetTemp - insideTemp

	// if we are more than 1 degree off, adjust for faster correction
	if math.Abs(adjustment) >= ADJUSTMENT_THRESHOLD {
		// square the adjustment to make it non-linear
		return math.Abs(adjustment) * adjustment
	}
	return 0
}

func (c *eBusClimate) getMinuteLoss() float64 {
	// TODO adjust based on inside target temp
	current_weather := c.state.OutsideTemp

	// reduce loss if target temp is lower than base temperature (we are in setback)
	current_weather += BASE_TEMP - c.state.TargetTemperature
	// adjust loss based on how far we are from target temp
	current_weather -= c.adjustTemp()

	loss3 := c.loss3
	loss7 := c.loss7

	currentLossW := float64(loss3-loss7)/10*(7-current_weather) + float64(loss7)
	c.stat.CurrentHeatLoss = currentLossW
	if currentLossW < 0 {
		if c.zeroLossTimer == nil {
			c.zeroLossTimer = time.NewTicker(time.Hour)
			go func() {
				for {
					select {
					case <-c.zeroLossTimer.C:
						c.zeroLossTimer.Stop()
						c.zeroLossTimer = nil
						c.state.HeatLoss = 1500.0 // reset to 1.5 kWh after an hour of zero loss
						log.Info().Msgf("Resetting heat loss to %f kWh after an hour of zero loss", c.state.HeatLoss)
						return
					case <-c.stopChan:
						return
					}
				}
			}()
		}
		currentLossW = 0
	} else {
		if c.zeroLossTimer != nil {
			c.zeroLossTimer.Stop()
			c.zeroLossTimer = nil
		}
	}
	return currentLossW / 60 // per minute
}

func (c *eBusClimate) calculateLoss() {
	if c.state.Mode != MODE_HEATING {
		// heating is off, no need to calculate loss
		return
	}

	currentLoss := c.getMinuteLoss()
	c.state.HeatLoss = c.state.HeatLoss - currentLoss
	log.Debug().Msgf("heat loss balance %f, at temps %f, with loss %f", c.state.HeatLoss, c.state.OutsideTemp, currentLoss)

	if c.state.HeatLoss < 0 {
		c.state.HeatLoss = c.state.HeatLoss + c.runCycle()
		log.Info().Msgf("Starting new heating cycle to cover heat loss, new balance %f", c.state.HeatLoss)
	}
}

func (c *eBusClimate) getRuntime(OutsideTemp float64) int {
	if OutsideTemp >= 10 {
		return MIN_RUNTIME
	}
	if OutsideTemp <= -3 {
		return MAX_RUNTIME
	}
	return MIN_RUNTIME + int((10-OutsideTemp)*float64(MAX_RUNTIME-MIN_RUNTIME)/13)
}

func (c *eBusClimate) runCycle() float64 {
	// TODO calculate cycle length increase duration for lower temps
	cycleLength := c.getRuntime(c.state.OutsideTemp)
	c.stat.Runtime = cycleLength
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
