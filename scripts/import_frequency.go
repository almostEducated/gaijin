//go:build ignore

package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"database/sql"

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

	// Step 1: Add frequency column if it doesn't exist
	log.Println("📦 Adding frequency column to words table...")
	_, err = db.Exec(`ALTER TABLE words ADD COLUMN IF NOT EXISTS frequency INTEGER`)
	if err != nil {
		log.Fatalf("Failed to add frequency column: %v", err)
	}
	log.Println("✅ Frequency column ready")

	// Step 2: Read the CSV file
	log.Println("📖 Reading frequency CSV file...")
	file, err := os.Open("Japanese Frequency Analysis.csv")
	if err != nil {
		log.Fatalf("Failed to open CSV file: %v", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)

	// Skip header row
	_, err = reader.Read()
	if err != nil {
		log.Fatalf("Failed to read CSV header: %v", err)
	}

	// Build a map of word -> rank from CSV
	frequencyMap := make(map[string]int)
	csvWordCount := 0

	for {
		record, err := reader.Read()
		if err != nil {
			break // End of file
		}

		if len(record) < 3 {
			continue
		}

		rank, err := strconv.Atoi(strings.TrimSpace(record[0]))
		if err != nil {
			continue
		}

		word := strings.TrimSpace(record[2])
		if word == "" {
			continue
		}

		// Only store the first occurrence (lowest rank = most frequent)
		if _, exists := frequencyMap[word]; !exists {
			frequencyMap[word] = rank
			csvWordCount++
		}
	}

	log.Printf("✅ Loaded %d unique words from CSV", csvWordCount)

	// Step 3: Get all words from database
	log.Println("📊 Fetching words from database...")
	rows, err := db.Query(`SELECT id, word FROM words`)
	if err != nil {
		log.Fatalf("Failed to query words: %v", err)
	}
	defer rows.Close()

	type dbWord struct {
		ID   int
		Word string
	}

	var dbWords []dbWord
	for rows.Next() {
		var w dbWord
		if err := rows.Scan(&w.ID, &w.Word); err != nil {
			log.Printf("Warning: failed to scan word: %v", err)
			continue
		}
		dbWords = append(dbWords, w)
	}

	log.Printf("✅ Found %d words in database", len(dbWords))

	// Step 4: Update frequency for each word
	log.Println("🔄 Updating frequency values...")

	updateStmt, err := db.Prepare(`UPDATE words SET frequency = $1 WHERE id = $2`)
	if err != nil {
		log.Fatalf("Failed to prepare update statement: %v", err)
	}
	defer updateStmt.Close()

	matchedCount := 0
	unmatchedCount := 0
	unmatchedWords := []string{}

	for _, w := range dbWords {
		if rank, exists := frequencyMap[w.Word]; exists {
			_, err := updateStmt.Exec(rank, w.ID)
			if err != nil {
				log.Printf("Warning: failed to update word %s: %v", w.Word, err)
				continue
			}
			matchedCount++
		} else {
			unmatchedCount++
			// Store first 50 unmatched words for reference
			if len(unmatchedWords) < 50 {
				unmatchedWords = append(unmatchedWords, w.Word)
			}
		}
	}

	// Step 5: Output results
	log.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	log.Println("📈 FREQUENCY IMPORT RESULTS")
	log.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	log.Printf("📚 Total words in database: %d", len(dbWords))
	log.Printf("✅ Words matched with frequency: %d", matchedCount)
	log.Printf("❌ Words WITHOUT frequency match: %d", unmatchedCount)
	log.Printf("📊 Match rate: %.1f%%", float64(matchedCount)/float64(len(dbWords))*100)
	log.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	if len(unmatchedWords) > 0 {
		log.Println("\n📝 Sample unmatched words (first 50):")
		for i, word := range unmatchedWords {
			log.Printf("   %d. %s", i+1, word)
		}
	}

	log.Println("\n✅ Frequency import complete!")
}
