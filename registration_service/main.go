// main.go
package main

import (
	"log"
	"os"
	"registration_service/dbService"
	"registration_service/routers"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables from .env file if available
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found; relying on system environment variables.")
	}

	dbService.InitDB()

	router := gin.Default()

	routers.RegisterRoutes(router)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	if err := router.Run(":" + port); err != nil {
		log.Fatal("Failed to run server: ", err)
	}
}
