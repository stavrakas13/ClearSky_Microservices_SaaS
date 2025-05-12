package rabbitmq

import (
	"fmt"
	"orchestrator/internal/config"

	amqp "github.com/rabbitmq/amqp091-go"
)

// Connect establishes a connection and channel to RabbitMQ.
func Connect() (*amqp.Connection, *amqp.Channel) {
	conn, err := amqp.Dial(config.Cfg.RabbitMQ.URL)
	if err != nil {
		panic(fmt.Errorf("Dial RabbitMQ failed: %w", err))
	}
	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		panic(fmt.Errorf("Open channel failed: %w", err))
	}
	return conn, ch
}

// SetupMessaging declares the exchange, queue with DLX, and bindings.
func SetupMessaging(ch *amqp.Channel) error {
	// Declare exchange
	if err := ch.ExchangeDeclare(
		config.Cfg.Exchange.Name,
		config.Cfg.Exchange.Type,
		true, false, false, false, nil,
	); err != nil {
		return fmt.Errorf("ExchangeDeclare failed: %w", err)
	}
	// Declare queue with DLX settings
	qArgs := amqp.Table{
		"x-dead-letter-exchange":    config.Cfg.Exchange.Name,
		"x-dead-letter-routing-key": config.Cfg.Queue.DLX,
	}
	if _, err := ch.QueueDeclare(
		config.Cfg.Queue.Name,
		true, false, false, false,
		qArgs,
	); err != nil {
		return fmt.Errorf("QueueDeclare failed: %w", err)
	}
	// Declare DLQ
	if _, err := ch.QueueDeclare(
		config.Cfg.Queue.DLX,
		true, false, false, false, nil,
	); err != nil {
		return fmt.Errorf("DLQ Declare failed: %w", err)
	}
	// Bindings
	for _, key := range config.Cfg.Bindings {
		if err := ch.QueueBind(
			config.Cfg.Queue.Name,
			key,
			config.Cfg.Exchange.Name,
			false, nil,
		); err != nil {
			return fmt.Errorf("QueueBind key '%s' failed: %w", key, err)
		}
	}
	return nil
}
