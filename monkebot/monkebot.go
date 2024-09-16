package monkebot

import (
	"database/sql"
	"fmt"
	"monkebot/command"
	"monkebot/config"
	"monkebot/database"
	"monkebot/twitchapi"
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
		normalizedMsg := command.NewMessage(message, db)
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

		tx, err := db.Begin()
		if err != nil {
			log.Err(err).Msg("failed to initialize transaction")
		}
		defer tx.Rollback()

		var cmdNames []string
		for _, cmd := range command.Commands {
			cmdNames = append(cmdNames, cmd.Name)
		}

		err = database.InsertCommands(tx, cmdNames...)
		if err != nil {
			log.Err(err).Msg("failed to insert commands")
			return
		}

		var twitchUsers []*twitch.User
		twitchUsers, err = twitchapi.GetUserByName(&cfg, cfg.InitialChannels...)
		if err != nil {
			log.Err(err).Strs("channels", cfg.InitialChannels).Msg("failed to get helix data for users")
			return
		}

		var users []struct {
			ID   string
			Name string
		}
		for _, twitchUser := range twitchUsers {
			users = append(users, struct {
				ID   string
				Name string
			}{twitchUser.ID, twitchUser.Name})
		}

		err = database.InsertUsers(tx, true, users...)
		if err != nil {
			log.Err(err).Msg("failed to insert initial users in the database")
			return
		}

		for _, twitchUser := range twitchUsers {
			err = database.InsertUserCommands(tx, twitchUser.ID, cmdNames...)
			if err != nil {
				log.Err(err).Str("name", twitchUser.Name).Str("id", twitchUser.ID).Msg("failed to insert user commands for user")
				return
			}
		}

		err = tx.Commit()
		if err != nil {
			log.Err(err).Msg("failed to commit transaction")
			return
		}

		log.Info().Msg("successfully inserted initial channels, you may remove the channels from the config now")
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
