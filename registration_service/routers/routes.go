// routers/routes.go
package routers

import (
	"registration_service/controllers"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(router *gin.Engine) {
	api := router.Group("/")

	api.PATCH("/register", controllers.RegisterController)

	router.GET("/", func(c *gin.Context) {
		c.String(200, "Simple root-route")
	})

	router.NoRoute(func(c *gin.Context) {
		c.JSON(404, gin.H{"error": "Not found"})
	})
}
