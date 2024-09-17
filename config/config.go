package config

import (
	"encoding/json"
	"fmt"
	"reflect"
)

type DBConfig struct {
	Driver         string `json:"Driver"`
	DataSourceName string `json:"DataSourceName"`
	Version        int    `json:"Version"` // used to keep track of migrations. 0 means the tables were not created yet.
}

// changes to this struct must be reflected in tests and config.json
type Config struct {
	TwitchToken     string   `json:"TwitchToken"`
	InitialChannels []string `json:"InitialChannels"`
	Prefix          string   `json:"Prefix"`
	UserID          string   `json:"UserID"`
	AdminUsernames  []string `json:"AdminUsernames"`
	Login           string   `json:"Login"`
	ClientID        string   `json:"ClientID"`
	DBConfig        DBConfig `json:"DBConfig"`
}

// unmarshal config and ensure every field is set or return an error
func LoadConfig(JSONData []byte) (*Config, error) {
	var cfg Config

	err := json.Unmarshal(JSONData, &cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal config template: %w", err)
	}

	fields := reflect.ValueOf(&cfg).Elem()
	for i := 0; i < fields.NumField(); i++ {
		if fields.Field(i).IsZero() {
			return nil, fmt.Errorf("missing field: %s", fields.Type().Field(i).Name)
		}
	}

	return &cfg, nil
}

func MarshalConfig(cfg *Config) ([]byte, error) {
	jsonBytes, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("error marshalling config: %w", err)
	}
	return jsonBytes, nil
}

func ConfigTemplateJSON() ([]byte, error) {
	cfg := Config{
		InitialChannels: []string{"hash_table"},
		TwitchToken:     "YOUR_OAUTH_TOKEN_HERE",
		Prefix:          "!",
		UserID:          "YOUR_USER_ID_HERE",
		AdminUsernames:  []string{"hash_table"},
		Login:           "YOUR_LOGIN_HERE",
		ClientID:        "YOUR_CLIENT_ID_HERE",
		DBConfig: DBConfig{
			Driver:         "sqlite3",
			DataSourceName: "file:data.db",
			Version:        0,
		},
	}

	jsonBytes, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("error marshalling template JSON: %w", err)
	}
	return jsonBytes, nil
}
