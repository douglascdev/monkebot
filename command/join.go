package command

import (
	"database/sql"
	"fmt"
	"monkebot/database"
	"monkebot/twitchapi"
	"monkebot/types"
	"strings"

	"github.com/rs/zerolog/log"
)

var join = types.Command{
	Name:              "join",
	Aliases:           []string{},
	Usage:             "join | join [channel]",
	Description:       "Join the message author's channel or the specified channel",
	Cooldown:          5,
	NoPrefix:          false,
	NoPrefixShouldRun: nil,
	CanDisable:        false,
	Execute: func(message *types.Message, sender types.MessageSender, args []string) error {
		tx, err := message.DB.Begin()
		if err != nil {
			sender.Say(message.Channel, "❌Command failed, please try again or contact an admin")
			return err
		}
		defer tx.Rollback()

		var channelsToJoin []struct {
			ID   string
			Name string
		}

		if len(args) == 2 && message.Chatter.Name == args[1] {
			channelsToJoin = append(channelsToJoin, struct {
				ID   string
				Name string
			}{ID: message.Chatter.ID, Name: message.Chatter.Name})
		} else if len(args) > 1 {
			isAdmin := false
			isAdmin, err = database.SelectIsUserAdmin(tx, message.Chatter.ID)

			if err != nil && err != sql.ErrNoRows {
				sender.Say(message.Channel, "❌Command failed, please try again or contact an admin")
				return err
			}

			if err == sql.ErrNoRows || !isAdmin {
				sender.Say(message.Channel, "❌You must be an admin to use this command")
				return nil
			}

			var twitchUsers *[]twitchapi.HelixUser
			twitchUsers, err = twitchapi.GetUserByName(message.Cfg, args[1:]...)
			if err != nil {
				sender.Say(message.Channel, "❌Command failed, please try again or contact an admin")
				return err
			}
			channelsToJoin = make([]struct {
				ID   string
				Name string
			}, 0, len(*twitchUsers))

			for _, user := range *twitchUsers {
				channelsToJoin = append(channelsToJoin, struct {
					ID   string
					Name string
				}{ID: user.ID, Name: user.Login})
			}
		} else {
			channelsToJoin = append(channelsToJoin, struct {
				ID   string
				Name string
			}{ID: message.Chatter.ID, Name: message.Chatter.Name})
		}

		if len(channelsToJoin) == 0 {
			sender.Say(message.Channel, "❌Channel(s) not found")
			return nil
		}

		// check if any of the channels are already in the database
		var (
			query    string
			rows     *sql.Rows
			channels []interface{}
		)
		query = fmt.Sprintf("SELECT name FROM user WHERE name IN (%s) AND bot_is_joined", strings.Repeat("?,", max(0, len(channelsToJoin)-1))+"?")
		channels = make([]interface{}, len(channelsToJoin))
		for i, channel := range channelsToJoin {
			channels[i] = channel.Name
		}

		rows, err = tx.Query(query, channels...)
		if err == nil {
			defer rows.Close()
			var foundChannels []string
			for rows.Next() {
				var name string
				err = rows.Scan(&name)
				if err != nil {
					return err
				}
				foundChannels = append(foundChannels, name)
			}
			err = rows.Err()
			if err != nil {
				return err
			}

			if len(foundChannels) > 0 {
				answer := fmt.Sprintf("❌The following channels were already joined: %s", strings.Join(foundChannels, ", "))
				sender.Say(message.Channel, answer)
				return nil
			}
		}

		err = database.InsertUsers(tx, true, channelsToJoin...)
		if err != nil {
			sender.Say(message.Channel, "❌Command failed, please try again or contact an admin")
			return err
		}

		// ensure all joined channels have bot_is_joined set to true if InsertUsers didn't just insert them(it skips existing users)
		var channelIDs []string
		for _, channel := range channelsToJoin {
			channelIDs = append(channelIDs, channel.ID)
		}
		err = database.UpdateIsBotJoined(tx, true, channelIDs...)
		if err != nil {
			return err
		}

		var commandNames []string
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
				log.Warn().Err(err).Str("channel", channel.Name).Msg("failed to insert user commands after join, skipping channel")
			}
		}

		err = tx.Commit()
		if err != nil {
			return err
		}

		channelNames := make([]string, len(channelsToJoin))
		for i, channel := range channelsToJoin {
			channelNames[i] = channel.Name
		}
		log.Info().Strs("channels", channelNames).Msg("successfully joined channels")
		sender.Join(channelNames...)
		sender.Say(message.Channel, fmt.Sprintf("✅ Joined channel(s) %s", strings.Join(channelNames, ", ")))
		for _, channel := range channelsToJoin {
			sender.Say(channel.Name, "ola")
		}
		return nil
	},
}
