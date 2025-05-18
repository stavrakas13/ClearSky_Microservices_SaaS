// instructor_review_reply_service
package main

import (
	"fmt"
	"instructor_review_reply_service/db"
	"instructor_review_reply_service/mq"

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

		routes.PostReply(router)
		routes.GetReviewReqeustList(router)
		routes.GetRequestInfo(router)

		router.Run(":8088")
	*/
	fmt.Println("Service started and waiting for RabbitMQ messages...")
	select {}

}
