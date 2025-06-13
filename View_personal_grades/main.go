package main

// Entry point for the personal grades microservice. It initialises the
// database and RabbitMQ connection then starts a consumer that listens for
// requests coming from the orchestrator.

import (
	"log"
	"time"

	"View_personal_grades/db"
	"View_personal_grades/rabbitmq"


	"github.com/joho/godotenv"
)

func main() {
        // Load environment variables when running locally. In Docker the
        // variables are already provided so missing .env is not an error.
        if err := godotenv.Load(); err != nil {
                log.Println("INFO: No .env file found, using environment variables.")
        }

        database, err := db.InitDB()
        if err != nil {
                log.Fatalf("❌ Failed to connect to DB: %v", err)
        }
	sqlDB, _ := database.DB()
	defer sqlDB.Close()

        // RabbitMQ might not be up yet when the container starts. Retry the
        // connection a few times before giving up so orchestration tools have
        // time to start the broker container.
        maxRetries := 15
        retryDelay := 3 * time.Second
        for i := 0; i < maxRetries; i++ {
                err = rabbitmq.Init()
                if err == nil {
                        log.Println("✅ Successfully connected to RabbitMQ.")

			break
		}
		log.Printf("WARNING: Failed to connect to RabbitMQ (attempt %d/%d): %v. Retrying in %v...", i+1, maxRetries, err, retryDelay)
		time.Sleep(retryDelay)
		if i == maxRetries-1 {
			log.Fatalf("❌ Could not connect to RabbitMQ after %d tries: %v", maxRetries, err)
		}
        }
        defer rabbitmq.Close()

        // Begin consuming messages. This call blocks until the program exits,
        // spawning goroutines to handle incoming requests.
        rabbitmq.StartConsumer(database)


	log.Println("✅ Personal grades consumer is running…")
	select {}
}
