package main

import (
	"log"
	"time"

	"View_personal_grades/db"
	"View_personal_grades/rabbitmq"

	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("INFO: No .env file found, using environment variables.")
	}

	database, err := db.InitDB()
	if err != nil {
		log.Fatalf("❌ Failed to connect to DB: %v", err)
	}
	sqlDB, _ := database.DB()
	defer sqlDB.Close()

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

	rabbitmq.StartConsumer(database)

	log.Println("✅ Personal grades consumer is running…")
	select {}
}
