package command

import (
	"database/sql"
	"fmt"
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
		err = database.InsertUsers(tx, true)
		if err != nil {
			sender.Say(message.Channel, "❌Command failed, please try again or contact an admin")
			return err
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

		err = database.InsertUserCommands(tx, message.Chatter.Name, commandNames...)
		if err != nil {
			return err
		}

		err = tx.Commit()
		if err != nil {
			return err
		}

		log.Info().Str("channel", message.Channel).Msg("successfully joined channel")
		sender.Say(message.Channel, fmt.Sprintf("✅ Joined channel %s", message.Channel))
		return nil
	},
}
