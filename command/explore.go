package command

import (
	"database/sql"
	"fmt"
	"math/rand/v2"
	"monkebot/database"
	"monkebot/types"

	"github.com/rs/zerolog/log"
)

var explore = types.Command{
	Name:              "explore",
	Aliases:           []string{"e"},
	Usage:             "explore",
	Description:       "Venture into mysterious islands, uncovering hidden treasures or encountering perilous dangers. Each exploration can result in gaining or losing items.",
	Cooldown:          30,
	NoPrefix:          false,
	NoPrefixShouldRun: nil,
	CanDisable:        true,
	Execute: func(message *types.Message, sender types.MessageSender, args []string) error {
		tx, err := message.DB.Begin()
		if err != nil {
			return err
		}
		defer tx.Rollback()

		// insert user if it doesn't exist
		user := struct{ ID, Name string }{ID: message.Chatter.ID, Name: message.Chatter.Name}
		err = database.InsertUsers(tx, false, user)
		if err != nil {
			return err
		}

		// randomly select an outcome and reward
		outcome := message.Cfg.RPGConfig.ExplorationResults[rand.IntN(len(message.Cfg.RPGConfig.ExplorationResults))]
		var (
			outcomeMultiplier int
			ok                bool
		)
		outcomeMultiplierMap := map[string]int{
			"VeryPositive": 2,
			"Positive":     1,
			"Negative":     -1,
			"VeryNegative": -2,
		}
		if outcomeMultiplier, ok = outcomeMultiplierMap[outcome.ResultType]; !ok {
			log.Warn().Msgf("unknown outcome type: %s", outcome.ResultType)
			outcomeMultiplier = 1
		}
		reward := outcomeMultiplier * max(1, rand.IntN(51))

		// get buttinho(coin) item id and name
		var (
			itemID   int
			itemName string
		)
		err = tx.QueryRow("SELECT id, name FROM rpg_item WHERE name = 'buttinho'").Scan(&itemID, &itemName)
		if err != nil {
			return err
		}

		// check if rpg_user_item exists
		var exists bool
		err = tx.QueryRow("SELECT 1 FROM rpg_user_item WHERE user_id = ? AND rpg_item_id = ?", user.ID, itemID).Scan(&exists)
		if err != nil && err != sql.ErrNoRows {
			return err
		}
		if exists {
			_, err = tx.Exec("UPDATE rpg_user_item SET amount = amount + ? WHERE user_id = ? AND rpg_item_id = ?", reward, user.ID, itemID)
			if err != nil {
				return err
			}
		} else {
			_, err = tx.Exec("INSERT INTO rpg_user_item (user_id, rpg_item_id, amount) VALUES (?, ?, ?)", user.ID, itemID, reward)
			if err != nil {
				return err
			}
		}

		// select the updated amount
		var amount int
		err = tx.QueryRow("SELECT amount FROM rpg_user_item WHERE user_id = ? AND rpg_item_id = ?", user.ID, itemID).Scan(&amount)
		if err != nil && err != sql.ErrNoRows {
			return err
		}

		err = tx.Commit()
		if err != nil {
			return err
		}

		var msg string
		if reward >= 0 {
			msg = fmt.Sprintf("%s [ +%d => %d %s ]", outcome.Message, reward, amount, itemName)
		} else {
			msg = fmt.Sprintf("%s [ %d => %d %s ]", outcome.Message, reward, amount, itemName)
		}

		sender.Say(message.Channel, msg, []struct {
			Param types.SenderParam
			Value string
		}{
			{types.ReplyMessageID, message.ID},
			{types.Me, "true"},
		}...)

		return nil
	},
}
