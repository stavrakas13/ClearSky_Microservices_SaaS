package rabbitmq

import (
	"encoding/json"
	"log"
	"stats_service/models"
	"stats_service/services"

	amqp "github.com/rabbitmq/amqp091-go"
	"gorm.io/gorm"
)

const RPCQueueName = "stats_service_rpc_queue" // Η ουρά που θα ακούει το service σου
// Ο Orchestrator θα πρέπει να ξέρει αυτό το όνομα για να στέλνει αιτήματα.
// Ή, μπορείς να χρησιμοποιήσεις ένα πιο γενικό routing key σε έναν direct exchange.

// Αυτό το struct θα μπορούσε να είναι το payload για την εισαγωγή δεδομένων
type PersistDataPayload struct {
	Exam   models.Exam    `json:"exam"`
	Grades []models.Grade `json:"grades"`
}

// Αυτό για την ανάκτηση
type GetDistributionsPayload struct {
	ClassID  string `json:"class_id"`
	ExamDate string `json:"exam_date"`
}

type RPCResponse struct {
	Status  string      `json:"status"` // "ok" or "error"
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"` // Για τα αποτελέσματα
}

func StartStatsRPCServer(db *gorm.DB) {
	// Δήλωσε τον exchange που θα χρησιμοποιεί ο Orchestrator για να στέλνει commands
	// (μπορεί να είναι ο ίδιος που χρησιμοποιούν και άλλα services, π.χ., "clearsky.commands")
	// ή ένας πιο γενικός "clearsky.events" αν ο Orchestrator στέλνει events που προορίζονται για RPC.
	// Ας υποθέσουμε έναν direct exchange "service_rpc_exchange" για απλότητα εδώ.
	err := Channel.ExchangeDeclare(
		"service_rpc_exchange", // Όνομα exchange
		"direct",               // Τύπος
		true,                   // durable
		false,                  // auto-deleted
		false,                  // internal
		false,                  // no-wait
		nil,                    // arguments
	)
	if err != nil {
		log.Fatalf("Failed to declare exchange 'service_rpc_exchange': %v", err)
	}

	//Δηλώνει την ουρά RPCQueueName όπου θα λαμβάνει τα αιτήματα το service σου.
	q, err := Channel.QueueDeclare(
		RPCQueueName, // name
		true,         // durable
		false,        // delete when unused
		false,        // exclusive
		false,        // no-wait
		nil,          // arguments
	)
	if err != nil {
		log.Fatalf("Failed to declare queue '%s': %v", q.Name, err)
	}

	// Κάνε bind την ουρά στον exchange με συγκεκριμένα routing keys
	// που θα χρησιμοποιεί ο Orchestrator για να καλέσει τις λειτουργίες σου.
	routingKeys := []string{
		"stats.persist_and_calculate",
		"stats.get_distributions",
	}

	//Συνδέει (bind) την ουρά σου (q.Name) με τον exchange ("service_rpc_exchange") για κάθε ένα από τα routingKeys.
	for _, rk := range routingKeys {
		err = Channel.QueueBind(
			q.Name,                 // queue name
			rk,                     // routing key
			"service_rpc_exchange", // exchange
			false,
			nil,
		)
		if err != nil {
			log.Fatalf("Failed to bind queue for key %s: %v", rk, err)
		}
	}

	// Κατανάλωση μηνυμάτων από την ουρά RPC
	msgs, err := Channel.Consume(
		q.Name, // queue
		"",     // consumer tag (το RabbitMQ θα γεννήσει ένα)
		false,  // auto-ack (false για χειροκίνητο ack/nack)
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	if err != nil {
		log.Fatalf("Failed to register a consumer: %v", err)
	}

	log.Printf(" [*] StatsService RPC Server waiting for messages on queue '%s'. To exit press CTRL+C", q.Name)

	//Ξεκινάει ένα νέο "νήμα" (goroutine) που επεξεργάζεται τα εισερχόμενα μηνύματα ένα-ένα.
	go func() {
		for d := range msgs {
			log.Printf("Received RPC request with routing key '%s': %s", d.RoutingKey, d.Body)
			var response RPCResponse
			var responseBody []byte

			switch d.RoutingKey {
			//Αυτή είναι η συνάρτηση που θα υλοποιήσεις εσύ στο services package σου,
			// παίρνοντας τη λογική από το παλιό PostData και την κλήση στο CalculateDistributions.
			case "stats.persist_and_calculate":
				var payload PersistDataPayload
				if err := json.Unmarshal(d.Body, &payload); err != nil {
					response = RPCResponse{Status: "error", Message: "Invalid payload: " + err.Error()}
				} else {
					// Κάλεσε την υπάρχουσα λογική σου, αλλά τροποποίησέ την
					// για να μην εξαρτάται από το gin.Context
					err := services.CalculateDistributions(db, payload.Exam, payload.Grades)
					if err != nil {
						response = RPCResponse{Status: "error", Message: "Failed to process data: " + err.Error()}
					} else {
						response = RPCResponse{Status: "ok", Message: "Data processed and distributions calculated."}
					}
				}

			case "stats.get_distributions":
				var payload GetDistributionsPayload
				if err := json.Unmarshal(d.Body, &payload); err != nil {
					response = RPCResponse{Status: "error", Message: "Invalid payload: " + err.Error()}
				} else {
					// Κάλεσε την υπάρχουσα λογική σου για να πάρεις τις κατανομές
					distributions, err := services.GetDistributions(db, payload.ClassID, payload.ExamDate)
					if err != nil {
						response = RPCResponse{Status: "error", Message: "Failed to get distributions: " + err.Error()}
					} else {
						response = RPCResponse{Status: "ok", Data: distributions}
					}
				}

			default:
				response = RPCResponse{Status: "error", Message: "Unknown RPC routing key: " + d.RoutingKey}
			}

			responseBody, _ = json.Marshal(response)

			// Στείλε την απάντηση πίσω στην ουρά που ορίστηκε στο ReplyTo
			err = Channel.Publish(
				"",        // exchange (default)
				d.ReplyTo, // routing key (το όνομα της reply queue)
				false,     // mandatory
				false,     // immediate
				amqp.Publishing{
					ContentType:   "application/json",
					CorrelationId: d.CorrelationId,
					Body:          responseBody,
				})
			if err != nil {
				log.Printf("Failed to publish RPC reply: %s", err)
			}

			d.Ack(false) // Επιβεβαίωσε την παραλαβή του μηνύματος
		}
	}()
}
