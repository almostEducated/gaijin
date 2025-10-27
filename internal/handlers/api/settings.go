package api

import (
	"gaijin/internal/auth"
	"gaijin/internal/database"
	"net/http"
	"strconv"
)

// SettingsHandler handles user settings API endpoints
type SettingsHandler struct {
	db   *database.Database
	auth *auth.Auth
}

// NewSettingsHandler creates a new settings handler with database and auth dependencies
func NewSettingsHandler(db *database.Database, auth *auth.Auth) *SettingsHandler {
	return &SettingsHandler{
		db:   db,
		auth: auth,
	}
}

// HandleUpdateSettings handles POST requests to update user settings
func (h *SettingsHandler) HandleUpdateSettings(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get current user
	userID, err := h.auth.GetCurrentUser(r)
	if err != nil {
		http.Error(w, "User not authenticated", http.StatusUnauthorized)
		return
	}

	// Parse form values
	srTimeJapanese, err := strconv.Atoi(r.FormValue("sr_time_japanese"))
	if err != nil || srTimeJapanese < 0 {
		http.Error(w, "Invalid sr_time_japanese", http.StatusBadRequest)
		return
	}

	srTimeEnglish, err := strconv.Atoi(r.FormValue("sr_time_english"))
	if err != nil || srTimeEnglish < 0 {
		http.Error(w, "Invalid sr_time_english", http.StatusBadRequest)
		return
	}

	submitKey := r.FormValue("submit_key")
	if submitKey == "" {
		http.Error(w, "submit_key is required", http.StatusBadRequest)
		return
	}

	key1 := r.FormValue("key_1")
	key2 := r.FormValue("key_2")
	key3 := r.FormValue("key_3")
	key4 := r.FormValue("key_4")
	key5 := r.FormValue("key_5")

	// Validate all keys are present
	if key1 == "" || key2 == "" || key3 == "" || key4 == "" || key5 == "" {
		http.Error(w, "All rating keys (key_1 through key_5) are required", http.StatusBadRequest)
		return
	}

	// Parse show_hiragana_mostly checkbox (if not checked, FormValue returns empty string)
	showHiraganaMostly := r.FormValue("show_hiragana_mostly") == "on"

	// Update user settings
	settings := &database.UserSettings{
		UserID:             userID,
		SRTimeJapanese:     srTimeJapanese,
		SRTimeEnglish:      srTimeEnglish,
		SubmitKey:          submitKey,
		Key1:               key1,
		Key2:               key2,
		Key3:               key3,
		Key4:               key4,
		Key5:               key5,
		ShowHiraganaMostly: showHiraganaMostly,
	}

	err = h.db.UpdateUserSettings(userID, settings)
	if err != nil {
		http.Error(w, "Failed to update settings: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Redirect back to profile page with success message
	http.Redirect(w, r, "/profile?success=1", http.StatusSeeOther)
}
