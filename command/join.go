package command

import (
	"database/sql"
	"monkebot/database"

	"github.com/rs/zerolog/log"
)

var join = Command{
	Name:        "join",
	Aliases:     []string{"j"},
	Usage:       "join",
	Description: "Join the message author's channel",
	Cooldown:    5,
	NoPrefix:    false,
	Execute: func(message *Message, sender MessageSender, args []string) error {
		tx, err := message.DB.Begin()
		if err != nil {
			sender.Say(message.Channel, "❌Command failed, please try again or contact an admin")
			return err
		}

		var channelsToJoin []struct {
			ID   string
			Name string
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

			for _, arg := range args {
				channelsToJoin = append(channelsToJoin, struct {
					ID   string
					Name string
				}{ID: arg, Name: arg})
			}
		} else {
			channelsToJoin = append(channelsToJoin, struct {
				ID   string
				Name string
			}{ID: message.Channel, Name: message.Channel})
		}

		err = database.InsertUsers(tx, true, channelsToJoin...)
		if err != nil {
			sender.Say(message.Channel, "❌Command failed, please try again or contact an admin")
		}

		var (
			commandNames []string
			rows         *sql.Rows
		)
		rows, err = tx.Query("SELECT name FROM command")
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			var name string
			err = rows.Scan(&name)
			if err != nil {
				return err
			}
			commandNames = append(commandNames, name)
		}

		for _, channel := range channelsToJoin {
			err = database.InsertUserCommands(tx, channel.ID, commandNames...)
			if err != nil {
				return err
			}
		}

		err = tx.Commit()
		if err != nil {
			return err
		}

		log.Info().Str("channel", message.Channel).Msg("successfully joined channel")
		sender.Say(message.Channel, "✅ Joined channel(s)")
		return nil
	},
}
