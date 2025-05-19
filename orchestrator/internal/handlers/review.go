package handlers

import (
	amqp "github.com/rabbitmq/amqp091-go"
)

// handleReviewRequested processes new request events
// -> sends 2 events: student.postNewRequest & instructor.insertStudentRequest
func handlePostNewRequest(d amqp.Delivery) {}

// handleGetRequestStatus processes student sees request status events
// -> sends 1 event: student.getRequestStatus
func handleGetRequestStatus(d amqp.Delivery) {}

// handlePostResponse processes responses on review requests
// -> sends 2 events: student.updateInstructorResponse & instructor.postResponse
func handlePostResponse(d amqp.Delivery) {}

// handleGetRequestList processes instructor get list of pending requests
// -> sends 1 event: instructor.getRequestsList
func handleGetRequestList(d amqp.Delivery) {}

// handleGetRequestInfo processes instructor sees request details
// -> sends 1 event: instructor.getRequestInfo
func handleGetRequestInfo(d amqp.Delivery) {}
