package command

import (
	"fmt"
	"monkebot/database"
	"monkebot/types"

	"github.com/rs/zerolog/log"
)

var setLevel = types.Command{
	Name:              "setlevel",
	Aliases:           []string{"permission", "perm", "level"},
	Usage:             "setlevel [username] [permission]",
	Description:       "Set a user's permission level",
	Cooldown:          5,
	NoPrefix:          false,
	NoPrefixShouldRun: nil,
	CanDisable:        false,
	Execute: func(message *types.Message, sender types.MessageSender, args []string) error {
		tx, err := message.DB.Begin()
		defer tx.Rollback()
		if err != nil {
			sender.Say(message.Channel, "❌Command failed, please try again or contact an admin")
			return err
		}

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

		if len(args) != 3 {
			sender.Say(message.Channel, "❌Usage: setlevel <username> <permission>")
			return nil
		}

		err = database.UpdateUserPermission(tx, args[1], args[2])
		if err != nil {
			sender.Say(message.Channel, "❌Command failed, please try again or contact an admin")
			return err
		}
		err = tx.Commit()
		if err != nil {
			sender.Say(message.Channel, "❌Command failed, please try again or contact an admin")
			return err
		}

		sender.Say(message.Channel, fmt.Sprintf("✅ Updated %s's permission to %s!", args[1], args[2]))
		log.Info().Str("channel", message.Channel).Str("user", message.Chatter.Name).Str("permission", args[2]).Msg("successfully updated user permission")

		return nil
	},
}
