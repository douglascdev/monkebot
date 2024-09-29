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

var part = types.Command{
	Name:              "part",
	Aliases:           []string{"leave"},
	Usage:             "part | part [channel]",
	Description:       "Leave the message author's channel or the specified channel",
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

		var channelsToLeave []struct {
			ID   string
			Name string
		}

		if len(args) == 2 && message.Chatter.Name == args[1] {
			channelsToLeave = append(channelsToLeave, struct {
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
			channelsToLeave = make([]struct {
				ID   string
				Name string
			}, 0, len(*twitchUsers))

			for _, user := range *twitchUsers {
				channelsToLeave = append(channelsToLeave, struct {
					ID   string
					Name string
				}{ID: user.ID, Name: user.Login})
			}
		} else {
			channelsToLeave = append(channelsToLeave, struct {
				ID   string
				Name string
			}{ID: message.Chatter.ID, Name: message.Chatter.Name})
		}

		if len(channelsToLeave) == 0 {
			sender.Say(message.Channel, "❌Channel(s) not found")
			return nil
		}

		// check if any of the channels are already in the database
		var (
			query    string
			rows     *sql.Rows
			channels []interface{}
		)
		query = fmt.Sprintf("SELECT name FROM user WHERE name IN (%s) AND bot_is_joined", strings.Repeat("?,", max(0, len(channelsToLeave)-1))+"?")
		channels = make([]interface{}, len(channelsToLeave))
		for i, channel := range channelsToLeave {
			channels[i] = channel.Name
		}

		rows, err = tx.Query(query, channels...)
		if err != nil {
			return err
		}

		defer rows.Close()
		foundChannels := map[string]struct{}{}
		for rows.Next() {
			var name string
			err = rows.Scan(&name)
			if err != nil {
				return err
			}
			foundChannels[name] = struct{}{}
		}
		err = rows.Err()
		if err != nil {
			return err
		}

		if len(foundChannels) != len(channelsToLeave) {
			channelsNotFound := make([]string, 0, len(channelsToLeave)-len(foundChannels))
			for _, channel := range channelsToLeave {
				if _, ok := foundChannels[channel.Name]; !ok {
					channelsNotFound = append(channelsNotFound, channel.Name)
				}
			}
			answer := fmt.Sprintf("❌The following channels were not joined: %s", strings.Join(channelsNotFound, ", "))
			sender.Say(message.Channel, answer)
			return nil
		}

		// ensure all joined channels have bot_is_joined set to false if InsertUsers didn't just insert them(it skips existing users)
		var channelIDs []string
		for _, channel := range channelsToLeave {
			channelIDs = append(channelIDs, channel.ID)
		}
		err = database.UpdateIsBotJoined(tx, false, channelIDs...)
		if err != nil {
			return err
		}

		err = tx.Commit()
		if err != nil {
			return err
		}

		channelNames := make([]string, len(channelsToLeave))
		for i, channel := range channelsToLeave {
			channelNames[i] = channel.Name
		}
		log.Info().Strs("channels", channelNames).Msg("successfully parted channels")
		sender.Part(channelNames...)
		sender.Say(message.Channel, fmt.Sprintf("✅Successfully parted %s", strings.Join(channelNames, ", ")))
		return nil
	},
}
