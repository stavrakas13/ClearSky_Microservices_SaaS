package db

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

var DB *sql.DB

func InitDB() error {

	// CONNECT TO REVIEWS DB
	// URL for docker connection
	//reviewsdbURL := "postgres://postgres:root@db:5432/reviewsdbinst?sslmode=disable"
	reviewsdbURL := "postgres://postgres:root@instructor_db:5432/reviewsdbinst?sslmode=disable"

	// URL for local connection
	//reviewsdbURL := "postgres://postgres:root@localhost:5432/reviewsinst?sslmode=disable"

	var err error
	DB, err = sql.Open("postgres", reviewsdbURL)
	if err != nil {
		fmt.Println("DB open error:", err)
		return err
	}
	err = DB.Ping()
	if err != nil {
		fmt.Println("DB ping error:", err)
		return err
	}
	fmt.Println("DB connection established.")
	return nil
}

func CloseDB() {
	DB.Close()
}
