package command

import (
	"fmt"
	"monkebot/types"
)

var help = types.Command{
	Name:              "help",
	Aliases:           []string{"commands"},
	Usage:             "help | help [command]",
	Description:       "Get the full list of commands, or help with a specific command",
	ChannelCooldown:   5,
	UserCooldown:      5,
	NoPrefix:          false,
	NoPrefixShouldRun: nil,
	CanDisable:        false,
	Execute: func(message *types.Message, sender types.MessageSender, args []string) error {
		if len(args) <= 1 {
			sender.Say(message.Channel, "🐒 Commands: https://douglascdev.github.io/monkebot/ ● For help with a specific command: help <command>")
			return nil
		}

		var (
			command types.Command
			ok      bool
		)
		if command, ok = commandMap[args[1]]; !ok {
			found := false
			for _, cmd := range commandsNoPrefix {
				if cmd.Name == args[1] {
					command = cmd
					found = true
					break
				}
			}
			if !found {
				sender.Say(message.Channel, fmt.Sprintf("❌Unknown command '%s'", args[1]))
				return nil
			}
		}

		sender.Say(message.Channel, fmt.Sprintf("🐒 Usage: %s", command.Usage))
		return nil
	},
}
