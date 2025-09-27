package climate

import (
	"encoding/json"
	"os"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
)

type ClimateState struct {
	Mode              string  `json:"mode"`               // off, heating
	TargetTemperature float64 `json:"target_temperature"` // target temperature for heating
	HWTargetTemp      int     `json:"hw_target_temp"`     // target temperature for hot water

	InsideTemp  float64 `json:"inside_temp"`  // current inside temperature
	OutsideTemp float64 `json:"outside_temp"` // current outside temperature

	LastActive string `json:"last_active"` // last active timestamp

	ConsumptionHeating float64 `json:"consumption_heating"` // total consumption for heating in kWh

	HeatLoss float64 `json:"heat_loss"` // current heat loss balance
}

type ClimateStateStore interface {
	Load() (*ClimateState, error)
	Save(state *ClimateState) error
}

type FileClimateStore struct {
	filePath     string
	mu           sync.Mutex
	pendingState *ClimateState
	timer        *time.Timer
}

const saveDelay = 120 * time.Second

func NewClimateStore() ClimateStateStore {
	return &FileClimateStore{
		filePath: "climate.data",
	}
}

// Load loads the climate state from the file.
func (s *FileClimateStore) Load() (*ClimateState, error) {
	file, err := os.Open(s.filePath)
	if err != nil {
		return &ClimateState{}, err
	}
	defer file.Close()

	var state ClimateState
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&state); err != nil {
		return &ClimateState{}, err
	}
	return &state, nil
}

// Save schedules saving the climate state to the file after a delay.
func (s *FileClimateStore) Save(state *ClimateState) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.timer != nil {
		s.pendingState = state
		return nil
	}

	s.timer = time.AfterFunc(saveDelay, func() {
		s.mu.Lock()
		defer s.mu.Unlock()
		if s.pendingState != nil {
			log.Debug().Msg("Saving pending climate state to file")
			file, err := os.Create(s.filePath)
			if err == nil {
				encoder := json.NewEncoder(file)
				encoder.SetIndent("", "  ")
				_ = encoder.Encode(s.pendingState)
				file.Close()
			}
			s.pendingState = nil
		}
		s.timer.Stop()
		s.timer = nil
	})

	return nil
}
