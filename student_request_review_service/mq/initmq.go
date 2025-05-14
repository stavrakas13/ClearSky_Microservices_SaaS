package mq

import (
	"fmt"

	"github.com/streadway/amqp"
)

var Mqconn *amqp.Connection
var Mqch *amqp.Channel

func InitRabbitMQ() {
	var err error
	Mqconn, err = amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		fmt.Println("Failed to connect to RabbitMQ")
		fmt.Println(err)

	}
	fmt.Println("RabbitMQ connection initialized.")
	Mqch, err = Mqconn.Channel()
	if err != nil {
		fmt.Println("Failed to open a channel")
		fmt.Println(err)
	}

	fmt.Println("RabbitMQ Channel initialized.")
}
