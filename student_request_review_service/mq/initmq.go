package mq

import (
	"fmt"

	"github.com/streadway/amqp"
)

var Mqconn *amqp.Connection
var Mqch *amqp.Channel

func InitRabbitMQ() error {
	var err error
	// FOR LOCAL TESTING ONLY.
	//Mqconn, err = amqp.Dial("amqp://guest:guest@localhost:5672/")
	Mqconn, err = amqp.Dial("amqp://guest:guest@rabbitmq:5672/")
	if err != nil {
		fmt.Println("Failed to connect to RabbitMQ:", err)
		return err
	}
	fmt.Println("RabbitMQ connection initialized.")

	Mqch, err = Mqconn.Channel()
	if err != nil {
		fmt.Println("Failed to open a channel:", err)
		return err
	}
	fmt.Println("RabbitMQ Channel initialized.")
	return nil
}
