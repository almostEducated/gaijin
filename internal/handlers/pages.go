package handlers

import (
	"gaijin/internal/database"
	"html/template"
	"net/http"
	"strconv"
)

var data = struct {
	Title string
}{
	Title: "Home",
}

type StudyData struct {
	Title       string
	SRWordID    int
	KanjiWord   string
	Furigana    string
	Romaji      string
	Definitions string
	Answered    bool
	NoWords     bool   // When user has no words due for review
	StudyMode   string // "reading" or "meaning"
}

type AnswerData struct {
	Title       string
	SRID        int
	KanjiWord   string
	Furigana    string
	Definitions string
	Type        string // "pronunciation" or "meaning"
	IsCorrect   bool   // whether the user's answer was correct
	UserAnswer  string // the user's actual answer
	Key0        string // keyboard shortcut for rating 0
	Key1        string // keyboard shortcut for rating 1
	Key2        string // keyboard shortcut for rating 2
	Key3        string // keyboard shortcut for rating 3
	Key4        string // keyboard shortcut for rating 4
	Key5        string // keyboard shortcut for rating 5
}

// MAYBE rename dashboard
func (h *PageHandler) HandleHome(w http.ResponseWriter, r *http.Request) {
	// Parse both the base layout and the page content
	tmpl, err := template.ParseFiles(
		"templates/layout/base.html",
		"templates/pages/home.html",
	)
	if err != nil {
		http.Error(w, "Template error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	data.Title = "Home"
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	// Execute the "base" template which will include the "content" template
	err = tmpl.ExecuteTemplate(w, "base", data)
	if err != nil {
		http.Error(w, "Template execution error: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

type ProfileData struct {
	Title        string
	UserInfo     *database.UserInfo
	UserSettings *database.UserSettings
	Success      bool // for showing success message after saving
}

func (h *PageHandler) HandleProfile(w http.ResponseWriter, r *http.Request) {
	// Get current user
	userID, err := h.auth.GetCurrentUser(r)
	if err != nil {
		http.Error(w, "User not authenticated", http.StatusUnauthorized)
		return
	}

	// Get user info
	userInfo, err := h.db.GetUserInfo(userID)
	if err != nil {
		http.Error(w, "Failed to get user info: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Get user settings
	userSettings, err := h.db.GetUserSettings(userID)
	if err != nil {
		http.Error(w, "Failed to get user settings: "+err.Error(), http.StatusInternalServerError)
		return
	}

	tmpl, err := template.ParseFiles(
		"templates/layout/base.html",
		"templates/pages/profile.html",
	)
	if err != nil {
		http.Error(w, "Template error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Check for success parameter
	success := r.URL.Query().Get("success") == "1"

	profileData := ProfileData{
		Title:        "Profile",
		UserInfo:     userInfo,
		UserSettings: userSettings,
		Success:      success,
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	err = tmpl.ExecuteTemplate(w, "base", profileData)
	if err != nil {
		http.Error(w, "Template error: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *PageHandler) HandleStudy(w http.ResponseWriter, r *http.Request) {
	// Get current user
	userID, err := h.auth.GetCurrentUser(r)
	if err != nil {
		http.Error(w, "User not authenticated", http.StatusUnauthorized)
		return
	}

	// Check if user has any SR words
	hasWords, err := h.db.HasUserSRWords(userID)
	if err != nil {
		http.Error(w, "Database error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// If user has no SR words, initialize with Level 5
	if !hasWords {
		err = h.db.InitializeUserSRWords(userID, 5)
		if err != nil {
			http.Error(w, "Failed to initialize study words: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}

	// Get the next word to study
	srWord, err := h.db.GetNextSRWord(userID)
	if err != nil {
		http.Error(w, "Failed to get study word: "+err.Error(), http.StatusInternalServerError)
		return
	}

	tmpl, err := template.ParseFiles(
		"templates/layout/base.html",
		"templates/pages/study.html",
	)
	if err != nil {
		http.Error(w, "Template error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Check if there are no words due for review
	if srWord == nil {
		studyData := StudyData{
			Title:     "Study",
			NoWords:   true,
			StudyMode: "",
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		err = tmpl.ExecuteTemplate(w, "base", studyData)
		if err != nil {
			http.Error(w, "Template execution error: "+err.Error(), http.StatusInternalServerError)
			return
		}
		return
	}

	// Map SR type to study mode
	// "japanese pronunciation" -> "reading"
	// "english meaning" -> "meaning"
	studyMode := "reading"
	if srWord.Type == "english meaning" {
		studyMode = "meaning"
	}

	studyData := StudyData{
		Title:       "Study",
		SRWordID:    srWord.SRID,
		KanjiWord:   srWord.Word.Word,
		Furigana:    srWord.Word.Furigana,
		Romaji:      srWord.Word.Romaji,
		Definitions: srWord.Word.Definitions,
		Answered:    false,
		NoWords:     false,
		StudyMode:   studyMode,
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	// Execute the "base" template which will include the "content" template
	err = tmpl.ExecuteTemplate(w, "base", studyData)
	if err != nil {
		http.Error(w, "Template execution error: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

// HandleStudyAnswer shows the answer page with the correct answer and rating options
func (h *PageHandler) HandleStudyAnswer(w http.ResponseWriter, r *http.Request) {
	// Get current user
	userID, err := h.auth.GetCurrentUser(r)
	if err != nil {
		http.Error(w, "User not authenticated", http.StatusUnauthorized)
		return
	}

	// Get SR ID from query params
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

	// Get type (pronunciation or meaning)
	studyType := r.URL.Query().Get("type")
	if studyType != "pronunciation" && studyType != "meaning" {
		http.Error(w, "Invalid type", http.StatusBadRequest)
		return
	}

	// Get whether answer was correct
	isCorrect := r.URL.Query().Get("correct") == "true"

	// Get the user's answer
	userAnswer := r.URL.Query().Get("answer")

	// Get user settings for keyboard shortcuts
	userSettings, err := h.db.GetUserSettings(userID)
	if err != nil {
		http.Error(w, "Failed to get user settings: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Verify user owns this SR record and get the word
	var ownerID int
	var wordID int
	query := `SELECT user_id, word_id FROM sr WHERE id = $1`
	err = h.db.DB.QueryRow(query, srID).Scan(&ownerID, &wordID)
	if err != nil {
		http.Error(w, "SR record not found: "+err.Error(), http.StatusNotFound)
		return
	}
	if ownerID != userID {
		http.Error(w, "Unauthorized access to SR record", http.StatusForbidden)
		return
	}

	// Get the word details
	word, err := h.db.LookupWordBySRId(srID)
	if err != nil {
		http.Error(w, "Failed to get word: "+err.Error(), http.StatusInternalServerError)
		return
	}

	tmpl, err := template.ParseFiles(
		"templates/layout/base.html",
		"templates/pages/answer.html",
	)
	if err != nil {
		http.Error(w, "Template error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	answerData := AnswerData{
		Title:       "Answer",
		SRID:        srID,
		KanjiWord:   word.Word,
		Furigana:    word.Furigana,
		Definitions: word.Definitions,
		Type:        studyType,
		IsCorrect:   isCorrect,
		UserAnswer:  userAnswer,
		Key0:        "0", // Default for now, can be made configurable later
		Key1:        userSettings.Key1,
		Key2:        userSettings.Key2,
		Key3:        userSettings.Key3,
		Key4:        userSettings.Key4,
		Key5:        userSettings.Key5,
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	err = tmpl.ExecuteTemplate(w, "base", answerData)
	if err != nil {
		http.Error(w, "Template execution error: "+err.Error(), http.StatusInternalServerError)
		return
	}
}
