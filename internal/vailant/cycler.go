// Cycled is responsible to calculate start and stop time to fire boiler
package vailant

import (
	"time"

	"github.com/rs/zerolog/log"
)

const CYCLE_CHECK_INTERVAL = 1
const BASE_TEMP = 20.0
const ADJUSTMENT_THRESHOLD = 0.5 // only adjust if we are more than this far from target

const ADJUSTMENT_RATE = 3

const MIN_RUNTIME = 15 // minimum runtime in minutes 10C
const MAX_RUNTIME = 30 // maximum runtime in minutes -3C

func (c *eBusClimate) isHwcDemandActive() bool {
	return c.stat.HwcDemand == "on" || c.stat.HwcDemand == "yes" || c.stat.HwcDemand == "1" || c.stat.HwcDemand == "true"
}

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

	return adjustment * ADJUSTMENT_RATE
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
	return currentLossW / 60 // per minute
}

func (c *eBusClimate) calculateLoss() {
	if c.state.Mode != MODE_HEATING {
		// heating is off, no need to calculate loss
		return
	}

	currentLoss := c.getMinuteLoss()
	c.stat.CurrentHeatLoss = currentLoss * 60 // in W
	c.state.HeatLoss = c.state.HeatLoss - currentLoss
	if c.state.HeatLoss > 1500 {
		c.state.HeatLoss = 1500 // cap the loss to avoid extreme values
	}
	log.Debug().Msgf("heat loss balance %f, at temps %f, with loss %f", c.state.HeatLoss, c.state.OutsideTemp, currentLoss)

	// update runtime based on current heat loss
	c.stat.Runtime = c.getRuntime()
	if c.state.HeatLoss < 0 {
		c.state.HeatLoss = c.state.HeatLoss + c.runCycle()
		log.Info().Msgf("Starting new heating cycle to cover heat loss, new balance %f", c.state.HeatLoss)
	}
}

func (c *eBusClimate) getRuntime() int {
	heatLoss := c.stat.CurrentHeatLoss
	if heatLoss < 1200.0 {
		return MIN_RUNTIME
	}
	if heatLoss > 3200.0 {
		return MAX_RUNTIME
	}
	return MIN_RUNTIME + int((heatLoss-1200)*float64(MAX_RUNTIME-MIN_RUNTIME)/2000)
}

func (c *eBusClimate) runCycle() float64 {
	// TODO calculate cycle length increase duration for lower temps
	cycleLength := c.stat.Runtime
	c.stat.Runtime = cycleLength
	c.runFor(cycleLength)
	return float64(c.power*cycleLength) / 60
}

func (c *eBusClimate) runFor(minutes int) {
	<-c.heatingTimerMutex                                // acquire lock
	defer func() { c.heatingTimerMutex <- struct{}{} }() // release lock

	if c.heatingActive {
		// Heating is already active, add minutes to existing cycle
		c.heatingEndTime = c.heatingEndTime.Add(time.Duration(minutes) * time.Minute)
		log.Info().Msgf("Extending heating cycle by %d minutes, new end time: %s", minutes, c.heatingEndTime.Format("15:04:05"))
		return
	}

	// Start new heating cycle
	c.heatingEndTime = time.Now().Add(time.Duration(minutes) * time.Minute)
	log.Info().Msgf("Start heating cycle for %d minutes (until %s)", minutes, c.heatingEndTime.Format("15:04:05"))

	go func() {
		c.StartHeating()
		interval := time.Minute
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for range ticker.C {
			// Check if HwcDemand is active and extend by 1 minute
			if c.isHwcDemandActive() {
				<-c.heatingTimerMutex
				c.heatingEndTime = c.heatingEndTime.Add(interval)
				c.heatingTimerMutex <- struct{}{}
				log.Info().Msgf("HwcDemand active during heating - extending cycle by 1 minute (until %s)", c.heatingEndTime.Format("15:04:05"))
			}

			if time.Now().After(c.heatingEndTime) || time.Now().Equal(c.heatingEndTime) {
				c.StopHeating()
				return
			}
		}
	}()
}
