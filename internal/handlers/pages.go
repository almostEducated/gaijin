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
