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

	// Get return URL (default to /study if not provided)
	returnURL := r.FormValue("return-url")
	if returnURL == "" {
		returnURL = "/study"
	}

	// If correct AND fast (knowIt), auto-rate as 5 and move to next word
	if isCorrect && knowIt {
		err = h.db.UpdateSRWord(srID, 5)
		if err != nil {
			http.Error(w, "Failed to update SR: "+err.Error(), http.StatusInternalServerError)
			return
		}
		// Redirect back to study page (will load next word)
		http.Redirect(w, r, returnURL, http.StatusSeeOther)
		return
	}

	// Otherwise, redirect to answer page for manual rating
	// Pass whether answer was correct for styling purposes
	correctParam := "false"
	if isCorrect {
		correctParam = "true"
	}
	http.Redirect(w, r, fmt.Sprintf("/study/answer?sr_id=%d&type=pronunciation&correct=%s&answer=%s&return-url=%s", srID, correctParam, answer, returnURL), http.StatusSeeOther)
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

	// Get return URL (default to /study if not provided)
	returnURL := r.FormValue("return-url")
	if returnURL == "" {
		returnURL = "/study"
	}

	// If correct AND fast (knowIt), auto-rate as 5 and move to next word
	if isCorrect && knowIt {
		err = h.db.UpdateSRWord(srID, 5)
		if err != nil {
			http.Error(w, "Failed to update SR: "+err.Error(), http.StatusInternalServerError)
			return
		}
		// Redirect back to study page (will load next word)
		http.Redirect(w, r, returnURL, http.StatusSeeOther)
		return
	}

	// Otherwise, redirect to answer page for manual rating
	// Pass whether answer was correct for styling purposes
	correctParam := "false"
	if isCorrect {
		correctParam = "true"
	}
	http.Redirect(w, r, fmt.Sprintf("/study/answer?sr_id=%d&type=meaning&correct=%s&answer=%s&return-url=%s", srID, correctParam, answer, returnURL), http.StatusSeeOther)
}

// HandleSubmitRating handles the manual quality rating submission (0-5)
func (h *StudyHandler) HandleSubmitRating(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, err := h.auth.GetCurrentUser(r)
	if err != nil {
		http.Error(w, "User not authenticated", http.StatusUnauthorized)
		return
	}

	// Verify user owns this SR record
	srID, err := strconv.Atoi(r.FormValue("sr_id"))
	if err != nil {
		http.Error(w, "Failed to parse SR ID: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Verify ownership by checking if the SR record belongs to this user
	var ownerID int
	query := `SELECT user_id FROM sr WHERE id = $1`
	err = h.db.DB.QueryRow(query, srID).Scan(&ownerID)
	if err != nil {
		http.Error(w, "SR record not found: "+err.Error(), http.StatusNotFound)
		return
	}
	if ownerID != userID {
		http.Error(w, "Unauthorized access to SR record", http.StatusForbidden)
		return
	}

	// Parse quality rating
	quality, err := strconv.Atoi(r.FormValue("quality"))
	if err != nil {
		http.Error(w, "Failed to parse quality: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Update SR record with the rating
	err = h.db.UpdateSRWord(srID, quality)
	if err != nil {
		http.Error(w, "Failed to update SR: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Get return URL from query params (default to /study if not provided)
	returnURL := r.FormValue("return-url")
	if returnURL == "" {
		returnURL = "/study"
	}

	// Redirect back to study page (will load next word)
	http.Redirect(w, r, returnURL, http.StatusSeeOther)
}
