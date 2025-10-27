package api

import (
	"fmt"
	"gaijin/internal/auth"
	"gaijin/internal/database"
	"net/http"
	"strconv"
	"strings"
)

// StudyHandler handles study-related API endpoints
type StudyHandler struct {
	db   *database.Database
	auth *auth.Auth
}

// NewStudyHandler creates a new study handler with database and auth dependencies
func NewStudyHandler(db *database.Database, auth *auth.Auth) *StudyHandler {
	return &StudyHandler{
		db:   db,
		auth: auth,
	}
}

func (h *StudyHandler) HandleAnswerPronunciation(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, err := h.auth.GetCurrentUser(r)
	if err != nil {
		http.Error(w, "User not authenticated", http.StatusUnauthorized)
		return
	}

	answer := r.FormValue("answer")
	if answer == "" {
		http.Error(w, "Answer is required", http.StatusBadRequest)
		return
	}

	srID, err := strconv.Atoi(r.FormValue("word-id"))
	if err != nil {
		http.Error(w, "Failed to parse word ID: "+err.Error(), http.StatusBadRequest)
		return
	}
	word, err := h.db.LookupWordBySRId(srID)
	if err != nil {
		http.Error(w, "Failed to lookup word by ID: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Validate answer against furigana (hiragana reading)
	isCorrect := strings.TrimSpace(answer) == strings.TrimSpace(word.Furigana)

	userSettings, err := h.db.GetUserSettings(userID)
	if err != nil {
		http.Error(w, "Failed to get user settings: "+err.Error(), http.StatusInternalServerError)
		return
	}

	time := r.FormValue("time")
	timeMs, err := strconv.Atoi(time)
	if err != nil {
		http.Error(w, "Failed to parse time: "+err.Error(), http.StatusBadRequest)
		return
	}
	knowIt := timeMs < userSettings.SRTimeJapanese

	// TODO: Implement SR update logic based on isCorrect and knowIt
	fmt.Printf("Reading answer: %s, correct: %v, expected: %s, time: %dms, knowIt: %v\n", answer, isCorrect, word.Furigana, timeMs, knowIt)

	// Redirect back to study page (will load next word)
	http.Redirect(w, r, "/study", http.StatusSeeOther)
}

func (h *StudyHandler) HandleAnswerMeaning(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, err := h.auth.GetCurrentUser(r)
	if err != nil {
		http.Error(w, "User not authenticated", http.StatusUnauthorized)
		return
	}

	answer := r.FormValue("answer")
	if answer == "" {
		http.Error(w, "Answer is required", http.StatusBadRequest)
		return
	}

	srID, err := strconv.Atoi(r.FormValue("word-id"))
	if err != nil {
		http.Error(w, "Failed to parse word ID: "+err.Error(), http.StatusBadRequest)
		return
	}
	word, err := h.db.LookupWordBySRId(srID)
	if err != nil {
		http.Error(w, "Failed to lookup word by ID: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Validate answer against definitions (case-insensitive, partial matching)
	answerLower := strings.ToLower(strings.TrimSpace(answer))
	definitionsLower := strings.ToLower(word.Definitions)

	// Check if the answer is contained in the definitions or vice versa
	isCorrect := strings.Contains(definitionsLower, answerLower) || strings.Contains(answerLower, definitionsLower)

	// Also check for comma-separated definitions
	if !isCorrect {
		definitions := strings.Split(definitionsLower, ",")
		for _, def := range definitions {
			def = strings.TrimSpace(def)
			if def == answerLower || strings.Contains(def, answerLower) || strings.Contains(answerLower, def) {
				isCorrect = true
				break
			}
		}
	}

	userSettings, err := h.db.GetUserSettings(userID)
	if err != nil {
		http.Error(w, "Failed to get user settings: "+err.Error(), http.StatusInternalServerError)
		return
	}

	time := r.FormValue("time")
	timeMs, err := strconv.Atoi(time)
	if err != nil {
		http.Error(w, "Failed to parse time: "+err.Error(), http.StatusBadRequest)
		return
	}
	knowIt := timeMs < userSettings.SRTimeEnglish

	// TODO: Implement SR update logic based on isCorrect and knowIt
	fmt.Printf("Meaning answer: %s, correct: %v, expected: %s, time: %dms, knowIt: %v\n", answer, isCorrect, word.Definitions, timeMs, knowIt)

	// Redirect back to study page (will load next word)
	http.Redirect(w, r, "/study", http.StatusSeeOther)
}
