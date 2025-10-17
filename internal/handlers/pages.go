package handlers

import (
	"gaijin/internal/auth"
	"gaijin/internal/database"
	"html/template"
	"net/http"
)

// PageHandler contains page-specific HTTP handlers
type PageHandler struct {
	db   *database.Database
	auth *auth.Auth
}

// NewPageHandler creates a new PageHandler with dependencies
func NewPageHandler(db *database.Database, auth *auth.Auth) *PageHandler {
	return &PageHandler{
		db:   db,
		auth: auth,
	}
}

var data = struct {
	Title string
}{
	Title: "Home",
}

type StudyData struct {
	Title       string
	KanjiWord   string
	Furigana    string
	Romaji      string
	Definitions string
	Answered    bool
	NoWords     bool // When user has no words due for review
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

func (h *PageHandler) HandleProfile(w http.ResponseWriter, r *http.Request) {

	tmpl, err := template.ParseFiles(
		"templates/layout/base.html",
		"templates/pages/profile.html",
	)
	if err != nil {
		http.Error(w, "Template error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	data.Title = "Profile"
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	err = tmpl.ExecuteTemplate(w, "base", data)
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

	// Handle POST request (user submitted answer)
	if r.Method == http.MethodPost {
		// Parse form data
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Failed to parse form", http.StatusBadRequest)
			return
		}

		// TODO: Process the user's answer and check correctness
		// For now, just show the answered view with the current word

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

		studyData := StudyData{
			Title:       "Study",
			KanjiWord:   srWord.Word.Word,
			Furigana:    srWord.Word.Furigana,
			Romaji:      srWord.Word.Romaji,
			Definitions: srWord.Word.Definitions,
			Answered:    true,
			NoWords:     false,
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		err = tmpl.ExecuteTemplate(w, "base", studyData)
		if err != nil {
			http.Error(w, "Template execution error: "+err.Error(), http.StatusInternalServerError)
			return
		}
		return
	}

	// Handle GET request (initial page load)
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
			Title:   "Study",
			NoWords: true,
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		err = tmpl.ExecuteTemplate(w, "base", studyData)
		if err != nil {
			http.Error(w, "Template execution error: "+err.Error(), http.StatusInternalServerError)
			return
		}
		return
	}

	studyData := StudyData{
		Title:       "Study",
		KanjiWord:   srWord.Word.Word,
		Furigana:    srWord.Word.Furigana,
		Romaji:      srWord.Word.Romaji,
		Definitions: srWord.Word.Definitions,
		Answered:    false,
		NoWords:     false,
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	// Execute the "base" template which will include the "content" template
	err = tmpl.ExecuteTemplate(w, "base", studyData)
	if err != nil {
		http.Error(w, "Template execution error: "+err.Error(), http.StatusInternalServerError)
		return
	}
}
