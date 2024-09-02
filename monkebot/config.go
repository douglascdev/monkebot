package monkebot

import (
	"encoding/json"
	"fmt"
)

// changes to this struct must be reflected in tests and config.json
type config struct {
	InitialChannels []string `json:"initial_channels"`
	TwitchToken     string   `json:"twitch_token"`
	Prefix          string   `json:"prefix"`
	UserId          string   `json:"user_id"`
	ClientId        string   `json:"client_id"`
}

var Config config

func LoadConfig(data []byte) error {
	err := json.Unmarshal(data, &Config)
	if err != nil {
		return fmt.Errorf("error parsing config: %w", err)
	}

	return nil
}
