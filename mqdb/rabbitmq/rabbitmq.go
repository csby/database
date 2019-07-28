package rabbitmq

import (
	"fmt"
	"github.com/csby/database/mqdb"
	"github.com/streadway/amqp"
	"strings"
)

type rabbitMq struct {
	connection mqdb.MqConnection
}

func NewDatabase(conn mqdb.MqConnection) mqdb.MqDatabase {
	return &rabbitMq{connection: conn}
}

func (s *rabbitMq) Test() (string, error) {
	conn, err := amqp.Dial(s.connection.Connection())
	if err != nil {
		mqErr, ok := err.(*amqp.Error)
		if ok {
			return "", fmt.Errorf("%d: %s", mqErr.Code, mqErr.Reason)
		}
		return "", err
	}
	defer conn.Close()

	sb := &strings.Builder{}
	product, ok := conn.Properties["product"]
	if ok {
		sb.WriteString(fmt.Sprint(product))
		sb.WriteString(" ")
	}
	version, ok := conn.Properties["version"]
	if ok {
		sb.WriteString(fmt.Sprint(version))
	}

	return sb.String(), nil
}

func (s *rabbitMq) NewAccess() (mqdb.MqAccess, error) {
	conn, err := amqp.Dial(s.connection.Connection())
	if err != nil {
		mqErr, ok := err.(*amqp.Error)
		if ok {
			return nil, fmt.Errorf("%d: %s", mqErr.Code, mqErr.Reason)
		}
		return nil, err
	}

	return &access{connection: conn}, nil
}
