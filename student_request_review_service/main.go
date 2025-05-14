// student_request_review_service
package main

import (
	"fmt"
	"student_request_review_service/db"
	"student_request_review_service/mq"

	_ "github.com/lib/pq"
)

func main() {

	mq.InitRabbitMQ()
	defer mq.Mqconn.Close()
	defer mq.Mqch.Close()

	// Start consuming messages from orchestrator
	mq.StartConsumer()

	db.InitDB()
	defer db.CloseDB()

	/*
		 	router := gin.Default()

			routes.PostNewReviewRequest(router)
			routes.GetAvailCourcesList(router)
			routes.GetReviewStatus(router)

			router.Run(":8087")
	*/
	fmt.Println("Service started and waiting for RabbitMQ messages...")

	select {}
}
