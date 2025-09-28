package vailant

import (
	"errors"
	"fmt"
	"time"

	"github.com/ksimuk/ebus-climate/internal/climate"
	"github.com/ksimuk/ebus-climate/internal/config"
	"github.com/ksimuk/ebus-climate/internal/ebusd/client"
	"github.com/rs/zerolog/log"
	"periph.io/x/conn/v3/gpio"
	host "periph.io/x/host/v3"
	"periph.io/x/host/v3/rpi"
)

const POOLING_INTERVAL = time.Second * 60

const DESIRED_FLOW_TEMPERATURE = 55 // max temp for flow

var READ_PARAMETERS = []string{
	"FlowTemp",
	"ReturnTemp",
	"FlowTempDesired",
	"ModulationTempDesired",
}

type eBusClimate struct {
	ebusClient *client.Client
	stateStore climate.ClimateStateStore
	state      *climate.ClimateState

	stopChan chan struct{}

	desiredFlowTemp int
	modulationTemp  int

	flowTemp   float64
	returnTemp float64

	loss3 int
	loss7 int
	power int

	heatingActive bool

	heatingRelay gpio.PinIO

	// TODO independant thermometers
	// external      *bluetooththermostat.BluetoothThermostat
	// internal      *bluetooththermostat.BluetoothThermostat
}

func New(config *config.Config) *eBusClimate {
	log.Debug().Msg("Creating new eBusClimate instance")
	if _, err := host.Init(); err != nil {
		log.Error().Err(err).Msg("Failed to initialize periph.io")
		return nil
	}
	ebusClient := client.New(config, READ_PARAMETERS)

	c := eBusClimate{
		ebusClient:      ebusClient,
		stopChan:        make(chan struct{}),
		stateStore:      climate.NewClimateStore(),
		loss3:           config.Climate.Loss3,
		loss7:           config.Climate.Loss7,
		power:           config.Climate.Power,
		heatingActive:   false,
		heatingRelay:    rpi.P1_31,
		desiredFlowTemp: DESIRED_FLOW_TEMPERATURE,
		// internal:   addThermometer(config.Climate.InternalSensorMAC),
		// external:   addThermometer(config.Climate.ExternalSensorMAC),
	}

	c.state, _ = c.stateStore.Load()
	lastActivity, err := time.Parse(time.RFC3339, c.state.LastActive)
	if err != nil {
		log.Warn().Msgf("Failed to parse last activity time: %v", err)
	} else {
		// estimate heat loss since last activity
		minutes := time.Since(lastActivity).Minutes()
		c.state.HeatLoss = c.state.HeatLoss - c.getMinuteLoss()*minutes
		if c.state.HeatLoss < 0 {
			c.state.HeatLoss = -1
		}
		log.Info().Msgf("Estimated heat loss of %f kWh since last activity %v (%f minutes)", c.state.HeatLoss, lastActivity, minutes)
	}

	c.StartPolling(POOLING_INTERVAL, c.readBoiler)
	c.startCycler()

	// start timer to save state every minute
	go func() {
		ticker := time.NewTicker(time.Minute)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				c.stateStore.Save(c.state)
			case <-c.stopChan:
				return
			}
		}
	}()

	return &c
}

func (c *eBusClimate) Info() ([]string, error) {
	return c.ebusClient.Info()
}

func (c *eBusClimate) readBoiler(client *client.Client) {
	//result :=
	result := client.ReadAll()
	c.onChange(result)
}

// StartPolling starts a timer to read data from ebusClient at the given interval.
func (c *eBusClimate) StartPolling(interval time.Duration, readFunc func(*client.Client)) {
	log.Debug().Msg("Start ebus pulling")
	c.readBoiler(c.ebusClient) // initial read

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

// TODO make it temporary override with expiration
func (c *eBusClimate) SetInsideOverride(temp float64) {
	c.state.InsideTemp = temp
	c.stateStore.Save(c.state)
}

func (c *eBusClimate) SetOutsideOverride(temp float64) {
	c.state.OutsideTemp = temp
	c.stateStore.Save(c.state)
}

func (c *eBusClimate) GetInsideTemp() float64 {
	return c.state.InsideTemp
}

func (c *eBusClimate) GetOutsideTemp() float64 {
	return c.state.OutsideTemp
}

func (c *eBusClimate) GetMode() string {
	return c.state.Mode
}

func (c *eBusClimate) GetTargetTemperature() float64 {
	return c.state.TargetTemperature
}

func (c *eBusClimate) GetHWTargetTemp() int {
	return c.state.HWTargetTemp
}

func (c *eBusClimate) GetFlowTemp() float64 {
	return c.flowTemp
}

func (c *eBusClimate) GetReturnTemp() float64 {
	return c.returnTemp
}

func (c *eBusClimate) GetDesiredFlowTemp() int {
	return c.desiredFlowTemp
}

func (c *eBusClimate) GetPower() int {
	return c.modulationTemp
}

func (c *eBusClimate) IsGasActive() bool {
	return c.heatingActive
}

func (c *eBusClimate) IsPumpActive() bool {
	return c.heatingActive
}

func (c *eBusClimate) IsConnected() bool {
	return false
}

func (c *eBusClimate) GetError() string {
	return ""
}

func (c *eBusClimate) GetConsumption() float64 {
	return c.state.ConsumptionHeating
}

func (c *eBusClimate) GetBoilerInfo() climate.BoilerInfo {
	return climate.BoilerInfo{
		Model:    "",
		Firmware: "",
	}
}

func (c *eBusClimate) SetHWTargetTemp(temp int) error {
	c.state.HWTargetTemp = temp
	c.stateStore.Save(c.state)
	return nil
}

func (c *eBusClimate) SetMode(mode string) error {
	if mode != "off" && mode != "heating" {
		return errors.New("invalid mode")
	}
	c.state.Mode = mode
	c.stateStore.Save(c.state)
	return nil
}

func (c *eBusClimate) SetTargetTemperature(temp float64) error {
	c.state.TargetTemperature = temp
	c.stateStore.Save(c.state)
	return nil
}

func (c *eBusClimate) StartHeating() {
	c.heatingActive = true
	log.Debug().Msg("Starting heating")
	if err := c.heatingRelay.Out(gpio.High); err != nil {
		log.Warn().Err(err).Msg("Failed to set relay pin high")
	}
}

func (c *eBusClimate) StopHeating() {
	c.heatingActive = false
	log.Debug().Msg("Stopping heating")
	if err := c.heatingRelay.Out(gpio.Low); err != nil {
		log.Warn().Err(err).Msg("Failed to set relay pin low")
	}
}

func (c *eBusClimate) pingHeating() {
	mode := "0"
	if c.heatingActive {
		mode = "0"
	}

	//SetModeOverride,
	// hcmode
	// flowtempdesired
	// hwctempdesired
	// hwcflowtempdesired
	// setmode1
	// disablehc
	// disablehwctapping
	// disablehwcload
	// setmode2
	// remoteControlHcPump
	// releaseBackup
	// releaseCooling
	command := fmt.Sprintf("%s;%d;%d;-;-;0;0;0;-;0;0;0", mode, c.desiredFlowTemp, c.state.HWTargetTemp)
	log.Debug().Msgf("Heating string: %s", command)
	c.ebusClient.Set("SetModeOverride", command)
}
