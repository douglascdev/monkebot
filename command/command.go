package command

import (
	"database/sql"
	"errors"
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
	optout,
	optin,
}

var UnknownCommandErr = errors.New("unknown command")

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
	isCmdEnabled           bool
	isCmdOnChannelCoolDown bool
	isCmdOnUserCoolDown    bool
	isUserIgnored          bool
	isOptedOut             bool
}

func getCommandData(tx *sql.Tx, message *types.Message, cmd types.Command) (*commandData, error) {
	// TODO: turn selects into separate goroutines after migrating to postgres
	result := &commandData{
		isCmdEnabled:           false,
		isCmdOnChannelCoolDown: false,
		isCmdOnUserCoolDown:    false,
		isUserIgnored:          false,
		isOptedOut:             false,
	}

	var err error

	if !cmd.CanDisable {
		result.isCmdEnabled = true
	} else {
		result.isCmdEnabled, err = database.SelectIsUserCommandEnabled(tx, message.RoomID, cmd.Name)
		if err != nil {
			return nil, fmt.Errorf("failed to select is_user_command_enabled: %w", err)
		}
	}

	result.isUserIgnored, err = database.SelectIsUserIgnored(tx, message.Chatter.ID)
	if err == sql.ErrNoRows {
		result.isUserIgnored = false
	} else if err != nil {
		return nil, fmt.Errorf("failed to select user's is_ignored: %w", err)
	}

	result.isCmdOnChannelCoolDown, err = database.SelectIsCommandOnChannelCooldown(tx, message.RoomID, cmd.Name, cmd.ChannelCooldown)
	if err != nil {
		return nil, fmt.Errorf("failed to select command cooldown: %w", err)
	}

	result.isCmdOnUserCoolDown, err = database.SelectIsCommandOnUserCooldown(tx, message.Chatter.ID, cmd.Name, cmd.UserCooldown)
	if err != nil {
		return nil, fmt.Errorf("failed to select user cooldown: %w", err)
	}

	result.isOptedOut, err = database.SelectIsCommandOptedOut(tx, message.Chatter.ID, cmd.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to select user's opted_out: %w", err)
	}

	return result, nil
}

func HandleCommands(message *types.Message, sender types.MessageSender, config *config.Config) error {
	var (
		cmdData *commandData
		args    []string
		tx      *sql.Tx
		err     error
	)

	tx, err = message.DB.Begin()
	if err != nil {
		log.Err(err).Msg("failed to start HandleCommands transaction")
		return err
	}
	defer tx.Rollback()

	hasPrefix := strings.HasPrefix(message.Message, config.Prefix)
	if hasPrefix {
		args = strings.Split(message.Message[len(config.Prefix):], " ")
	} else {
		args = strings.Split(message.Message, " ")

		// check if command is no prefix
		for _, noPrefixCmd := range commandsNoPrefix {
			if noPrefixCmd.NoPrefixShouldRun != nil && noPrefixCmd.NoPrefixShouldRun(message, sender, args) {
				err = database.InsertUsers(tx, false, struct{ ID, Name string }{message.Chatter.ID, message.Chatter.Name})
				if err != nil {
					return err
				}

				cmdData, err = getCommandData(tx, message, noPrefixCmd)
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

				if cmdData.isCmdOnChannelCoolDown {
					log.Debug().Str("command", noPrefixCmd.Name).Str("channel", message.Channel).Msg("command ignored due to channel cooldown")
					return nil
				}

				if cmdData.isCmdOnUserCoolDown {
					log.Debug().Str("command", noPrefixCmd.Name).Str("channel", message.Channel).Msg("command ignored due to user command cooldown")
					return nil
				}

				if cmdData.isOptedOut {
					log.Debug().Str("command", noPrefixCmd.Name).Str("channel", message.Channel).Msg("command ignored due to opt out")
					return nil
				}

				err = database.UpdateUserCommandLastUsed(tx, message.RoomID, noPrefixCmd.Name, message.Chatter.ID)
				if err != nil {
					return fmt.Errorf("failed to update last_used for command %s: %w", noPrefixCmd.Name, err)
				}

				err = tx.Commit()
				if err != nil {
					return fmt.Errorf("failed to commit transaction to update last_used for command %s: %w", noPrefixCmd.Name, err)
				}

				err = noPrefixCmd.Execute(message, sender, args)
				if err != nil {
					return err
				}

				break
			}
		}

		return nil
	}

	if cmd, ok := commandMap[args[0]]; ok {
		err = database.InsertUsers(tx, false, struct{ ID, Name string }{message.Chatter.ID, message.Chatter.Name})
		if err != nil {
			return err
		}

		cmdData, err = getCommandData(tx, message, cmd)
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

		if cmdData.isCmdOnChannelCoolDown {
			log.Debug().Str("command", cmd.Name).Str("channel", message.Channel).Msg("command ignored due to channel cooldown")
			return nil
		}

		if cmdData.isCmdOnUserCoolDown {
			log.Debug().Str("command", cmd.Name).Str("channel", message.Channel).Msg("command ignored due to user command cooldown")
			return nil
		}

		if cmdData.isOptedOut {
			log.Debug().Str("command", cmd.Name).Str("channel", message.Channel).Msg("command ignored due to opt out")
			return nil
		}

		err = database.UpdateUserCommandLastUsed(tx, message.RoomID, cmd.Name, message.Chatter.ID)
		if err != nil {
			return fmt.Errorf("failed to update last_used for command %s: %w", cmd.Name, err)
		}

		err = tx.Commit()
		if err != nil {
			return fmt.Errorf("failed to commit transaction to update last_used for command %s: %w", cmd.Name, err)
		}

		if err = cmd.Execute(message, sender, args); err != nil {
			return err
		}

	} else if hasPrefix {
		return fmt.Errorf("%w: '%s' called by '%s'", UnknownCommandErr, args, message.Chatter.Name)
	}

	return nil
}
