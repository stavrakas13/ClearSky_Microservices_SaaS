package main

import (
	"log"
	"stats_service/db"
	"stats_service/routes"

	"github.com/gin-gonic/gin"
)

func main() {
	database, err := db.InitDB()
	if err != nil {
		log.Fatalf("❌ Failed to connect to DB: %v", err)
	}

	r := gin.Default()
	routes.SetupRoutes(r, database)

	log.Println("🚀 Server running on http://localhost:8080")
	r.Run(":8080")
}
