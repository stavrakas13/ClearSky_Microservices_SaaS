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
		handlers.HandlePostNewRequest(c, ch)
	})
	r.PATCH("/student/status", func(c *gin.Context) {
		handlers.HandleGetRequestStatus(c, ch)
	})
	r.PATCH("/instructor/review-list", func(c *gin.Context) {
		handlers.HandleGetRequestList(c, ch)
	})
	r.PATCH("/instructor/reply", func(c *gin.Context) {
		handlers.HandlePostResponse(c, ch)
	})
	return r
}
