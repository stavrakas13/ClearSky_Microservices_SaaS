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

	return r
}
