package monkebot

import (
	"fmt"
	"strings"
	"time"

	"github.com/gempir/go-twitch-irc/v4"
)

var commands = []Command{
	{
		Name:        "ping",
		Aliases:     []string{"pong"},
		Usage:       "ping",
		Description: "Ping!",
		Cooldown:    5,
		Execute: func(message *Message, mb *Monkebot) error {
			latency := fmt.Sprintf("%d ms", time.Now().Sub(message.Time).Milliseconds())
			mb.Say(message.Channel, fmt.Sprintf("🐒 Pong! Latency: %s", latency))
			return nil
		},
	},
}

var commandMap = createCommandMap(commands)

// Maps command names and aliases to Command structs
func createCommandMap(commands []Command) map[string]Command {
	commandMap := make(map[string]Command)
	for _, cmd := range commands {
		commandMap[cmd.Name] = cmd
		for _, alias := range cmd.Aliases {
			commandMap[alias] = cmd
		}
	}
	return commandMap
}

func HandleCommands(message *Message, mb *Monkebot, config *Config) error {
	if !strings.HasPrefix(message.Message, config.Prefix) {
		return nil
	}

	args := strings.Split(message.Message[len(config.Prefix):], " ")

	if cmd, ok := commandMap[args[0]]; ok {
		if err := cmd.Execute(message, mb); err != nil {
			return err
		}
	} else {
		return fmt.Errorf("unknown command: '%s' called by '%s'", args, message.Chatter.Name)
	}

	return nil
}

type Chatter struct {
	Name string
	ID   string
}

// Message normalized to be platform agnostic
type Message struct {
	Message string
	Time    time.Time
	Channel string
	Chatter Chatter
}

func NewMessage(msg twitch.PrivateMessage) *Message {
	return &Message{
		Message: msg.Message,
		Time:    msg.Time,
		Channel: msg.Channel,
		Chatter: Chatter{
			Name: msg.User.Name,
			ID:   msg.User.ID,
		},
	}
}

type Command struct {
	Name        string
	Aliases     []string
	Usage       string
	Description string
	Cooldown    int
	Execute     func(message *Message, mb *Monkebot) error
}
