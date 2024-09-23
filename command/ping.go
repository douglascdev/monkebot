package command

import (
	"fmt"
	"monkebot/types"
	"runtime/metrics"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
)

var ping = types.Command{
	Name:              "ping",
	Aliases:           []string{},
	Usage:             "ping",
	Description:       "Responds with pong and latency to twitch in milliseconds",
	Cooldown:          5,
	NoPrefix:          false,
	NoPrefixShouldRun: nil,
	CanDisable:        false,
	Execute: func(message *types.Message, sender types.MessageSender, args []string) error {
		responses := []string{
			"🐒 Pong!",
		}

		latency, err := sender.Ping()
		if err != nil {
			log.Warn().Err(err).Msg("failed to get latency for ping message")
		} else if latency == 0 {
			log.Warn().Msg("failed to get latency for ping message, client sent no pings yet")
		} else {
			responses = append(responses, fmt.Sprintf("Latency: %dms", latency.Milliseconds()))
		}

		memSamples := []metrics.Sample{
			{Name: "/memory/classes/total:bytes"},
		}
		metrics.Read(memSamples)

		responses = append(responses,
			fmt.Sprintf("Memory: %d MiB", memSamples[0].Value.Uint64()/1024/1024),
			fmt.Sprintf("Uptime: %s", sender.Uptime().Round(time.Second)),
		)

		sender.Say(message.Channel, strings.Join(responses, " 🍌 "))
		return nil
	},
}
