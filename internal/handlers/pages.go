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
	ReturnURL   string // URL to return to after answering (e.g., "/study" or "/study/adverbs")
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
	ReturnURL   string // URL to return to after rating (e.g., "/study" or "/study/adverbs")
	Key0        string // keyboard shortcut for rating 0
	Key1        string // keyboard shortcut for rating 1
	Key2        string // keyboard shortcut for rating 2
	Key3        string // keyboard shortcut for rating 3
	Key4        string // keyboard shortcut for rating 4
	Key5        string // keyboard shortcut for rating 5
}

type VisualConfusionData struct {
	Title    string
	NoPairs  bool
	Kanji1   string
	Kanji2   string
	Reading1 string
	Reading2 string
	Meaning1 string
	Meaning2 string
	Word1    string
	Word2    string
}

// KanaStudyData holds data for the kana study page
type KanaStudyData struct {
	Title            string
	SRKanaID         int
	Character        string
	Romaji           string
	Category         string
	Answered         bool
	NoKana           bool   // When user has no kana due for review
	NeverInitialized bool   // True if user has never added kana to their deck
	KanaType         string // "hiragana" or "katakana"
	ReturnURL        string // URL to return to after answering
}

// KanaAnswerData holds data for the kana answer/rating page
type KanaAnswerData struct {
	Title      string
	SRID       int
	Character  string
	Romaji     string
	Category   string
	KanaType   string // "hiragana" or "katakana"
	IsCorrect  bool   // whether the user's answer was correct
	UserAnswer string // the user's actual answer
	ReturnURL  string // URL to return to after rating
	Key0       string
	Key1       string
	Key2       string
	Key3       string
	Key4       string
	Key5       string
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
		ReturnURL:   "/study",
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	// Execute the "base" template which will include the "content" template
	err = tmpl.ExecuteTemplate(w, "base", studyData)
	if err != nil {
		http.Error(w, "Template execution error: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

// HandleStudyAdverbs handles the adverb-specific study page
func (h *PageHandler) HandleStudyAdverbs(w http.ResponseWriter, r *http.Request) {
	// Get current user
	userID, err := h.auth.GetCurrentUser(r)
	if err != nil {
		http.Error(w, "User not authenticated", http.StatusUnauthorized)
		return
	}

	// Get the next adverb word to study
	srWord, err := h.db.GetNextSRWordAdverbs(userID)
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
			Title:     "Study Adverbs",
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
		Title:       "Study Adverbs",
		SRWordID:    srWord.SRID,
		KanjiWord:   srWord.Word.Word,
		Furigana:    srWord.Word.Furigana,
		Romaji:      srWord.Word.Romaji,
		Definitions: srWord.Word.Definitions,
		Answered:    false,
		NoWords:     false,
		StudyMode:   studyMode,
		ReturnURL:   "/study/adverbs",
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

	// Get return URL (default to /study if not provided)
	returnURL := r.URL.Query().Get("return-url")
	if returnURL == "" {
		returnURL = "/study"
	}

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
		ReturnURL:   returnURL,
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

// HandleAbout shows the About page with the Way of Thinking content
func (h *PageHandler) HandleAbout(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles(
		"templates/layout/base.html",
		"templates/pages/about.html",
	)
	if err != nil {
		http.Error(w, "Template error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	data.Title = "About"
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	// Execute the "base" template which will include the "content" template
	err = tmpl.ExecuteTemplate(w, "base", data)
	if err != nil {
		http.Error(w, "Template execution error: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

// HandleStudyHiragana handles the hiragana study page
func (h *PageHandler) HandleStudyHiragana(w http.ResponseWriter, r *http.Request) {
	h.handleKanaStudy(w, r, "hiragana")
}

// HandleStudyKatakana handles the katakana study page
func (h *PageHandler) HandleStudyKatakana(w http.ResponseWriter, r *http.Request) {
	h.handleKanaStudy(w, r, "katakana")
}

// handleKanaStudy is the shared handler for both hiragana and katakana study
func (h *PageHandler) handleKanaStudy(w http.ResponseWriter, r *http.Request, kanaType string) {
	// Get current user
	userID, err := h.auth.GetCurrentUser(r)
	if err != nil {
		http.Error(w, "User not authenticated", http.StatusUnauthorized)
		return
	}

	// Get the next kana to study
	srKana, err := h.db.GetNextSRKana(userID, kanaType)
	if err != nil {
		http.Error(w, "Failed to get study kana: "+err.Error(), http.StatusInternalServerError)
		return
	}

	tmpl, err := template.ParseFiles(
		"templates/layout/base.html",
		"templates/pages/study_kana.html",
	)
	if err != nil {
		http.Error(w, "Template error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	title := "Study Hiragana"
	if kanaType == "katakana" {
		title = "Study Katakana"
	}

	// Check if there are no kana due for review
	if srKana == nil {
		// Check if user has never initialized kana
		hasKana, err := h.db.HasUserSRKana(userID, kanaType)
		if err != nil {
			http.Error(w, "Database error: "+err.Error(), http.StatusInternalServerError)
			return
		}

		studyData := KanaStudyData{
			Title:            title,
			NoKana:           true,
			NeverInitialized: !hasKana,
			KanaType:         kanaType,
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		err = tmpl.ExecuteTemplate(w, "base", studyData)
		if err != nil {
			http.Error(w, "Template execution error: "+err.Error(), http.StatusInternalServerError)
			return
		}
		return
	}

	studyData := KanaStudyData{
		Title:     title,
		SRKanaID:  srKana.SRID,
		Character: srKana.Kana.Character,
		Romaji:    srKana.Kana.Romaji,
		Category:  srKana.Kana.Category,
		Answered:  false,
		NoKana:    false,
		KanaType:  kanaType,
		ReturnURL: "/study/" + kanaType,
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	err = tmpl.ExecuteTemplate(w, "base", studyData)
	if err != nil {
		http.Error(w, "Template execution error: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

// HandleStudyKanaAnswer shows the answer page for kana with rating options
func (h *PageHandler) HandleStudyKanaAnswer(w http.ResponseWriter, r *http.Request) {
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

	// Get kana type
	kanaType := r.URL.Query().Get("type")
	if kanaType != "hiragana" && kanaType != "katakana" {
		http.Error(w, "Invalid kana type", http.StatusBadRequest)
		return
	}

	// Get whether answer was correct
	isCorrect := r.URL.Query().Get("correct") == "true"

	// Get the user's answer
	userAnswer := r.URL.Query().Get("answer")

	// Get return URL
	returnURL := r.URL.Query().Get("return-url")
	if returnURL == "" {
		returnURL = "/study/" + kanaType
	}

	// Get user settings for keyboard shortcuts
	userSettings, err := h.db.GetUserSettings(userID)
	if err != nil {
		http.Error(w, "Failed to get user settings: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Get the kana details
	kana, _, err := h.db.LookupKanaBySRId(srID)
	if err != nil {
		http.Error(w, "Failed to get kana: "+err.Error(), http.StatusInternalServerError)
		return
	}

	tmpl, err := template.ParseFiles(
		"templates/layout/base.html",
		"templates/pages/answer_kana.html",
	)
	if err != nil {
		http.Error(w, "Template error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	answerData := KanaAnswerData{
		Title:      "Answer",
		SRID:       srID,
		Character:  kana.Character,
		Romaji:     kana.Romaji,
		Category:   kana.Category,
		KanaType:   kanaType,
		IsCorrect:  isCorrect,
		UserAnswer: userAnswer,
		ReturnURL:  returnURL,
		Key0:       "0",
		Key1:       userSettings.Key1,
		Key2:       userSettings.Key2,
		Key3:       userSettings.Key3,
		Key4:       userSettings.Key4,
		Key5:       userSettings.Key5,
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	err = tmpl.ExecuteTemplate(w, "base", answerData)
	if err != nil {
		http.Error(w, "Template execution error: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

// HandleVisualConfusion shows a page for practicing visually similar kanji
func (h *PageHandler) HandleVisualConfusion(w http.ResponseWriter, r *http.Request) {
	// Get current user
	userID, err := h.auth.GetCurrentUser(r)
	if err != nil {
		http.Error(w, "User not authenticated", http.StatusUnauthorized)
		return
	}

	// Get a random kanji confusion pair for this user
	confusionPair, err := h.db.GetRandomKanjiConfusionPair(userID)
	if err != nil {
		http.Error(w, "Failed to get confusion pair: "+err.Error(), http.StatusInternalServerError)
		return
	}

	tmpl, err := template.ParseFiles(
		"templates/layout/base.html",
		"templates/pages/visual_confusion.html",
	)
	if err != nil {
		http.Error(w, "Template error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// If no pairs exist
	if confusionPair == nil {
		visualConfusionData := VisualConfusionData{
			Title:   "Visual Confusion Practice",
			NoPairs: true,
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		err = tmpl.ExecuteTemplate(w, "base", visualConfusionData)
		if err != nil {
			http.Error(w, "Template execution error: "+err.Error(), http.StatusInternalServerError)
			return
		}
		return
	}

	visualConfusionData := VisualConfusionData{
		Title:    "Visual Confusion Practice",
		NoPairs:  false,
		Kanji1:   confusionPair.Kanji1,
		Kanji2:   confusionPair.Kanji2,
		Reading1: confusionPair.Furigana1,
		Reading2: confusionPair.Furigana2,
		Meaning1: confusionPair.Definitions1,
		Meaning2: confusionPair.Definitions2,
		Word1:    confusionPair.Word1,
		Word2:    confusionPair.Word2,
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	err = tmpl.ExecuteTemplate(w, "base", visualConfusionData)
	if err != nil {
		http.Error(w, "Template execution error: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

// LevelInfo represents info about a JLPT level for the template
type LevelInfo struct {
	Level        int
	TotalCount   int
	LearnedCount int
	IsActive     bool
}

// LearnData holds data for the Learn page
type LearnData struct {
	Title           string
	Words           []database.LearnWord
	Level           int         // Current JLPT level being viewed (5 = N5)
	Page            int         // Current page number
	TotalPages      int         // Total pages for this level
	TotalWords      int         // Total words at this level
	LearnedCount    int         // How many words user has learned at this level
	Levels          []LevelInfo // Info for each level tab
	BatchSize       int         // Words per page
	ProgressPercent int         // Progress percentage for progress bar
	PrevPage        int         // Previous page number
	NextPage        int         // Next page number
}

// HandleLearn shows the Learn page where users can discover new words in batches
func (h *PageHandler) HandleLearn(w http.ResponseWriter, r *http.Request) {
	// Get current user
	userID, err := h.auth.GetCurrentUser(r)
	if err != nil {
		http.Error(w, "User not authenticated", http.StatusUnauthorized)
		return
	}

	// Get query parameters
	levelStr := r.URL.Query().Get("level")
	pageStr := r.URL.Query().Get("page")

	// Default to N5 (level 5) and page 1
	// level=0 means "all levels by frequency"
	level := 5
	page := 1
	batchSize := 10 // Words per page

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

	// Calculate offset
	offset := (page - 1) * batchSize

	// Get words for this level
	words, totalWords, err := h.db.GetWordsForLearning(userID, level, batchSize, offset)
	if err != nil {
		http.Error(w, "Failed to get words: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Calculate total pages
	totalPages := (totalWords + batchSize - 1) / batchSize
	if totalPages == 0 {
		totalPages = 1
	}

	// If requested page is beyond available pages, redirect to last page
	if page > totalPages {
		page = totalPages
		offset = (page - 1) * batchSize
		words, _, err = h.db.GetWordsForLearning(userID, level, batchSize, offset)
		if err != nil {
			http.Error(w, "Failed to get words: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}

	// Get level word counts
	levelCounts, err := h.db.GetLevelWordCounts()
	if err != nil {
		http.Error(w, "Failed to get level counts: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Get user's learned counts per level
	learnedCounts, err := h.db.GetUserLearnedCountByLevel(userID)
	if err != nil {
		http.Error(w, "Failed to get learned counts: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Count how many words user has learned at current level (or all levels if level=0)
	var learnedCount int
	if level == 0 {
		// Sum all learned counts for "all" mode
		for _, count := range learnedCounts {
			learnedCount += count
		}
	} else {
		learnedCount = learnedCounts[level]
	}

	// Calculate total words with frequency for "All" tab
	totalWithFrequency := 0
	totalLearnedAll := 0
	for _, count := range levelCounts {
		totalWithFrequency += count
	}
	for _, count := range learnedCounts {
		totalLearnedAll += count
	}

	// Build level info for tabs (N5 to N1, then All)
	var levels []LevelInfo
	for _, lvl := range []int{5, 4, 3, 2, 1} {
		levels = append(levels, LevelInfo{
			Level:        lvl,
			TotalCount:   levelCounts[lvl],
			LearnedCount: learnedCounts[lvl],
			IsActive:     lvl == level,
		})
	}
	// Add "All" tab (level=0)
	levels = append(levels, LevelInfo{
		Level:        0,
		TotalCount:   totalWithFrequency,
		LearnedCount: totalLearnedAll,
		IsActive:     level == 0,
	})

	// Calculate progress percentage
	progressPercent := 0
	if totalWords > 0 {
		progressPercent = (learnedCount * 100) / totalWords
	}

	tmpl, err := template.ParseFiles(
		"templates/layout/base.html",
		"templates/pages/learn.html",
	)
	if err != nil {
		http.Error(w, "Template error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	learnData := LearnData{
		Title:           "Learn",
		Words:           words,
		Level:           level,
		Page:            page,
		TotalPages:      totalPages,
		TotalWords:      totalWords,
		LearnedCount:    learnedCount,
		Levels:          levels,
		BatchSize:       batchSize,
		ProgressPercent: progressPercent,
		PrevPage:        page - 1,
		NextPage:        page + 1,
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	err = tmpl.ExecuteTemplate(w, "base", learnData)
	if err != nil {
		http.Error(w, "Template execution error: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

// WordsByLevel groups words by their JLPT level
type WordsByLevel struct {
	Level int
	Words []database.KanjiWord
}

// KanjiLookupData holds data for the kanji lookup page
type KanjiLookupData struct {
	Title        string
	Kanji        string
	WordsByLevel []WordsByLevel
	WordCount    int
	NoResults    bool
}

// HandleKanjiLookup shows all words containing a specific kanji
func (h *PageHandler) HandleKanjiLookup(w http.ResponseWriter, r *http.Request) {
	// Get the kanji from query parameter
	kanji := r.URL.Query().Get("kanji")
	if kanji == "" {
		http.Error(w, "Kanji parameter is required", http.StatusBadRequest)
		return
	}

	// Search for words containing this kanji
	words, err := h.db.GetWordsByKanji(kanji)
	if err != nil {
		http.Error(w, "Failed to search words: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Group words by level
	levelGroups := make(map[int][]database.KanjiWord)
	for _, word := range words {
		levelGroups[word.Level] = append(levelGroups[word.Level], word)
	}

	// Convert to ordered slice (N5 first, then N4, etc.)
	var wordsByLevel []WordsByLevel
	for _, lvl := range []int{5, 4, 3, 2, 1} {
		if wordsAtLevel, ok := levelGroups[lvl]; ok {
			wordsByLevel = append(wordsByLevel, WordsByLevel{
				Level: lvl,
				Words: wordsAtLevel,
			})
		}
	}

	tmpl, err := template.ParseFiles(
		"templates/layout/base.html",
		"templates/pages/kanji_lookup.html",
	)
	if err != nil {
		http.Error(w, "Template error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	lookupData := KanjiLookupData{
		Title:        "Words with " + kanji,
		Kanji:        kanji,
		WordsByLevel: wordsByLevel,
		WordCount:    len(words),
		NoResults:    len(words) == 0,
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	err = tmpl.ExecuteTemplate(w, "base", lookupData)
	if err != nil {
		http.Error(w, "Template execution error: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

// SearchResultsData holds data for the search results page
type SearchResultsData struct {
	Title        string
	Query        string
	SearchType   string // "japanese" or "english"
	WordsByLevel []WordsByLevel
	WordCount    int
	NoResults    bool
}

// HandleSearch shows search results for words
func (h *PageHandler) HandleSearch(w http.ResponseWriter, r *http.Request) {
	// Get the search query
	query := r.URL.Query().Get("q")
	if query == "" {
		// If no query, redirect to learn page
		http.Redirect(w, r, "/learn", http.StatusSeeOther)
		return
	}

	// Search for words
	words, searchType, err := h.db.SearchWords(query)
	if err != nil {
		http.Error(w, "Failed to search words: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Group words by level
	levelGroups := make(map[int][]database.KanjiWord)
	for _, word := range words {
		levelGroups[word.Level] = append(levelGroups[word.Level], word)
	}

	// Convert to ordered slice (N5 first, then N4, etc.)
	var wordsByLevel []WordsByLevel
	for _, lvl := range []int{5, 4, 3, 2, 1} {
		if wordsAtLevel, ok := levelGroups[lvl]; ok {
			wordsByLevel = append(wordsByLevel, WordsByLevel{
				Level: lvl,
				Words: wordsAtLevel,
			})
		}
	}

	tmpl, err := template.ParseFiles(
		"templates/layout/base.html",
		"templates/pages/search_results.html",
	)
	if err != nil {
		http.Error(w, "Template error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	searchData := SearchResultsData{
		Title:        "Search: " + query,
		Query:        query,
		SearchType:   searchType,
		WordsByLevel: wordsByLevel,
		WordCount:    len(words),
		NoResults:    len(words) == 0,
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	err = tmpl.ExecuteTemplate(w, "base", searchData)
	if err != nil {
		http.Error(w, "Template execution error: "+err.Error(), http.StatusInternalServerError)
		return
	}
}
