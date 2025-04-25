package main

import (
	"log"
	"os"

	"user_management_service/internal/config"
	"user_management_service/internal/middleware"

	"user_management_service/internal/handler"

	"github.com/gin-gonic/gin"
)

func main() {
	db := config.SetupDatabase()
	sqlDB, _ := db.DB()
	defer sqlDB.Close()

	r := gin.Default()
	r.POST("/register", handler.Register(db))
	r.POST("/login", handler.Login(db))

	auth := r.Group("/auth")
	auth.Use(middleware.JWTAuthMiddleware())
	auth.GET("/validate", handler.Validate())

	port := os.Getenv("PORT")
	if port == "" {
		port = "8082"
	}

	log.Println("Server is running on port", port)
	r.Run(":" + port)
}
