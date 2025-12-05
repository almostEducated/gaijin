package api

import (
	"encoding/json"
	"gaijin/internal/auth"
	"gaijin/internal/database"
	"net/http"
	"strconv"
)

type LearnHandler struct {
	db   *database.Database
	auth *auth.Auth
}

func NewLearnHandler(db *database.Database, auth *auth.Auth) *LearnHandler {
	return &LearnHandler{db: db, auth: auth}
}

// AddWordRequest represents the request body for adding a single word
type AddWordRequest struct {
	WordID int `json:"word_id"`
}

// AddWordsRequest represents the request body for adding multiple words
type AddWordsRequest struct {
	WordIDs []int `json:"word_ids"`
}

// HandleAddWord adds a single word to the user's SR deck
func (h *LearnHandler) HandleAddWord(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, err := h.auth.GetCurrentUser(r)
	if err != nil {
		http.Error(w, "User not authenticated", http.StatusUnauthorized)
		return
	}

	// Try to get word_id from form data first (for form submissions)
	wordIDStr := r.FormValue("word_id")
	var wordID int

	if wordIDStr != "" {
		wordID, err = strconv.Atoi(wordIDStr)
		if err != nil {
			http.Error(w, "Invalid word ID", http.StatusBadRequest)
			return
		}
	} else {
		// Try JSON body
		var req AddWordRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}
		wordID = req.WordID
	}

	if wordID <= 0 {
		http.Error(w, "Invalid word ID", http.StatusBadRequest)
		return
	}

	err = h.db.AddWordToSR(userID, wordID)
	if err != nil {
		http.Error(w, "Failed to add word: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Check if this is an HTMX request
	if r.Header.Get("HX-Request") == "true" {
		// Return a small HTML snippet indicating success
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write([]byte(`<span class="learn-status learned">✓ Added</span>`))
		return
	}

	// For regular requests, return JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Word added to study deck",
	})
}

// HandleAddWords adds multiple words to the user's SR deck
func (h *LearnHandler) HandleAddWords(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, err := h.auth.GetCurrentUser(r)
	if err != nil {
		http.Error(w, "User not authenticated", http.StatusUnauthorized)
		return
	}

	var req AddWordsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if len(req.WordIDs) == 0 {
		http.Error(w, "No word IDs provided", http.StatusBadRequest)
		return
	}

	err = h.db.AddMultipleWordsToSR(userID, req.WordIDs)
	if err != nil {
		http.Error(w, "Failed to add words: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Words added to study deck",
		"count":   len(req.WordIDs),
	})
}

// HandleAddAllOnPage adds all words from the current page to the user's SR deck
func (h *LearnHandler) HandleAddAllOnPage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, err := h.auth.GetCurrentUser(r)
	if err != nil {
		http.Error(w, "User not authenticated", http.StatusUnauthorized)
		return
	}

	// Get level and page from query parameters
	levelStr := r.URL.Query().Get("level")
	pageStr := r.URL.Query().Get("page")

	level := 5
	page := 1
	batchSize := 10

	if levelStr != "" {
		if levelStr == "all" || levelStr == "0" {
			level = 0 // All levels mode
		} else if l, err := strconv.Atoi(levelStr); err == nil && l >= 1 && l <= 5 {
			level = l
		}
	}
	if pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p >= 1 {
			page = p
		}
	}

	offset := (page - 1) * batchSize

	// Get words on this page
	words, _, err := h.db.GetWordsForLearning(userID, level, batchSize, offset)
	if err != nil {
		http.Error(w, "Failed to get words: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Collect word IDs that aren't already learned
	var wordIDs []int
	for _, word := range words {
		if !word.IsLearned {
			wordIDs = append(wordIDs, word.ID)
		}
	}

	if len(wordIDs) > 0 {
		err = h.db.AddMultipleWordsToSR(userID, wordIDs)
		if err != nil {
			http.Error(w, "Failed to add words: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}

	// Redirect back to the same page
	http.Redirect(w, r, r.Header.Get("Referer"), http.StatusSeeOther)
}

// HandleToggleSuspended toggles the suspended status for a word in the user's SR deck
func (h *LearnHandler) HandleToggleSuspended(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, err := h.auth.GetCurrentUser(r)
	if err != nil {
		http.Error(w, "User not authenticated", http.StatusUnauthorized)
		return
	}

	// Get word ID from query parameter
	wordIDStr := r.URL.Query().Get("word_id")
	if wordIDStr == "" {
		http.Error(w, "word_id is required", http.StatusBadRequest)
		return
	}

	wordID, err := strconv.Atoi(wordIDStr)
	if err != nil {
		http.Error(w, "Invalid word_id", http.StatusBadRequest)
		return
	}

	// Toggle suspended status
	newState, err := h.db.ToggleWordSuspended(userID, wordID)
	if err != nil {
		http.Error(w, "Failed to toggle suspended: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Check if this is an HTMX request
	if r.Header.Get("HX-Request") == "true" {
		// Return a button with the new state for HTMX to swap
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if newState {
			// Word is now suspended (known)
			w.Write([]byte(`<button class="know-btn known" 
				hx-post="/api/learn/toggle-suspended?word_id=` + wordIDStr + `" 
				hx-swap="outerHTML"
				style="padding: 6px 12px; font-size: 12px; background: #28a745; color: white; border: none; border-radius: 15px; cursor: pointer; font-weight: 500;">
				✓ Known
			</button>`))
		} else {
			// Word is now unsuspended (not known)
			w.Write([]byte(`<button class="know-btn" 
				hx-post="/api/learn/toggle-suspended?word_id=` + wordIDStr + `" 
				hx-swap="outerHTML"
				style="padding: 6px 12px; font-size: 12px; background: #f0f0f0; color: #666; border: none; border-radius: 15px; cursor: pointer; font-weight: 500;">
				Know it
			</button>`))
		}
		return
	}

	// Non-HTMX request - redirect back
	http.Redirect(w, r, r.Header.Get("Referer"), http.StatusSeeOther)
}

// HandleSuspendAllOnPage suspends all words on the current page (marks them as "known")
func (h *LearnHandler) HandleSuspendAllOnPage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, err := h.auth.GetCurrentUser(r)
	if err != nil {
		http.Error(w, "User not authenticated", http.StatusUnauthorized)
		return
	}

	// Get level and page from query parameters
	levelStr := r.URL.Query().Get("level")
	pageStr := r.URL.Query().Get("page")

	level := 5
	page := 1
	batchSize := 10

	if levelStr != "" {
		if levelStr == "all" || levelStr == "0" {
			level = 0 // All levels mode
		} else if l, err := strconv.Atoi(levelStr); err == nil && l >= 1 && l <= 5 {
			level = l
		}
	}
	if pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p >= 1 {
			page = p
		}
	}

	offset := (page - 1) * batchSize

	// Get words on this page
	words, _, err := h.db.GetWordsForLearning(userID, level, batchSize, offset)
	if err != nil {
		http.Error(w, "Failed to get words: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Suspend all words that are in the SR deck
	suspendedCount := 0
	for _, word := range words {
		if word.IsLearned && !word.IsSuspended {
			_, err := h.db.SuspendWord(userID, word.ID)
			if err != nil {
				// Log but continue
				continue
			}
			suspendedCount++
		}
	}

	// Redirect back to the same page
	http.Redirect(w, r, r.Header.Get("Referer"), http.StatusSeeOther)
}
