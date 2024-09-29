package command

import (
	"database/sql"
	"fmt"

	"github.com/rs/zerolog/log"
)

var optoutOptions = make(map[string]func(tx *sql.Tx, userID string, optOut bool) error)

func init() {
	optoutOptions["all"] = func(tx *sql.Tx, userID string, optOut bool) error {
		_, err := tx.Exec("UPDATE user_command_data SET opted_out = ? WHERE user_id = ?", optOut, userID)
		if err != nil {
			return err
		}
		log.Debug().Str("user_id", userID).Bool("opted_out", optOut).Msg("updated opt out for all commands")
		return nil
	}

	for _, command := range Commands {
		optoutOptions[command.Name] = func(tx *sql.Tx, userID string, optOut bool) error {
			result, err := tx.Exec(`
				UPDATE user_command_data SET opted_out = ?
				WHERE user_id = ? AND command_id = (
					SELECT id FROM command WHERE name = ?
				)`, optOut, userID, command.Name)
			if err != nil {
				return err
			}

			rowsAffected, err := result.RowsAffected()
			if err != nil {
				return err
			}

			if rowsAffected != 1 {
				return fmt.Errorf("invalid number of rows affected: %d", rowsAffected)
			}

			log.Debug().Str("user_id", userID).Str("command", command.Name).Bool("opted_out", optOut).Msg("updated opt out")
			return nil
		}
	}
}
