package rabbitmq

import (
	"encoding/json"
	"fmt"
	"github.com/csby/database/mqdb"
	"testing"
)

func TestRabbitMq_Test(t *testing.T) {
	db := NewDatabase(testConnection())
	info, err := db.Test()
	if err != nil {
		t.Fatal(err)
	}
	t.Log("info:", info)
}

func TestAccess_Publish(t *testing.T) {
	db := NewDatabase(testConnection())
	ac, err := db.NewAccess()
	if err != nil {
		t.Fatal(err)
	}
	defer ac.Close()

	argument := &testArgument{
		PatientID: "A20070910",
		VisitID:   "1",
	}
	body, err := json.Marshal(argument)
	if err != nil {
		t.Fatal(err)
	}

	name := "unit-test"
	msg := &mqdb.MqMessage{
		Body: body,
	}
	err = ac.Publish(name, msg)
	if err != nil {
		t.Fatal(err)
	}
}

func TestAccess_Consume1(t *testing.T) {
	testAccess_Consume(t)
}

func TestAccess_Consume2(t *testing.T) {
	testAccess_Consume(t)
}

func testAccess_Consume(t *testing.T) {
	db := NewDatabase(testConnection())
	ac, err := db.NewAccess()
	if err != nil {
		t.Fatal(err)
	}
	defer ac.Close()

	name := "unit-test"
	err = ac.Consume(name, func(receiver mqdb.MqReceiver) {
		fmt.Println("ContentType:", receiver.ContentType())
		fmt.Println("ContentEncoding:", receiver.ContentEncoding())
		fmt.Println("MessageId:", receiver.MessageId())
		fmt.Println("Type:", receiver.Type())
		fmt.Println("Body:", string(receiver.Body()))
		fmt.Println("")
		receiver.Ack(true)
	})
	if err != nil {
		t.Fatal(err)
	}
}

func testConnection() *Connection {
	return &Connection{
		Server:      "192.168.123.5",
		Port:        5672,
		User:        "dev",
		Password:    "pwd",
		VirtualHost: "host-dev",
	}
}

type testArgument struct {
	PatientID string `json:"patientId"`
	VisitID   string `json:"visitId"`
}
