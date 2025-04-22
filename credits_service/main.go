// main.go
package main

import (
	"credits_service/dbService"
	"credits_service/routers"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found; relying on system environment variables.")
	}

	dbService.InitDB()

	router := gin.Default()

	routers.CreditRoutes(router)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	if err := router.Run(":" + port); err != nil {
		log.Fatal("Failed to run server: ", err)
	}
}
