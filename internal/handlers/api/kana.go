package api

import (
	"fmt"
	"gaijin/internal/auth"
	"gaijin/internal/database"
	"net/http"
	"strconv"
	"strings"
)

// KanaHandler handles kana study-related API endpoints
type KanaHandler struct {
	db   *database.Database
	auth *auth.Auth
}

// NewKanaHandler creates a new kana handler with database and auth dependencies
func NewKanaHandler(db *database.Database, auth *auth.Auth) *KanaHandler {
	return &KanaHandler{
		db:   db,
		auth: auth,
	}
}

// HandleAnswerKana handles romaji answer submission for kana study
func (h *KanaHandler) HandleAnswerKana(w http.ResponseWriter, r *http.Request) {
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

	srID, err := strconv.Atoi(r.FormValue("kana-id"))
	if err != nil {
		http.Error(w, "Failed to parse kana ID: "+err.Error(), http.StatusBadRequest)
		return
	}

	kana, kanaType, err := h.db.LookupKanaBySRId(srID)
	if err != nil {
		http.Error(w, "Failed to lookup kana by ID: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Validate answer against romaji (case-insensitive)
	answerLower := strings.ToLower(strings.TrimSpace(answer))
	romajiLower := strings.ToLower(strings.TrimSpace(kana.Romaji))
	isCorrect := answerLower == romajiLower

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
	// Use Japanese SR time for kana (typically faster recognition)
	knowIt := timeMs < userSettings.SRTimeJapanese

	// Get return URL (default based on kana type)
	returnURL := r.FormValue("return-url")
	if returnURL == "" {
		returnURL = "/study/" + kanaType
	}

	// If correct AND fast (knowIt), auto-rate as 5 and move to next kana
	if isCorrect && knowIt {
		err = h.db.UpdateSRKana(srID, 5)
		if err != nil {
			http.Error(w, "Failed to update SR kana: "+err.Error(), http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, returnURL, http.StatusSeeOther)
		return
	}

	// Otherwise redirect to answer page for manual rating
	redirectURL := fmt.Sprintf("/study/kana/answer?sr_id=%d&type=%s&correct=%t&answer=%s&return-url=%s",
		srID, kanaType, isCorrect, answer, returnURL)
	http.Redirect(w, r, redirectURL, http.StatusSeeOther)
}

// HandleSubmitKanaRating handles the submission of a quality rating for kana SR
func (h *KanaHandler) HandleSubmitKanaRating(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	_, err := h.auth.GetCurrentUser(r)
	if err != nil {
		http.Error(w, "User not authenticated", http.StatusUnauthorized)
		return
	}

	srIDStr := r.FormValue("sr_id")
	if srIDStr == "" {
		http.Error(w, "SR ID is required", http.StatusBadRequest)
		return
	}

	srID, err := strconv.Atoi(srIDStr)
	if err != nil {
		http.Error(w, "Invalid SR ID", http.StatusBadRequest)
		return
	}

	qualityStr := r.FormValue("quality")
	if qualityStr == "" {
		http.Error(w, "Quality is required", http.StatusBadRequest)
		return
	}

	quality, err := strconv.Atoi(qualityStr)
	if err != nil {
		http.Error(w, "Invalid quality rating", http.StatusBadRequest)
		return
	}

	// Update the SR kana record
	err = h.db.UpdateSRKana(srID, quality)
	if err != nil {
		http.Error(w, "Failed to update SR kana: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Get return URL (default to /study/hiragana if not provided)
	returnURL := r.FormValue("return-url")
	if returnURL == "" {
		returnURL = "/study/hiragana"
	}

	http.Redirect(w, r, returnURL, http.StatusSeeOther)
}

// HandleInitializeKana initializes all kana of a type for the user
func (h *KanaHandler) HandleInitializeKana(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, err := h.auth.GetCurrentUser(r)
	if err != nil {
		http.Error(w, "User not authenticated", http.StatusUnauthorized)
		return
	}

	kanaType := r.FormValue("kana_type")
	if kanaType != "hiragana" && kanaType != "katakana" {
		http.Error(w, "Invalid kana type", http.StatusBadRequest)
		return
	}

	// Initialize kana for the user
	err = h.db.InitializeUserSRKana(userID, kanaType)
	if err != nil {
		http.Error(w, "Failed to initialize kana: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Redirect back to the study page
	http.Redirect(w, r, "/study/"+kanaType, http.StatusSeeOther)
}
