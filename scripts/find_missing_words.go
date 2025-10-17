//go:build ignore

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"gaijin/internal/database"
)

// MissingWord represents a word that's in jlpt_vocabulary but not in words
type MissingWord struct {
	ID       int    `json:"id"`
	Word     string `json:"word"`
	Meaning  string `json:"meaning"`
	Furigana string `json:"furigana"`
	Romaji   string `json:"romaji"`
	Level    int    `json:"level"`
}

func main() {
	// Connect to database
	db, err := database.ConnectDB()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	log.Println("Finding words in jlpt_vocabulary that are NOT in words table...")

	// Query to find missing words
	query := `
		SELECT jv.id, jv.word, jv.meaning, jv.furigana, jv.romaji, jv.level
		FROM jlpt_vocabulary jv
		LEFT JOIN words w ON jv.word = w.word AND jv.level = w.level
		WHERE w.id IS NULL
		ORDER BY jv.level, jv.id
	`

	rows, err := db.Query(query)
	if err != nil {
		log.Fatalf("Failed to query: %v", err)
	}
	defer rows.Close()

	var missingWords []MissingWord
	for rows.Next() {
		var word MissingWord
		err := rows.Scan(&word.ID, &word.Word, &word.Meaning, &word.Furigana, &word.Romaji, &word.Level)
		if err != nil {
			log.Printf("Error scanning row: %v", err)
			continue
		}
		missingWords = append(missingWords, word)
	}

	log.Printf("Found %d missing words", len(missingWords))

	if len(missingWords) > 0 {
		// Save to JSON file
		jsonData, err := json.MarshalIndent(missingWords, "", "  ")
		if err != nil {
			log.Fatalf("Failed to marshal JSON: %v", err)
		}

		err = os.WriteFile("missing_words.json", jsonData, 0644)
		if err != nil {
			log.Fatalf("Failed to write file: %v", err)
		}

		log.Printf("✅ Saved missing words to missing_words.json")

		// Print first 10 as a sample
		log.Println("\nSample of missing words:")
		for i, word := range missingWords {
			if i >= 10 {
				break
			}
			fmt.Printf("  [%d] %s (%s) - Level %d\n", word.ID, word.Word, word.Furigana, word.Level)
		}
	}
}
