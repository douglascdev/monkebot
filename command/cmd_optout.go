package command

import (
	"database/sql"
	"monkebot/types"
)

var optout = types.Command{
	Name:              "optout",
	Aliases:           []string{},
	Usage:             "optout [all] | optout [command]",
	Description:       "Opt out of one or all commands",
	ChannelCooldown:   5,
	UserCooldown:      5,
	NoPrefix:          false,
	NoPrefixShouldRun: nil,
	CanDisable:        false,
	Execute: func(message *types.Message, sender types.MessageSender, args []string) error {
		if len(args) != 2 {
			sender.Say(message.Channel, "🐒 Usage: optout [all] | optout [command]")
			return nil
		}

		tx, err := message.DB.Begin()
		if err != nil {
			return err
		}
		defer tx.Rollback()

		var (
			fn func(tx *sql.Tx, userID string, optOut bool) error
			ok bool
		)
		if fn, ok = optoutOptions[args[1]]; !ok {
			sender.Say(message.Channel, "❌ Unknown command")
			return nil
		}

		err = fn(tx, message.Chatter.ID, true)
		if err != nil {
			return err
		}

		err = tx.Commit()
		if err != nil {
			return err
		}

		sender.Say(message.Channel, "✅ Opted out", struct {
			Param types.SenderParam
			Value string
		}{Param: types.ReplyMessageID, Value: message.ID})
		return nil
	},
}
