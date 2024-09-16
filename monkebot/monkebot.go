package monkebot

import (
	"database/sql"
	"fmt"
	"monkebot/command"
	"monkebot/config"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/Potat-Industries/go-potatFilters"
	"github.com/gempir/go-twitch-irc/v4"
)

type Monkebot struct {
	TwitchClient *twitch.Client
	Cfg          config.Config
	db           *sql.DB
}

func NewMonkebot(cfg config.Config, db *sql.DB) (*Monkebot, error) {
	client := twitch.NewClient(cfg.Login, "oauth:"+cfg.TwitchToken)
	mb := &Monkebot{
		TwitchClient: client,
		Cfg:          cfg,
		db:           db,
	}

	client.OnPrivateMessage(func(message twitch.PrivateMessage) {
		startTime := time.Now()
		normalizedMsg := command.NewMessage(message)
		if err := command.HandleCommands(normalizedMsg, mb, &cfg); err != nil {
			log.Info().Err(err)
		}
		internalLatency := fmt.Sprintf("%d ms", time.Since(startTime).Milliseconds())
		log.Info().
			Str("channel", message.Channel).
			Str("user", message.User.Name).
			Str("msg", message.Message).
			Str("internalLatency", internalLatency).
			Msg("new message")
	})

	client.OnConnect(func() {
		log.Info().
			Str("login", cfg.Login).
			Msg("connected to Twitch")
		mb.Join(cfg.InitialChannels...)
	})

	client.OnSelfJoinMessage(func(message twitch.UserJoinMessage) {
		log.Info().Str("channel", message.Channel).Msg("joined channel")
	})

	client.OnSelfPartMessage(func(message twitch.UserPartMessage) {
		log.Info().Str("channel", message.Channel).Msg("parted channel")
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
		log.Warn().
			Str("channel", channel).
			Str("message", message).
			Msg("message filtered")
		t.TwitchClient.Say(channel, "⚠ Message withheld for containing a banned phrase...")
		return
	}
	const invisPrefix = "󠀀�" // prevents command injection
	t.TwitchClient.Say(channel, invisPrefix+message)
}
