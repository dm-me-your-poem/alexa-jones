package twitch

import (
	"crypto/tls"
	"net"
	"strings"
	"time"

	"github.com/dm-me-your-poem/alexa-jones/alexa_jones"
	"github.com/dm-me-your-poem/alexa-jones/irc"
)

const (
	MessageRate = time.Duration((30 * time.Second) / 100)
	IrcHost     = "irc.chat.twitch.tv:6667"
	IrcHostTLS  = "irc.chat.twitch.tv:6697"

	MessageRateLimitTimeWindow = 30 * time.Second

	// Adds additional metadata to the command and membership messages. For the
	// list of metadata available with each message, see Twitch tags. To request
	// the tags capability, you must also request the commands capability.
	TagsCapability = "twitch.tv/tags"

	// Lets your bot send PRIVMSG messages that include Twitch chat commands and
	// receive Twitch-specific IRC messages.
	CommandsCapability = "twitch.tv/commands"

	// Lets your bot receive JOIN and PART messages when users join and leave
	// the chat room.
	MembershipCapability = "twitch.tv/membership"
)

type Bot struct {
	channel string
	name    string
	oauth   string

	conn        net.Conn
	ircClient   irc.Client
	startupTime time.Time

	maxMessagesPerTimeWindow int
	messages                 chan alexa_jones.Message
}

func NewBot(name string, channel string, oauth string) *Bot {
	if !strings.HasPrefix(channel, "#") {
		channel = "#" + channel
	}

	return &Bot{
		channel:                  channel,
		oauth:                    oauth,
		name:                     name,
		startupTime:              time.Now(),
		maxMessagesPerTimeWindow: 20,
		messages:                 make(chan alexa_jones.Message),
	}
}

func (bot *Bot) Connect() error {
	conn, err := tls.Dial("tcp", IrcHostTLS, &tls.Config{})
	if err != nil {
		return err
	}
	bot.conn = conn

	bot.ircClient = *irc.NewClient(
		irc.WithRateLimit(MessageRateLimitTimeWindow, bot.maxMessagesPerTimeWindow),
	)

	bot.ircClient.Connect(conn)
	err = bot.ircClient.Authenticate(bot.name, bot.oauth)
	if err != nil {
		return err
	}

	bot.ircClient.RequestCapabilities([]string{
		TagsCapability,
		CommandsCapability,
		MembershipCapability,
	})

	err = bot.ircClient.Join(bot.channel)
	if err != nil {
		return err
	}

	go bot.listen()

	return nil
}

func (bot *Bot) Say(msg string) error {
	return bot.ircClient.Say(msg)
}

func (bot *Bot) Messages() <-chan alexa_jones.Message {
	return bot.messages
}

func (bot *Bot) listen() {
	for ircMessage := range bot.ircClient.Messages() {
		bot.messages <- &Message{
			ircMessage: ircMessage,
			bot:        bot,
		}
	}
	close(bot.messages)
}
