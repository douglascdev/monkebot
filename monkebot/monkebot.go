package monkebot

import (
	"fmt"
	"log"
	"time"

	"github.com/Potat-Industries/go-potatFilters"
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
		normalizedMsg := NewMessage(message)
		if err := HandleCommands(normalizedMsg, mb, &cfg); err != nil {
			log.Println(err)
		}
		internalLatency := fmt.Sprintf("%d ms", time.Since(startTime).Milliseconds())
		log.Printf("message in %s -> '%s: %s'. Internal latency: %s.", message.Channel, message.User.Name, message.Message, internalLatency)
	})

	client.OnConnect(func() {
		log.Println("connected to Twitch, joining initial channels")
		mb.Join(cfg.InitialChannels...)
	})

	client.OnSelfJoinMessage(func(message twitch.UserJoinMessage) {
		log.Printf("joined channel %s", message.Channel)
	})

	client.OnSelfPartMessage(func(message twitch.UserPartMessage) {
		log.Printf("parted channel %s", message.Channel)
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
	if potatFilters.Test(message, potatFilters.FilterAll) {
		log.Printf("Message filtered in channel '%s': '%s'", channel, message)
		t.TwitchClient.Say(channel, "⚠ Message withheld for containing a banned phrase...")
		return
	}
	const invisPrefix = "󠀀�" // prevents command injection
	t.TwitchClient.Say(channel, invisPrefix+message)
}
