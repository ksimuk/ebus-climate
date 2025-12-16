package config

import (
	"os"

	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Name string `yaml:"name"`
	Ebus struct {
		Host    string `yaml:"host"`
		Port    string `yaml:"port"`
		Circuit string `yaml:"circuit"`
	} `yaml:"ebus"`

	WebPort int `yaml:"web_port"`
	Climate struct {
		Power             int     `yaml:"power"`        // boiler power in kwh
		MinRunTime        float64 `yaml:"min_run_time"` // minimum boiler run time in minutes
		MaxRunTime        float64 `yaml:"max_run_time"` // maximum boiler run time in minutes
		Loss3             int     `yaml:"loss3"`        // heatloss at -3C
		Loss7             int     `yaml:"loss7"`        // heatloss at 7C
		AdjustmentRate    float64 `yaml:"adjustment_rate"`
		InternalSensorMAC string  `yaml:"internal_sensor_mac"`
		ExternalSensorMAC string  `yaml:"external_sensor_mac"`
	}

	LogLevel string `yaml:"log_level"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	// set defaults
	cfg.Climate.AdjustmentRate = 3.0

	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	log.Trace().Msgf("Loaded config from %s: %+v", path, cfg)
	return &cfg, nil
}
