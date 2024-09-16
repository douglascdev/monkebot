package client

// import (
// 	"monkebot/config"
//
// 	"github.com/gempir/go-twitch-irc/v4"
// 	"github.com/rs/zerolog/log"
// )
//
// type TwitchClient struct {
// 	client *twitch.Client
// }
//
// func NewTwitchClient(client *twitch.Client, cfg *config.Config) *TwitchClient {
// 	c := &TwitchClient{
// 		client: client,
// 	}
// 	c.client.OnConnect(func() {
// 		c.OnConnect()
// 	})
//
// 	c.client.OnSelfJoinMessage(func(message twitch.UserJoinMessage) {
// 		c.OnSelfJoin(message.Channel)
// 	})
// 	c.client.OnSelfPartMessage(func(message twitch.UserPartMessage) {
// 		c.OnSelfPart(message.Channel)
// 	})
// 	return c
// }
//
// func (tc *TwitchClient) Join(channels ...PlatformUser) {
// 	channelNames := make([]string, len(channels))
// 	for _, channel := range channels {
// 		channelNames = append(channelNames, channel.Name)
// 	}
//
// 	tc.client.Join(channelNames...)
// }
//
// func (tc *TwitchClient) Part(channels ...PlatformUser) {
// 	for _, channel := range channels {
// 		tc.client.Depart(channel.Name)
// 	}
// }
//
// func (tc *TwitchClient) Say(channel PlatformUser, message string) {
// 	tc.client.Say(channel.Name, message)
// }
//
// func (tc *TwitchClient) OnConnect() {
// 	log.Info().
// 		Str("login", cfg.Login).
// 		Msg("connected to Twitch")
// }
//
// func (tc *TwitchClient) OnSelfJoin(cb func(channel string)) {
// 	log.Info().Str("channel", channel).Msg("joined channel")
// }
//
// func (tc *TwitchClient) OnSelfPart(cb func(channel string)) {
// 	log.Info().Str("channel", channel).Msg("parted channel")
// }
//
// func (tc *TwitchClient) Connect() error {
// 	return tc.client.Connect()
// }
