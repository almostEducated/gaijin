//go:build ignore

package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
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

// JLPTVocab represents a row from the jlpt_vocabulary table
type JLPTVocab struct {
	ID       int
	Word     string
	Meaning  string
	Furigana string
	Romaji   string
	Level    int
}

func main() {
	// Connect to database
	db, err := database.ConnectDB()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	log.Println("Starting migration from jlpt_vocabulary to words table with Jisho API enrichment...")

	// Fetch all existing JLPT vocabulary
	rows, err := db.Query(`
		SELECT id, word, meaning, furigana, romaji, level 
		FROM jlpt_vocabulary 
		ORDER BY level, id
	`)
	if err != nil {
		log.Fatalf("Failed to query jlpt_vocabulary: %v", err)
	}
	defer rows.Close()

	var vocabList []JLPTVocab
	for rows.Next() {
		var v JLPTVocab
		err := rows.Scan(&v.ID, &v.Word, &v.Meaning, &v.Furigana, &v.Romaji, &v.Level)
		if err != nil {
			log.Printf("Error scanning row: %v", err)
			continue
		}
		vocabList = append(vocabList, v)
	}

	log.Printf("Found %d vocabulary words to migrate", len(vocabList))

	// Skip already processed words (update this number as needed)
	skipCount := 4416
	successCount := 0
	errorCount := 0

	// Process each vocabulary word
	for i, vocab := range vocabList {
		// Skip already processed
		if i < skipCount {
			continue
		}
		log.Printf("[%d/%d] Processing: %s (%s)", i+1, len(vocabList), vocab.Word, vocab.Furigana)

		var wordID int

		// Query Jisho API
		jishoData, err := queryJishoAPI(vocab.Word)
		if err != nil {
			log.Printf("  ⚠️  Error querying Jisho API: %v", err)
			errorCount++
			continue
		}

		if len(jishoData.Data) == 0 {
			log.Printf("  ⚠️  No results from Jisho API")
			errorCount++
			continue
		}

		// Find the entry where slug exactly matches our word
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

		for i := range jishoData.Data {
			if jishoData.Data[i].Slug == vocab.Word {
				entry = &jishoData.Data[i]
				break
			}
		}

		if entry == nil {
			log.Printf("  ⚠️  No exact match found (slug != %s)", vocab.Word)
			errorCount++
			continue
		}

		// Collect all definitions and parts of speech into semicolon-separated strings
		var allDefinitions []string
		var allPartsOfSpeech []string

		for _, sense := range entry.Senses {
			// Combine english definitions for this sense
			for _, def := range sense.EnglishDefinitions {
				allDefinitions = append(allDefinitions, def)
			}

			// Collect parts of speech for this sense
			for _, pos := range sense.PartsOfSpeech {
				// Avoid duplicates
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
		for i, def := range allDefinitions {
			if i > 0 {
				definitionsStr += "; "
			}
			definitionsStr += def
		}

		partsOfSpeechStr := ""
		for i, pos := range allPartsOfSpeech {
			if i > 0 {
				partsOfSpeechStr += "; "
			}
			partsOfSpeechStr += pos
		}

		// Insert into words table with definitions and parts of speech
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

		// Rate limiting - be nice to Jisho API
		time.Sleep(1 * time.Second)
	}

	log.Printf("\n=== Migration Complete ===")
	log.Printf("✅ Success: %d words", successCount)
	log.Printf("❌ Errors: %d words", errorCount)
}

func queryJishoAPI(word string) (*JishoResponse, error) {
	// URL encode the word
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
