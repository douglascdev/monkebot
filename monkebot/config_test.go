package monkebot

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"
)

func generateMockJSON() []byte {
	cfg := map[string]interface{}{
		"InitialChannels": []string{"hash_table"},
		"TwitchToken":     "YOUR_OAUTH_TOKEN_HERE",
		"Prefix":          "!",
		"UserID":          "YOUR_USER_ID_HERE",
		"Login":           "YOUR_LOGIN_HERE",
		"ClientID":        "YOUR_CLIENT_ID_HERE",
	}

	jsonBytes, err := json.Marshal(cfg)
	if err != nil {
		panic(fmt.Errorf("error marshalling mock json: %v", err))
	}

	return jsonBytes
}

func validateMockJSONConfig(cfg *Config) error {
	if cfg.InitialChannels[0] != "hash_table" {
		return fmt.Errorf("failed to parse initial_channels")
	}
	if cfg.Prefix != "!" {
		return fmt.Errorf("failed to parse prefix")
	}
	if cfg.UserID != "YOUR_USER_ID_HERE" {
		return fmt.Errorf("failed to parse user_id")
	}
	if cfg.Login != "YOUR_LOGIN_HERE" {
		return fmt.Errorf("failed to parse login")
	}
	if cfg.ClientID != "YOUR_CLIENT_ID_HERE" {
		return fmt.Errorf("failed to parse client_id")
	}
	return nil
}

func TestLoadConfig(t *testing.T) {
	mockJSONBytes := generateMockJSON()

	cfg, err := LoadConfig(mockJSONBytes)
	if err != nil {
		t.Errorf("failed to load config: %v", err)
	}

	err = validateMockJSONConfig(cfg)
	if err != nil {
		t.Errorf("failed to validate config: %v", err)
	}
}

func TestLoadConfigFromFile(t *testing.T) {
	file, err := os.CreateTemp(os.TempDir(), "monkebotTestJson")
	if err != nil {
		t.Errorf("failed to create temp file: %v", err)
	}
	defer file.Close()

	mockJSONBytes := generateMockJSON()
	file.WriteString(string(mockJSONBytes))

	config, err := LoadConfigFromFile(file.Name())
	if err != nil {
		t.Errorf("failed to load config from file: %v", err)
	}

	err = validateMockJSONConfig(config)
	if err != nil {
		t.Errorf("failed to validate config: %v", err)
	}
}
