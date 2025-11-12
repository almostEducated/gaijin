package api

import (
	"encoding/json"
	"gaijin/internal/auth"
	"gaijin/internal/database"
	"net/http"
	"strconv"
)

// KanjiConfusionHandler handles kanji confusion-related API endpoints
type KanjiConfusionHandler struct {
	db   *database.Database
	auth *auth.Auth
}

// NewKanjiConfusionHandler creates a new kanji confusion handler
func NewKanjiConfusionHandler(db *database.Database, auth *auth.Auth) *KanjiConfusionHandler {
	return &KanjiConfusionHandler{
		db:   db,
		auth: auth,
	}
}

// HandleGetSimilarKanji returns a list of kanji that might be visually similar to the current word
func (h *KanjiConfusionHandler) HandleGetSimilarKanji(w http.ResponseWriter, r *http.Request) {
	userID, err := h.auth.GetCurrentUser(r)
	if err != nil {
		http.Error(w, "User not authenticated", http.StatusUnauthorized)
		return
	}

	srIDStr := r.URL.Query().Get("sr_id")
	if srIDStr == "" {
		http.Error(w, "SR ID is required", http.StatusBadRequest)
		return
	}

	srID, err := strconv.Atoi(srIDStr)
	if err != nil {
		http.Error(w, "Invalid SR ID", http.StatusBadRequest)
		return
	}

	// Get the word from SR ID
	word, err := h.db.LookupWordBySRId(srID)
	if err != nil {
		http.Error(w, "Failed to get word: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Verify user owns this SR record
	var ownerID int
	query := `SELECT user_id FROM sr WHERE id = $1`
	err = h.db.DB.QueryRow(query, srID).Scan(&ownerID)
	if err != nil {
		http.Error(w, "SR record not found: "+err.Error(), http.StatusNotFound)
		return
	}
	if ownerID != userID {
		http.Error(w, "Unauthorized access", http.StatusForbidden)
		return
	}

	// Extract individual kanji from the word
	kanjiChars := extractKanji(word.Word)
	if len(kanjiChars) == 0 {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"kanji": []interface{}{},
		})
		return
	}

	// For now, return words that contain any of these kanji
	// This is a simple approach - you can enhance this with a proper similarity algorithm
	similarWords, err := h.findWordsWithSimilarKanji(kanjiChars, word.ID, word.Level)
	if err != nil {
		http.Error(w, "Failed to find similar kanji: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"kanji": similarWords,
	})
}

// HandleLinkKanji creates a confusion pair between two kanji
func (h *KanjiConfusionHandler) HandleLinkKanji(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, err := h.auth.GetCurrentUser(r)
	if err != nil {
		http.Error(w, "User not authenticated", http.StatusUnauthorized)
		return
	}

	srIDStr := r.FormValue("sr_id")
	if srIDStr == "" {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "SR ID is required",
		})
		return
	}

	srID, err := strconv.Atoi(srIDStr)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Invalid SR ID",
		})
		return
	}

	// Verify ownership
	var wordID int
	query := `SELECT word_id FROM sr WHERE id = $1 AND user_id = $2`
	err = h.db.DB.QueryRow(query, srID, userID).Scan(&wordID)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "SR record not found or unauthorized",
		})
		return
	}

	similarKanji := r.FormValue("similar_kanji")
	similarWordID, err := strconv.Atoi(r.FormValue("similar_word_id"))
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Invalid similar word ID",
		})
		return
	}

	_ = r.FormValue("similar_word")     // May be used later
	_ = r.FormValue("similar_furigana") // May be used later

	// Get the current word's kanji
	currentWord, err := h.db.LookupWordBySRId(srID)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Failed to get current word",
		})
		return
	}

	// Extract first kanji from each word
	kanji1 := extractFirstKanji(currentWord.Word)
	kanji2 := similarKanji

	// Insert confusion pair
	insertQuery := `
		INSERT INTO kanji_confusion (kanji_1, kanji_2, word1_id, word2_id, user_id)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT DO NOTHING
	`
	_, err = h.db.DB.Exec(insertQuery, kanji1, kanji2, wordID, similarWordID, userID)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Failed to link kanji: " + err.Error(),
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
	})
}

// Helper function to extract kanji from a string
func extractKanji(s string) []rune {
	var kanji []rune
	for _, r := range s {
		// Kanji range: U+4E00 to U+9FFF, with some extensions
		if (r >= 0x4E00 && r <= 0x9FFF) ||
			(r >= 0x3400 && r <= 0x4DBF) || // CJK Extension A
			(r >= 0x20000 && r <= 0x2A6DF) { // CJK Extension B (rare)
			kanji = append(kanji, r)
		}
	}
	return kanji
}

// Helper function to extract first kanji
func extractFirstKanji(s string) string {
	kanji := extractKanji(s)
	if len(kanji) > 0 {
		return string(kanji[0])
	}
	return ""
}

// Helper function to find words with similar kanji
func (h *KanjiConfusionHandler) findWordsWithSimilarKanji(kanji []rune, excludeWordID, level int) ([]map[string]interface{}, error) {
	if len(kanji) == 0 {
		return []map[string]interface{}{}, nil
	}

	// For each kanji, find words that contain it
	var results []map[string]interface{}
	seen := make(map[int]bool)

	for _, k := range kanji {
		// Search for words containing this kanji
		query := `
			SELECT DISTINCT id, word, furigana
			FROM words
			WHERE id != $1
			AND level = $2
			AND word LIKE '%' || $3 || '%'
			LIMIT 5
		`
		rows, err := h.db.DB.Query(query, excludeWordID, level, string(k))
		if err != nil {
			continue
		}

		for rows.Next() {
			var wordID int
			var word, furigana string
			err := rows.Scan(&wordID, &word, &furigana)
			if err != nil {
				continue
			}

			// Avoid duplicates
			if seen[wordID] {
				continue
			}
			seen[wordID] = true

			// Get first kanji from this word
			firstKanji := extractFirstKanji(word)

			results = append(results, map[string]interface{}{
				"word_id":  wordID,
				"word":     word,
				"furigana": furigana,
				"kanji":    firstKanji,
			})
		}
		rows.Close()

		// Limit total results
		if len(results) >= 10 {
			break
		}
	}

	return results, nil
}
