package client

import (
	"database/sql"
	"monkebot/config"

	"github.com/gempir/go-twitch-irc/v4"
	"github.com/rs/zerolog/log"
)

type TwitchClient struct {
	client *twitch.Client
	config *config.Config
	db     *sql.DB
}

func NewTwitchClient(client *twitch.Client, cfg *config.Config) *TwitchClient {
	c := &TwitchClient{
		client: client,
		config: cfg,
	}

	c.client.OnConnect(func() {
		c.OnConnect()
	})

	c.client.OnSelfJoinMessage(func(message twitch.UserJoinMessage) {
		c.OnSelfJoin(message.Channel)
	})
	c.client.OnSelfPartMessage(func(message twitch.UserPartMessage) {
		c.OnSelfPart(message.Channel)
	})

	return c
}

func (tc *TwitchClient) Join(channels ...PlatformUser) {
	channelNames := make([]string, len(channels))
	for _, channel := range channels {
		channelNames = append(channelNames, channel.Name)
	}

	tc.client.Join(channelNames...)
}

func (tc *TwitchClient) Part(channels ...PlatformUser) {
	for _, channel := range channels {
		tc.client.Depart(channel.Name)
	}
}

func (tc *TwitchClient) Say(channel PlatformUser, message string) {
	tc.client.Say(channel.Name, message)
}

func (tc *TwitchClient) OnConnect() {
	log.Info().
		Str("login", tc.config.Login).
		Msg("connected to Twitch")
}

func (tc *TwitchClient) OnSelfJoin(channel string) {
	log.Info().Str("channel", channel).Msg("joined channel")
}

func (tc *TwitchClient) OnSelfPart(channel string) {
	log.Info().Str("channel", channel).Msg("parted channel")
}

func (tc *TwitchClient) OnMessage(channel string, message string) {
	log.Info().
		Str("channel", channel).
		Str("message", message).
		Msg("new message")
}

func (tc *TwitchClient) BeginTransaction() (*sql.Tx, error) {
	return tc.db.Begin()
}

func (tc *TwitchClient) Connect() error {
	return tc.client.Connect()
}
