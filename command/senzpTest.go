package command

import (
	"monkebot/types"
	"regexp"
	"strings"
	"unicode"
)

var expr = regexp.MustCompile(`^senzpTest`)

func shouldRun(message *types.Message, sender types.MessageSender, args []string) bool {
	return expr.MatchString(message.Message)
}

var senzpTest = types.Command{
	Name:              "senzpTest",
	Aliases:           []string{},
	Usage:             "senzpTest <text>",
	Description:       "Translates senzp language to english",
	ChannelCooldown:   5,
	UserCooldown:      5,
	NoPrefix:          true,
	NoPrefixShouldRun: shouldRun,
	CanDisable:        true,
	Execute: func(message *types.Message, sender types.MessageSender, args []string) error {
		cleanString := func(s string) string {
			cleaned := []rune{}
			for _, r := range s {
				if !unicode.Is(unicode.Mn, r) { // Filter out combining marks
					cleaned = append(cleaned, r)
				}
			}
			trimmed := strings.Trim(string(cleaned), " ")
			return strings.ReplaceAll(trimmed, "  ", " ")
		}
		for i, word := range args {
			allLetter := true
			for _, r := range word {
				if !unicode.IsLetter(r) {
					allLetter = false
					break
				}
			}
			if allLetter && word != " " && word != "" && word != "senzpTest" {
				emoteMap := map[string]string{
					"elisAsk":       "catAsk",
					"mysztiHmmm":    "hmm",
					"peeepoHUH":     "wtfwtfwtf",
					"exemYes":       "Yes",
					"mysztiOh":      "Oh",
					"vulpNou":       "NoU",
					"vulpSoCute":    "SoCute",
					"mysztiOopsie":  "oopsie",
					"senzpNOWAYING": "NOWAYING",
					"elisEHEHE":     "EHEHE",
					"sammim1HEHE":   "EHEHE",
					"sammim1Wade":   "wade",
					"hvdrasWowie":   "Wowie",
					"hvdrasWoah":    "NOWAY",
					"neeshSadge":    "sadg",
				}
				if emote, ok := emoteMap[word]; ok {
					args[i] = " " + emote + " "
					continue
				}
				args[i] = " <emote> "
			}
		}
		senzpWords := strings.Split(strings.Join(args, ""), "senzpTest")

		senzpAlphabet := []string{
			"🅰️", "🅱️", "©️", "↩️", "📧", "🎏", "🗜️", "♓", "ℹ️", "🗾", "🎋", "👢", "〽️", "♑", "🅾️", "🅿️", "♌", "®️", "⚡", "🌴", "⛎", "♈", "〰️", "❌", "🌱", "💤",
		}
		senzpRuneAlphabet := [][]rune{}
		for _, letter := range senzpAlphabet {
			senzpRuneAlphabet = append(senzpRuneAlphabet, []rune(letter))
		}
		alphabet := []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "l", "m", "n", "o", "p", "q", "r", "s", "t", "u", "v", "w", "x", "y", "z"}
		for i, word := range senzpWords {
			newWord := strings.Map(func(c rune) rune {
				for j, senzpRune := range senzpRuneAlphabet {
					if c == senzpRune[0] {
						return []rune(alphabet[j])[0]
					}
				}
				return c
			}, word)
			senzpWords[i] = newWord
		}
		result := cleanString(strings.Join(senzpWords, " "))
		sender.Say(message.Channel, result)
		return nil
	},
}
