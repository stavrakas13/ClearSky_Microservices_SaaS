package rabbitmq

import (
	"fmt"
	"log"
	"orchestrator/internal/config"
	"orchestrator/internal/handlers"

	amqp "github.com/rabbitmq/amqp091-go"
)

// StartOrchestratorConsumer opens the consumer and dispatches deliveries.
func StartOrchestratorConsumer(ch *amqp.Channel) error {
	msgs, err := ch.Consume(
		config.Cfg.Queue.Name,
		"orchestrator-consumer",
		false, false, false, false, nil,
	)
	if err != nil {
		return fmt.Errorf("Consume failed: %w", err)
	}
	go func() {
		for d := range msgs {
			switch d.RoutingKey {
			case "user.created":
				handlers.HandleUserCreated(d)
			// other cases omitted for brevity
			default:
				log.Printf("[Orchestrator] Unknown key: %s", d.RoutingKey)
				d.Nack(false, false)
			}
		}
	}()
	return nil
}
