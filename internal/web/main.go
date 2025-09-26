package web

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/ksimuk/ebus-climate/internal/climate"
	"github.com/ksimuk/ebus-climate/internal/config"
	"github.com/ksimuk/ebus-climate/internal/vailant"
	"github.com/rs/zerolog/log"
)

type Server struct {
	climate climate.Climate
	config  config.Config
}

type Set struct {
	Mode              string `json:"mode"` // off, heating
	TargetTemperature int    `json:"target_temperature"`
	Boost             bool   `json:"boost"`
}

type Get struct {
	FlowTemp        float64 `json:"flow_temp"`
	ReturnTemp      float64 `json:"return_temp"`
	DesiredFlowTemp int     `json:"desired_flow_temp"`
	Power           int     `json:"power"`

	GasActive  bool `json:"gas_active"`
	PumpActive bool `json:"pump_active"`
	Boiler     struct {
		Model    string `json:"model"`
		Firmware string `json:"firmware"`
	}
	Connected bool   `json:"connected"`
	Error     string `json:"error"`
}

func Start(config config.Config) {
	println("Starting web server on port", config.WebPort)
	climate := vailant.New(&config)
	server := Server{
		config:  config,
		climate: climate,
	}
	server.start()
}

func (s *Server) start() {
	log.Info().Msg("Starting web server...")
	http.HandleFunc("/set", s.handleSet)
	http.HandleFunc("/get", s.handleGet)
	log.Fatal().Err(http.ListenAndServe(fmt.Sprintf(":%d", s.config.WebPort), nil)).Msg("Web server stopped")
}

func (s *Server) handleSet(w http.ResponseWriter, r *http.Request) {
	var state Set
	if err := json.NewDecoder(r.Body).Decode(&state); err != nil {
		log.Error().Err(err).Msg("Failed to decode request body")
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	log.Debug().Msgf("Set: %+v", state)
}

func (s *Server) handleGet(w http.ResponseWriter, r *http.Request) {
	climateInfo, err := s.climate.Info()
	if err != nil {
		log.Error().Err(err).Msg("Failed to get climate info")
		http.Error(w, "Failed to get climate info", http.StatusInternalServerError)
		return
	}
	log.Debug().Msgf("Climate info: %+v", climateInfo)
	state := Get{}
	state.Connected = false
	state.Error = "Not connected to ebusd"
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(state); err != nil {
		log.Error().Err(err).Msg("Failed to encode response")
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
	log.Debug().Msgf("State: %+v", state)
}
