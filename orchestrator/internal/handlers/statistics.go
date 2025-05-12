package handlers

import (
	amqp "github.com/rabbitmq/amqp091-go"
)

// handleStatisticsViewed processes statistics.viewed events
func handleStatisticsViewed(d amqp.Delivery) {}

// handleGradesViewed processes grades.viewed events
func handleGradesViewed(d amqp.Delivery) {}
