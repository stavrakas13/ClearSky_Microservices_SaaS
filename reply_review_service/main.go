package main

import "github.com/gin-gonic/gin"

func main() {
	server := gin.Default()

	// GET ENDPOINT
	server.GET("/reply-review", func(ctx *gin.Context) {
		ctx.String(200, "hello from reply_review_service")
	})

	server.Run(":8080")
}
