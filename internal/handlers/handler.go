package handlers

import (
	"gaijin/internal/auth"
	"gaijin/internal/database"
)

// PageHandler handles HTML page rendering endpoints
type PageHandler struct {
	db   *database.Database
	auth *auth.Auth
}

// NewPageHandler creates a new page handler with database and auth dependencies
func NewPageHandler(db *database.Database, auth *auth.Auth) *PageHandler {
	return &PageHandler{
		db:   db,
		auth: auth,
	}
}
