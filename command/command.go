package command

import (
	"database/sql"
	"fmt"
	"monkebot/config"
	"monkebot/database"
	"regexp"
	"strings"
	"time"

	"github.com/gempir/go-twitch-irc/v4"
	"github.com/rs/zerolog/log"
)

type MessageSender interface {
	Say(channel string, message string)
	Join(channels ...string)
	Part(channels ...string)
}

type Command struct {
	Name           string
	Aliases        []string
	Usage          string
	Description    string
	Cooldown       int
	NoPrefix       bool
	NoPrefixRegexp *regexp.Regexp
	CanDisable     bool
	Execute        func(message *Message, sender MessageSender, args []string) error
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
	RoomID  string
	Chatter Chatter
	DB      *sql.DB
}

func NewMessage(msg twitch.PrivateMessage, db *sql.DB) *Message {
	return &Message{
		Message: msg.Message,
		Time:    msg.Time,
		Channel: msg.Channel,
		RoomID:  msg.RoomID,
		Chatter: Chatter{
			Name:          msg.User.Name,
			ID:            msg.User.ID,
			IsMod:         msg.Tags["mod"] == "mod",
			IsVIP:         msg.Tags["vip"] == "vip",
			IsBroadcaster: msg.RoomID == msg.User.ID,
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
	setenabled,
}

var (
	commandMap         map[string]Command
	commandMapNoPrefix map[string]Command
)

func init() {
	commandMap = createCommandMap(Commands, true)
	commandMapNoPrefix = createCommandMap(Commands, false)
}

// Maps command names and aliases to Command structs
// If prefixedOnly is true, only commands with NoPrefix=false will be added
func createCommandMap(commands []Command, prefixedOnly bool) map[string]Command {
	cmdMap := make(map[string]Command)
	for _, cmd := range commands {
		if prefixedOnly == cmd.NoPrefix {
			continue
		}
		cmdMap[cmd.Name] = cmd
		for _, alias := range cmd.Aliases {
			cmdMap[alias] = cmd
		}
	}
	return cmdMap
}

func HandleCommands(message *Message, sender MessageSender, config *config.Config) error {
	var args []string

	hasPrefix := strings.HasPrefix(message.Message, config.Prefix)
	if hasPrefix {
		args = strings.Split(message.Message[len(config.Prefix):], " ")
	} else {
		args = strings.Split(message.Message, " ")
	}

	if cmd, ok := commandMap[args[0]]; ok {
		if cmd.CanDisable {

			tx, err := message.DB.Begin()
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
		if len(args) > 1 {
			argsStart := strings.Index(message.Message, " ")
			args = strings.Split(message.Message[argsStart:], " ")
		} else {
			args = []string{}
		}
		if err := cmd.Execute(message, sender, args); err != nil {
			return err
		}
	} else if hasPrefix {
		return fmt.Errorf("unknown command: '%s' called by '%s'", args, message.Chatter.Name)
	}

	return nil
}
