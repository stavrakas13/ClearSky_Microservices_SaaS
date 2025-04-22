package dbService

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx"
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

func Diminish(inst_name string, credits int) (bool, error) {
	ctx := context.Background()

	tx, err := Pool.Begin(ctx)
	if err != nil {
		log.Printf("Failed to begin transaction: %v", err)
		return false, err
	}
	defer tx.Rollback(ctx) // Safely rollback if anything fails

	checkQuery := `SELECT credits FROM credits_inst WHERE name = $1 FOR UPDATE` //LOCK TO AVOID RACE CONDITION IN DB
	var current_credits int
	err = tx.QueryRow(ctx, checkQuery, inst_name).Scan(&current_credits)

	if err != nil {
		if err == pgx.ErrNoRows {
			return false, fmt.Errorf("institution '%s' not found", inst_name)
		}
		log.Printf("Failed to check credits: %v", err)
		return false, err
	}

	if current_credits < credits {
		return false, fmt.Errorf("insufficient credits (current: %d)", current_credits)
	}

	insertQuery := `UPDATE credits_inst SET credits = credits - $1 WHERE name = $2`

	_, err = tx.Exec(ctx, insertQuery, credits, inst_name)

	if err != nil {
		log.Printf("Failed to make the diminsh: %v", err)
		return false, err
	}

	//Commit the transaction
	if err := tx.Commit(ctx); err != nil {
		log.Printf("Failed to commit transaction: %v", err)
		return false, err
	}

	return true, nil

}

func BuyCredits(inst_name string, credits int) (bool, error) {
	ctx := context.Background()

	tx, err := Pool.Begin(ctx)
	if err != nil {
		log.Printf("Failed to begin transaction: %v", err)
		return false, err
	}
	defer tx.Rollback(ctx) // Safely rollback if anything fails

	insertQuery := `UPDATE credits_inst SET credits = credits + $1 WHERE name = $2`

	_, err = tx.Exec(ctx, insertQuery, credits, inst_name)

	if err != nil {
		log.Printf("Failed to make the purchase: %v", err)
		return false, err
	}

	if err := tx.Commit(ctx); err != nil {
		log.Printf("Failed to commit transaction: %v", err)
		return false, err
	}

	return true, nil

}
