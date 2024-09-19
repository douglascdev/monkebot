package command

import (
	"fmt"
)

var help = Command{
	Name:              "help",
	Aliases:           []string{},
	Usage:             "help | help <command>",
	Description:       "Responds with pong and latency to twitch in milliseconds",
	Cooldown:          5,
	NoPrefix:          false,
	NoPrefixShouldRun: nil,
	CanDisable:        false,
	Execute: func(message *Message, sender MessageSender, args []string) error {
		if len(args) <= 1 {
			sender.Say(message.Channel, "🐒 Commands: https://douglascdev.github.io/monkebot/ ● For help with a specific command: help <command>")
			return nil
		}

		cmd, ok := commandMap[args[1]]
		if !ok {
			sender.Say(message.Channel, fmt.Sprintf("❌Unknown command '%s'", args[1]))
			return nil
		}

		sender.Say(message.Channel, fmt.Sprintf("🐒 Usage: %s", cmd.Usage))
		return nil
	},
}
