package command

import (
	"fmt"
	"monkebot/types"
	"time"
)

var ping = types.Command{
	Name:              "ping",
	Aliases:           []string{},
	Usage:             "ping",
	Description:       "Responds with pong and latency to twitch in milliseconds",
	Cooldown:          5,
	NoPrefix:          false,
	NoPrefixShouldRun: nil,
	CanDisable:        false,
	Execute: func(message *types.Message, sender types.MessageSender, args []string) error {
		latency := fmt.Sprintf("%d ms", time.Since(message.Time).Milliseconds())
		sender.Say(message.Channel, fmt.Sprintf("🐒 Pong! Latency: %s", latency))
		return nil
	},
}
