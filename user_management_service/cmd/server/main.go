package main

import (
	"log"
	"os"

	"user_management_service/internal/config"

	"github.com/gin-gonic/gin"
)

func main() {
	db := config.SetupDatabase()
	sqlDB, _ := db.DB()
	defer sqlDB.Close()

	r := gin.Default()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Println("Server is running on port", port)
	r.Run(":" + port)
}
