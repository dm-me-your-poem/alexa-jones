package twitch

import "github.com/dm-me-your-poem/alexa-jones/irc"

type Message struct {
	ircMessage irc.Message
	bot        *Bot
}

func (m *Message) GetContent() string {
	return m.ircMessage.Content
}

func (m *Message) Reply(reply string) error {
	return m.bot.Say(reply)
}
