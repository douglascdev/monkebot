package command

var buttsbot = Command{
	Name:        "buttsbot",
	Aliases:     []string{},
	Usage:       "send any message in chat",
	Description: "Replaces random syllables with butt",
	Cooldown:    5,
	NoPrefix:    true,
	// TODO: this affects the probability, Buttify shoudn't have to run twice
	NoPrefixShouldRun: func(message *Message, sender MessageSender, args []string) bool {
		_, didButtify := sender.Buttify(message.Message)
		return didButtify
	},
	CanDisable: false,
	Execute: func(message *Message, sender MessageSender, args []string) error {
		newSentence, didButtify := sender.Buttify(message.Message)
		if didButtify {
			sender.Say(message.Channel, newSentence)
		}
		return nil
	},
}
