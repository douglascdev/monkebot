package command

var buttsbot = Command{
	Name:        "buttsbot",
	Aliases:     []string{},
	Usage:       "send any message in chat",
	Description: "Replaces random syllables with butt",
	Cooldown:    5,
	NoPrefix:    true,
	NoPrefixShouldRun: func(message *Message, sender MessageSender, args []string) bool {
		return sender.ShouldButtify()
	},
	CanDisable: false,
	Execute: func(message *Message, sender MessageSender, args []string) error {
		newSentence := sender.Buttify(message.Message)
		if newSentence != message.Message {
			sender.Say(message.Channel, newSentence)
		}
		return nil
	},
}
