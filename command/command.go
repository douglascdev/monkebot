package command

import (
	"database/sql"
	"fmt"
	"monkebot/config"
	"monkebot/database"
	"monkebot/types"
	"strings"

	"github.com/rs/zerolog/log"
)

var Commands = []types.Command{
	ping,
	senzpTest,
	join,
	part,
	setLevel,
	buttsbot,
	butt,
	help,
	explore,
	enable,
	disable,
}

var (
	commandMap       map[string]types.Command
	commandsNoPrefix []types.Command
)

func init() {
	commandMap = createCommandMap(Commands)

	for _, cmd := range Commands {
		if cmd.NoPrefix {
			commandsNoPrefix = append(commandsNoPrefix, cmd)
		}
	}
}

// Maps command names and aliases to types.Command structs
// If prefixedOnly is true, only commands with NoPrefix=false will be added
func createCommandMap(commands []types.Command) map[string]types.Command {
	cmdMap := make(map[string]types.Command)
	for _, cmd := range commands {
		if cmd.NoPrefix {
			continue
		}
		cmdMap[cmd.Name] = cmd
		for _, alias := range cmd.Aliases {
			cmdMap[alias] = cmd
		}
	}
	return cmdMap
}

type commandData struct {
	isCmdEnabled    bool
	isCmdOnCoolDown bool
	isUserIgnored   bool
}

func getCommandData(message *types.Message, cmd types.Command) (*commandData, error) {
	result := &commandData{
		isCmdEnabled:    false,
		isCmdOnCoolDown: false,
		isUserIgnored:   false,
	}

	tx, err := message.DB.Begin()
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	if !cmd.CanDisable {
		result.isCmdEnabled = true
	} else {
		result.isCmdEnabled, err = database.SelectIsUserCommandEnabled(tx, message.RoomID, cmd.Name)
		if err != nil {
			return nil, err
		}
	}

	result.isUserIgnored, err = database.SelectIsUserIgnored(tx, message.Chatter.ID)
	if err != nil {
		return nil, err
	}

	result.isCmdOnCoolDown, err = database.SelectIsCommandOnCooldown(tx, message.RoomID, cmd.Name, cmd.Cooldown)
	if err != nil {
		return nil, err
	}

	return result, tx.Commit()
}

func HandleCommands(message *types.Message, sender types.MessageSender, config *config.Config) error {
	var (
		cmdData *commandData
		args    []string
		err     error
	)

	hasPrefix := strings.HasPrefix(message.Message, config.Prefix)
	if hasPrefix {
		args = strings.Split(message.Message[len(config.Prefix):], " ")
	} else {
		args = strings.Split(message.Message, " ")

		// check if command is no prefix
		for _, noPrefixCmd := range commandsNoPrefix {
			if noPrefixCmd.NoPrefixShouldRun != nil && noPrefixCmd.NoPrefixShouldRun(message, sender, args) {
				cmdData, err = getCommandData(message, noPrefixCmd)
				if err != nil {
					return err
				}
				if !cmdData.isCmdEnabled {
					log.Debug().Str("command", noPrefixCmd.Name).Str("channel", message.Channel).Msg("ignored disabled no-prefix command")
					return nil
				}

				if cmdData.isUserIgnored {
					log.Debug().Str("user", message.Chatter.Name).Str("channel", message.Channel).Msg("ignored user")
					return nil
				}

				if cmdData.isCmdOnCoolDown {
					log.Debug().Str("command", noPrefixCmd.Name).Str("channel", message.Channel).Msg("command ignored due to channel cooldown")
					return nil
				}

				err = noPrefixCmd.Execute(message, sender, args)
				if err != nil {
					return err
				}

				var tx *sql.Tx
				tx, err = message.DB.Begin()
				if err != nil {
					return fmt.Errorf("failed to start last_used update transaction: %w", err)
				}
				defer tx.Rollback()

				err = database.UpdateUserCommandLastUsed(tx, message.RoomID, noPrefixCmd.Name)
				if err != nil {
					return fmt.Errorf("failed to update last_used for command %s: %w", noPrefixCmd.Name, err)
				}

				err = tx.Commit()
				if err != nil {
					return fmt.Errorf("failed to commit transaction to update last_used for command %s: %w", noPrefixCmd.Name, err)
				}

				break
			}
		}

		return nil
	}

	if cmd, ok := commandMap[args[0]]; ok {
		cmdData, err = getCommandData(message, cmd)
		if err != nil {
			return err
		}
		if !cmdData.isCmdEnabled {
			log.Debug().Str("command", cmd.Name).Str("channel", message.Channel).Msg("ignored disabled command")
			return nil
		}

		if cmdData.isUserIgnored {
			log.Debug().Str("user", message.Chatter.Name).Str("channel", message.Channel).Msg("ignored user")
			return nil
		}

		if cmdData.isCmdOnCoolDown {
			log.Debug().Str("command", cmd.Name).Str("channel", message.Channel).Msg("command ignored due to channel cooldown")
			return nil
		}

		if err = cmd.Execute(message, sender, args); err != nil {
			return err
		}

		var tx *sql.Tx
		tx, err = message.DB.Begin()
		if err != nil {
			return fmt.Errorf("failed to start last_used update transaction: %w", err)
		}
		defer tx.Rollback()

		err = database.UpdateUserCommandLastUsed(tx, message.RoomID, cmd.Name)
		if err != nil {
			return fmt.Errorf("failed to update last_used for command %s: %w", cmd.Name, err)
		}

		err = tx.Commit()
		if err != nil {
			return fmt.Errorf("failed to commit transaction to update last_used for command %s: %w", cmd.Name, err)
		}
	} else if hasPrefix {
		return fmt.Errorf("unknown command: '%s' called by '%s'", args, message.Chatter.Name)
	}

	return nil
}
