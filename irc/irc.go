package irc

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net/textproto"
	"strings"
	"sync/atomic"
	"time"
)

const MaxMessageLength = 510

type Client struct {
	MessageCount int64
	conn         io.ReadWriter
	messages     chan Message
	channel      string

	rateLimited          bool
	rateLimitMsgCounter  int64
	maxMsgsPerTimeWindow int64
	timeWindow           time.Duration
	beginTimeWindow      time.Time
}

func NewClient(opts ...Option) *Client {
	c := &Client{
		messages: make(chan Message),
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

type Option func(*Client)

func WithRateLimit(timeWindow time.Duration, maxMsgsPerTimeWindow int) Option {
	return func(c *Client) {
		c.timeWindow = timeWindow
		c.maxMsgsPerTimeWindow = int64(maxMsgsPerTimeWindow)
		c.rateLimited = true
	}
}

func (client *Client) Connect(conn io.ReadWriter) {
	client.conn = conn
	go client.listen()
}

func (client *Client) Disconnect() {
	client.sendln("QUIT")
}

func (client *Client) RequestCapabilities(capabilities []string) error {
	s := strings.Join(capabilities, " ")
	return client.sendln("CAP REQ :" + s)
}

func (client *Client) Authenticate(nick string, pass string) error {
	err := client.sendln("PASS " + pass)
	if err != nil {
		return err
	}
	return client.sendln("NICK " + nick)
}

func (client *Client) Join(channel string) error {
	err := client.sendln("JOIN " + channel)
	if err != nil {
		return err
	}
	client.channel = channel
	return nil
}

func (client *Client) Say(msg string) error {
	return client.sendln(fmt.Sprintf("PRIVMSG %s :%s", client.channel, msg))
}

func (client *Client) Messages() <-chan Message {
	return client.messages
}

func (client *Client) sendln(line string) error {
	if client.rateLimited {
		timeLeft := client.timeWindow - time.Since(client.beginTimeWindow)
		if timeLeft > 0 {
			if client.rateLimitMsgCounter >= client.maxMsgsPerTimeWindow {
				time.Sleep(timeLeft)
				client.beginTimeWindow = time.Now()
				atomic.StoreInt64(&client.rateLimitMsgCounter, 0)
			}
		} else {
			client.beginTimeWindow = time.Now()
			atomic.StoreInt64(&client.rateLimitMsgCounter, 0)
		}
	}

	upperBound := len(line)
	if len(line) > MaxMessageLength-2 {
		upperBound = MaxMessageLength - 2
	}
	_, err := client.conn.Write([]byte(line[:upperBound] + "\r\n"))
	if err != nil {
		return err
	}

	atomic.AddInt64(&client.MessageCount, 1)
	atomic.AddInt64(&client.rateLimitMsgCounter, 1)
	log.Println(line)
	return nil
}

func (client *Client) listen() {
	tp := textproto.NewReader(bufio.NewReader(client.conn))

	for line, err := tp.ReadLine(); err == nil; line, err = tp.ReadLine() {
		log.Println(line)

		if strings.HasPrefix(line, "PING") {
			err := client.sendln(strings.Replace(line, "PING", "PONG", 1))
			if err != nil {
				break
			}
		} else {
			parsedMessage := parseMessage(line)
			if parsedMessage != nil {
				client.messages <- *parsedMessage
			}
		}
	}
	close(client.messages)
}
