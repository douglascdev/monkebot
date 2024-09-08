package monkebot

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
)

// changes to this struct must be reflected in tests and config.json
type Config struct {
	TwitchToken     string   `json:"TwitchToken"`
	InitialChannels []string `json:"InitialChannels"`
	Prefix          string   `json:"Prefix"`
	UserID          string   `json:"UserID"`
	Login           string   `json:"Login"`
	ClientID        string   `json:"ClientID"`
}

// unmarshal config and ensure every field is set or return an error
func LoadConfig(JSONData []byte) (*Config, error) {
	var cfg Config

	err := json.Unmarshal(JSONData, &cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal config template: %v", err)
	}

	fields := reflect.ValueOf(&cfg).Elem()
	for i := 0; i < fields.NumField(); i++ {
		if fields.Field(i).IsZero() {
			return nil, fmt.Errorf("missing field: %s", fields.Type().Field(i).Name)
		}
	}

	return &cfg, nil
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

func ConfigTemplateJSON() ([]byte, error) {
	cfg := Config{
		InitialChannels: []string{"hash_table"},
		TwitchToken:     "YOUR_OAUTH_TOKEN_HERE",
		Prefix:          "!",
		UserID:          "YOUR_USER_ID_HERE",
		Login:           "YOUR_LOGIN_HERE",
		ClientID:        "YOUR_CLIENT_ID_HERE",
	}

	jsonBytes, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("error marshalling template JSON: %v", err)
	}
	return jsonBytes, nil
}
