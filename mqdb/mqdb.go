package mqdb

type MqConnection interface {
	Connection() string
}

type MqDatabase interface {
	Test() (string, error)
	NewAccess() (MqAccess, error)
}

type MqAccess interface {
	Close() error
	Publish(queueName string, msg *MqMessage) error
	Consume(queueName string, received func(receiver MqReceiver)) error
}

type MqReceiver interface {
	Ack(multiple bool) error

	ContentType() string
	ContentEncoding() string
	MessageId() string
	Type() string
	Body() []byte
}

type MqMessage struct {
	ContentType     string // MIME content type, default is 'application/json'
	ContentEncoding string // MIME content encoding, default is 'utf-8'
	MessageId       string // message identifier
	Type            string // message type name

	Body []byte
}
