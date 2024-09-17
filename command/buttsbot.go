package command

import (
	"fmt"
	"time"
)

var buttsbot = Command{
	Name:        "buttsbot",
	Usage:       "buttsbot <on | off>",
	Description: "Randomly replaces message syllables with butt",
	Cooldown:    5,
	NoPrefix:    false,
	Execute: func(message *Message, sender MessageSender, args []string) error {
		latency := fmt.Sprintf("%d ms", time.Since(message.Time).Milliseconds())
		sender.Say(message.Channel, fmt.Sprintf("🐒 Pong! Latency: %s", latency))
		return nil
	},
}
