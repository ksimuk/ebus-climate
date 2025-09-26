package config

import (
	"log"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Ebus struct {
		Host string `yaml:"host"`
		Port int    `yaml:"port"`
	} `yaml:"ebus"`

	WebPort int `yaml:"web_port"`
	Climate struct {
		Power                int     `yaml:"power"`        // boiler power in kwh
		MinRunTime           float64 `yaml:"min_run_time"` // minimum boiler run time in minutes
		MaxRunTime           float64 `yaml:"max_run_time"` // maximum boiler run time in minutes
		FlowCurve            string  `yaml:"flow_curve"`   // flow temperature curve
		TemperatureSensorMAC string  `yaml:"temperature_sensor_mac"`
	}
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	log.Printf("Loaded config: %+v\n", cfg)
	return &cfg, nil
}
