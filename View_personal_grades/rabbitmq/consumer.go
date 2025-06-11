package rabbitmq

import (
	"encoding/json"
	"log"

	"View_personal_grades/services"

	amqp "github.com/rabbitmq/amqp091-go"
	"gorm.io/gorm"
)

const (
	ExchangeName = "clearSky.events"
	QueueName    = "personal_grades_queue"
	RKGetCourses = "personal.get_courses"
	RKGetGrades  = "personal.get_grades"
)

// StartConsumer sets up the queue and spawns workers to handle messages.
func StartConsumer(db *gorm.DB) {
	if Channel == nil {
		log.Fatal("RabbitMQ channel not initialized")
	}

	if err := Channel.ExchangeDeclare(ExchangeName, "direct", true, false, false, false, nil); err != nil {
		log.Fatalf("Failed to declare exchange %s: %v", ExchangeName, err)
	}

	q, err := Channel.QueueDeclare(QueueName, true, false, false, false, nil)
	if err != nil {
		log.Fatalf("Failed to declare queue %s: %v", QueueName, err)
	}

	for _, rk := range []string{RKGetCourses, RKGetGrades} {
		if err := Channel.QueueBind(q.Name, rk, ExchangeName, false, nil); err != nil {
			log.Fatalf("Failed to bind queue %s with key %s: %v", q.Name, rk, err)
		}
	}

	if err := Channel.Qos(1, 0, false); err != nil {
		log.Fatalf("Failed to set QoS: %v", err)
	}

	msgs, err := Channel.Consume(q.Name, "personal_grades_consumer", false, false, false, false, nil)
	if err != nil {
		log.Fatalf("Failed to register consumer: %v", err)
	}

	const workers = 2
	log.Printf("[*] PersonalGradesService waiting for messages on %s", q.Name)
	for i := 0; i < workers; i++ {
		go func(id int) {
			log.Printf("personal grades worker %d ready", id)
			for d := range msgs {
				handleMessage(db, d)
			}
		}(i)
	}
}

// handleMessage dispatches incoming messages based on their routing key and
// publishes a response if the sender provided a reply queue.
func handleMessage(db *gorm.DB, d amqp.Delivery) {
	switch d.RoutingKey {
	case RKGetCourses:
		var p struct {
			StudentID string `json:"student_id"`
		}
		if err := json.Unmarshal(d.Body, &p); err != nil {
			log.Printf("invalid get_courses payload: %v", err)
			d.Nack(false, false)
			return
		}
		courses, err := services.GetStudentCoursesWithStatus(db, p.StudentID)
		reply(d, courses, err)
	case RKGetGrades:
		var p struct {
			ClassID   string `json:"class_id"`
			ExamDate  string `json:"exam_date"`
			StudentID string `json:"student_id"`
		}
		if err := json.Unmarshal(d.Body, &p); err != nil {
			log.Printf("invalid get_grades payload: %v", err)
			d.Nack(false, false)
			return
		}
		grades, err := services.GetStudentPersonalGrades(db, p.ClassID, p.ExamDate, p.StudentID)
		reply(d, grades, err)
	default:
		log.Printf("unknown routing key %s", d.RoutingKey)
		d.Nack(false, false)
	}
}

// reply sends the encoded response back to the ReplyTo queue using the same
// correlation ID so the caller can match it to the request.
func reply(d amqp.Delivery, data interface{}, err error) {
	if d.ReplyTo == "" {
		if err == nil {
			d.Ack(false)
		} else {
			d.Nack(false, false)
		}
		return
	}

	var body []byte
	if err != nil {
		body, _ = json.Marshal(map[string]string{"error": err.Error()})
	} else {
		body, _ = json.Marshal(data)
	}

	if err := Channel.Publish("", d.ReplyTo, false, false, amqp.Publishing{
		ContentType:   "application/json",
		CorrelationId: d.CorrelationId,
		Body:          body,
	}); err != nil {
		log.Printf("failed to publish reply: %v", err)
	}

	if err != nil {
		d.Nack(false, false)
	} else {
		d.Ack(false)
	}
}
