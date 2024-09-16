package command

import (
	"fmt"
	"time"
)

var ping = Command{
	Name:        "ping",
	Aliases:     []string{},
	Usage:       "ping",
	Description: "Responds with pong and latency to twitch in milliseconds",
	Cooldown:    5,
	NoPrefix:    false,
	Execute: func(message *Message, sender MessageSender, args []string) error {
		latency := fmt.Sprintf("%d ms", time.Since(message.Time).Milliseconds())
		sender.Say(message.Channel, fmt.Sprintf("🐒 Pong! Latency: %s", latency))
		return nil
	},
}
