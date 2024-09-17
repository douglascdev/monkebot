package command

import (
	"database/sql"
	"fmt"
	"monkebot/config"
	"strings"
	"time"

	"github.com/gempir/go-twitch-irc/v4"
)

type MessageSender interface {
	Say(channel string, message string)
	Join(channels ...string)
	Part(channels ...string)
}

type Command struct {
	Name        string
	Aliases     []string
	Usage       string
	Description string
	Cooldown    int
	NoPrefix    bool
	Execute     func(message *Message, sender MessageSender, args []string) error
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
	buttsbot,
}

var commandMap = createCommandMap(Commands)

// Maps command names and aliases to Command structs
func createCommandMap(commands []Command) map[string]Command {
	cmdMap := make(map[string]Command)
	for _, cmd := range commands {
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
