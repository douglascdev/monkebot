package command

import (
	"fmt"
	"monkebot/database"
	"monkebot/twitchapi"
	"monkebot/types"

	"github.com/rs/zerolog/log"
)

var setLevel = types.Command{
	Name:              "setlevel",
	Aliases:           []string{"permission", "perm", "level"},
	Usage:             "setlevel [username] [permission]",
	Description:       "Set a user's permission level",
	ChannelCooldown:   5,
	UserCooldown:      5,
	NoPrefix:          false,
	NoPrefixShouldRun: nil,
	CanDisable:        false,
	Execute: func(message *types.Message, sender types.MessageSender, args []string) error {
		if len(args) != 3 {
			sender.Say(message.Channel, "❌Usage: setlevel <username> <permission>")
			return nil
		}

		tx, err := message.DB.Begin()
		defer tx.Rollback()
		if err != nil {
			return err
		}

		var isAdmin bool
		isAdmin, err = database.SelectIsUserAdmin(tx, message.Chatter.ID)
		if err != nil {
			return err
		}
		if !isAdmin {
			sender.Say(message.Channel, "❌You must be an admin to use this command")
			return nil
		}

		var userExists bool
		userExists, err = database.SelectUserExists(tx, args[1])
		if err != nil {
			return err
		}

		if !userExists {
			var users *[]twitchapi.HelixUser
			users, err = twitchapi.GetUserByName(message.Cfg, args[1])
			if err != nil {
				sender.Say(message.Channel, fmt.Sprintf("❌User '%s' not found", args[1]))
				return err
			}
			user := (*users)[0]
			// user isn't in the db but exists on twitch, so it's a new user
			err = database.InsertUsers(tx, false, struct{ ID, Name string }{user.ID, user.Login})
			if err != nil {
				return err
			}
		}

		err = database.UpdateUserPermission(tx, args[1], args[2])
		if err != nil {
			return err
		}
		err = tx.Commit()
		if err != nil {
			return err
		}

		sender.Say(message.Channel, fmt.Sprintf("✅ Updated %s's permission to %s!", args[1], args[2]))
		log.Info().Str("channel", message.Channel).Str("user", message.Chatter.Name).Str("permission", args[2]).Msg("successfully updated user permission")

		return nil
	},
}
