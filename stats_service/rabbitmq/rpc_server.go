package rabbitmq

import (
	"encoding/json"          // Πακέτο για κωδικοποίηση και αποκωδικοποίηση JSON δεδομένων.
	"fmt"                    // Πακέτο για μορφοποιημένη είσοδο/έξοδο, όπως εκτύπωση στην κονσόλα.
	"log"                    // Πακέτο για logging.
	"stats_service/models"   // Εισάγει τα μοντέλα δεδομένων (Exam, Grade, GradeDistribution) από το project σου.
	"stats_service/services" // Εισάγει τα business logic services (HandlePersistAndCalculate, GetDistributions).

	amqp "github.com/rabbitmq/amqp091-go" // Η επίσημη βιβλιοθήκη για RabbitMQ στη Go.
	"gorm.io/gorm"                        // Η βιβλιοθήκη GORM για ORM λειτουργίες με τη βάση δεδομένων.
)

// Σταθερές για τα ονόματα της ουράς και του exchange, και των routing keys.
// Καλό είναι να ορίζονται ως σταθερές για εύκολη αλλαγή και αποφυγή τυπογραφικών.
const RPCQueueName = "stats_service_rpc_queue" // Όνομα της ουράς που θα ακούει αυτό το service για RPC αιτήματα.
const ExchangeName = "clearsky.rpc.exchange"   // Όνομα του exchange που θα χρησιμοποιηθεί για RPC.
// Προτείνεται ξεχωριστό exchange για RPC για καλύτερη οργάνωση.

// Routing keys για τις διαφορετικές RPC κλήσεις που μπορεί να δεχτεί το stats_service.
const RKPing = "stats.ping"                                // Για ένα απλό ping/pong test.
const RKPesistAndCalculate = "stats.persist_and_calculate" // Για αποθήκευση βαθμών και υπολογισμό κατανομών.
const RKGetDistributions = "stats.get_distributions"       // Για ανάκτηση των υπολογισμένων κατανομών.

// PersistDataPayload είναι η δομή του αναμενόμεนู payload για την ενέργεια persist_and_calculate.
type PersistDataPayload struct {
	Exam   models.Exam    `json:"exam"`   // Τα μεταδεδομένα της εξέτασης.
	Grades []models.Grade `json:"grades"` // Μια λίστα με τις βαθμολογίες των φοιτητών.
}

// GetDistributionsPayload είναι η δομή του αναμενόμενου payload για την ενέργεια get_distributions.
type GetDistributionsPayload struct {
	ClassID  string `json:"class_id"`  // Το ID του μαθήματος.
	ExamDate string `json:"exam_date"` // Η ημερομηνία της εξέτασης.
}

// RPCResponse είναι η γενική δομή για τις απαντήσεις των RPC κλήσεων.
type RPCResponse struct {
	Status  string      `json:"status"`            // "ok" ή "error" για την κατάσταση της απάντησης.
	Message string      `json:"message,omitempty"` // Προαιρετικό μήνυμα, π.χ., περιγραφή σφάλματος.
	Data    interface{} `json:"data,omitempty"`    // Τα πραγματικά δεδομένα της απάντησης (αν υπάρχουν).
}

