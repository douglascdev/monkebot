package command

import (
	"fmt"
	"monkebot/database"
	"monkebot/types"
)

var disable = types.Command{
	Name:              "disable",
	Aliases:           []string{},
	Usage:             "disable [command]",
	Description:       "Disables a command for all users in the channel",
	ChannelCooldown:   5,
	UserCooldown:      5,
	NoPrefix:          false,
	NoPrefixShouldRun: nil,
	CanDisable:        false,
	Execute: func(message *types.Message, sender types.MessageSender, args []string) error {
		if len(args) != 2 {
			sender.Say(message.Channel, "Usage: disable [command]")
			return nil
		}

		var (
			command types.Command
			ok      bool
			err     error
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

		if !command.CanDisable {
			sender.Say(message.Channel, "❌This command cannot be disabled")
			return nil
		}

		if !(message.Chatter.IsMod || message.Chatter.IsBroadcaster) {
			sender.Say(message.Channel, "❌You must be a moderator to use this command")
			return nil
		}

		tx, err := message.DB.Begin()
		if err != nil {
			return err
		}
		defer tx.Rollback()

		err = database.UpdateIsUserCommandEnabled(tx, false, message.RoomID, command.Name)
		if err != nil {
			return err
		}

		err = tx.Commit()
		if err != nil {
			return err
		}

		sender.Say(message.Channel, fmt.Sprintf("✅Disabled command '%s'", command.Name))
		return nil
	},
}
