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

	"github.com/joho/godotenv"
)

func main() {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	// Validate required environment variables
	requiredEnvVars := []string{
		"GOOGLE_CLIENT_ID",
		"GOOGLE_CLIENT_SECRET",
		"GOOGLE_REDIRECT_URL",
		"JWT_SECRET",
	}

	for _, envVar := range requiredEnvVars {
		if os.Getenv(envVar) == "" {
			log.Fatalf("Required environment variable %s is not set", envVar)
		}
	}

	// Log loaded configuration (without secrets)
	log.Printf("Google Client ID: %s", os.Getenv("GOOGLE_CLIENT_ID"))
	log.Printf("Redirect URL: %s", os.Getenv("GOOGLE_REDIRECT_URL"))

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
	fmt.Printf("Google OAuth redirect URL: %s\n", os.Getenv("GOOGLE_REDIRECT_URL"))
	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
