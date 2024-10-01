package command

import (
	"fmt"
	"monkebot/locale"
	"monkebot/types"

	"github.com/nicksnyder/go-i18n/v2/i18n"
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
		userLocale := "pt-BR"

		if len(args) <= 1 {
			msg := &i18n.LocalizeConfig{
				DefaultMessage: &i18n.Message{
					ID:    "help.commands",
					Other: "🐒 Commands: {{.Url}} ● For help with a specific command: help <command>",
				},
				TemplateData: map[string]string{
					"Url": "https://douglascdev.github.io/monkebot/",
				},
			}
			sender.Say(message.Channel, locale.PT.MustLocalize(msg))
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
