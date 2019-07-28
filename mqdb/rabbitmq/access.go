package rabbitmq

import (
	"fmt"
	"github.com/csby/database/mqdb"
	"github.com/streadway/amqp"
)

type access struct {
	connection *amqp.Connection
}

func (s *access) Close() error {
	if s.connection == nil {
		return nil
	}

	return s.connection.Close()
}

func (s *access) Publish(queueName string, msg *mqdb.MqMessage) error {
	if msg == nil {
		return fmt.Errorf("invalid parameter: msg is nil")
	}

	ch, err := s.connection.Channel()
	if err != nil {
		mqErr, ok := err.(*amqp.Error)
		if ok {
			return fmt.Errorf("%d: %s", mqErr.Code, mqErr.Reason)
		}
		return err
	}
	defer ch.Close()

	q, err := ch.QueueDeclare(queueName, // name
		false, // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		mqErr, ok := err.(*amqp.Error)
		if ok {
			return fmt.Errorf("%d: %s", mqErr.Code, mqErr.Reason)
		}
		return err
	}

	contentType := "application/json"
	if len(msg.ContentType) > 0 {
		contentType = msg.ContentType
	}
	contentEncoding := "utf-8"
	if len(msg.ContentEncoding) > 0 {
		contentEncoding = msg.ContentEncoding
	}

	err = ch.Publish(
		"",     // exchange
		q.Name, // routing key
		false,  // mandatory
		false,  // immediate
		amqp.Publishing{
			ContentType:     contentType,
			ContentEncoding: contentEncoding,
			MessageId:       msg.MessageId,
			Type:            msg.Type,
			Body:            msg.Body,
		})

	if err != nil {
		mqErr, ok := err.(*amqp.Error)
		if ok {
			return fmt.Errorf("%d: %s", mqErr.Code, mqErr.Reason)
		}
	}

	return err
}

func (s *access) Consume(queueName string, received func(mqReceiver mqdb.MqReceiver)) error {
	ch, err := s.connection.Channel()
	if err != nil {
		mqErr, ok := err.(*amqp.Error)
		if ok {
			return fmt.Errorf("%d: %s", mqErr.Code, mqErr.Reason)
		}
		return err
	}
	defer ch.Close()

	q, err := ch.QueueDeclare(queueName, // name
		false, // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		mqErr, ok := err.(*amqp.Error)
		if ok {
			return fmt.Errorf("%d: %s", mqErr.Code, mqErr.Reason)
		}
		return err
	}

	msgs, err := ch.Consume(q.Name, // queue
		"",    // consumer
		false, // auto-ack
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,   // args
	)
	if err != nil {
		mqErr, ok := err.(*amqp.Error)
		if ok {
			return fmt.Errorf("%d: %s", mqErr.Code, mqErr.Reason)
		}
		return err
	}

	for d := range msgs {
		if received != nil {
			mr := &receiver{
				contentType:     d.ContentType,
				contentEncoding: d.ContentEncoding,
				messageId:       d.MessageId,
				bodyType:        d.Type,
				body:            d.Body,
				ack:             d.Ack,
			}
			received(mr)
		}
	}

	return nil
}
