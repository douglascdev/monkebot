package command

import (
	"database/sql"
	"fmt"
	"monkebot/database"
	"strings"

	"github.com/rs/zerolog/log"
)

var join = Command{
	Name:              "join",
	Aliases:           []string{},
	Usage:             "join | join <channel>",
	Description:       "Join the message author's channel",
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

		var channelsToJoin []struct {
			ID   string
			Name string
		}

		if len(args) > 1 {
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

			for _, arg := range args[1:] {
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

		log.Info().Msg("successfully joined channels")
		channelNames := make([]string, len(channelsToJoin))
		for i, channel := range channelsToJoin {
			channelNames[i] = channel.Name
		}
		sender.Join(channelNames...)
		sender.Say(message.Channel, fmt.Sprintf("✅ Joined channel(s) %s", strings.Join(channelNames, ", ")))
		for _, channel := range channelsToJoin {
			sender.Say(channel.Name, "ola")
		}
		return nil
	},
}
