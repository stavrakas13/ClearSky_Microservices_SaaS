package main

import (
	"log"
	"orchestrator/internal/config"
	"orchestrator/internal/rabbitmq"
)

func main() {
	// Load config (auto via init)
	log.Println("Starting Orchestrator...")

	// Connect to RabbitMQ
	conn, ch := rabbitmq.Connect()
	defer conn.Close()
	defer ch.Close()

	// Setup exchanges, queues, bindings
	if err := rabbitmq.SetupMessaging(ch); err != nil {
		log.Fatalf("SetupMessaging failed: %v", err)
	}

	// Start consuming
	if err := rabbitmq.StartOrchestratorConsumer(ch); err != nil {
		log.Fatalf("Consumer failed: %v", err)
	}

	log.Printf("Orchestrator listening on exchange '%s', queue '%s'...", config.Cfg.Exchange.Name, config.Cfg.Queue.Name)
	select {} // block forever
}
