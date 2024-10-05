package command

import (
	"database/sql"
	"monkebot/types"
)

var optin = types.Command{
	Name:              "optin",
	Aliases:           []string{},
	Usage:             "optin [all] | optin [command]",
	Description:       "Opt in to one or all commands",
	ChannelCooldown:   5,
	UserCooldown:      5,
	NoPrefix:          false,
	NoPrefixShouldRun: nil,
	CanDisable:        false,
	Execute: func(message *types.Message, sender types.MessageSender, args []string) error {
		if len(args) != 2 {
			sender.Say(message.Channel, "🐒 Usage: optin [all] | optin [command]")
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

		err = fn(tx, message.Chatter.ID, false)
		if err != nil {
			return err
		}

		err = tx.Commit()
		if err != nil {
			return err
		}

		sender.Say(message.Channel, "✅ Opted in", struct {
			Param types.SenderParam
			Value string
		}{Param: types.ReplyMessageID, Value: message.ID})
		return nil
	},
}
