package rabbitmq

// This package implements the RabbitMQ consumer used by the stats service.
// It listens for requests on a dedicated queue and replies when needed.


import (
	"encoding/json"
	"log"
	"stats_service/models"
	"stats_service/services"

	amqp "github.com/rabbitmq/amqp091-go"
	"gorm.io/gorm"
)


// Names for the exchange, queue and routing keys used by stats_service.
const (
        // ExchangeName is the common direct exchange used by all services.
        ExchangeName = "clearSky.events"
        // QueueName receives RPC messages destined for the stats service.
        QueueName = "stats_queue"
        // RKPersistCalc instructs the service to store grades and calculate
        // distributions.
        RKPersistCalc = "stats.persist_and_calculate"
        // RKGetDists requests the already calculated distributions for a given
        // exam.
        RKGetDists    = "stats.get_distributions"
)

// StartConsumer configures the exchange and queue used by the stats service and
// starts a few goroutines that process messages from RabbitMQ. Each worker
// calls handleMessage for every delivery it receives.

func StartConsumer(db *gorm.DB) {
	if Channel == nil {
		log.Fatal("RabbitMQ channel not initialized")
	}

	if err := Channel.ExchangeDeclare(
		ExchangeName, "direct", true, false, false, false, nil,
	); err != nil {
		log.Fatalf("Failed to declare exchange %s: %v", ExchangeName, err)
	}

	q, err := Channel.QueueDeclare(QueueName, true, false, false, false, nil)
	if err != nil {
		log.Fatalf("Failed to declare queue %s: %v", QueueName, err)
	}

	for _, rk := range []string{RKPersistCalc, RKGetDists} {
		if err := Channel.QueueBind(q.Name, rk, ExchangeName, false, nil); err != nil {
			log.Fatalf("Failed to bind queue %s with key %s: %v", q.Name, rk, err)
		}
	}

	if err := Channel.Qos(1, 0, false); err != nil {
		log.Fatalf("Failed to set QoS: %v", err)
	}

	msgs, err := Channel.Consume(q.Name, "stats_consumer", false, false, false, false, nil)
	if err != nil {
		log.Fatalf("Failed to register consumer: %v", err)
	}

	const workers = 2
	log.Printf("[*] StatsService waiting for messages on %s", q.Name)
	for i := 0; i < workers; i++ {
		go func(id int) {
			log.Printf("Stats worker %d ready", id)
			for d := range msgs {
				handleMessage(db, d)
			}
		}(i)
	}
}

// handleMessage performs the business logic for a single delivery. Depending on
// the routing key it either persists incoming grades or returns previously
// calculated distributions.

func handleMessage(db *gorm.DB, d amqp.Delivery) {
	switch d.RoutingKey {
	case RKPersistCalc:
		var payload struct {
			Exam   models.Exam    `json:"exam"`
			Grades []models.Grade `json:"grades"`
		}
		if err := json.Unmarshal(d.Body, &payload); err != nil {
			log.Printf("invalid persist payload: %v", err)
			d.Nack(false, false)
			return
		}
		if err := services.HandlePersistAndCalculate(db, payload.Exam, payload.Grades); err != nil {
			log.Printf("persist_and_calculate error: %v", err)
			d.Nack(false, false)
			return
		}
		d.Ack(false)
	case RKGetDists:
		var payload struct {
			ClassID  string `json:"class_id"`
			ExamDate string `json:"exam_date"`
		}
		if err := json.Unmarshal(d.Body, &payload); err != nil {
			log.Printf("invalid get_distributions payload: %v", err)
			d.Nack(false, false)
			return
		}
		distributions, err := services.GetDistributions(db, payload.ClassID, payload.ExamDate)
		if err != nil {
			log.Printf("get_distributions error: %v", err)
			d.Nack(false, false)
			return
		}
                // For RPC style requests send the JSON encoded result back to
                // the provided reply queue.
                if d.ReplyTo != "" {
                        body, _ := json.Marshal(distributions)
                        if err := Channel.Publish("", d.ReplyTo, false, false, amqp.Publishing{
                                ContentType:   "application/json",
                                CorrelationId: d.CorrelationId,
                                Body:          body,
                        }); err != nil {
                                log.Printf("failed to publish reply: %v", err)
                        }
                }

		d.Ack(false)
	default:
		log.Printf("unknown routing key %s", d.RoutingKey)
		d.Nack(false, false)
	}
}
