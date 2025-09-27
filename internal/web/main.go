package web

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

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
	Mode              string  `json:"mode"` // off, heating
	TargetTemperature float64 `json:"target_temperature"`
	HWTargetTemp      int     `json:"hw_target_temp"`
}

type Get struct {
	// set parameters
	Mode              string  `json:"mode"` // off, heating
	TargetTemperature float64 `json:"target_temperature"`
	HWTargetTemp      int     `json:"hw_target_temp"`

	// info from  climate
	OutsideTemp float64 `json:"outside_temp"`
	InsideTemp  float64 `json:"inside_temp"`

	FlowTemp        float64 `json:"flow_temp"`
	ReturnTemp      float64 `json:"return_temp"`
	DesiredFlowTemp int     `json:"desired_flow_temp"`
	Power           int     `json:"power"`

	GasActive  bool `json:"gas_active"`
	PumpActive bool `json:"pump_active"`
	Boiler     struct {
		Model    string `json:"model"`
		Firmware string `json:"firmware"`
	} `json:"boiler"`

	ConsumptionHeating float64 `json:"consumption_heating"` // total consumption for heating in kWh
	Connected          bool    `json:"connected"`
	Error              string  `json:"error"`
}

func Start(config config.Config) {
	println("Starting web server on port", config.WebPort)
	climate := vailant.New(&config) // TODO: make load boiler type from config when change boiler
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

	// Override endpoints for temperature sensors
	http.HandleFunc("/override", s.handleOverride)

	log.Fatal().Err(http.ListenAndServe(fmt.Sprintf(":%d", s.config.WebPort), nil)).Msg("Web server stopped")
}

func (s *Server) handleSet(w http.ResponseWriter, r *http.Request) {
	var state Set
	if err := json.NewDecoder(r.Body).Decode(&state); err != nil {
		log.Error().Err(err).Msg("Failed to decode request body")
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if err := s.climate.SetMode(state.Mode); err != nil {
		log.Error().Err(err).Msgf("Failed to set mode to %s", state.Mode)
		http.Error(w, "Failed to set mode", http.StatusInternalServerError)
		return
	}
	if err := s.climate.SetTargetTemperature(state.TargetTemperature); err != nil {
		log.Error().Err(err).Msgf("Failed to set target temperature to %f", state.TargetTemperature)
		http.Error(w, "Failed to set target temperature", http.StatusInternalServerError)
		return
	}
	if err := s.climate.SetHWTargetTemp(state.HWTargetTemp); err != nil {
		log.Error().Err(err).Msgf("Failed to set hot water target temperature to %d", state.HWTargetTemp)
		http.Error(w, "Failed to set hot water target temperature", http.StatusInternalServerError)
		return
	}
	log.Info().Msgf("Set mode to %s, target temperature to %f, hot water target temperature to %d", state.Mode, state.TargetTemperature, state.HWTargetTemp)

	w.WriteHeader(http.StatusOK)
}

func (s *Server) handleOverride(w http.ResponseWriter, r *http.Request) {
	inside := r.URL.Query().Get("inside_temp")
	if inside != "" {
		log.Info().Msgf("Overriding inside temperature to %s", inside)
		inside_float, err := strconv.ParseFloat(inside, 64)
		if err != nil {
			log.Error().Err(err).Msgf("Failed to parse inside temperature: %s", inside)
			http.Error(w, "Invalid inside temperature", http.StatusBadRequest)
			return
		}
		s.climate.SetInsideOverride(inside_float) // TODO: get value from request
	}
	outside := r.URL.Query().Get("outside_temp")
	if outside != "" {
		log.Info().Msgf("Overriding outside temperature to %s", outside)
		outside_float, err := strconv.ParseFloat(outside, 64)
		if err != nil {
			log.Error().Err(err).Msgf("Failed to parse outside temperature: %s", outside)
			http.Error(w, "Invalid outside temperature", http.StatusBadRequest)
			return
		}
		s.climate.SetOutsideOverride(outside_float)
	}

	forceHeating := r.URL.Query().Get("force_heating")
	switch forceHeating {
	case "1":
		log.Info().Msg("Forcing heating ON")
		s.climate.StartHeating()
	case "0":
		log.Info().Msg("Forcing heating OFF")
		s.climate.StopHeating()
	}

	w.WriteHeader(http.StatusOK)
}

func (s *Server) handleGet(w http.ResponseWriter, r *http.Request) {
	state := Get{
		Mode:              s.climate.GetMode(),
		TargetTemperature: s.climate.GetTargetTemperature(),
		HWTargetTemp:      s.climate.GetHWTargetTemp(),

		InsideTemp:  s.climate.GetInsideTemp(),
		OutsideTemp: s.climate.GetOutsideTemp(),

		FlowTemp:        s.climate.GetFlowTemp(),
		ReturnTemp:      s.climate.GetReturnTemp(),
		DesiredFlowTemp: s.climate.GetDesiredFlowTemp(),
		Power:           s.climate.GetPower(),

		GasActive:  s.climate.IsGasActive(),
		PumpActive: s.climate.IsPumpActive(),

		Connected:          s.climate.IsConnected(),
		Error:              s.climate.GetError(),
		ConsumptionHeating: s.climate.GetConsumption(),
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(state); err != nil {
		log.Error().Err(err).Msg("Failed to encode response")
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}
