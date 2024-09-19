package command

import (
	"regexp"
)

var buttRegexp = regexp.MustCompile(`^butt`)

var butt = Command{
	Name:        "butt",
	Aliases:     []string{},
	Usage:       "butt[anything]",
	Description: "Responds with butt to messages starting with butt",
	Cooldown:    5,
	NoPrefix:    true,
	NoPrefixShouldRun: func(message *Message, sender MessageSender, args []string) bool {
		return buttRegexp.MatchString(message.Message)
	},
	CanDisable: true,
	Execute: func(message *Message, sender MessageSender, args []string) error {
		sender.Say(message.Channel, "butt")
		return nil
	},
}
