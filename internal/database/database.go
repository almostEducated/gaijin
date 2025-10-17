package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type Database struct {
	DB *sql.DB
}

func connect(connectionString string) (*Database, error) {
	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return &Database{DB: db}, nil
}

func ConnectDB() (*Database, error) {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	user := os.Getenv("PG_USER")
	password := os.Getenv("PG_PASSWORD")
	host := os.Getenv("PG_HOST")
	database := os.Getenv("PG_DATABASE")
	port := os.Getenv("PG_PORT")

	defaultConnStr := fmt.Sprintf("postgres://%s:%s@%s:%s/postgres?sslmode=disable&client_encoding=UTF8", user, password, host, port)
	defaultDB, err := connect(defaultConnStr)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to default postgres database: %w", err)
	}

	var exists bool
	query := "SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = $1)"
	err = defaultDB.DB.QueryRow(query, database).Scan(&exists)
	if err != nil {
		defaultDB.Close()
		return nil, fmt.Errorf("failed to check database existence: %w", err)
	}

	if !exists {
		log.Printf("Database '%s' does not exist. Creating it...", database)
		_, err = defaultDB.DB.Exec(fmt.Sprintf("CREATE DATABASE %s WITH ENCODING 'UTF8' LC_COLLATE='C' LC_CTYPE='C' TEMPLATE=template0", database))
		if err != nil {
			defaultDB.Close()
			return nil, fmt.Errorf("failed to create database: %w", err)
		}
		log.Printf("✅ Database '%s' created successfully with UTF8 encoding", database)
	}

	defaultDB.Close()

	connectionString := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable&client_encoding=UTF8", user, password, host, port, database)
	db, err := connect(connectionString)
	if err != nil {
		return nil, err
	}

	log.Printf("✅ Successfully connected to database: %s@%s:%s", database, host, port)
	return db, nil
}

func (db *Database) Close() error {
	return db.DB.Close()
}

func (db *Database) Query(query string, args ...interface{}) (*sql.Rows, error) {
	return db.DB.Query(query, args...)
}

// InitializeTables creates all necessary database tables
func (db *Database) InitializeTables() error {
	if db.DB == nil {
		return fmt.Errorf("database connection is nil")
	}
	createSessionsTable := `
	CREATE TABLE IF NOT EXISTS sessions (
		id SERIAL PRIMARY KEY,
		user_id INTEGER NOT NULL,
		session_token VARCHAR(255) NOT NULL,
		expires_at TIMESTAMP NOT NULL,
		user_agent VARCHAR(255) NOT NULL,
		ip_address VARCHAR(255) NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	)
	`
	// Create a users table
	createUsersTable := `
	CREATE TABLE IF NOT EXISTS users (
		id SERIAL PRIMARY KEY,
		username VARCHAR(50) NOT NULL UNIQUE,
		password_hash VARCHAR(255) NOT NULL,
		email VARCHAR(255) NOT NULL UNIQUE,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);`

	createSRTable := `
	CREATE TABLE IF NOT EXISTS sr (
		id SERIAL PRIMARY KEY,
		user_id INTEGER NOT NULL,
		word_id INTEGER NOT NULL,
		repetitions INTEGER DEFAULT 0,
		ef FLOAT DEFAULT 2.5,
		interval INTEGER DEFAULT 0,
		last_reviewed TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		next_review TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		note TEXT,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);`

	// Create a kanji table
	createKanjiTable := `
	CREATE TABLE IF NOT EXISTS kanji (
		id SERIAL PRIMARY KEY,
		character VARCHAR(10) NOT NULL UNIQUE,
		meaning VARCHAR(255) NOT NULL,
		reading VARCHAR(255) NOT NULL,
		stroke_count INTEGER NOT NULL,
		grade_level INTEGER,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);`

	// Create a grammar table
	createGrammarTable := `
	CREATE TABLE IF NOT EXISTS grammar (
		id SERIAL PRIMARY KEY,
		pattern VARCHAR(255) NOT NULL,
		explanation TEXT NOT NULL,
		example_japanese TEXT NOT NULL,
		example_english TEXT NOT NULL,
		level VARCHAR(50) NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);`

	// Create JLPT vocabulary table
	createJLPTTable := `
	CREATE TABLE IF NOT EXISTS jlpt_vocabulary (
		id SERIAL PRIMARY KEY,
		word VARCHAR(255) NOT NULL,
		meaning TEXT NOT NULL,
		furigana VARCHAR(255),
		romaji VARCHAR(255),
		level INTEGER NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		UNIQUE(word, level)
	);`

	// Main words table - stores all word information
	// definitions and parts_of_speech are stored as semicolon-separated strings
	createWordsTable := `
	CREATE TABLE IF NOT EXISTS words (
		id SERIAL PRIMARY KEY,
		word VARCHAR(255) NOT NULL,
		furigana VARCHAR(255),
		romaji VARCHAR(255),
		level INTEGER NOT NULL,
		definitions TEXT,
		parts_of_speech TEXT,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		UNIQUE(word, level)
	);`

	// Execute table creation
	_, err := db.DB.Exec(createSessionsTable)
	if err != nil {
		return fmt.Errorf("error creating sessions table: %w", err)
	}
	_, err = db.DB.Exec(createSRTable)
	if err != nil {
		return fmt.Errorf("error creating sr table: %w", err)
	}
	_, err = db.DB.Exec(createKanjiTable)
	if err != nil {
		return fmt.Errorf("error creating kanji table: %w", err)
	}
	_, err = db.DB.Exec(createGrammarTable)
	if err != nil {
		return fmt.Errorf("error creating grammar table: %w", err)
	}
	_, err = db.DB.Exec(createUsersTable)
	if err != nil {
		return fmt.Errorf("error creating users table: %w", err)
	}
	_, err = db.DB.Exec(createJLPTTable)
	if err != nil {
		return fmt.Errorf("error creating JLPT vocabulary table: %w", err)
	}
	_, err = db.DB.Exec(createWordsTable)
	if err != nil {
		return fmt.Errorf("error creating words table: %w", err)
	}
	log.Println("All tables created successfully")

	return nil
}

