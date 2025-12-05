//go:build ignore

package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func main() {
	// Load environment variables
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// Connect to database
	user := os.Getenv("PG_USER")
	password := os.Getenv("PG_PASSWORD")
	host := os.Getenv("PG_HOST")
	database := os.Getenv("PG_DATABASE")
	port := os.Getenv("PG_PORT")

	connectionString := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable&client_encoding=UTF8", user, password, host, port, database)
	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	log.Println("✅ Connected to database")

	// Add suspended column to sr table
	log.Println("📦 Adding suspended column to sr table...")
	_, err = db.Exec(`ALTER TABLE sr ADD COLUMN IF NOT EXISTS suspended BOOLEAN DEFAULT FALSE`)
	if err != nil {
		log.Fatalf("Failed to add suspended column: %v", err)
	}
	log.Println("✅ Suspended column added to sr table")

	// Verify the column was added
	var columnExists bool
	err = db.QueryRow(`
		SELECT EXISTS (
			SELECT 1 FROM information_schema.columns 
			WHERE table_name = 'sr' AND column_name = 'suspended'
		)
	`).Scan(&columnExists)
	if err != nil {
		log.Fatalf("Failed to verify column: %v", err)
	}

	if columnExists {
		log.Println("✅ Verified: suspended column exists in sr table")
	} else {
		log.Println("❌ Warning: suspended column not found in sr table")
	}

	log.Println("\n✅ Migration complete!")
}
