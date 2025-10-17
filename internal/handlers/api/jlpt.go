package api

import (
	"encoding/json"
	"fmt"
	"gaijin/internal/database"
	"gaijin/internal/handlers/models"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"
)

// JLPTHandler handles JLPT vocabulary API endpoints
type JLPTHandler struct {
	db *database.Database
}

// NewJLPTHandler creates a new JLPT handler with database dependency
func NewJLPTHandler(db *database.Database) *JLPTHandler {
	return &JLPTHandler{db: db}
}

// HandleWords handles the general JLPT words endpoint with flexible parameters
func (h *JLPTHandler) HandleWords(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	word := query.Get("word")
	level := query.Get("level")
	offset := query.Get("offset")
	limit := query.Get("limit")

	// Build API URL
	apiURL := "https://jlpt-vocab-api.vercel.app/api/words"
	params := []string{}

	if word != "" {
		params = append(params, "word="+word)
	}
	if level != "" {
		params = append(params, "level="+level)
	}
	if offset != "" {
		params = append(params, "offset="+offset)
	}
	if limit != "" {
		params = append(params, "limit="+limit)
	}

	if len(params) > 0 {
		apiURL += "?" + strings.Join(params, "&")
	}

	// Fetch from API
	words, err := h.fetchJLPTFromAPI(apiURL)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error fetching JLPT words: %v", err), http.StatusInternalServerError)
		return
	}

	// Cache in database if available
	if h.db != nil {
		h.cacheJLPTWordsInDB(words)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(words)
}

// HandlePaginated handles the paginated JLPT words endpoint
func (h *JLPTHandler) HandlePaginated(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	level := query.Get("level")
	offset := query.Get("offset")
	limit := query.Get("limit")

	// Set defaults
	if offset == "" {
		offset = "0"
	}
	if limit == "" {
		limit = "10"
	}

	// Build API URL with pagination
	apiURL := fmt.Sprintf("https://jlpt-vocab-api.vercel.app/api/words?offset=%s&limit=%s", offset, limit)
	if level != "" {
		apiURL += "&level=" + level
	}

	// Fetch from API
	words, err := h.fetchJLPTFromAPI(apiURL)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error fetching paginated JLPT words: %v", err), http.StatusInternalServerError)
		return
	}

	// Cache in database if available
	if h.db != nil {
		h.cacheJLPTWordsInDB(words)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(words)
}

// HandleRandom handles the random JLPT word endpoint
func (h *JLPTHandler) HandleRandom(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	level := query.Get("level")

	apiURL := "https://jlpt-vocab-api.vercel.app/api/words/random"
	if level != "" {
		apiURL += "?level=" + level
	}

	// Fetch random word (returns single word, not array)
	word, err := h.fetchJLPTRandomFromAPI(apiURL)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error fetching random JLPT word: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(word)
}

// HandleAll handles the endpoint to fetch all JLPT words
func (h *JLPTHandler) HandleAll(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	level := query.Get("level")

	apiURL := "https://jlpt-vocab-api.vercel.app/api/words/all"
	if level != "" {
		apiURL += "?level=" + level
	}

	// Fetch all words (returns array, not object)
	words, err := h.fetchJLPTAllFromAPI(apiURL)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error fetching all JLPT words: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(words)
}

// HandleLevel returns a handler function for a specific JLPT level
func (h *JLPTHandler) HandleLevel(level int) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// First try to get from database cache
		if h.db != nil {
			if words := h.getJLPTWordsFromDB(level); words != nil {
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(words)
				return
			}
		}

		// If not in database, fetch from API
		apiURL := fmt.Sprintf("https://jlpt-vocab-api.vercel.app/api/words?level=%d", level)
		words, err := h.fetchJLPTFromAPI(apiURL)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error fetching JLPT N%d words: %v", level, err), http.StatusInternalServerError)
			return
		}

		// Cache in database if available
		if h.db != nil {
			h.cacheJLPTWordsInDB(words)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(words)
	}
}

