package command

import (
	"database/sql"
	"fmt"
	"monkebot/config"
	"monkebot/database"
	"strings"
	"time"

	"github.com/gempir/go-twitch-irc/v4"
	"github.com/rs/zerolog/log"
)

type MessageSender interface {
	Say(channel string, message string)

	Join(channels ...string)
	Part(channels ...string)

	Buttify(message string) string
	ShouldButtify() bool
}

type Command struct {
	Name        string
	Aliases     []string
	Usage       string
	Description string
	Cooldown    int
	NoPrefix    bool
	CanDisable  bool

	// `json:"-"` excludes these 2 fields from being serialized into the command list json
	NoPrefixShouldRun func(message *Message, sender MessageSender, args []string) bool  `json:"-"`
	Execute           func(message *Message, sender MessageSender, args []string) error `json:"-"`
}

type Chatter struct {
	Name string
	ID   string

	IsMod         bool
	IsVIP         bool
	IsBroadcaster bool
}

// Message normalized to be platform agnostic
type Message struct {
	Message string
	Time    time.Time
	Channel string
	Cfg     *config.Config
	RoomID  string
	Chatter Chatter
	DB      *sql.DB
}

func NewMessage(msg twitch.PrivateMessage, db *sql.DB, cfg *config.Config) *Message {
	IsBroadcaster := msg.RoomID == msg.User.ID
	_, IsMod := msg.Tags["mod"]
	_, IsVIP := msg.Tags["vip"]
	return &Message{
		Message: msg.Message,
		Time:    msg.Time,
		Channel: msg.Channel,
		RoomID:  msg.RoomID,
		Cfg:     cfg,
		Chatter: Chatter{
			Name:          msg.User.Name,
			ID:            msg.User.ID,
			IsMod:         IsMod,
			IsVIP:         IsVIP,
			IsBroadcaster: IsBroadcaster,
		},
		DB: db,
	}
}

var Commands = []Command{
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
	commandMap       map[string]Command
	commandsNoPrefix []Command
)

func init() {
	commandMap = createCommandMap(Commands)

	for _, cmd := range Commands {
		if cmd.NoPrefix {
			commandsNoPrefix = append(commandsNoPrefix, cmd)
		}
	}
}

// Maps command names and aliases to Command structs
// If prefixedOnly is true, only commands with NoPrefix=false will be added
func createCommandMap(commands []Command) map[string]Command {
	cmdMap := make(map[string]Command)
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

func isCommandEnabled(message *Message, cmd Command) (bool, error) {
	if !cmd.CanDisable {
		return true, nil
	}

	tx, err := message.DB.Begin()
	if err != nil {
		return false, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	var enabled bool
	enabled, err = database.SelectIsUserCommandEnabled(tx, message.RoomID, cmd.Name)
	if err != nil {
		return false, err
	}

	return enabled, nil
}

func HandleCommands(message *Message, sender MessageSender, config *config.Config) error {
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
				var isEnabled bool
				isEnabled, err = isCommandEnabled(message, noPrefixCmd)
				if err != nil {
					return err
				}
				if !isEnabled {
					log.Debug().Str("command", noPrefixCmd.Name).Str("channel", message.Channel).Msg("ignored disabled no-prefix command")
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
		if cmd.CanDisable {

			var tx *sql.Tx
			tx, err = message.DB.Begin()
			if err != nil {
				return fmt.Errorf("failed to begin transaction: %w", err)
			}
			defer tx.Rollback()

			var enabled bool
			enabled, err = database.SelectIsUserCommandEnabled(tx, message.RoomID, cmd.Name)
			if err != nil {
				return err
			}

			if !enabled {
				log.Debug().Str("command", cmd.Name).Str("channel", message.Channel).Msg("ignored disabled command")
				return nil
			}
		}

		if err := cmd.Execute(message, sender, args); err != nil {
			return err
		}
	} else if hasPrefix {
		return fmt.Errorf("unknown command: '%s' called by '%s'", args, message.Chatter.Name)
	}

	return nil
}
