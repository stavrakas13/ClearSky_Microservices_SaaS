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

	// 1. Start transaction
	tx, err := Pool.Begin(ctx)
	if err != nil {
		log.Printf("[Diminish] Failed to begin transaction: %v", err)
		return false, err
	}
	defer tx.Rollback(ctx) // automatic rollback on error

	// 2. Lock and read current credits
	checkQuery := `SELECT credits FROM credits_inst WHERE name = $1 FOR UPDATE`
	var current_credits int
	err = tx.QueryRow(ctx, checkQuery, inst_name).Scan(&current_credits)
	if err != nil {
		if err == pgx.ErrNoRows {
			log.Printf("[Diminish] Institution not found: %s", inst_name)
			return false, fmt.Errorf("institution '%s' not found", inst_name)
		}
		log.Printf("[Diminish] Failed to check credits: %v", err)
		return false, err
	}
	log.Printf("[Diminish] Current credits for %s: %d", inst_name, current_credits)

	// 3. Check for sufficient credits
	if current_credits < credits {
		log.Printf("[Diminish] Not enough credits: has %d, needs %d", current_credits, credits)
		return false, fmt.Errorf("insufficient credits (current: %d)", current_credits)
	}

	// 4. Perform update
	updateQuery := `UPDATE credits_inst SET credits = credits - $1 WHERE name = $2`
	res, err := tx.Exec(ctx, updateQuery, credits, inst_name)
	if err != nil {
		log.Printf("[Diminish] Failed to decrement credits: %v", err)
		return false, err
	}

	rows := res.RowsAffected()
	log.Printf("[Diminish] Rows affected by update: %d", rows)
	if rows != 1 {
		return false, fmt.Errorf("unexpected number of rows affected: %d", rows)
	}

	// 5. Commit
	if err := tx.Commit(ctx); err != nil {
		log.Printf("[Diminish] Failed to commit transaction: %v", err)
		return false, err
	}

	log.Printf("[Diminish] Credits decremented successfully for %s", inst_name)
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

func AvailableCredits(instName string) (int, error) {
	ctx := context.Background()

	const qry = `
        SELECT credits
        FROM credits_inst
        WHERE name = $1
    `
	var current int
	err := Pool.QueryRow(ctx, qry, instName).Scan(&current)
	if err != nil {
		if err == pgx.ErrNoRows {
			return 0, fmt.Errorf("institution %q not found", instName)
		}
		log.Printf("Failed to query available credits for %q: %v", instName, err)
		return 0, err
	}

	return current, nil
}

func NewInstitution(instName string, initialCredits int) (bool, error) {
	ctx := context.Background()

	tx, err := Pool.Begin(ctx)
	if err != nil {
		log.Printf("Failed to begin transaction: %v", err)
		return false, err
	}
	defer tx.Rollback(ctx) // Ensures rollback on failure

	const checkQuery = `SELECT 1 FROM credits_inst WHERE name = $1`
	var exists int
	err = tx.QueryRow(ctx, checkQuery, instName).Scan(&exists)
	if err == nil {
		log.Printf("Error checking institution existence: %v", err)
		return false, fmt.Errorf("institution %q already exists", instName)
	}

	const insertQuery = `INSERT INTO credits_inst (name, credits) VALUES ($1, $2)`
	_, err = tx.Exec(ctx, insertQuery, instName, initialCredits)
	if err != nil {
		log.Printf("Failed to insert new institution: %v", err)
		return false, err
	}

	if err := tx.Commit(ctx); err != nil {
		log.Printf("Failed to commit transaction: %v", err)
		return false, err
	}

	return true, nil
}
