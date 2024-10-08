package types

import (
	"database/sql"
	"time"

	"monkebot/config"

	"github.com/gempir/go-twitch-irc/v4"
)

// Command is a struct defining a command.
//
// An ArgSpec is an argument specification, a struct defining the expected arguments for this command.
// Each ArgSpec is an alternative valid set of arguments for this command, meaning
// that if any of the ArgSpecs are matched, the Command will be executed.
// The tags required(true/false) and argtype(positional/named) may be specified
// for each field defined in an ArgSpec struct.
type Command struct {
	Name            string
	Aliases         []string
	Usage           string
	Description     string
	ChannelCooldown int
	UserCooldown    int
	NoPrefix        bool
	CanDisable      bool

	// `json:"-"` excludes these fields from being serialized into the command list json
	ArgSpecs          []interface{}                                                     `json:"-"`
	NoPrefixShouldRun func(message *Message, sender MessageSender, args []string) bool  `json:"-"`
	Execute           func(message *Message, sender MessageSender, args []string) error `json:"-"`
}

type SortByPrefixAndName []Command

func (a SortByPrefixAndName) Len() int      { return len(a) }
func (a SortByPrefixAndName) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a SortByPrefixAndName) Less(i, j int) bool {
	if a[i].Name == a[j].Name {
		return a[i].NoPrefix && !a[j].NoPrefix
	}
	return a[i].Name < a[j].Name
}

type Chatter struct {
	Name string
	ID   string

	IsMod         bool
	IsVIP         bool
	IsBroadcaster bool
}

// Message normalized to be platform agnostic
type Message struct {
	ID      string
	Message string
	Time    time.Time
	Channel string
	Cfg     *config.Config
	RoomID  string
	Chatter Chatter
	DB      *sql.DB
}

func NewMessage(msg twitch.PrivateMessage, db *sql.DB, cfg *config.Config) *Message {
	return &Message{
		ID:      msg.ID,
		Message: msg.Message,
		Time:    msg.Time,
		Channel: msg.Channel,
		RoomID:  msg.RoomID,
		Cfg:     cfg,
		Chatter: Chatter{
			Name:          msg.User.Name,
			ID:            msg.User.ID,
			IsMod:         msg.User.IsMod,
			IsVIP:         msg.User.IsVip,
			IsBroadcaster: msg.User.IsBroadcaster,
		},
		DB: db,
	}
}

type SenderParam int

const (
	ReplyMessageID SenderParam = iota
	Me
)

type MessageSender interface {
	Say(channel string, message string, params ...struct {
		Param SenderParam
		Value string
	})

	Join(channels ...string)
	Part(channels ...string)
	Ping() (time.Duration, error)

	Uptime() time.Duration

	Buttify(message string) string
	ShouldButtify() bool
}
