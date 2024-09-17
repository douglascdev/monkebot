package command

import (
	"fmt"
	"monkebot/database"
)

var buttsbot = Command{
	Name:        "buttsbot",
	Usage:       "buttsbot <on | off>",
	Description: "Randomly replaces message syllables with butt",
	Cooldown:    5,
	NoPrefix:    false,
	Execute: func(message *Message, sender MessageSender, args []string) error {
		if len(args) != 2 || (args[1] != "on" && args[1] != "off") {
			sender.Say(message.Chatter.Name, "Usage: buttsbot <on | off>")
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

		err = database.UpdateIsUserCommandEnabled(tx, args[1] == "on", message.RoomID, "buttsbot")
		if err != nil {
			sender.Say(message.Channel, "❌Command failed, please try again or contact an admin")
			return err
		}

		err = tx.Commit()
		if err != nil {
			sender.Say(message.Channel, "❌Command failed, please try again or contact an admin")
			return err
		}

		sender.Say(message.Channel, fmt.Sprintf("✅Set command buttsbot to '%s'", args[1]))
		return nil
	},
}
