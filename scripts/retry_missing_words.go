//go:build ignore

package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"

	"gaijin/internal/database"
)

// JishoResponse represents the API response from Jisho.org
type JishoResponse struct {
	Meta struct {
		Status int `json:"status"`
	} `json:"meta"`
	Data []struct {
		Slug     string   `json:"slug"`
		IsCommon bool     `json:"is_common"`
		Tags     []string `json:"tags"`
		JLPT     []string `json:"jlpt"`
		Japanese []struct {
			Word    string `json:"word"`
			Reading string `json:"reading"`
		} `json:"japanese"`
		Senses []struct {
			EnglishDefinitions []string `json:"english_definitions"`
			PartsOfSpeech      []string `json:"parts_of_speech"`
			Tags               []string `json:"tags"`
		} `json:"senses"`
	} `json:"data"`
}

// MissingWord represents a word from the missing_words.json file
type MissingWord struct {
	ID       int    `json:"id"`
	Word     string `json:"word"`
	Meaning  string `json:"meaning"`
	Furigana string `json:"furigana"`
	Romaji   string `json:"romaji"`
	Level    int    `json:"level"`
}

func main() {
	// Check if missing_words.json exists
	if _, err := os.Stat("missing_words.json"); os.IsNotExist(err) {
		log.Fatal("missing_words.json not found. Run find_missing_words.go first!")
	}

	// Read missing words from JSON
	data, err := os.ReadFile("missing_words.json")
	if err != nil {
		log.Fatalf("Failed to read missing_words.json: %v", err)
	}

	var missingWords []MissingWord
	if err := json.Unmarshal(data, &missingWords); err != nil {
		log.Fatalf("Failed to parse JSON: %v", err)
	}

	log.Printf("Loaded %d missing words from missing_words.json", len(missingWords))

	// Connect to database
	db, err := database.ConnectDB()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	successCount := 0
	errorCount := 0
	skipCount := 0

	// Process each missing word
	for i, vocab := range missingWords {
		// Skip first N words if needed (set this manually to resume after rate limiting)
		if skipCount > 0 && i < skipCount {
			continue
		}

		log.Printf("[%d/%d] Processing: %s (%s)", i+1, len(missingWords), vocab.Word, vocab.Furigana)

		// Query Jisho API
		jishoData, err := queryJishoAPI(vocab.Word)
		if err != nil {
			log.Printf("  ⚠️  Error querying Jisho API: %v", err)
			errorCount++
			// Longer wait on error (might be rate limited)
			time.Sleep(2 * time.Second)
			continue
		}

		if len(jishoData.Data) == 0 {
			log.Printf("  ⚠️  No results from Jisho API - using JLPT fallback data")
			// Use fallback: insert with original JLPT vocabulary data
			var wordID int
			err = db.DB.QueryRow(`
				INSERT INTO words (word, furigana, romaji, level, definitions, parts_of_speech, created_at)
				VALUES ($1, $2, $3, $4, $5, $6, NOW())
				ON CONFLICT (word, level) DO UPDATE 
				SET furigana = EXCLUDED.furigana, 
				    romaji = EXCLUDED.romaji,
				    definitions = EXCLUDED.definitions,
				    parts_of_speech = EXCLUDED.parts_of_speech
				RETURNING id
			`, vocab.Word, vocab.Furigana, vocab.Romaji, vocab.Level, vocab.Meaning, "Unknown").Scan(&wordID)

			if err != nil {
				log.Printf("  ❌ Error inserting fallback word: %v", err)
				errorCount++
			} else {
				log.Printf("  ✅ Success (fallback): using JLPT meaning")
				successCount++
			}
			time.Sleep(500 * time.Millisecond)
			continue
		}

		// Find exact match
		var entry *struct {
			Slug     string   `json:"slug"`
			IsCommon bool     `json:"is_common"`
			Tags     []string `json:"tags"`
			JLPT     []string `json:"jlpt"`
			Japanese []struct {
				Word    string `json:"word"`
				Reading string `json:"reading"`
			} `json:"japanese"`
			Senses []struct {
				EnglishDefinitions []string `json:"english_definitions"`
				PartsOfSpeech      []string `json:"parts_of_speech"`
				Tags               []string `json:"tags"`
			} `json:"senses"`
		}

		for j := range jishoData.Data {
			if jishoData.Data[j].Slug == vocab.Word {
				entry = &jishoData.Data[j]
				break
			}
		}

		if entry == nil {
			log.Printf("  ⚠️  No exact match found (slug != %s) - using JLPT fallback data", vocab.Word)
			// Use fallback: insert with original JLPT vocabulary data
			var wordID int
			err = db.DB.QueryRow(`
				INSERT INTO words (word, furigana, romaji, level, definitions, parts_of_speech, created_at)
				VALUES ($1, $2, $3, $4, $5, $6, NOW())
				ON CONFLICT (word, level) DO UPDATE 
				SET furigana = EXCLUDED.furigana, 
				    romaji = EXCLUDED.romaji,
				    definitions = EXCLUDED.definitions,
				    parts_of_speech = EXCLUDED.parts_of_speech
				RETURNING id
			`, vocab.Word, vocab.Furigana, vocab.Romaji, vocab.Level, vocab.Meaning, "Unknown").Scan(&wordID)

			if err != nil {
				log.Printf("  ❌ Error inserting fallback word: %v", err)
				errorCount++
			} else {
				log.Printf("  ✅ Success (fallback): using JLPT meaning")
				successCount++
			}
			time.Sleep(500 * time.Millisecond)
			continue
		}

		// Collect definitions and parts of speech
		var allDefinitions []string
		var allPartsOfSpeech []string

		for _, sense := range entry.Senses {
			for _, def := range sense.EnglishDefinitions {
				allDefinitions = append(allDefinitions, def)
			}

			for _, pos := range sense.PartsOfSpeech {
				found := false
				for _, existing := range allPartsOfSpeech {
					if existing == pos {
						found = true
						break
					}
				}
				if !found {
					allPartsOfSpeech = append(allPartsOfSpeech, pos)
				}
			}
		}

		// Join with semicolons
		definitionsStr := ""
		for j, def := range allDefinitions {
			if j > 0 {
				definitionsStr += "; "
			}
			definitionsStr += def
		}

		partsOfSpeechStr := ""
		for j, pos := range allPartsOfSpeech {
			if j > 0 {
				partsOfSpeechStr += "; "
			}
			partsOfSpeechStr += pos
		}

		// Insert into words table
		var wordID int
		err = db.DB.QueryRow(`
			INSERT INTO words (word, furigana, romaji, level, definitions, parts_of_speech, created_at)
			VALUES ($1, $2, $3, $4, $5, $6, NOW())
			ON CONFLICT (word, level) DO UPDATE 
			SET furigana = EXCLUDED.furigana, 
			    romaji = EXCLUDED.romaji,
			    definitions = EXCLUDED.definitions,
			    parts_of_speech = EXCLUDED.parts_of_speech
			RETURNING id
		`, vocab.Word, vocab.Furigana, vocab.Romaji, vocab.Level, definitionsStr, partsOfSpeechStr).Scan(&wordID)

		if err != nil {
			log.Printf("  ❌ Error inserting word: %v", err)
			errorCount++
			continue
		}

		log.Printf("  ✅ Success: %d definitions, %d parts of speech",
			len(allDefinitions), len(allPartsOfSpeech))
		successCount++

		// Longer rate limiting - be extra nice to Jisho API (500ms = ~2 requests per second)
		time.Sleep(500 * time.Millisecond)
	}

	log.Printf("\n=== Retry Complete ===")
	log.Printf("✅ Success: %d words", successCount)
	log.Printf("❌ Errors: %d words", errorCount)
}

func queryJishoAPI(word string) (*JishoResponse, error) {
	keyword := url.QueryEscape(word)
	apiURL := fmt.Sprintf("https://jisho.org/api/v1/search/words?keyword=%s", keyword)

	resp, err := http.Get(apiURL)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var jishoResp JishoResponse
	err = json.Unmarshal(body, &jishoResp)
	if err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	return &jishoResp, nil
}
