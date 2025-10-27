package handlers

import (
	"gaijin/internal/auth"
	"gaijin/internal/database"
	"html/template"
	"net/http"
	"regexp"
	"strings"
)

type AuthHandler struct {
	db   *database.Database
	auth *auth.Auth
}

func NewAuthHandler(db *database.Database, auth *auth.Auth) *AuthHandler {
	return &AuthHandler{
		db:   db,
		auth: auth,
	}
}

func (h *AuthHandler) HandleLogin(w http.ResponseWriter, r *http.Request) {
	// Handle POST request (form submission)
	if r.Method == http.MethodPost {
		username := r.FormValue("username")
		password := r.FormValue("password")
		if username != "" && password != "" {
			// Query database for user authentication - get both id and password hash
			rows, err := h.db.Query("SELECT id, password_hash FROM users WHERE username = $1", username)
			if err != nil {
				http.Error(w, "Database error", http.StatusInternalServerError)
				return
			}
			defer rows.Close()

			if rows.Next() {
				var userID int
				var storedHash string
				if err := rows.Scan(&userID, &storedHash); err != nil {
					http.Error(w, "Database error", http.StatusInternalServerError)
					return
				}

				// Verify password against stored hash (using auth method)
				if h.auth.CheckPassword(password, storedHash) {
					// Create session using auth method
					userAgent := r.Header.Get("User-Agent")
					ipAddress := r.RemoteAddr

					// Handle X-Forwarded-For header for proxies
					if forwardedFor := r.Header.Get("X-Forwarded-For"); forwardedFor != "" {
						ipAddress = forwardedFor
					}

					token, err := h.auth.CreateSession(userID, userAgent, ipAddress)
					if err != nil {
						http.Error(w, "Failed to create session", http.StatusInternalServerError)
						return
					}

					// Set session cookie
					cookie := &http.Cookie{
						Name:     "session_token",
						Value:    token,
						Path:     "/",
						HttpOnly: true,                 // Prevent XSS attacks
						Secure:   false,                // Set to true in production with HTTPS
						SameSite: http.SameSiteLaxMode, // CSRF protection
						MaxAge:   7 * 24 * 60 * 60,     // 7 days in seconds
					}
					http.SetCookie(w, cookie)

					// Successful login - redirect to home
					http.Redirect(w, r, "/", http.StatusSeeOther)
					return
				}
			}

			// Invalid credentials - show login page with error
			tmpl, err := template.ParseFiles("templates/layout/login.html")
			if err != nil {
				http.Error(w, "Template error", http.StatusInternalServerError)
				return
			}

			data := struct {
				Title        string
				Error        string
				ShowRegister bool
			}{
				Title:        "Login",
				Error:        "Invalid username or password",
				ShowRegister: false,
			}
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			tmpl.ExecuteTemplate(w, "login", data)
			return
		} else {
			// Empty username or password - show login page with error
			tmpl, err := template.ParseFiles("templates/layout/login.html")
			if err != nil {
				http.Error(w, "Template error", http.StatusInternalServerError)
				return
			}

			data := struct {
				Title        string
				Error        string
				ShowRegister bool
			}{
				Title:        "Login",
				Error:        "Please enter both username and password",
				ShowRegister: false,
			}
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			tmpl.ExecuteTemplate(w, "login", data)
			return
		}
	}

	// Handle GET request (show login form)
	tmpl, err := template.ParseFiles("templates/layout/login.html")
	if err != nil {
		http.Error(w, "Template error", http.StatusInternalServerError)
		return
	}

	data := struct {
		Title           string
		Error           string
		RegisterError   string
		RegisterSuccess string
		ShowRegister    bool
	}{
		Title:           "Login",
		Error:           "",
		RegisterError:   "",
		RegisterSuccess: "",
		ShowRegister:    false,
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	tmpl.ExecuteTemplate(w, "login", data)
}

func (h *AuthHandler) HandleRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		username := r.FormValue("username")
		password := r.FormValue("password")
		confirmPassword := r.FormValue("confirm")
		email := r.FormValue("email")

		// Validation
		if username == "" || password == "" || email == "" {
			h.showRegisterForm(w, "", "All fields are required")
			return
		}

		// Validate email format
		email = strings.TrimSpace(strings.ToLower(email))
		emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
		if !emailRegex.MatchString(email) {
			h.showRegisterForm(w, "", "Please enter a valid email address")
			return
		}

		if password != confirmPassword {
			h.showRegisterForm(w, "", "Passwords do not match")
			return
		}

		if len(password) < 6 {
			h.showRegisterForm(w, "", "Password must be at least 6 characters long")
			return
		}

		// Check if username already exists
		rows, err := h.db.Query("SELECT COUNT(*) FROM users WHERE username = $1", username)
		if err != nil {
			h.showRegisterForm(w, "", "Database error")
			return
		}
		defer rows.Close()

		var count int
		if rows.Next() {
			if err := rows.Scan(&count); err != nil {
				h.showRegisterForm(w, "", "Database error")
				return
			}
		}

		if count > 0 {
			h.showRegisterForm(w, "", "Username already exists")
			return
		}

		// Check if email already exists
		rows2, err := h.db.Query("SELECT COUNT(*) FROM users WHERE email = $1", email)
		if err != nil {
			h.showRegisterForm(w, "", "Database error")
			return
		}
		defer rows2.Close()

		var emailCount int
		if rows2.Next() {
			if err := rows2.Scan(&emailCount); err != nil {
				h.showRegisterForm(w, "", "Database error")
				return
			}
		}

		if emailCount > 0 {
			h.showRegisterForm(w, "", "Email address is already registered")
			return
		}

		// Create the user using auth method
		err = h.auth.CreateUser(username, password, email)
		if err != nil {
			h.showRegisterForm(w, "", "Failed to create user")
			return
		}

		// Success - show success message and redirect to login
		h.showRegisterForm(w, "Account created successfully! You can now log in.", "")
		return
	}

	// GET request - show registration form
	h.showRegisterForm(w, "", "")
}

func (h *AuthHandler) showRegisterForm(w http.ResponseWriter, success, error string) {
	tmpl, err := template.ParseFiles("templates/layout/login.html")
	if err != nil {
		http.Error(w, "Template error", http.StatusInternalServerError)
		return
	}

	data := struct {
		Title           string
		Error           string
		RegisterError   string
		RegisterSuccess string
		ShowRegister    bool
	}{
		Title:           "Login",
		Error:           "",
		RegisterError:   error,
		RegisterSuccess: success,
		ShowRegister:    true,
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	tmpl.ExecuteTemplate(w, "login", data)
}

func (h *AuthHandler) HandleLogout(w http.ResponseWriter, r *http.Request) {
	// Use auth logout method
	h.auth.Logout(w, r)

	// Redirect to login page
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}
