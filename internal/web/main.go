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

type Boiler struct {
	Name      string `json:"name"`
	Model     string `json:"model"`
	Firmware  string `json:"firmware"`
	Connected bool   `json:"connected"`
	Error     string `json:"error"`
}

type Get struct {
	// set parameters
	Mode              string  `json:"mode"` // off, heating
	TargetTemperature float64 `json:"target_temperature"`
	HWTargetTemp      int     `json:"hw_target_temp"`

	// info from  climate
	OutsideTemp     float64 `json:"outside_temp"`
	InsideTemp      float64 `json:"inside_temp"`
	HeatLossBalance float64 `json:"heat_loss_balance"`

	FlowTemp        float64 `json:"flow_temp"`
	ReturnTemp      float64 `json:"return_temp"`
	DesiredFlowTemp int     `json:"desired_flow_temp"`
	Power           int     `json:"power"`

	GasActive  bool   `json:"gas_active"`
	PumpActive bool   `json:"pump_active"`
	Boiler     Boiler `json:"boiler"`

	Stat climate.Stat `json:"stat"`
}

func GetServer(config config.Config) Server {
	climate := vailant.New(&config) // TODO: make load boiler type from config when change boiler
	server := Server{
		config:  config,
		climate: climate,
	}
	return server
}

func (s *Server) Start() {
	log.Info().Msgf("Starting web server on port %d...", s.config.WebPort)
	http.HandleFunc("/set", s.handleSet)
	http.HandleFunc("/get", s.handleGet)

	// Override endpoints for temperature sensors
	http.HandleFunc("/override", s.handleOverride)
	http.HandleFunc("/force_heating", s.handleForceHeating)
	http.HandleFunc("/check", func(w http.ResponseWriter, r *http.Request) {
		// todo authentication check
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	http.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	log.Fatal().Err(http.ListenAndServe(fmt.Sprintf(":%d", s.config.WebPort), nil)).Msg("Web server stopped")
}

func (s *Server) handleForceHeating(w http.ResponseWriter, r *http.Request) {
	duration := r.URL.Query().Get("duration")
	duration_int, err := strconv.Atoi(duration)
	if err != nil {
		log.Error().Err(err).Msg("Failed to parse duration")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	log.Info().Msgf("Forcing heating for %d minutes", duration_int)
	s.climate.RunFor(duration_int)
	w.WriteHeader(http.StatusOK)
}

func (s *Server) handleSet(w http.ResponseWriter, r *http.Request) {
	var state Set

	if err := json.NewDecoder(r.Body).Decode(&state); err != nil {
		log.Error().Err(err).Msg("Failed to decode request body")
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	log.Debug().Msgf("Received set request: %+v", state)

	if state.Mode != "" {
		log.Info().Msgf("Setting mode to %s", state.Mode)
		if err := s.climate.SetMode(state.Mode); err != nil {
			log.Error().Err(err).Msgf("Failed to set mode to %s", state.Mode)
			http.Error(w, "Failed to set mode", http.StatusInternalServerError)
			return
		}
	} else {
		log.Debug().Msgf("Mode not set")
	}

	if state.TargetTemperature > 0 {
		log.Info().Msgf("Setting target temperature to %f", state.TargetTemperature)
		if err := s.climate.SetTargetTemperature(state.TargetTemperature); err != nil {
			log.Error().Err(err).Msgf("Failed to set target temperature to %f", state.TargetTemperature)
			http.Error(w, "Failed to set target temperature", http.StatusInternalServerError)
			return
		}
	} else {
		log.Debug().Msgf("Target temperature not set")
	}

	if state.HWTargetTemp > 0 {
		log.Info().Msgf("Setting hot water target temperature to %d", state.HWTargetTemp)
		if err := s.climate.SetHWTargetTemp(state.HWTargetTemp); err != nil {
			log.Error().Err(err).Msgf("Failed to set hot water target temperature to %d", state.HWTargetTemp)
			http.Error(w, "Failed to set hot water target temperature", http.StatusInternalServerError)
			return
		}
	} else {
		log.Debug().Msgf("Hot water target temperature not set")
	}

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
		s.climate.OverrideHeating(600) // force heating for 10 minutes
	case "0":
		log.Info().Msg("Forcing heating OFF")
		s.climate.StopHeating()
	}

	w.WriteHeader(http.StatusOK)
}

func (s *Server) handleGet(w http.ResponseWriter, r *http.Request) {
	if s.config.Name == "" {
		s.config.Name = "Glow Worm Ultimate 3 35C"
	}
	state := Get{
		Mode:              s.climate.GetMode(),
		TargetTemperature: s.climate.GetTargetTemperature(),
		HWTargetTemp:      s.climate.GetHWTargetTemp(),

		Boiler: Boiler{
			Name:      s.config.Name,
			Connected: s.climate.IsConnected(),
			Error:     s.climate.GetError(),
		},

		InsideTemp:  s.climate.GetInsideTemp(),
		OutsideTemp: s.climate.GetOutsideTemp(),

		FlowTemp:        s.climate.GetFlowTemp(),
		ReturnTemp:      s.climate.GetReturnTemp(),
		DesiredFlowTemp: s.climate.GetDesiredFlowTemp(),
		Power:           s.climate.GetPower(),

		GasActive:  s.climate.IsGasActive(),
		PumpActive: s.climate.IsPumpActive(),

		HeatLossBalance: s.climate.GetHeatLossBalance(),
		Stat:            s.climate.GetStat(),
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(state); err != nil {
		log.Error().Err(err).Msg("Failed to encode response")
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

func (s *Server) Shutdown() {
	s.climate.Shutdown()
}
