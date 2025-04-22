package dbService

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

var Pool *pgxpool.Pool

func InitDB() {
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	dbname := os.Getenv("DB_NAME")

	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		user, password, host, port, dbname)

	var err error
	Pool, err = pgxpool.New(context.Background(), dsn)
	if err != nil {
		log.Fatal("Unable to connect to database:", err)
	}

	if err = Pool.Ping(context.Background()); err != nil {
		log.Fatal("Unable to ping the database:", err)
	}

	log.Println("Connected to PostgreSQL via pgxpool.")
}

func diminish(inst_name, credits) (bool, error) {
	ctx := context.Background()

	checkQuery := `SELECT credits FROM credits_inst WHERE name = $1`
	var current_credits int
	err := Pool.QueryRow(ctx, checkQuery, inst_name).Scan(&current_credits)

	if err == nil && current_credits == 0 {
		return false, fmt.Errorf("Not enough credits for this operation")
	}
	elseif err != nil {
		log.Printf("Failed to diminish credits: %v", err)
		return false, err
	}

	insertQuery := `UPDATE credits_inst SET credits = credits - 1 WHERE name = $1`;

	_, err = Pool.Exec(ctx, insertQuery, inst_name)

	if err != nil {
		log.Printf("Failed to make the diminsh: %v", err)
		return 0, err
	}

	return true, nil

}
