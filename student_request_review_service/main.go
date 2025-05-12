// student_request_review_service
package main

import (
	"student_request_review_service/db"
	"student_request_review_service/routes"

	"os"
	"student_request_review_service/mq"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

func main() {

	os.Setenv("RABBITMQ_URL", "amqp://guest:guest@rabbitmq:5672/")

	mq.InitRabbitMQ()
	defer mq.MQChannel.Close()

	db.InitDB()
	defer db.CloseDB()

	router := gin.Default()

	routes.PostNewReviewRequest(router)
	routes.GetAvailCourcesList(router)
	routes.GetReviewStatus(router)

	router.Run(":8087")
}
