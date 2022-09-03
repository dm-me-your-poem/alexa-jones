package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/dm-me-your-poem/alexa-jones/alexa_jones"
	"github.com/dm-me-your-poem/alexa-jones/twitch"
)

type Config struct {
	CommandPrefix     string `json:"command_prefix"`
	TwitchBotName     string `json:"twitch_bot_name"`
	TwitchChannelName string `json:"twitch_channel_name"`
	TwitchOAuth       string `json:"twitch_oauth"`
}

var quotes []string
var tb *twitch.Bot

func main() {
	rand.Seed(time.Now().Unix())

	configFile, err := os.ReadFile("config.json")
	if err != nil {
		log.Fatalf("[ERROR]: %s\n", err)
	}

	var config Config
	dec := json.NewDecoder(strings.NewReader((string(configFile))))
	if err = dec.Decode(&config); err != nil && err != io.EOF {
		log.Fatalf("[ERROR]: %s\n", err)
	}

	quotesFile, err := os.ReadFile("data/quotes.json")
	if err != nil {
		log.Fatalf("[ERROR]: %s\n", err)
	}

	dec = json.NewDecoder(strings.NewReader((string(quotesFile))))
	if err = dec.Decode(&quotes); err != nil && err != io.EOF {
		log.Fatalf("[ERROR]: %s\n", err)
	}

	tb = twitch.NewBot(
		config.TwitchBotName,
		config.TwitchChannelName,
		config.TwitchOAuth,
	)
	tb.Connect()

	for msg := range tb.Messages() {
		handleMessage(config.CommandPrefix, msg)
	}
}

func handleMessage(commandPrefix string, msg alexa_jones.Message) {
	if strings.HasPrefix(msg.GetContent(), commandPrefix) {
		fields := strings.Fields(msg.GetContent()[len(commandPrefix):])
		if len(fields) >= 1 {
			switch fields[0] {
			case "quote":
				msg.Reply(quotes[rand.Intn(len(quotes))])
			default:
				msg.Reply(fmt.Sprintf("Unknown command '%s'\n", fields[0]))
			}
		}
	}
}
