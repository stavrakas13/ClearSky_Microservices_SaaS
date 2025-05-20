package routes

import (
	"orchestrator/internal/handlers"

	"github.com/gin-gonic/gin"
	amqp "github.com/rabbitmq/amqp091-go"
)

func SetupRouter(ch *amqp.Channel) *gin.Engine {
	r := gin.Default()

	// PATCH /purchase → credit purchase RPC
	r.PATCH("/purchase", func(c *gin.Context) {
		handlers.HandleCreditsPurchased(c, ch)
	})

	r.GET("/mycredits", func(c *gin.Context) {
		handlers.HandleCreditsAvail(c, ch)
	})

	r.PATCH("/spending", func(c *gin.Context) {
		handlers.HandleCreditsSpent(c, ch)
	})

	// Student and Instructor API calls.

	r.PATCH("/student/reviewRequest", func(c *gin.Context) {
		// Calls HandlePostNewRequest:
		//   Expects JSON body {course_id, user_id, student_message, exam_period}
		//   Returns JSON: {"data": map[string]interface{}} where data is the service response
		handlers.HandlePostNewRequest(c, ch)
	})
	r.PATCH("/student/status", func(c *gin.Context) {
		// Calls HandleGetRequestStatus:
		//   Expects JSON body {course_id, user_id, exam_period}
		//   Returns JSON: {"data": map[string]interface{}} containing status details
		handlers.HandleGetRequestStatus(c, ch)
	})
	r.PATCH("/instructor/review-list", func(c *gin.Context) {
		// Calls HandleGetRequestList:
		//   Expects JSON body {course_id, exam_period}
		//   Returns JSON: {"data": map[string]interface{}} listing pending reviews
		handlers.HandleGetRequestList(c, ch)
	})
	r.PATCH("/instructor/reply", func(c *gin.Context) {
		// Calls HandlePostResponse:
		//   Expects JSON body {course_id, user_id, exam_period, instructor_reply_message, instructor_action}
		//   Returns JSON: {"data": map[string]interface{}} acknowledgments from services
		handlers.HandlePostResponse(c, ch)
	})
	return r
}
