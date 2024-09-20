package command

import (
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

// return if the command is enabled and if the user is ignored
func getCommandData(message *types.Message, cmd types.Command) (bool, bool, error) {
	tx, err := message.DB.Begin()
	if err != nil {
		return false, false, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	var isEnabled bool
	if !cmd.CanDisable {
		isEnabled = true
	} else {
		isEnabled, err = database.SelectIsUserCommandEnabled(tx, message.RoomID, cmd.Name)
		if err != nil {
			return false, false, err
		}
	}

	var isIgnored bool
	isIgnored, err = database.SelectIsUserIgnored(tx, message.Chatter.ID)
	if err != nil {
		return false, false, err
	}

	return isEnabled, isIgnored, tx.Commit()
}

func HandleCommands(message *types.Message, sender types.MessageSender, config *config.Config) error {
	var (
		args []string
		err  error
	)

	hasPrefix := strings.HasPrefix(message.Message, config.Prefix)
	if hasPrefix {
		args = strings.Split(message.Message[len(config.Prefix):], " ")
	} else {
		args = strings.Split(message.Message, " ")

		// check if command is no prefix
		for _, noPrefixCmd := range commandsNoPrefix {
			if noPrefixCmd.NoPrefixShouldRun != nil && noPrefixCmd.NoPrefixShouldRun(message, sender, args) {
				var (
					isEnabled     bool
					isUserIgnored bool
				)
				isEnabled, isUserIgnored, err = getCommandData(message, noPrefixCmd)
				if err != nil {
					return err
				}
				if !isEnabled {
					log.Debug().Str("command", noPrefixCmd.Name).Str("channel", message.Channel).Msg("ignored disabled no-prefix command")
					return nil
				}

				if isUserIgnored {
					log.Debug().Str("user", message.Chatter.Name).Str("channel", message.Channel).Msg("ignored user")
					return nil
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
		var (
			isEnabled     bool
			isUserIgnored bool
		)
		isEnabled, isUserIgnored, err = getCommandData(message, cmd)
		if err != nil {
			return err
		}
		if !isEnabled {
			log.Debug().Str("command", cmd.Name).Str("channel", message.Channel).Msg("ignored disabled command")
			return nil
		}

		if isUserIgnored {
			log.Debug().Str("user", message.Chatter.Name).Str("channel", message.Channel).Msg("ignored user")
			return nil
		}

		if err := cmd.Execute(message, sender, args); err != nil {
			return err
		}
	} else if hasPrefix {
		return fmt.Errorf("unknown command: '%s' called by '%s'", args, message.Chatter.Name)
	}

	return nil
}
