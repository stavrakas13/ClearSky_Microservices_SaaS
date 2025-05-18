// student_request_review_service
package main

import (
	"fmt"
	"student_request_review_service/db"
	"student_request_review_service/mq"
	"time"

	_ "github.com/lib/pq"
)

func main() {

	for i := 0; i < 15; i++ {
		err := mq.InitRabbitMQ()
		if err == nil {
			break
		}
		fmt.Printf("Waiting for RabbitMQ... (%d/15)\n", i+1)
		time.Sleep(3 * time.Second)
		if i == 14 {
			panic("Could not connect to RabbitMQ after 15 tries")
		}
	}
	defer mq.Mqconn.Close()
	defer mq.Mqch.Close()

	for i := 0; i < 5; i++ {
		err := db.InitDB()
		if err == nil {
			break
		}
		fmt.Printf("Waiting for DB... (%d/5)\n", i+1)
		time.Sleep(3 * time.Second)
		if i == 4 {
			panic("Could not connect to DB after 5 tries")
		}
	}
	defer db.CloseDB()

	mq.StartConsumer()

	fmt.Println("Student Service started and waiting for RabbitMQ messages...")
	select {}
}
