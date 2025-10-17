package main

import (
	"gaijin/internal/database"
	"gaijin/internal/server"
	"log"

	_ "github.com/lib/pq"
)

func main() {
	// Connect to database
	db, err := database.ConnectDB()
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// Initialize database tables (safe to run every time - uses CREATE TABLE IF NOT EXISTS)
	if err := db.InitializeTables(); err != nil {
		log.Printf("Warning: Failed to initialize tables: %v", err)
	}

	srv := server.New(db)

	log.Println("Starting server...")
	if err := srv.Start(); err != nil {
		log.Fatal("Server failed to start:", err)
	}
}
