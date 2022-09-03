package irc

import (
	"log"
	"regexp"
	"strings"
)

type Message struct {
	Tags    map[string]string
	Sender  string
	Content string
}

func parseMessage(msg string) *Message {
	re := regexp.MustCompile(`@(?P<tags>\S+) :(?P<sender>\S+)!\S+@\S+\.tmi\.twitch\.tv PRIVMSG #\S+ :(?P<content>.*)$`)

	if re.MatchString(msg) {
		subexpressions := re.FindStringSubmatch(msg)

		result := &Message{
			Tags:    parseTags(subexpressions[re.SubexpIndex("tags")]),
			Sender:  subexpressions[re.SubexpIndex("sender")],
			Content: subexpressions[re.SubexpIndex("content")],
		}
		return result
	} else {
		return nil
	}
}

func parseTags(tags string) map[string]string {
	result := make(map[string]string)

	for _, rawTag := range strings.Split(tags, ";") {
		parsedTag := strings.Split(rawTag, "=")
		if len(parsedTag) == 2 {
			result[parsedTag[0]] = parsedTag[1]
		} else {
			log.Printf("[WARNING]: Malformed tag: `%s`", rawTag)
		}
	}

	return result
}
