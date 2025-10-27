package handlers

import (
	"html/template"
	"net/http"
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
