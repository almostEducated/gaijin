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

	createUserSettingsTable := `
	CREATE TABLE IF NOT EXISTS user_settings (
		id SERIAL PRIMARY KEY,
		user_id INTEGER NOT NULL,
		sr_time_japanese INTEGER NOT NULL,
		sr_time_english INTEGER NOT NULL,
		submit_key VARCHAR(10) NOT NULL,
		key_0 VARCHAR(10) NOT NULL,
		key_1 VARCHAR(10) NOT NULL,
		key_2 VARCHAR(10) NOT NULL,
		key_3 VARCHAR(10) NOT NULL,
		key_4 VARCHAR(10) NOT NULL,
		key_5 VARCHAR(10) NOT NULL,
		show_hiragana_mostly BOOLEAN DEFAULT TRUE
	);`

	createSRTable := `
	CREATE TABLE IF NOT EXISTS sr (
		id SERIAL PRIMARY KEY,
		user_id INTEGER NOT NULL,
		word_id INTEGER NOT NULL,
		repetitions INTEGER DEFAULT 0,
		ef FLOAT DEFAULT 2.5,
		interval INTEGER DEFAULT 0,
		type VARCHAR(50) NOT NULL,
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
		hiragana_only BOOLEAN DEFAULT FALSE,
		katakana_only BOOLEAN DEFAULT FALSE,
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
	_, err = db.DB.Exec(createUserSettingsTable)
	if err != nil {
		return fmt.Errorf("error creating user settings table: %w", err)
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
// Each word gets an english meaning entry, and non-katakana_only words also get a japanese pronunciation entry
func (db *Database) InitializeUserSRWords(userID int, level int) error {
	// Insert meaning entries for all words, and pronunciation entries only for non-katakana_only words
	query := `
		INSERT INTO sr (user_id, word_id, repetitions, ef, interval, type, last_reviewed, next_review)
		SELECT $1::INTEGER, id, 0, 2.5, 0, 'english meaning', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
		FROM words
		WHERE level = $2::INTEGER
		UNION ALL
		SELECT $1::INTEGER, id, 0, 2.5, 0, 'japanese pronunciation', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
		FROM words
		WHERE level = $2::INTEGER AND katakana_only = FALSE
		ON CONFLICT DO NOTHING
	`
	_, err := db.DB.Exec(query, userID, level)
	if err != nil {
		return fmt.Errorf("failed to initialize user SR words: %w", err)
	}
	log.Printf("✅ Initialized SR words for user %d with level %d (meaning for all, pronunciation for non-katakana words)", userID, level)
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
	HiraganaOnly  bool
	CreatedAt     string
}

// SRWord represents a word in the SR system with metadata
type SRWord struct {
	SRID         int
	UserID       int
	WordID       int
	Repetitions  int
	EF           float64
	Interval     int
	Type         string
	LastReviewed string
	NextReview   string
	Word         Word
}

// GetNextSRWord retrieves the next word to study for a user (words due for review)
// It considers user settings to skip pronunciation study for hiragana_only words if ShowHiraganaMostly is disabled
func (db *Database) GetNextSRWord(userID int) (*SRWord, error) {
	// First, get user settings to check ShowHiraganaMostly preference
	userSettings, err := db.GetUserSettings(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user settings: %w", err)
	}

	// Build query with conditional filtering based on ShowHiraganaMostly setting
	var query string
	if userSettings.ShowHiraganaMostly {
		// Show all words including pronunciation for hiragana_only words
		query = `
			SELECT 
				sr.id, sr.user_id, sr.word_id, sr.repetitions, sr.ef, sr.interval, sr.type,
				sr.last_reviewed, sr.next_review,
				w.id, w.word, w.furigana, w.romaji, w.level, w.definitions, w.parts_of_speech, w.hiragana_only, w.created_at
			FROM sr
			JOIN words w ON sr.word_id = w.id
			WHERE sr.user_id = $1 AND sr.next_review <= CURRENT_TIMESTAMP
			ORDER BY sr.next_review ASC
			LIMIT 1
		`
	} else {
		// Skip pronunciation study for hiragana_only words
		query = `
			SELECT 
				sr.id, sr.user_id, sr.word_id, sr.repetitions, sr.ef, sr.interval, sr.type,
				sr.last_reviewed, sr.next_review,
				w.id, w.word, w.furigana, w.romaji, w.level, w.definitions, w.parts_of_speech, w.hiragana_only, w.created_at
			FROM sr
			JOIN words w ON sr.word_id = w.id
			WHERE sr.user_id = $1 
				AND sr.next_review <= CURRENT_TIMESTAMP
				AND NOT (w.hiragana_only = TRUE AND sr.type = 'japanese pronunciation')
			ORDER BY sr.next_review ASC
			LIMIT 1
		`
	}

	var srWord SRWord
	err = db.DB.QueryRow(query, userID).Scan(
		&srWord.SRID, &srWord.UserID, &srWord.WordID, &srWord.Repetitions,
		&srWord.EF, &srWord.Interval, &srWord.Type, &srWord.LastReviewed, &srWord.NextReview,
		&srWord.Word.ID, &srWord.Word.Word, &srWord.Word.Furigana, &srWord.Word.Romaji,
		&srWord.Word.Level, &srWord.Word.Definitions, &srWord.Word.PartsOfSpeech, &srWord.Word.HiraganaOnly, &srWord.Word.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil // No words due for review
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get next SR word: %w", err)
	}

	return &srWord, nil
}

func (db *Database) LookupWordBySRId(srID int) (*Word, error) {
	query := `
	SELECT w.id, w.word, w.furigana, w.romaji, w.level, w.definitions, w.parts_of_speech, w.hiragana_only, w.created_at
	FROM sr
	JOIN words w ON sr.word_id = w.id
	WHERE sr.id = $1
	`
	var word Word
	err := db.DB.QueryRow(query, srID).Scan(&word.ID, &word.Word, &word.Furigana, &word.Romaji, &word.Level, &word.Definitions, &word.PartsOfSpeech, &word.HiraganaOnly, &word.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to lookup word by ID: %w", err)
	}
	return &word, nil
}

type UserSettings struct {
	UserID             int
	SRTimeJapanese     int
	SRTimeEnglish      int
	SubmitKey          string
	Key1               string
	Key2               string
	Key3               string
	Key4               string
	Key5               string
	ShowHiraganaMostly bool
}

type UserInfo struct {
	ID        int
	Username  string
	Email     string
	CreatedAt string
}

func (db *Database) GetUserSettings(userID int) (*UserSettings, error) {
	var userSettings UserSettings
	query := `
		SELECT id, user_id, sr_time_japanese, sr_time_english, submit_key, key_1, key_2, key_3, key_4, key_5, show_hiragana_mostly 
		FROM user_settings 
		WHERE user_id = $1
	`
	var id int // temporary variable to scan the id column
	err := db.DB.QueryRow(query, userID).Scan(&id, &userSettings.UserID, &userSettings.SRTimeJapanese, &userSettings.SRTimeEnglish, &userSettings.SubmitKey, &userSettings.Key1, &userSettings.Key2, &userSettings.Key3, &userSettings.Key4, &userSettings.Key5, &userSettings.ShowHiraganaMostly)
	if err != nil {
		return nil, fmt.Errorf("failed to get user settings: %w", err)
	}
	return &userSettings, nil
}

func (db *Database) UpdateUserSettings(userID int, settings *UserSettings) error {
	query := `
		UPDATE user_settings 
		SET sr_time_japanese = $2, 
		    sr_time_english = $3, 
		    submit_key = $4, 
		    key_1 = $5, 
		    key_2 = $6, 
		    key_3 = $7, 
		    key_4 = $8, 
		    key_5 = $9,
		    show_hiragana_mostly = $10
		WHERE user_id = $1
	`
	_, err := db.DB.Exec(query, userID, settings.SRTimeJapanese, settings.SRTimeEnglish, settings.SubmitKey, settings.Key1, settings.Key2, settings.Key3, settings.Key4, settings.Key5, settings.ShowHiraganaMostly)
	if err != nil {
		return fmt.Errorf("failed to update user settings: %w", err)
	}
	log.Printf("✅ Updated user settings for user %d", userID)
	return nil
}

func (db *Database) GetUserInfo(userID int) (*UserInfo, error) {
	var userInfo UserInfo
	var createdAt sql.NullTime
	query := `
		SELECT id, username, email, created_at
		FROM users
		WHERE id = $1
	`
	err := db.DB.QueryRow(query, userID).Scan(&userInfo.ID, &userInfo.Username, &userInfo.Email, &createdAt)
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}

	// Format the date as "January 2, 2006"
	if createdAt.Valid {
		userInfo.CreatedAt = createdAt.Time.Format("January 2, 2006")
	} else {
		userInfo.CreatedAt = "Unknown"
	}

	return &userInfo, nil
}

// UpdateSRWord updates an SR record using the SM-2 algorithm
// quality: 0-5 rating (5=perfect, 4=correct after hesitation, 3=difficult, 2=incorrect, 1=barely, 0=blackout)
func (db *Database) UpdateSRWord(srID int, quality int) error {
	// Validate quality rating
	if quality < 0 || quality > 5 {
		return fmt.Errorf("quality must be between 0 and 5")
	}

	// Get current SR data
	var currentEF float64
	var currentInterval int
	var currentRepetitions int
	query := `SELECT ef, interval, repetitions FROM sr WHERE id = $1`
	err := db.DB.QueryRow(query, srID).Scan(&currentEF, &currentInterval, &currentRepetitions)
	if err != nil {
		return fmt.Errorf("failed to get current SR data: %w", err)
	}

	// Calculate new EF using SM-2 formula
	// EF' = EF + (0.1 - (5 - q) * (0.08 + (5 - q) * 0.02))
	newEF := currentEF + (0.1 - float64(5-quality)*(0.08+float64(5-quality)*0.02))
	if newEF < 1.3 {
		newEF = 1.3
	}

	var newInterval int
	var newRepetitions int

	// Calculate new interval based on quality
	if quality < 3 {
		// Incorrect answer - reset
		newRepetitions = 0
		newInterval = 1
	} else {
		// Correct answer
		newRepetitions = currentRepetitions + 1
		if newRepetitions == 1 {
			newInterval = 1
		} else if newRepetitions == 2 {
			newInterval = 6
		} else {
			newInterval = int(float64(currentInterval) * newEF)
		}
	}

	// Update SR record
	updateQuery := `
		UPDATE sr 
		SET ef = $1, 
		    interval = $2, 
		    repetitions = $3,
		    last_reviewed = CURRENT_TIMESTAMP,
		    next_review = CURRENT_TIMESTAMP + INTERVAL '1 day' * $2::INTEGER
		WHERE id = $4
	`
	_, err = db.DB.Exec(updateQuery, newEF, newInterval, newRepetitions, srID)
	if err != nil {
		return fmt.Errorf("failed to update SR word: %w", err)
	}

	log.Printf("✅ Updated SR word %d: quality=%d, EF=%.2f→%.2f, interval=%d→%d days, reps=%d→%d",
		srID, quality, currentEF, newEF, currentInterval, newInterval, currentRepetitions, newRepetitions)

	return nil
}

// TODO: Implement additional database operations
// - User management
// - Session storage
