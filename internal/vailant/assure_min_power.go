// ensures flow temperature is adjusted to run boiler on min power
// instead of firing on full power and then shutting down as max flow is reached
// boiler is drastically oversized for my case, so we do not use flow temp to modulate,
// but rather cycle boiler based  on heat loss
//
// Update, set the boiler to lowest power for heating, not needed anymore
package vailant

import "github.com/rs/zerolog/log"

const MAX_FLOW_TEMP = 50.0
const MIN_FLOW_TEMP = 30.0 // check later,  looks unstable below 30
const DIFF_TEMP = 10.0

// func assureMinPower(flowTemp float64, returnTemp float64) int {
// 	calculatedFlow := returnTemp + DIFF_TEMP

// 	if flowTemp < calculatedFlow {
// 		calculatedFlow = flowTemp
// 	}

// 	// limit to max flow temp

// 	if calculatedFlow > MAX_FLOW_TEMP {
// 		return MAX_FLOW_TEMP
// 	}

// 	// limit to min flow temp
// 	if calculatedFlow < MIN_FLOW_TEMP {
// 		return MIN_FLOW_TEMP
// 	}

// 	return int(calculatedFlow) // this truncates towards zero
// }

func (c *eBusClimate) onReturnTemperatureChange() {
	// newTemp := assureMinPower(c.flowTemp, c.returnTemp)
	// if newTemp != c.desiredTemp {
	// 	c.desiredTemp = newTemp
	// c.pingHeating()
	// }
	log.Info().Msgf("Desired Flow:  %d, return %f, flow %f", c.desiredFlowTemp, c.returnTemp, c.flowTemp)
}
