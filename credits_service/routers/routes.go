// routers/routes.go
package routers

import (
	"credits_service/controllers"

	"github.com/gin-gonic/gin"
)

func CreditRoutes(router *gin.Engine) {
	api := router.Group("/")

	api.PUT("/spend_credit", controllers.SpendController)

	router.GET("/", func(c *gin.Context) {
		c.String(200, "Simple root-route")
	})

	router.NoRoute(func(c *gin.Context) {
		c.JSON(404, gin.H{"error": "Not found"})
	})
}
