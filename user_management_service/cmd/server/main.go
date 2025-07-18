package main

import (
	"log"
	"os"
	"user_management_service/internal/config"
	"user_management_service/internal/handler"
	"user_management_service/internal/messaging"
	"user_management_service/internal/middleware"

	"github.com/gin-gonic/gin"
)

func main() {
	// 1) Database setup
	db := config.SetupDatabase()
	sqlDB, _ := db.DB()
	defer sqlDB.Close()

	// 2) RabbitMQ init (messaging.Init() δηλώνει exchanges & queues/bindings)
	messaging.Init()
	defer messaging.Conn.Close()
	defer messaging.Channel.Close()

	// 3) Start consumer για auth requests
	//    (δεν χρειάζεται να δίνουμε πια το Channel, το έχει ήδη global)
	messaging.ConsumeAuthQueue(db)

	// 4) HTTP server
	r := gin.Default()
	r.POST("/register", handler.Register(db))
	r.POST("/login", handler.Login(db))
	r.POST("/upsert", handler.UpsertUser(db))

	auth := r.Group("/auth")
	auth.Use(middleware.JWTAuthMiddleware()) // sets user_id, email, role in context
	auth.GET("/validate", handler.Validate())
	auth.GET("/profile", handler.Profile(db))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8082"
	}

	log.Println("🟢 User-Management Service listening on port", port)
	r.Run(":" + port)
}
