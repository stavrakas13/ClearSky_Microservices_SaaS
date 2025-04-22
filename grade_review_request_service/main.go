package main

import "github.com/gin-gonic/gin"

func main() {
	server := gin.Default()

	// GET ENDPOINT
	server.GET("/review-request", func(ctx *gin.Context) {
		ctx.String(200, "hello from grade_review_request_service")
	})

	server.Run(":8081")
}
