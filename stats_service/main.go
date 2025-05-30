package main

import (
	"log"
	"stats_service/db"
	"time"

	//"stats_service/routes"
	"stats_service/rabbitmq" //τρεχω ενα go mod tidy για να το προσθεσει

	"github.com/gin-gonic/gin"
)

func main() {
	database, err := db.InitDB()
	if err != nil {
		log.Fatalf("❌ Failed to connect to DB: %v", err)
	}

	for i := 0; i < 15; i++ {
		err = rabbitmq.Init()
		if err == nil {
			break
		}
		log.Printf("Waiting for RabbitMQ... (%d/15): %v\n", i+1, err)
		time.Sleep(3 * time.Second)
		if i == 14 {
			log.Fatalf("❌ Could not connect to RabbitMQ after 15 tries: %v", err)
		}
	}
	defer rabbitmq.Close()

	rabbitmq.StartStatsRPCServer(database)

	r := gin.Default()
	//routes.SetupRoutes(r, database)*/

	log.Println("🚀 Server running on http://localhost:8080")
	r.Run(":8080")
}
