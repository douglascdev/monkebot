package command

import (
	"fmt"
	"monkebot/database"
)

var setenabled = Command{
	Name:           "setenabled",
	Usage:          "setenabled <command> <on | off>",
	Description:    "Enables or disables a command for all users in the channel",
	Cooldown:       5,
	NoPrefix:       false,
	NoPrefixRegexp: nil,
	CanDisable:     false,
	Execute: func(message *Message, sender MessageSender, args []string) error {
		if len(args) != 3 || (args[2] != "on" && args[2] != "off") {
			sender.Say(message.Chatter.Name, "Usage: setenabled <command> <on | off>")
			return nil
		}

		var (
			command Command
			ok      bool
			err     error
		)
		if command, ok = commandMap[args[1]]; !ok {
			sender.Say(message.Channel, "❌Unknown command")
			return nil
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
			sender.Say(message.Channel, "❌Command failed, please try again or contact an admin")
			return err
		}
		defer tx.Rollback()

		err = database.UpdateIsUserCommandEnabled(tx, args[2] == "on", message.RoomID, command.Name)
		if err != nil {
			sender.Say(message.Channel, "❌Command failed, please try again or contact an admin")
			return err
		}

		err = tx.Commit()
		if err != nil {
			sender.Say(message.Channel, "❌Command failed, please try again or contact an admin")
			return err
		}

		sender.Say(message.Channel, fmt.Sprintf("✅Set command '%s' to '%s'", command.Name, args[2]))
		return nil
	},
}
