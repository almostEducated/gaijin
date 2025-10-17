package server

import (
	"context"
	"fmt"
	"gaijin/internal/database"
	"gaijin/internal/router"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type Server struct {
	router *router.Router
	port   string
	server *http.Server
}

func New(db *database.Database) *Server {
	return &Server{
		router: router.New(db),
		port:   "localhost:8080",
	}
}

func (s *Server) Start() error {
	// Setup routes
	s.router.SetupRoutes()

	// Create HTTP server with proper configuration
	s.server = &http.Server{
		Addr:         s.port,
		Handler:      s.router.Mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Channel to listen for interrupt or terminate signals
	done := make(chan bool, 1)
	quit := make(chan os.Signal, 1)

	// Register signals to listen for
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	// Start server in a goroutine
	go func() {
		fmt.Printf("Server starting on http://localhost%s\n", s.port)
		fmt.Println("Press Ctrl+C to stop the server gracefully")

		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	// Wait for interrupt signal
	<-quit
	log.Println("Server is shutting down...")

	// Create context with timeout for graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Attempt graceful shutdown
	if err := s.server.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
		return err
	}

	close(done)
	log.Println("Server exited gracefully")
	return nil
}
