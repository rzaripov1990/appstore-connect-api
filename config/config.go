package config

import (
	"encoding/json"
	"os"
)

type Config struct {
	Kid   string `json:"kid"`
	Iis   string `json:"iis"`
	P8Key string `json:"p8key"`
}

func New() *Config {
	configFile, err := os.ReadFile("./config.json")
	if err != nil {
		panic(err)
	}

	cfg := &Config{}
	err = json.Unmarshal(configFile, cfg)
	if err != nil {
		panic(err)
	}

	return cfg
}
