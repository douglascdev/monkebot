package command

import "monkebot/types"

var buttsbot = types.Command{
	Name:            "buttsbot",
	Aliases:         []string{},
	Usage:           "send any message in chat",
	Description:     "Replaces random syllables with butt",
	ChannelCooldown: 60,
	UserCooldown:    60 * 5,
	NoPrefix:        true,
	NoPrefixShouldRun: func(message *types.Message, sender types.MessageSender, args []string) bool {
		return sender.ShouldButtify()
	},
	CanDisable: true,
	Execute: func(message *types.Message, sender types.MessageSender, args []string) error {
		newSentence := sender.Buttify(message.Message)
		if newSentence != message.Message {
			sender.Say(message.Channel, newSentence)
		}
		return nil
	},
}