// RegisterRoutes registers all JLPT-related routes to the router
func (h *JLPTHandler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/api/jlpt/words", h.HandleWords).Methods("GET")
	router.HandleFunc("/api/jlpt/words/paginated", h.HandlePaginated).Methods("GET")
	router.HandleFunc("/api/jlpt/words/random", h.HandleRandom).Methods("GET")
	router.HandleFunc("/api/jlpt/words/all", h.HandleAll).Methods("GET")
	router.HandleFunc("/api/jlpt/n5", h.HandleLevel(5)).Methods("GET")
	router.HandleFunc("/api/jlpt/n4", h.HandleLevel(4)).Methods("GET")
	router.HandleFunc("/api/jlpt/n3", h.HandleLevel(3)).Methods("GET")
	router.HandleFunc("/api/jlpt/n2", h.HandleLevel(2)).Methods("GET")
	router.HandleFunc("/api/jlpt/n1", h.HandleLevel(1)).Methods("GET")
}

// fetchJLPTFromAPI fetches JLPT words from the external API
func (h *JLPTHandler) fetchJLPTFromAPI(apiURL string) (*models.JLPTAPIResponse, error) {
	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Get(apiURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var jlptResp models.JLPTAPIResponse
	err = json.Unmarshal(body, &jlptResp)
	if err != nil {
		return nil, err
	}

	return &jlptResp, nil
}

// fetchJLPTRandomFromAPI fetches a random JLPT word from the external API
func (h *JLPTHandler) fetchJLPTRandomFromAPI(apiURL string) (*models.JLPTWord, error) {
	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Get(apiURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var word models.JLPTWord
	err = json.Unmarshal(body, &word)
	if err != nil {
		return nil, err
	}

	return &word, nil
}

// fetchJLPTAllFromAPI fetches all JLPT words from the external API
func (h *JLPTHandler) fetchJLPTAllFromAPI(apiURL string) ([]models.JLPTWord, error) {
	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Get(apiURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var words []models.JLPTWord
	err = json.Unmarshal(body, &words)
	if err != nil {
		return nil, err
	}

	return words, nil
}

// getJLPTWordsFromDB retrieves JLPT words from the database cache
func (h *JLPTHandler) getJLPTWordsFromDB(level int) *models.JLPTAPIResponse {
	if h.db == nil || h.db.DB == nil {
		return nil
	}

	rows, err := h.db.DB.Query(`
		SELECT word, meaning, furigana, romaji, level 
		FROM jlpt_vocabulary WHERE level = $1 
		ORDER BY id`, level)
	if err != nil {
		return nil
	}
	defer rows.Close()

	var words []models.JLPTWord
	for rows.Next() {
		var word models.JLPTWord
		err := rows.Scan(&word.Word, &word.Meaning, &word.Furigana,
			&word.Romaji, &word.Level)
		if err != nil {
			continue
		}
		words = append(words, word)
	}

	if len(words) == 0 {
		return nil
	}

	return &models.JLPTAPIResponse{
		Words:  words,
		Total:  len(words),
		Offset: 0,
		Limit:  len(words),
	}
}

// cacheJLPTWordsInDB caches JLPT words in the database
func (h *JLPTHandler) cacheJLPTWordsInDB(words *models.JLPTAPIResponse) {
	if h.db == nil || h.db.DB == nil || words == nil {
		return
	}

	for _, word := range words.Words {
		_, err := h.db.DB.Exec(`
			INSERT INTO jlpt_vocabulary (word, meaning, furigana, romaji, level) 
			VALUES ($1, $2, $3, $4, $5) 
			ON CONFLICT (word, level) DO UPDATE SET
			meaning = EXCLUDED.meaning,
			furigana = EXCLUDED.furigana,
			romaji = EXCLUDED.romaji`,
			word.Word, word.Meaning, word.Furigana, word.Romaji, word.Level)

		if err != nil {
			log.Printf("Error caching JLPT word %s: %v", word.Word, err)
		}
	}
}
