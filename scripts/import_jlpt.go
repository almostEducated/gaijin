//go:build ignore

package main

import (
	"encoding/json"
	"log"
	"os"

	"gaijin/internal/database"
	"gaijin/internal/handlers/models"
)

func main() {
	// Connect to database
	db, err := database.ConnectDB()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Ensure tables exist
	if err := db.InitializeTables(); err != nil {
		log.Fatalf("Failed to initialize tables: %v", err)
	}

	// Read JSON file
	jsonFile, err := os.ReadFile("internal/database/all.json")
	if err != nil {
		log.Fatalf("Failed to read JSON file: %v", err)
	}

	// Parse JSON
	var words []models.JLPTWord
	if err := json.Unmarshal(jsonFile, &words); err != nil {
		log.Fatalf("Failed to parse JSON: %v", err)
	}

	log.Printf("Loaded %d words from JSON file", len(words))

	// Insert words into database
	insertQuery := `
		INSERT INTO jlpt_vocabulary (word, meaning, furigana, romaji, level)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (word, level) DO UPDATE 
		SET meaning = EXCLUDED.meaning,
		    furigana = EXCLUDED.furigana,
		    romaji = EXCLUDED.romaji`

	successCount := 0
	errorCount := 0

	for i, word := range words {
		_, err := db.DB.Exec(insertQuery, word.Word, word.Meaning, word.Furigana, word.Romaji, word.Level)
		if err != nil {
			log.Printf("Error inserting word %d ('%s'): %v", i+1, word.Word, err)
			errorCount++
		} else {
			successCount++
		}

		// Log progress every 1000 words
		if (i+1)%1000 == 0 {
			log.Printf("Progress: %d/%d words processed", i+1, len(words))
		}
	}

	log.Printf("✅ Import completed!")
	log.Printf("   Successfully inserted: %d words", successCount)
	log.Printf("   Errors: %d", errorCount)
}
