package monkebot

import (
	"strings"
	"testing"
)

// implementation of MessageSender for testing
type MockSender struct {
	responses []string
}

func (m *MockSender) Say(channel string, message string) {
	m.responses = append(m.responses, message)
}

func TestCommandMap(t *testing.T) {
	if len(commandMap) != len(commands) {
		t.Errorf("expected %d commands, got %d", len(commands), len(commandMap))
	}

	for _, cmd := range commands {
		if _, ok := commandMap[cmd.Name]; !ok {
			t.Errorf("command '%s' not found in commandMap", cmd.Name)
		}

		for _, alias := range cmd.Aliases {
			if _, ok := commandMap[alias]; !ok {
				t.Errorf("command alias '%s' not found in commandMap", alias)
			}
		}
	}
}

func TestCommandSenzp(t *testing.T) {
	cmd := commandMap["senzpTest"]
	expectedResponses := map[string]string{
		"🅰️ 🅱️ ©️ ↩️ 📧 🎏 🗜️ ♓ ℹ️ 🗾 🎋 👢 〽️ ♑ 🅾️ 🅿️ ♌ ®️ ⚡ 🌴 ⛎ ♈ 〰️ ❌ 🌱 💤":                                          "abcdefghijklmnopqrstuvwxyz",
		"♓ 🅰️ ⚡ senzpTest 🌴 🅾️ senzpTest ↩️ 🅾️ senzpTest 〰️ ℹ️ 🌴 ♓ senzpTest 〽️ ℹ️ ↩️ ↩️ 👢 📧 senzpTest ♑ 🅰️ 〽️ 📧": "has to do with middle name",
		"🅿️ 🅰️ 👢 👢 🌱": "pally",
		"©️ 🅾️ ↩️":    "cod",
		"🅰️ 🅿️ 📧 ❌":   "apex",
		"exemYes ℹ️ senzpTest ©️ 🅰️ ♑ senzpTest ⛎ ⚡ 📧 senzpTest ©️ ♓ ®️ 🅾️ 〽️ 📧":                                                                                  "Yes i can use chrome",
		"ℹ️ ⚡ senzpTest 🌴 ♓ 📧 ®️ 📧 senzpTest 🅰️ senzpTest 🎏 ®️ 📧 ♌ ⛎ 📧 ♑ 🌴 👢 🌱 senzpTest ⛎ ⚡ 📧 ↩️ senzpTest 📧 〽️ 🅾️ 🌴 📧 senzpTest 🌴 ♓ ℹ️ ♑ 🗜️ elisAsk mysztiHmmm": "is there a frequently used emote thing catAsk hmm",
		"peeepoHUH": "wtfwtfwtf",
	}

	sender := &MockSender{
		responses: []string{},
	}

	for input, expected := range expectedResponses {
		err := cmd.Execute(&Message{Channel: "test"}, sender, strings.Split(input, " "))
		if err != nil {
			t.Errorf("unexpected error: %s", err)
		}

		if sender.responses[len(sender.responses)-1] != expected {
			t.Errorf("expected '%s' for input '%s', got '%s'", expected, input, sender.responses[len(sender.responses)-1])
		}
	}
}
