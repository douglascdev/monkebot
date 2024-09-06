package monkebot

import (
	"fmt"
	"log"
	"time"

	"github.com/gempir/go-twitch-irc/v4"
)

type Monkebot struct {
	TwitchClient *twitch.Client
	Cfg          Config
}

func NewMonkebot(cfg Config, token string) (*Monkebot, error) {
	client := twitch.NewClient(cfg.Login, "oauth:"+token)
	mb := &Monkebot{
		TwitchClient: client,
		Cfg:          cfg,
	}

	client.OnPrivateMessage(func(message twitch.PrivateMessage) {
		startTime := time.Now()
		if message.Message == cfg.Prefix+"ping" {
			latency := time.Now().Sub(message.Time)
			response := fmt.Sprintf("🐒 Pong! Latency: %dms", latency.Milliseconds())
			mb.Say(message.Channel, response)
		}
		internalLatency := fmt.Sprintf("%d ms", time.Now().Sub(startTime).Milliseconds())
		log.Printf("Message in %s -> '%s: %s'. Internal latency: %s.", message.Channel, message.User.Name, message.Message, internalLatency)
	})

	client.OnConnect(func() {
		log.Println("Connected to Twitch, joining initial channels")
		mb.Join(cfg.InitialChannels...)
	})

	client.OnSelfJoinMessage(func(message twitch.UserJoinMessage) {
		log.Printf("Joined channel %s", message.Channel)
	})

	client.OnSelfPartMessage(func(message twitch.UserPartMessage) {
		log.Printf("Parted channel %s", message.Channel)
	})
	return mb, nil
}

func (t *Monkebot) Connect() error {
	return t.TwitchClient.Connect()
}

func (t *Monkebot) Join(channels ...string) {
	t.TwitchClient.Join(channels...)
}

func (t *Monkebot) Part(channels ...string) {
	for _, channel := range channels {
		t.TwitchClient.Depart(channel)
	}
}

func (t *Monkebot) Say(channel string, message string) {
	t.TwitchClient.Say(channel, message)
}