// SR (Spaced Repetition) Operations

// HasUserSRWords checks if a user has any words in their SR table
func (db *Database) HasUserSRWords(userID int) (bool, error) {
	var count int
	query := `SELECT COUNT(*) FROM sr WHERE user_id = $1`
	err := db.DB.QueryRow(query, userID).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check user SR words: %w", err)
	}
	return count > 0, nil
}

// InitializeUserSRWords populates SR table with all words from a specific level for a user
func (db *Database) InitializeUserSRWords(userID int, level int) error {
	query := `
		INSERT INTO sr (user_id, word_id, repetitions, ef, interval, last_reviewed, next_review)
		SELECT $1, id, 0, 2.5, 0, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
		FROM words
		WHERE level = $2
		ON CONFLICT DO NOTHING
	`
	_, err := db.DB.Exec(query, userID, level)
	if err != nil {
		return fmt.Errorf("failed to initialize user SR words: %w", err)
	}
	log.Printf("✅ Initialized SR words for user %d with level %d", userID, level)
	return nil
}

// Word represents a word from the words table
type Word struct {
	ID            int
	Word          string
	Furigana      string
	Romaji        string
	Level         int
	Definitions   string
	PartsOfSpeech string
}

// SRWord represents a word in the SR system with metadata
type SRWord struct {
	SRID         int
	UserID       int
	WordID       int
	Repetitions  int
	EF           float64
	Interval     int
	LastReviewed string
	NextReview   string
	Word         Word
}

// GetNextSRWord retrieves the next word to study for a user (words due for review)
func (db *Database) GetNextSRWord(userID int) (*SRWord, error) {
	query := `
		SELECT 
			sr.id, sr.user_id, sr.word_id, sr.repetitions, sr.ef, sr.interval,
			sr.last_reviewed, sr.next_review,
			w.id, w.word, w.furigana, w.romaji, w.level, w.definitions, w.parts_of_speech
		FROM sr
		JOIN words w ON sr.word_id = w.id
		WHERE sr.user_id = $1 AND sr.next_review <= CURRENT_TIMESTAMP
		ORDER BY sr.next_review ASC
		LIMIT 1
	`

	var srWord SRWord
	err := db.DB.QueryRow(query, userID).Scan(
		&srWord.SRID, &srWord.UserID, &srWord.WordID, &srWord.Repetitions,
		&srWord.EF, &srWord.Interval, &srWord.LastReviewed, &srWord.NextReview,
		&srWord.Word.ID, &srWord.Word.Word, &srWord.Word.Furigana, &srWord.Word.Romaji,
		&srWord.Word.Level, &srWord.Word.Definitions, &srWord.Word.PartsOfSpeech,
	)

	if err == sql.ErrNoRows {
		return nil, nil // No words due for review
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get next SR word: %w", err)
	}

	return &srWord, nil
}

// TODO: Implement additional database operations
// - User management
// - Session storage
// - Update SR word after review
