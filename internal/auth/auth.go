package auth

import (
	"crypto/rand"
	"encoding/hex"
	"gaijin/internal/database"
	"net/http"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type Auth struct {
	db *database.Database
}

func New(db *database.Database) *Auth {
	return &Auth{db: db}
}

func (a *Auth) Middleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !a.IsAuthenticated(r) {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		next(w, r)
	}
}

func (a *Auth) IsAuthenticated(r *http.Request) bool {
	userID, err := a.GetCurrentUser(r)
	return err == nil && userID > 0
}

func (a *Auth) GetCurrentUser(r *http.Request) (int, error) {
	cookie, err := r.Cookie("session_token")
	if err != nil {
		return 0, err // No session cookie
	}

	return a.ValidateSession(cookie.Value)
}

func generateSessionToken() (string, error) {
	bytes := make([]byte, 32) // 256 bits
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func (a *Auth) CreateSession(userID int, userAgent, ipAddress string) (string, error) {
	token, err := generateSessionToken()
	if err != nil {
		return "", err
	}

	// Sessions expire in 7 days
	expiresAt := time.Now().Add(7 * 24 * time.Hour)

	_, err = a.db.DB.Exec(`
		INSERT INTO sessions (user_id, session_token, expires_at, user_agent, ip_address)
		VALUES ($1, $2, $3, $4, $5)`,
		userID, token, expiresAt, userAgent, ipAddress)

	if err != nil {
		return "", err
	}

	return token, nil
}

func (a *Auth) ValidateSession(token string) (int, error) {
	var userID int
	var expiresAt time.Time

	err := a.db.DB.QueryRow(`
		SELECT user_id, expires_at
		FROM sessions
		WHERE session_token = $1 AND expires_at > NOW()`,
		token).Scan(&userID, &expiresAt)

	if err != nil {
		return 0, err
	}

	return userID, nil
}

func (a *Auth) Logout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session_token")
	if err == nil {
		_, _ = a.db.DB.Exec("DELETE FROM sessions WHERE session_token = $1", cookie.Value)
	}

	clearCookie := &http.Cookie{
		Name:     "session_token",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1, // Expire immediately
	}
	http.SetCookie(w, clearCookie)
}

func hashPassword(password string) (string, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedBytes), nil
}

func (a *Auth) CheckPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func (a *Auth) CreateUser(username, password, email string) error {
	hashedPassword, err := hashPassword(password)
	if err != nil {
		return err
	}

	_, err = a.db.DB.Exec("INSERT INTO users (username, password_hash, email) VALUES ($1, $2, $3)",
		username, hashedPassword, email)
	return err
}
