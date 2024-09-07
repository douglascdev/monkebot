package monkebot

import (
	"fmt"
	"strings"
	"time"
	"unicode"

	"github.com/gempir/go-twitch-irc/v4"
)

type MessageSender interface {
	Say(channel string, message string)
}

var commands = []Command{
	{
		Name:        "ping",
		Aliases:     []string{},
		Usage:       "ping",
		Description: "Responds with pong and latency to twitch in milliseconds",
		Cooldown:    5,
		Execute: func(message *Message, sender MessageSender, args []string) error {
			latency := fmt.Sprintf("%d ms", time.Since(message.Time).Milliseconds())
			sender.Say(message.Channel, fmt.Sprintf("🐒 Pong! Latency: %s", latency))
			return nil
		},
	},
	{
		Name:        "senzp",
		Aliases:     []string{},
		Usage:       "senzp <text>",
		Description: "Translates senzp language to english",
		Cooldown:    5,
		Execute: func(message *Message, sender MessageSender, args []string) error {
			cleanString := func(s string) string {
				cleaned := []rune{}
				for _, r := range s {
					if !unicode.Is(unicode.Mn, r) { // Filter out combining marks
						cleaned = append(cleaned, r)
					}
				}
				trimmed := strings.Trim(string(cleaned), " ")
				return strings.ReplaceAll(trimmed, "  ", " ")
			}
			for i, word := range args {
				allLetter := true
				for _, r := range word {
					if !unicode.IsLetter(r) {
						allLetter = false
						break
					}
				}
				if allLetter && word != " " && word != "" && word != "senzpTest" {
					emoteMap := map[string]string{
						"elisAsk":    "catAsk",
						"mysztiHmmm": "hmm",
						"peeepoHUH":  "wtfwtfwtf",
						"exemYes":    "Yes",
					}
					if emote, ok := emoteMap[word]; ok {
						args[i] = " " + emote + " "
						continue
					}
					args[i] = " <emote> "
				}
			}
			senzpWords := strings.Split(strings.Join(args, ""), "senzpTest")

			senzpAlphabet := []string{
				"🅰️", "🅱️", "©️", "↩️", "📧", "🎏", "🗜️", "♓", "ℹ️", "🗾", "🎋", "👢", "〽️", "♑", "🅾️", "🅿️", "♌", "®️", "⚡", "🌴", "⛎", "♈", "〰️", "❌", "🌱", "💤",
			}
			senzpRuneAlphabet := [][]rune{}
			for _, letter := range senzpAlphabet {
				senzpRuneAlphabet = append(senzpRuneAlphabet, []rune(letter))
			}
			alphabet := []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "l", "m", "n", "o", "p", "q", "r", "s", "t", "u", "v", "w", "x", "y", "z"}
			for i, word := range senzpWords {
				newWord := strings.Map(func(c rune) rune {
					for j, senzpRune := range senzpRuneAlphabet {
						if c == senzpRune[0] {
							return []rune(alphabet[j])[0]
						}
					}
					return c
				}, word)
				senzpWords[i] = newWord
			}
			result := cleanString(strings.Join(senzpWords, " "))
			sender.Say(message.Channel, result)
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

func HandleCommands(message *Message, sender MessageSender, config *Config) error {
	if len(message.Message) <= len(config.Prefix) || !strings.HasPrefix(message.Message, config.Prefix) {
		return nil
	}

	args := strings.Split(message.Message[len(config.Prefix):], " ")

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
	Execute     func(message *Message, sender MessageSender, args []string) error
}
