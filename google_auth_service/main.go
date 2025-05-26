package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"google_auth_service/database"
	"google_auth_service/handlers"
	"google_auth_service/middlewares"
	"google_auth_service/rabbitmq"
)

func main() {
	database.ConnectDatabase()
	rabbitmq.Connect()
	rabbitmq.StartGoogleAuthConsumer()
	port := "8086"
	if os.Getenv("PORT") != "" {
		port = os.Getenv("PORT")
	}

	http.HandleFunc("/auth/google/login", handlers.GoogleLoginHandler)
	http.HandleFunc("/auth/google/callback", handlers.GoogleCallbackHandler)
	http.HandleFunc("/auth/logout", handlers.LogoutHandler)
	protected := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Welcome to the protected route!"))
	})

	http.Handle("/protected", middlewares.AuthMiddleware(protected))

	fmt.Printf("Starting google_auth_service on port %s...\n", port)
	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
