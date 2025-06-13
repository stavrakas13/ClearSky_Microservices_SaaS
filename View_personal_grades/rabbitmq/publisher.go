package rabbitmq

import (
	"encoding/json"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

// StatsPersistCalc is the routing key used by the stats service for persisting
// grades and computing distributions.
const StatsPersistCalc = "stats.persist_and_calculate"

// PublishPersistAndCalculate forwards the provided payload to the stats service
// through RabbitMQ so that it can store the grades and calculate statistics.
func PublishPersistAndCalculate(payload interface{}) error {
	if Channel == nil {
		return fmt.Errorf("RabbitMQ channel not initialized")
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal payload: %w", err)
	}
	return Channel.Publish(
		ExchangeName,
		StatsPersistCalc,
		false, false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)
}
