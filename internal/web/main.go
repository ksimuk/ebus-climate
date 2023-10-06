package web

import (
	"github.com/ksimuk/ebus-climate/internal/climate"
	"github.com/rs/zerolog/log"
)

type Server struct {
	climate *climate.Climate
}

type State struct {
	// Sensors
	MinTemp     float64
	MaxTemp     float64
	CurrentTemp float64

	// Boiler
	HotWaterTemp    int
	DesiredFlowTemp int

	HotWaterEnabled bool
	HeatingEnabled  bool

	//State
	FlowTemp float64

	isHeating bool
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

func (s *Server) setState() {
}

func (s *Server) getState() {
	state := State{
		MinTemp:     s.climate.MinTemperature(),
		MaxTemp:     s.climate.MaxTemperature(),
		CurrentTemp: s.climate.CurrentTemperature(),

		HotWaterTemp:    s.climate.HotWaterTemperature(),
		DesiredFlowTemp: s.climate.DesiredFlowTemperature(),
		FlowTemp:        s.climate.FlowTemperature(),

		isHeating: s.climate.IsHeating(),
	}
	log.Debug().Msgf("State: %+v", state)
}
