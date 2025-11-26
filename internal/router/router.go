package router

import (
	"gaijin/internal/auth"
	"gaijin/internal/database"
	"gaijin/internal/handlers"
	"gaijin/internal/handlers/api"
	"log"
	"net/http"
	"os"
	"time"
)

type Router struct {
	Mux                   *http.ServeMux
	DB                    *database.Database
	auth                  *auth.Auth
	logger                *Logger
	authHandler           *handlers.AuthHandler
	pageHandler           *handlers.PageHandler
	studyHandler          *api.StudyHandler
	jlptHandler           *api.JLPTHandler
	settingsHandler       *api.SettingsHandler
	verbHandler           *api.VerbHandler
	kanjiConfusionHandler *api.KanjiConfusionHandler
}

func New(db *database.Database) *Router {
	authService := auth.New(db)
	return &Router{
		Mux:                   http.NewServeMux(),
		auth:                  authService,
		logger:                NewLogger(),
		authHandler:           handlers.NewAuthHandler(db, authService),
		pageHandler:           handlers.NewPageHandler(db, authService),
		studyHandler:          api.NewStudyHandler(db, authService),
		jlptHandler:           api.NewJLPTHandler(db),
		settingsHandler:       api.NewSettingsHandler(db, authService),
		verbHandler:           api.NewVerbHandler(db),
		kanjiConfusionHandler: api.NewKanjiConfusionHandler(db, authService),
	}
}

// Logger provides request logging middleware
type Logger struct {
	enabled bool
}

// NewLogger creates a new logger instance
func NewLogger() *Logger {
	// Enable logging in development mode
	enabled := os.Getenv("GO_ENV") != "production"
	return &Logger{enabled: enabled}
}

// Middleware logs HTTP requests
func (l *Logger) Middleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		if l.enabled {
			log.Printf("[REQUEST] %s %s from %s", r.Method, r.URL.Path, r.RemoteAddr)
		}

		rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		next(rw, r)

		duration := time.Since(start)
		if l.enabled {
			log.Printf("[RESPONSE] %s %s - %d (%v)", r.Method, r.URL.Path, rw.statusCode, duration)
			if rw.statusCode == 500 {
				log.Printf("[ERROR] %s %s from %s - %s", r.Method, r.URL.Path, r.RemoteAddr, string(rw.body))
			}
		}
	}
}

// responseWriter wraps http.ResponseWriter to capture status code and body
type responseWriter struct {
	http.ResponseWriter
	statusCode int
	body       []byte
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(data []byte) (int, error) {
	// Capture the response body for error logging
	rw.body = append(rw.body, data...)
	return rw.ResponseWriter.Write(data)
}

func (r *Router) SetupRoutes() {
	// Public routes (no authentication required)
	r.Mux.HandleFunc("/", r.logger.Middleware(r.auth.Middleware(r.pageHandler.HandleHome)))

	// Auth routes
	r.Mux.HandleFunc("/login", r.logger.Middleware(r.authHandler.HandleLogin))
	r.Mux.HandleFunc("/register", r.logger.Middleware(r.authHandler.HandleRegister))
	r.Mux.HandleFunc("/logout", r.logger.Middleware(r.authHandler.HandleLogout))

	// Page routes
	r.Mux.HandleFunc("/study", r.logger.Middleware(r.auth.Middleware(r.pageHandler.HandleStudy)))
	r.Mux.HandleFunc("/study/adverbs", r.logger.Middleware(r.auth.Middleware(r.pageHandler.HandleStudyAdverbs)))
	r.Mux.HandleFunc("/visual-confusion", r.logger.Middleware(r.auth.Middleware(r.pageHandler.HandleVisualConfusion)))
	r.Mux.HandleFunc("/profile", r.logger.Middleware(r.auth.Middleware(r.pageHandler.HandleProfile)))

	// Study routes
	r.Mux.HandleFunc("/answer/pronunciation", r.logger.Middleware(r.auth.Middleware(r.studyHandler.HandleAnswerPronunciation)))
	r.Mux.HandleFunc("/answer/meaning", r.logger.Middleware(r.auth.Middleware(r.studyHandler.HandleAnswerMeaning)))
	r.Mux.HandleFunc("/study/answer", r.logger.Middleware(r.auth.Middleware(r.pageHandler.HandleStudyAnswer)))
	r.Mux.HandleFunc("/study/rate", r.logger.Middleware(r.auth.Middleware(r.studyHandler.HandleSubmitRating)))

	// Settings routes
	r.Mux.HandleFunc("/api/settings", r.logger.Middleware(r.auth.Middleware(r.settingsHandler.HandleUpdateSettings)))

	// Kanji confusion routes
	r.Mux.HandleFunc("/api/similar-kanji", r.logger.Middleware(r.auth.Middleware(r.kanjiConfusionHandler.HandleGetSimilarKanji)))
	r.Mux.HandleFunc("/api/link-kanji", r.logger.Middleware(r.auth.Middleware(r.kanjiConfusionHandler.HandleLinkKanji)))

	// Verb conjugation routes (public - no auth required)
	r.Mux.HandleFunc("/api/verb/conjugate", r.logger.Middleware(r.verbHandler.HandleConjugate))

	// Static files - no logging for performance (optional)
	r.Mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
}
