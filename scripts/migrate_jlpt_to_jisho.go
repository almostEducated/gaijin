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

// isKatakana checks if a rune is a katakana character
func isKatakana(r rune) bool {
	return (r >= '\u30A0' && r <= '\u30FF') || // Katakana block
		(r >= '\u31F0' && r <= '\u31FF') || // Katakana Phonetic Extensions
		(r >= '\uFF65' && r <= '\uFF9F') // Halfwidth Katakana
}

// isKatakanaOnly checks if a string contains only katakana characters and common punctuation
func isKatakanaOnly(s string) bool {
	if len(s) == 0 {
		return false
	}

	hasKatakana := false
	for _, r := range s {
		if isKatakana(r) {
			hasKatakana = true
		} else if !(r == ' ' || r == '・' || r == 'ー' || r == '〜' || r == '～') {
			// Not katakana and not allowed punctuation
			return false
		}
	}
	return hasKatakana
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
	skipCount := 0
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

		// Check if word is katakana-only
		isKatakanaWord := isKatakanaOnly(vocab.Word)
		if isKatakanaWord {
			log.Printf("  📝 Detected katakana-only word")
		}

		// Query Jisho API
		jishoData, err := queryJishoAPI(vocab.Word)
		if err != nil {
			log.Printf("  ⚠️  Error querying Jisho API: %v", err)
			errorCount++
			continue
		}

		if len(jishoData.Data) == 0 {
			log.Printf("  ⚠️  No results from Jisho API - using JLPT fallback data")
			// Use fallback: insert with original JLPT vocabulary data
			err = db.DB.QueryRow(`
				INSERT INTO words (word, furigana, romaji, level, definitions, parts_of_speech, katakana_only, hiragana_only, created_at)
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW())
				ON CONFLICT (word, level) DO UPDATE 
				SET furigana = EXCLUDED.furigana, 
				    romaji = EXCLUDED.romaji,
				    definitions = EXCLUDED.definitions,
				    parts_of_speech = EXCLUDED.parts_of_speech,
				    katakana_only = EXCLUDED.katakana_only,
				    hiragana_only = EXCLUDED.hiragana_only
				RETURNING id
			`, vocab.Word, vocab.Furigana, vocab.Romaji, vocab.Level, vocab.Meaning, "Unknown", isKatakanaWord, false).Scan(&wordID)

			if err != nil {
				log.Printf("  ❌ Error inserting fallback word: %v", err)
				errorCount++
			} else {
				log.Printf("  ✅ Success (fallback): using JLPT meaning")
				successCount++
			}
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
			// Before falling back, check if the JLPT word matches any reading in the results
			// This handles cases where JLPT gives us hiragana-only but Jisho has the kanji version
			log.Printf("  🔍 No slug match - checking if JLPT word matches any reading...")

			for i := range jishoData.Data {
				for _, japanese := range jishoData.Data[i].Japanese {
					if japanese.Reading == vocab.Word {
						log.Printf("  ✨ Found reading match! JLPT has hiragana '%s', Jisho has kanji '%s'", vocab.Word, japanese.Word)
						entry = &jishoData.Data[i]
						break
					}
				}
				if entry != nil {
					break
				}
			}
		}

		if entry == nil {
			log.Printf("  ⚠️  No exact match or reading match found - using JLPT fallback data")
			// Use fallback: insert with original JLPT vocabulary data
			err = db.DB.QueryRow(`
				INSERT INTO words (word, furigana, romaji, level, definitions, parts_of_speech, katakana_only, hiragana_only, created_at)
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW())
				ON CONFLICT (word, level) DO UPDATE 
				SET furigana = EXCLUDED.furigana, 
				    romaji = EXCLUDED.romaji,
				    definitions = EXCLUDED.definitions,
				    parts_of_speech = EXCLUDED.parts_of_speech,
				    katakana_only = EXCLUDED.katakana_only,
				    hiragana_only = EXCLUDED.hiragana_only
				RETURNING id
			`, vocab.Word, vocab.Furigana, vocab.Romaji, vocab.Level, vocab.Meaning, "Unknown", isKatakanaWord, false).Scan(&wordID)

			if err != nil {
				log.Printf("  ❌ Error inserting fallback word: %v", err)
				errorCount++
			} else {
				log.Printf("  ✅ Success (fallback): using JLPT meaning")
				successCount++
			}
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

		// Determine the actual word to use and whether this is a hiragana-only case
		actualWord := vocab.Word
		actualFurigana := vocab.Furigana
		isHiraganaOnly := false

		// Check if we matched by reading (JLPT had hiragana, Jisho has kanji)
		if entry.Slug != vocab.Word && len(entry.Japanese) > 0 {
			// Look for a Japanese entry that has our vocab.Word as the reading
			for _, japanese := range entry.Japanese {
				if japanese.Reading == vocab.Word && japanese.Word != "" && japanese.Word != vocab.Word {
					// We found the case where JLPT gives hiragana but Jisho has kanji
					actualWord = japanese.Word
					actualFurigana = vocab.Word // Use the original hiragana as furigana
					isHiraganaOnly = true
					log.Printf("  📚 Using kanji '%s' with hiragana '%s' as furigana (hiragana_only=true)", actualWord, actualFurigana)
					break
				}
			}
		}

		// Insert into words table with definitions and parts of speech
		err = db.DB.QueryRow(`
			INSERT INTO words (word, furigana, romaji, level, definitions, parts_of_speech, katakana_only, hiragana_only, created_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW())
			ON CONFLICT (word, level) DO UPDATE 
			SET furigana = EXCLUDED.furigana, 
			    romaji = EXCLUDED.romaji,
			    definitions = EXCLUDED.definitions,
			    parts_of_speech = EXCLUDED.parts_of_speech,
			    katakana_only = EXCLUDED.katakana_only,
			    hiragana_only = EXCLUDED.hiragana_only
			RETURNING id
		`, actualWord, actualFurigana, vocab.Romaji, vocab.Level, definitionsStr, partsOfSpeechStr, isKatakanaWord, isHiraganaOnly).Scan(&wordID)

		if err != nil {
			log.Printf("  ❌ Error inserting word: %v", err)
			errorCount++
			continue
		}

		log.Printf("  ✅ Success: %d definitions, %d parts of speech",
			len(allDefinitions), len(allPartsOfSpeech))
		successCount++

		// Rate limiting - be nice to Jisho API (increased to avoid 423 errors)
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
