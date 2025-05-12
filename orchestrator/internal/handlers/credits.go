package handlers

import (
	amqp "github.com/rabbitmq/amqp091-go"
)

// handleCreditsAvail processes credits.avail events
func handleCreditsAvail(d amqp.Delivery) {}

// handleCreditsSpent processes credits.spent events
func handleCreditsSpent(d amqp.Delivery) {}

// handleCreditsPurchased processes credits.purchased events
func handleCreditsPurchased(d amqp.Delivery) {}
