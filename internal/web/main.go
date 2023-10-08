package web

import (
	"github.com/ksimuk/ebus-climate/internal/climate"
	"github.com/rs/zerolog/log"
)

type Server struct {
	climate *climate.Climate
}

type State struct {
	// Config
	MinTemp float64 `json:"min_temp"`
	MaxTemp float64 `json:"max_temp"`

	DesiredHotWaterTemp int `json:"desired_hot_water_temp"`
	DesiredFlowTemp     int `json:"desired_flow_temp"`

	HotWaterEnabled bool `json:"hot_water_enabled"`
	HeatingEnabled  bool `json:"heating_enabled"`

	// Sensors
	CurrentTemp     float64 `json:"current_temp"`
	CurrentHumidity float64 `json:"current_humidity"`

	//State
	FlowTemp  float64 `json:"flow_temp"`
	IsHeating bool    `json:"is_heating"`
}

func Start(port int, climate *climate.Climate) {
	println("Starting web server on port", port)
	server := Server{
		climate: climate,
	}
	server.start()
}

func (s *Server) start() {
}

func (s *Server) setState(state State) {
	s.climate.SetMaxTemperature(state.MaxTemp)
	s.climate.SetMinTemperature(state.MinTemp)
	s.climate.SetDesiredHotWaterTemperature(state.DesiredHotWaterTemp)
	s.climate.SetDesiredFlowTemperature(state.DesiredFlowTemp)

}

func (s *Server) getState() {
	state := State{
		MinTemp:             s.climate.MinTemperature(),
		MaxTemp:             s.climate.MaxTemperature(),
		CurrentTemp:         s.climate.CurrentTemperature(),
		CurrentHumidity:     s.climate.CurrentHumidity(),
		DesiredHotWaterTemp: s.climate.DesiredHotWaterTemperature(),
		DesiredFlowTemp:     s.climate.DesiredFlowTemperature(),
		FlowTemp:            s.climate.FlowTemperature(),

		IsHeating: s.climate.IsHeating(),
	}
	log.Debug().Msgf("State: %+v", state)
}
