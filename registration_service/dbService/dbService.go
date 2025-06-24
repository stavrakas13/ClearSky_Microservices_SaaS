package dbService

import (
	"context"
	"database/sql"
	"errors"
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

func AddInstitution(inst_name, email, director string) (int, error) {
	log.Printf("→ AddInstitution called with name=%q, email=%q, director=%q", inst_name, email, director)
	ctx := context.Background()

	// 1. Check for existing institution
	log.Println("… Checking if institution already exists")
	checkQuery := `SELECT name FROM institution WHERE name = $1;`
	var existing string
	err := Pool.QueryRow(ctx, checkQuery, inst_name).Scan(&existing)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		// anything except “no rows” is fatal
		log.Printf("❌ Error during existence check: %v", err)
		return 0, err
	}
	if err == nil {
		// found a match, don’t insert
		log.Printf("⚠ Institution %q already exists", inst_name)
		return 2, fmt.Errorf("institution already exists")
	}
	log.Println("✅ No existing institution found, proceeding to insert")

	// 2. Insert new institution
	insertQuery := `INSERT INTO institution (name, email, director) VALUES ($1, $2, $3);`
	log.Printf("… Inserting institution %q into database", inst_name)
	_, err = Pool.Exec(ctx, insertQuery, inst_name, email, director)
	if err != nil {
		log.Printf("❌ Failed to insert institution %q: %v", inst_name, err)
		return 0, err
	}
	log.Printf("✅ Institution %q inserted successfully", inst_name)

	return 1, nil
}
