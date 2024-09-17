package command

import (
	"fmt"
	"monkebot/database"
	"strings"

	"github.com/rs/zerolog/log"
)

var part = Command{
	Name:              "part",
	Aliases:           []string{"leave"},
	Usage:             "part <channel>",
	Description:       "Leave the message author's channel",
	Cooldown:          5,
	NoPrefix:          false,
	NoPrefixShouldRun: nil,
	CanDisable:        false,
	Execute: func(message *Message, sender MessageSender, args []string) error {
		tx, err := message.DB.Begin()
		if err != nil {
			sender.Say(message.Channel, "❌Command failed, please try again or contact an admin")
			return err
		}

		if len(args) > 0 {
			var isAdmin bool
			isAdmin, err = database.SelectIsUserAdmin(tx, message.Chatter.ID)
			if err != nil {
				sender.Say(message.Channel, "❌Command failed, please try again or contact an admin")
				return err
			}
			if !isAdmin {
				sender.Say(message.Channel, "❌You must be an admin to use this command")
				return nil
			}
		}

		err = database.UpdateIsBotJoined(tx, false, args[1:]...)
		if err != nil {
			sender.Say(message.Channel, "❌Command failed, please try again or contact an admin")
			return err
		}

		err = tx.Commit()
		if err != nil {
			sender.Say(message.Channel, "❌Command failed, please try again or contact an admin")
			return err
		}

		sender.Part(args[1:]...)
		log.Info().Str("channel", message.Channel).Msg("successfully left channels")
		sender.Say(message.Channel, fmt.Sprintf("✅ Parted channel(s) %s", strings.Join(args[1:], ", ")))
		return nil
	},
}
