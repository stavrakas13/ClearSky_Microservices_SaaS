package db

import (
	"database/sql"
	"log"

	_ "github.com/lib/pq"
)

var DB *sql.DB

func InitDB() {

	// CONNECT TO REVIEWS DB
	// URL for docker connection
	reviewsdbURL := "postgres://postgres:root@db:5432/reviewsdb?sslmode=disable"
	// URL for local connection
	// reviewsdbURL := "postgres://postgres:root@localhost:5432/reviews?sslmode=disable"

	var err error
	DB, err = sql.Open("postgres", reviewsdbURL)
	if err != nil {
		log.Fatal("Failed to connect to DB:", err)
	}

	log.Println("Connected to reviewsdbinst.")
}

func CloseDB() {
	DB.Close()
}
