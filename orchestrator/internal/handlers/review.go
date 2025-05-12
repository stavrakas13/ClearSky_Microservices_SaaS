package handlers

import (
	amqp "github.com/rabbitmq/amqp091-go"
)

// handleReviewRequested processes grades.review.requested events
func handleReviewRequested(d amqp.Delivery) {}

// handleReviewResponded processes grades.review.responded events
func handleReviewResponded(d amqp.Delivery) {}
