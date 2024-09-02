package monkebot_test

import (
	"encoding/json"
	"monkebot/monkebot"
	"testing"
)

func TestMonkebotLoadConfig(t *testing.T) {
	jsonMap := map[string]interface{}{
		"initial_channels": []string{"hash_table"},
		"twitch_token":     "oauth:YOUR_OAUTH_TOKEN_HERE",
		"prefix":           "!",
		"user_id":          "YOUR_USER_ID_HERE",
		"client_id":        "YOUR_CLIENT_ID_HERE",
	}
	jsonBytes, err := json.Marshal(jsonMap)
	if err != nil {
		t.Errorf("error marshalling test json: %v", err)
	}

	err = monkebot.LoadConfig(jsonBytes)
	if err != nil {
		t.Errorf("error parsing config: %v", err)
	}
}
