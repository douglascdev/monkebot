package monkebot

import "github.com/gempir/go-twitch-irc/v4"

type Monkebot struct {
	TwitchClient    *twitch.Client
	InitialChannels []string
}

func NewMonkebot(initialChannels []string) *Monkebot {
	// client := twitch.NewClient("monkebot", "oauth:YOUR_OAUTH_TOKEN_HERE")

	return &Monkebot{
		// TwitchClient: twitchClient,
		// InitialChannels: initialChannels,
	}
}

func (t *Monkebot) JoinInitialChannels() {
	for _, channel := range t.InitialChannels {
		go func() {
			t.TwitchClient.Join(channel)
		}()
	}
}