// StartStatsRPCServer αρχικοποιεί και ξεκινά τον RPC server που ακούει για μηνύματα στο RabbitMQ.
// Παίρνει ως όρισμα μια σύνδεση με τη βάση δεδομένων (db *gorm.DB).
func StartStatsRPCServer(db *gorm.DB) {
	// Έλεγχος αν το καθολικό κανάλι Channel (από το connection.go) έχει αρχικοποιηθεί.
	if Channel == nil {
		log.Fatal("FATAL: RabbitMQ channel is not initialized. Call rabbitmq.Init() first.")
	}

	// Δήλωση του exchange.
	// Ένα exchange δρομολογεί μηνύματα στις ουρές με βάση το routing key και τον τύπο του exchange.
	err := Channel.ExchangeDeclare(
		ExchangeName, // Όνομα του exchange.
		"direct",     // Τύπος του exchange. 'direct' σημαίνει ότι το μήνυμα πάει σε ουρές που το routing key τους ταιριάζει ακριβώς.
		true,         // durable: το exchange θα επιβιώσει αν ο RabbitMQ server κάνει restart.
		false,        // autoDelete: το exchange δεν θα διαγραφεί όταν δεν υπάρχουν ουρές συνδεδεμένες.
		false,        // internal: δεν είναι internal, μπορεί να δεχτεί μηνύματα από publishers.
		false,        // noWait: ο client δεν θα περιμένει επιβεβαίωση από τον server.
		nil,          // arguments: επιπλέον arguments (συνήθως nil).
	)
	if err != nil {
		log.Fatalf("FATAL: Failed to declare exchange '%s': %v", ExchangeName, err)
	}

	// Δήλωση της ουράς (queue) όπου θα λαμβάνουμε τα RPC αιτήματα.
	q, err := Channel.QueueDeclare(
		RPCQueueName, // Όνομα της ουράς.
		true,         // durable: η ουρά θα επιβιώσει αν ο RabbitMQ server κάνει restart.
		false,        // autoDelete: η ουρά δεν θα διαγραφεί όταν δεν υπάρχουν consumers.
		false,        // exclusive: η ουρά δεν είναι exclusive (μπορεί να χρησιμοποιηθεί από πολλούς consumers).
		false,        // noWait: ο client δεν θα περιμένει επιβεβαίωση.
		nil,          // arguments: επιπλέον arguments.
	)
	if err != nil {
		log.Fatalf("FATAL: Failed to declare queue '%s': %v", q.Name, err)
	}

	// Σύνδεση (binding) της ουράς με το exchange για κάθε routing key που μας ενδιαφέρει.
	// Αυτό λέει στο exchange "όταν έρθει μήνυμα με αυτό το routing key, στείλ' το σε αυτή την ουρά".
	routingKeys := []string{RKPing, RKPesistAndCalculate, RKGetDistributions}
	for _, rk := range routingKeys {
		log.Printf("Binding queue %s to exchange %s with routing key %s", q.Name, ExchangeName, rk)
		err = Channel.QueueBind(
			q.Name,       // queueName
			rk,           // routingKey
			ExchangeName, // exchangeName
			false,        // noWait
			nil,          // arguments
		)
		if err != nil {
			log.Fatalf("FATAL: Failed to bind queue for key %s: %v", rk, err)
		}
	}

	// Ορισμός Quality of Service (QoS).
	// Αυτό λέει στον RabbitMQ server να μην στέλνει περισσότερα από 'prefetchCount' μηνύματα
	// σε έναν consumer μέχρι αυτός να κάνει acknowledge (Ack/Nack) τα προηγούμενα.
	// Βοηθά στην ομοιόμορφη κατανομή φορτίου μεταξύ των workers.
	err = Channel.Qos(
		1,     // prefetchCount: κάθε worker παίρνει 1 μήνυμα τη φορά.
		0,     // prefetchSize: 0 σημαίνει χωρίς όριο στο συνολικό μέγεθος των μηνυμάτων.
		false, // global: false σημαίνει ότι το QoS εφαρμόζεται ανά consumer (channel), όχι global για όλους τους consumers της ουράς.
	)
	if err != nil {
		log.Fatalf("FATAL: Failed to set QoS: %v", err)
	}

	// Έναρξη κατανάλωσης μηνυμάτων από την ουρά.
	// Αυτό επιστρέφει ένα Go channel (`<-chan amqp.Delivery`) από το οποίο διαβάζουμε τα μηνύματα.
	msgs, err := Channel.Consume(
		q.Name, // queue: το όνομα της ουράς από την οποία θα καταναλώσουμε.
		"",     // consumer: ένα client-generated consumer tag για να αναγνωρίζεται ο consumer. Αν κενό, ο server γεννά ένα.
		false,  // autoAck: false σημαίνει ότι πρέπει να κάνουμε χειροκίνητα Ack/Nack τα μηνύματα. Πολύ σημαντικό για αξιοπιστία.
		false,  // exclusive: false σημαίνει ότι και άλλοι consumers μπορούν να καταναλώνουν από αυτή την ουρά.
		false,  // noLocal: true αν ο consumer δεν πρέπει να λαμβάνει μηνύματα που δημοσίευσε ο ίδιος (δεν υποστηρίζεται από όλους τους brokers).
		false,  // noWait: ο client δεν θα περιμένει επιβεβαίωση.
		nil,    // args: επιπλέον arguments.
	)
	if err != nil {
		log.Fatalf("FATAL: Failed to register a consumer: %v", err)
	}

	log.Printf(" [*] StatsService RPC Server waiting for messages on queue '%s'. To exit press CTRL+C", q.Name)

	// Ξεκινάμε πολλαπλούς workers (goroutines) για να επεξεργάζονται μηνύματα παράλληλα.
	numWorkers := 2 // Μπορείς να το κάνεις παραμετροποιήσιμο.
	for i := 0; i < numWorkers; i++ {
		go func(workerID int) { // Κάθε worker τρέχει στη δική του goroutine.
			log.Printf("INFO: RPC Worker %d started", workerID)
			for d := range msgs { // Loop που περιμένει για μηνύματα από το RabbitMQ channel 'msgs'.
				log.Printf("INFO: Worker %d received RPC request with routing key '%s', CorrelationID: %s, ReplyTo: %s", workerID, d.RoutingKey, d.CorrelationId, d.ReplyTo)

				var response RPCResponse  // Η απάντηση που θα σταλεί πίσω.
				var processingError error // Για να παρακολουθούμε αν η επεξεργασία απέτυχε.

				// Διαχείριση του μηνύματος ανάλογα με το routing key του.
				switch d.RoutingKey {
				case RKPing:
					response = RPCResponse{Status: "ok", Message: "stats_service pong"}

				case RKPesistAndCalculate:
					var payload PersistDataPayload
					// Αποκωδικοποίηση του JSON payload του μηνύματος στο struct PersistDataPayload.
					if err := json.Unmarshal(d.Body, &payload); err != nil {
						response = RPCResponse{Status: "error", Message: "Invalid payload for persist_and_calculate: " + err.Error()}
						processingError = err
					} else {
						// Κλήση της κύριας λογικής του service.
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
								// Αν δεν βρεθούν κατανομές, στέλνουμε "ok" με μήνυμα και κενά δεδομένα.
								response = RPCResponse{Status: "ok", Message: "No distributions found for the given criteria.", Data: []models.GradeDistribution{}}
							} else {
								response = RPCResponse{Status: "ok", Data: distributions}
							}
						}
					}
				default:
					// Άγνωστο routing key.
					response = RPCResponse{Status: "error", Message: "Unknown RPC routing key: " + d.RoutingKey}
					processingError = fmt.Errorf("unknown RPC routing key: %s", d.RoutingKey)
				}

				// Κωδικοποίηση της απάντησης σε JSON.
				responseBody, err := json.Marshal(response)
				if err != nil {
					log.Printf("ERROR: Worker %d failed to marshal RPC response: %v", workerID, err)
					// Αν αποτύχει το marshal, δεν μπορούμε να στείλουμε απάντηση.
					// Κάνουμε Nack το μήνυμα για να μην χαθεί (αν έχει ρυθμιστεί DLQ) ή να ξαναπροσπαθήσει (αν το requeue είναι true).
					d.Nack(false, false) // false για requeue, ώστε να μην μπει σε ατέρμονο βρόχο αν το μήνυμα είναι προβληματικό.
					continue             // Προχωράμε στο επόμενο μήνυμα.
				}

				// Έλεγχος αν ο αποστολέας περιμένει απάντηση (αν έχει ορίσει ReplyTo ουρά).
				if d.ReplyTo != "" {
					// Δημοσίευση της απάντησης στην ουρά ReplyTo που όρισε ο client.
					// Χρησιμοποιούμε το default exchange ("") που στέλνει απευθείας στην ουρά που ονομάζεται στο routing key (εδώ το d.ReplyTo).
					err = Channel.Publish(
						"",        // exchange: default exchange
						d.ReplyTo, // routing key: το όνομα της ουράς απάντησης
						false,     // mandatory: αν true και δεν μπορεί να δρομολογηθεί, επιστρέφεται στον publisher.
						false,     // immediate: αν true και δεν υπάρχει consumer έτοιμος, επιστρέφεται. (deprecated)
						amqp.Publishing{
							ContentType:   "application/json",
							CorrelationId: d.CorrelationId, // Σημαντικό για να αντιστοιχίσει ο client την απάντηση με το αίτημα.
							Body:          responseBody,
						})
					if err != nil {
						log.Printf("ERROR: Worker %d failed to publish RPC reply to %s: %v", workerID, d.ReplyTo, err)
						// Εδώ, η επεξεργασία μπορεί να έχει γίνει, αλλά η απάντηση απέτυχε.
						// Η απόφαση για Ack/Nack εξαρτάται από την πολιτική σου.
						// Αν η επεξεργασία έγινε και είναι μη-αντιστρέψιμη, ίσως Ack.
						// Αν η αποτυχία αποστολής απάντησης σημαίνει αποτυχία της όλης RPC, τότε Nack.
						if processingError == nil { // Αν η επεξεργασία ήταν ΟΚ, αλλά η απάντηση απέτυχε
							d.Ack(false) // Κάνουμε Ack γιατί η δουλειά έγινε, απλά ο client δεν πήρε απάντηση.
						} else {
							d.Nack(false, false) // Αν και η επεξεργασία είχε σφάλμα, Nack.
						}
						continue // Προχωράμε στο επόμενο μήνυμα.
					}
					log.Printf("INFO: Worker %d sent RPC reply to %s for CorrelationID: %s", workerID, d.ReplyTo, d.CorrelationId)
				} else {
					// Αν δεν υπάρχει ReplyTo, απλά καταγράφουμε (ίσως είναι fire-and-forget μήνυμα).
					log.Printf("WARNING: Worker %d received message without ReplyTo field. RoutingKey: %s", workerID, d.RoutingKey)
				}

				// Επιβεβαίωση (Acknowledgement) στο RabbitMQ.
				if processingError != nil {
					// Αν υπήρξε σφάλμα στην επεξεργασία, στέλνουμε Nack.
					// Το `false` στο δεύτερο όρισμα σημαίνει "μην το ξαναβάλεις στην ουρά (requeue)".
					// Ιδανικά, θα πήγαινε σε μια Dead Letter Queue (DLQ) αν έχει ρυθμιστεί.
					log.Printf("ERROR: Worker %d Nacking message due to processing error: %v", workerID, processingError)
					d.Nack(false, false)
				} else {
					// Αν όλα πήγαν καλά, στέλνουμε Ack.
					log.Printf("INFO: Worker %d Acking message for CorrelationID: %s", workerID, d.CorrelationId)
					d.Ack(false)
				}
			}
			log.Printf("INFO: RPC Worker %d stopped.", workerID)
		}(i) // Περνάμε το i για να έχει κάθε goroutine το δικό της μοναδικό workerID (για logging κυρίως).
	}
}
