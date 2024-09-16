package config

import (
	"fmt"
	"testing"
)

func validateTemplateJSONConfig(cfg *Config) error {
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
	mockJSONBytes, err := ConfigTemplateJSON()
	if err != nil {
		t.Errorf("failed to generate config template: %v", err)
	}

	cfg, err := LoadConfig(mockJSONBytes)
	if err != nil {
		t.Errorf("failed to load config: %v", err)
	}

	err = validateTemplateJSONConfig(cfg)
	if err != nil {
		t.Errorf("failed to validate config: %v", err)
	}
}

func TestMarshalConfig(t *testing.T) {
	mockJSONBytes, err := ConfigTemplateJSON()
	if err != nil {
		t.Errorf("failed to generate config template: %v", err)
	}

	cfg, err := LoadConfig(mockJSONBytes)
	if err != nil {
		t.Errorf("failed to load config: %v", err)
	}

	_, err = MarshalConfig(cfg)
	if err != nil {
		t.Errorf("failed to marshal config: %v", err)
	}
}

// check if all Config values are set in the template using reflection.
// guarantees that changes made in Config are saved in new templates.
func TestConfigTemplateJSON(t *testing.T) {
	// see if the function works against a known invalid template.
	// incompleteTemplate doesn't have TwitchToken and Login fields.
	incompleteTemplate := []byte(`
{
	"InitialChannels": ["abc"],
  "ClientID": "123",
  "UserID": "sdasd",
  "Prefix": "!"
}
	`)
	_, err := LoadConfig(incompleteTemplate)
	if err == nil {
		t.Errorf("partially initialized config template not caught by LoadConfig validation. Template: %s", string(incompleteTemplate))
	}

	templateBytes, err := ConfigTemplateJSON()
	if err != nil {
		t.Errorf("failed to generate config template: %v", err)
	}

	// test against the actual template
	_, err = LoadConfig(templateBytes)
	if err != nil {
		t.Errorf("failed to validate config template: %v", err)
	}
}
