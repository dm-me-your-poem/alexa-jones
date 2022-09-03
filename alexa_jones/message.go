package alexa_jones

type Message interface {
	GetContent() string
	Reply(msg string) error
}
