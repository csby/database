package rabbitmq

type receiver struct {
	contentType     string
	contentEncoding string
	messageId       string
	bodyType        string
	body            []byte

	ack func(multiple bool) error
}

func (s *receiver) Ack(multiple bool) error {
	if s.ack == nil {
		return nil
	}

	return s.ack(multiple)
}

func (s *receiver) ContentType() string {
	return s.contentType
}

func (s *receiver) ContentEncoding() string {
	return s.contentEncoding
}

func (s *receiver) MessageId() string {
	return s.messageId
}

func (s *receiver) Type() string {
	return s.bodyType
}

func (s *receiver) Body() []byte {
	return s.body
}
