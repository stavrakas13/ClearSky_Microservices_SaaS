package rabbitmq

import (
	"encoding/json"
	"fmt"
	"log"
	"stats_service/models"
	"stats_service/services" // Τα δικά σου services

	amqp "github.com/rabbitmq/amqp091-go"
	"gorm.io/gorm"
)

const RPCQueueName = "stats_service_rpc_queue"
const ExchangeName = "clearsky.rpc.exchange" // Προτείνω ξεχωριστό exchange για RPC

// Routing keys για τις RPC κλήσεις προς το stats_service
const RKPing = "stats.ping"
const RKPesistAndCalculate = "stats.persist_and_calculate"
const RKGetDistributions = "stats.get_distributions"

type PersistDataPayload struct {
	Exam   models.Exam    `json:"exam"`
	Grades []models.Grade `json:"grades"`
}

type GetDistributionsPayload struct {
	ClassID  string `json:"class_id"`
	ExamDate string `json:"exam_date"`
}

type RPCResponse struct {
	Status  string      `json:"status"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

func StartStatsRPCServer(db *gorm.DB) {
	if Channel == nil {
		log.Fatal("FATAL: RabbitMQ channel is not initialized. Call rabbitmq.Init() first.")
	}

	err := Channel.ExchangeDeclare(
		ExchangeName, "direct", true, false, false, false, nil,
	)
	if err != nil {
		log.Fatalf("FATAL: Failed to declare exchange '%s': %v", ExchangeName, err)
	}

	q, err := Channel.QueueDeclare(
		RPCQueueName, true, false, false, false, nil,
	)
	if err != nil {
		log.Fatalf("FATAL: Failed to declare queue '%s': %v", q.Name, err)
	}

	routingKeys := []string{RKPing, RKPesistAndCalculate, RKGetDistributions}
	for _, rk := range routingKeys {
		log.Printf("Binding queue %s to exchange %s with routing key %s", q.Name, ExchangeName, rk)
		err = Channel.QueueBind(q.Name, rk, ExchangeName, false, nil)
		if err != nil {
			log.Fatalf("FATAL: Failed to bind queue for key %s: %v", rk, err)
		}
	}

	err = Channel.Qos(1, 0, false)
	if err != nil {
		log.Fatalf("FATAL: Failed to set QoS: %v", err)
	}

	msgs, err := Channel.Consume(
		q.Name, "", false, false, false, false, nil,
	)
	if err != nil {
		log.Fatalf("FATAL: Failed to register a consumer: %v", err)
	}

	log.Printf(" [*] StatsService RPC Server waiting for messages on queue '%s'. To exit press CTRL+C", q.Name)

	numWorkers := 2
	for i := 0; i < numWorkers; i++ {
		go func(workerID int) {
			log.Printf("INFO: RPC Worker %d started", workerID)
			for d := range msgs {
				log.Printf("INFO: Worker %d received RPC request with routing key '%s', CorrelationID: %s, ReplyTo: %s", workerID, d.RoutingKey, d.CorrelationId, d.ReplyTo)

				var response RPCResponse
				var processingError error

				switch d.RoutingKey {
				case RKPing:
					response = RPCResponse{Status: "ok", Message: "stats_service pong"}
				case RKPesistAndCalculate:
					var payload PersistDataPayload
					if err := json.Unmarshal(d.Body, &payload); err != nil {
						response = RPCResponse{Status: "error", Message: "Invalid payload for persist_and_calculate: " + err.Error()}
						processingError = err
					} else {
						// **ΣΩΣΤΗ ΚΛΗΣΗ ΕΔΩ**
						err := services.HandlePersistAndCalculate(db, payload.Exam, payload.Grades)
						if err != nil {
							response = RPCResponse{Status: "error", Message: "Failed to process data: " + err.Error()}
							processingError = err
						} else {
							response = RPCResponse{Status: "ok", Message: "Data processed and distributions calculated."}
						}
					}
				case RKGetDistributions:
					var payload GetDistributionsPayload
					if err := json.Unmarshal(d.Body, &payload); err != nil {
						response = RPCResponse{Status: "error", Message: "Invalid payload for get_distributions: " + err.Error()}
						processingError = err
					} else {
						distributions, err := services.GetDistributions(db, payload.ClassID, payload.ExamDate)
						if err != nil {
							response = RPCResponse{Status: "error", Message: "Failed to get distributions: " + err.Error()}
							processingError = err
						} else {
							if len(distributions) == 0 {
								response = RPCResponse{Status: "ok", Message: "No distributions found for the given criteria.", Data: []models.GradeDistribution{}}
							} else {
								response = RPCResponse{Status: "ok", Data: distributions}
							}
						}
					}
				default:
					response = RPCResponse{Status: "error", Message: "Unknown RPC routing key: " + d.RoutingKey}
					processingError = fmt.Errorf("unknown RPC routing key: %s", d.RoutingKey)
				}

				responseBody, err := json.Marshal(response)
				if err != nil {
					log.Printf("ERROR: Worker %d failed to marshal RPC response: %v", workerID, err)
					d.Nack(false, false)
					continue
				}

				if d.ReplyTo != "" {
					err = Channel.Publish(
						"", d.ReplyTo, false, false,
						amqp.Publishing{
							ContentType:   "application/json",
							CorrelationId: d.CorrelationId,
							Body:          responseBody,
						})
					if err != nil {
						log.Printf("ERROR: Worker %d failed to publish RPC reply to %s: %v", workerID, d.ReplyTo, err)
						if processingError == nil {
							d.Ack(false)
						} else {
							d.Nack(false, false)
						}
						continue
					}
					log.Printf("INFO: Worker %d sent RPC reply to %s for CorrelationID: %s", workerID, d.ReplyTo, d.CorrelationId)
				} else {
					log.Printf("WARNING: Worker %d received message without ReplyTo field. RoutingKey: %s", workerID, d.RoutingKey)
				}

				// Επιβεβαίωση (Acknowledgement) στο RabbitMQ.
				if processingError != nil {
					// Αν υπήρξε σφάλμα στην επεξεργασία, στέλνουμε Nack.
					// Το `false` στο δεύτερο όρισμα σημαίνει "μην το ξαναβάλεις στην ουρά (requeue)".
					// Ιδανικά, θα πήγαινε σε μια Dead Letter Queue (DLQ) αν έχει ρυθμιστεί.
					d.Nack(false, false)
				} else {
					// Αν όλα πήγαν καλά, στέλνουμε Ack.
					d.Ack(false)
				}
			}
			log.Printf("INFO: RPC Worker %d stopped.", workerID)
		}(i) // Περνάμε το i για να έχει κάθε goroutine το δικό της workerID.
	}
}
