package monkebot

import (
	"encoding/json"
	"fmt"
	"os"
)

// changes to this struct must be reflected in tests and config.json
type Config struct {
	InitialChannels []string `json:"InitialChannels"`
	Prefix          string   `json:"Prefix"`
	UserID          string   `json:"UserID"`
	Login           string   `json:"Login"`
	ClientID        string   `json:"ClientID"`
}

func LoadConfig(data []byte) (*Config, error) {
	var config Config
	err := json.Unmarshal(data, &config)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &config, err
}

func LoadConfigFromFile(filename string) (*Config, error) {
	bytes, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open config file: %w", err)
	}

	config, err := LoadConfig(bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return config, nil
}
