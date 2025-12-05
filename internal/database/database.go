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
		suspended BOOLEAN DEFAULT FALSE,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);`

	// Create a kanji table
	createKanjiConfusionTable := `
	CREATE TABLE IF NOT EXISTS kanji_confusion (
		id SERIAL PRIMARY KEY,
		kanji_1 VARCHAR(10) NOT NULL,
		kanji_2 VARCHAR(10) NOT NULL,
		word1_id INTEGER NOT NULL,
		word2_id INTEGER NOT NULL,
		user_id INTEGER NOT NULL,
		note TEXT,
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
		frequency INTEGER,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		UNIQUE(word, level)
	);`

	// Hiragana table - stores all hiragana mora for beginners deck
	createHiraganaTable := `
	CREATE TABLE IF NOT EXISTS hiragana (
		id SERIAL PRIMARY KEY,
		character VARCHAR(10) NOT NULL UNIQUE,
		romaji VARCHAR(10) NOT NULL,
		category VARCHAR(50) NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);`

	// Katakana table - stores all katakana mora for beginners deck
	createKatakanaTable := `
	CREATE TABLE IF NOT EXISTS katakana (
		id SERIAL PRIMARY KEY,
		character VARCHAR(10) NOT NULL UNIQUE,
		romaji VARCHAR(10) NOT NULL,
		category VARCHAR(50) NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);`

	// SR Kana table - spaced repetition tracking for hiragana/katakana
	createSRKanaTable := `
	CREATE TABLE IF NOT EXISTS sr_kana (
		id SERIAL PRIMARY KEY,
		user_id INTEGER NOT NULL,
		kana_id INTEGER NOT NULL,
		kana_type VARCHAR(10) NOT NULL,
		repetitions INTEGER DEFAULT 0,
		ef FLOAT DEFAULT 2.5,
		interval INTEGER DEFAULT 0,
		last_reviewed TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		next_review TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		UNIQUE(user_id, kana_id, kana_type)
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
	_, err = db.DB.Exec(createKanjiConfusionTable)
	if err != nil {
		return fmt.Errorf("error creating kanji confusion table: %w", err)
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
	_, err = db.DB.Exec(createHiraganaTable)
	if err != nil {
		return fmt.Errorf("error creating hiragana table: %w", err)
	}
	_, err = db.DB.Exec(createKatakanaTable)
	if err != nil {
		return fmt.Errorf("error creating katakana table: %w", err)
	}
	_, err = db.DB.Exec(createSRKanaTable)
	if err != nil {
		return fmt.Errorf("error creating sr_kana table: %w", err)
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
	// Also filter out suspended words
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
			WHERE sr.user_id = $1 
				AND sr.next_review <= CURRENT_TIMESTAMP
				AND (sr.suspended = FALSE OR sr.suspended IS NULL)
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
				AND (sr.suspended = FALSE OR sr.suspended IS NULL)
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

// GetNextSRWordAdverbs retrieves the next adverb word to study for a user (words due for review)
// It filters words where "adverb" appears in the parts_of_speech column (semicolon-separated)
// It considers user settings to skip pronunciation study for hiragana_only words if ShowHiraganaMostly is disabled
func (db *Database) GetNextSRWordAdverbs(userID int) (*SRWord, error) {
	// First, get user settings to check ShowHiraganaMostly preference
	userSettings, err := db.GetUserSettings(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user settings: %w", err)
	}

	// Build query with conditional filtering based on ShowHiraganaMostly setting
	// Filter for adverbs by checking if "adverb" appears in the semicolon-separated parts_of_speech
	// Also filter out suspended words
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
			WHERE sr.user_id = $1 
				AND sr.next_review <= CURRENT_TIMESTAMP
				AND w.parts_of_speech IS NOT NULL
				AND ((';' || w.parts_of_speech || ';') LIKE '%;adverb;%')
				AND (sr.suspended = FALSE OR sr.suspended IS NULL)
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
				AND w.parts_of_speech IS NOT NULL
				AND ((';' || w.parts_of_speech || ';') LIKE '%;adverb;%')
				AND (sr.suspended = FALSE OR sr.suspended IS NULL)
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
		return nil, fmt.Errorf("failed to get next SR adverb word: %w", err)
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

// KanjiConfusionPair represents a pair of visually similar kanji
type KanjiConfusionPair struct {
	Kanji1       string
	Kanji2       string
	Word1        string
	Word2        string
	Furigana1    string
	Furigana2    string
	Definitions1 string
	Definitions2 string
	Word1ID      int
	Word2ID      int
}

// GetRandomKanjiConfusionPair retrieves a random kanji confusion pair for a user
func (db *Database) GetRandomKanjiConfusionPair(userID int) (*KanjiConfusionPair, error) {
	query := `
		SELECT 
			kc.kanji_1, kc.kanji_2,
			w1.word, w2.word,
			w1.furigana, w2.furigana,
			w1.definitions, w2.definitions,
			w1.id, w2.id
		FROM kanji_confusion kc
		JOIN words w1 ON kc.word1_id = w1.id
		JOIN words w2 ON kc.word2_id = w2.id
		WHERE kc.user_id = $1
		ORDER BY RANDOM()
		LIMIT 1
	`

	var pair KanjiConfusionPair
	err := db.DB.QueryRow(query, userID).Scan(
		&pair.Kanji1, &pair.Kanji2,
		&pair.Word1, &pair.Word2,
		&pair.Furigana1, &pair.Furigana2,
		&pair.Definitions1, &pair.Definitions2,
		&pair.Word1ID, &pair.Word2ID,
	)

	if err == sql.ErrNoRows {
		return nil, nil // No pairs found
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get confusion pair: %w", err)
	}

	return &pair, nil
}

// Kana represents a hiragana or katakana character
type Kana struct {
	ID        int
	Character string
	Romaji    string
	Category  string
	CreatedAt string
}

// SRKana represents a kana in the SR system with metadata
type SRKana struct {
	SRID         int
	UserID       int
	KanaID       int
	KanaType     string // "hiragana" or "katakana"
	Repetitions  int
	EF           float64
	Interval     int
	LastReviewed string
	NextReview   string
	Kana         Kana
}

// HasUserSRKana checks if a user has any kana in their SR kana table for a specific type
func (db *Database) HasUserSRKana(userID int, kanaType string) (bool, error) {
	var count int
	query := `SELECT COUNT(*) FROM sr_kana WHERE user_id = $1 AND kana_type = $2`
	err := db.DB.QueryRow(query, userID, kanaType).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check user SR kana: %w", err)
	}
	return count > 0, nil
}

// InitializeUserSRKana populates SR kana table with all kana of a specific type for a user
func (db *Database) InitializeUserSRKana(userID int, kanaType string) error {
	var query string
	if kanaType == "hiragana" {
		query = `
			INSERT INTO sr_kana (user_id, kana_id, kana_type, repetitions, ef, interval, last_reviewed, next_review)
			SELECT $1::INTEGER, id, 'hiragana', 0, 2.5, 0, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
			FROM hiragana
			ON CONFLICT (user_id, kana_id, kana_type) DO NOTHING
		`
	} else {
		query = `
			INSERT INTO sr_kana (user_id, kana_id, kana_type, repetitions, ef, interval, last_reviewed, next_review)
			SELECT $1::INTEGER, id, 'katakana', 0, 2.5, 0, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
			FROM katakana
			ON CONFLICT (user_id, kana_id, kana_type) DO NOTHING
		`
	}
	_, err := db.DB.Exec(query, userID)
	if err != nil {
		return fmt.Errorf("failed to initialize user SR kana: %w", err)
	}
	log.Printf("✅ Initialized SR kana for user %d with type %s", userID, kanaType)
	return nil
}

// GetNextSRKana retrieves the next kana to study for a user (kana due for review)
func (db *Database) GetNextSRKana(userID int, kanaType string) (*SRKana, error) {
	var query string
	if kanaType == "hiragana" {
		query = `
			SELECT 
				sk.id, sk.user_id, sk.kana_id, sk.kana_type, sk.repetitions, sk.ef, sk.interval,
				sk.last_reviewed, sk.next_review,
				h.id, h.character, h.romaji, h.category, h.created_at
			FROM sr_kana sk
			JOIN hiragana h ON sk.kana_id = h.id
			WHERE sk.user_id = $1 AND sk.kana_type = 'hiragana' AND sk.next_review <= CURRENT_TIMESTAMP
			ORDER BY sk.next_review ASC
			LIMIT 1
		`
	} else {
		query = `
			SELECT 
				sk.id, sk.user_id, sk.kana_id, sk.kana_type, sk.repetitions, sk.ef, sk.interval,
				sk.last_reviewed, sk.next_review,
				k.id, k.character, k.romaji, k.category, k.created_at
			FROM sr_kana sk
			JOIN katakana k ON sk.kana_id = k.id
			WHERE sk.user_id = $1 AND sk.kana_type = 'katakana' AND sk.next_review <= CURRENT_TIMESTAMP
			ORDER BY sk.next_review ASC
			LIMIT 1
		`
	}

	var srKana SRKana
	err := db.DB.QueryRow(query, userID).Scan(
		&srKana.SRID, &srKana.UserID, &srKana.KanaID, &srKana.KanaType, &srKana.Repetitions,
		&srKana.EF, &srKana.Interval, &srKana.LastReviewed, &srKana.NextReview,
		&srKana.Kana.ID, &srKana.Kana.Character, &srKana.Kana.Romaji, &srKana.Kana.Category, &srKana.Kana.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil // No kana due for review
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get next SR kana: %w", err)
	}

	return &srKana, nil
}

// LookupKanaBySRId retrieves kana information by SR ID
func (db *Database) LookupKanaBySRId(srID int) (*Kana, string, error) {
	// First get the kana type
	var kanaType string
	var kanaID int
	query := `SELECT kana_id, kana_type FROM sr_kana WHERE id = $1`
	err := db.DB.QueryRow(query, srID).Scan(&kanaID, &kanaType)
	if err != nil {
		return nil, "", fmt.Errorf("failed to lookup SR kana: %w", err)
	}

	var kana Kana
	if kanaType == "hiragana" {
		query = `SELECT id, character, romaji, category, created_at FROM hiragana WHERE id = $1`
	} else {
		query = `SELECT id, character, romaji, category, created_at FROM katakana WHERE id = $1`
	}
	err = db.DB.QueryRow(query, kanaID).Scan(&kana.ID, &kana.Character, &kana.Romaji, &kana.Category, &kana.CreatedAt)
	if err != nil {
		return nil, "", fmt.Errorf("failed to lookup kana by ID: %w", err)
	}
	return &kana, kanaType, nil
}

// UpdateSRKana updates an SR kana record using the SM-2 algorithm
// quality: 0-5 rating (5=perfect, 4=correct after hesitation, 3=difficult, 2=incorrect, 1=barely, 0=blackout)
func (db *Database) UpdateSRKana(srID int, quality int) error {
	// Validate quality rating
	if quality < 0 || quality > 5 {
		return fmt.Errorf("quality must be between 0 and 5")
	}

	// Get current SR data
	var currentEF float64
	var currentInterval int
	var currentRepetitions int
	query := `SELECT ef, interval, repetitions FROM sr_kana WHERE id = $1`
	err := db.DB.QueryRow(query, srID).Scan(&currentEF, &currentInterval, &currentRepetitions)
	if err != nil {
		return fmt.Errorf("failed to get current SR kana data: %w", err)
	}

	// Calculate new EF using SM-2 formula
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
		UPDATE sr_kana 
		SET ef = $1, 
		    interval = $2, 
		    repetitions = $3,
		    last_reviewed = CURRENT_TIMESTAMP,
		    next_review = CURRENT_TIMESTAMP + INTERVAL '1 day' * $2::INTEGER
		WHERE id = $4
	`
	_, err = db.DB.Exec(updateQuery, newEF, newInterval, newRepetitions, srID)
	if err != nil {
		return fmt.Errorf("failed to update SR kana: %w", err)
	}

	log.Printf("✅ Updated SR kana %d: quality=%d, EF=%.2f→%.2f, interval=%d→%d days, reps=%d→%d",
		srID, quality, currentEF, newEF, currentInterval, newInterval, currentRepetitions, newRepetitions)

	return nil
}

// GetKanaCount returns the count of kana for a specific type
func (db *Database) GetKanaCount(kanaType string) (int, error) {
	var count int
	var query string
	if kanaType == "hiragana" {
		query = `SELECT COUNT(*) FROM hiragana`
	} else {
		query = `SELECT COUNT(*) FROM katakana`
	}
	err := db.DB.QueryRow(query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get kana count: %w", err)
	}
	return count, nil
}

// LearnWord represents a word for the Learn page with user-specific learning status
type LearnWord struct {
	ID            int
	Word          string
	Furigana      string
	Romaji        string
	Level         int
	Definitions   string
	PartsOfSpeech string
	HiraganaOnly  bool
	KatakanaOnly  bool
	IsLearned     bool // Whether the user has added this word to their SR deck
	Frequency     *int // Frequency rank (lower = more common), nil if no data
	IsSuspended   bool // Whether the word is suspended (user marked as "known")
}

// GetWordsForLearning retrieves words by level with pagination for the Learn page
// Returns words along with whether the user has already added them to their SR deck
func (db *Database) GetWordsForLearning(userID int, level int, limit int, offset int) ([]LearnWord, int, error) {
	// Get total count of words (at this level, or all levels if level=0)
	var totalCount int
	var countQuery string
	var err error

	if level == 0 {
		// All levels mode - only count words with frequency data
		countQuery = `SELECT COUNT(*) FROM words WHERE frequency IS NOT NULL`
		err = db.DB.QueryRow(countQuery).Scan(&totalCount)
	} else {
		countQuery = `SELECT COUNT(*) FROM words WHERE level = $1`
		err = db.DB.QueryRow(countQuery, level).Scan(&totalCount)
	}
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count words: %w", err)
	}

	// Get words with learning status, ordered by frequency (most common first)
	var query string
	var rows *sql.Rows

	if level == 0 {
		// All levels mode - sort purely by frequency, only include words with frequency
		query = `
			SELECT 
				w.id, w.word, w.furigana, w.romaji, w.level, 
				w.definitions, w.parts_of_speech, w.hiragana_only, 
				COALESCE(w.katakana_only, false) as katakana_only,
				CASE WHEN EXISTS (
					SELECT 1 FROM sr WHERE sr.user_id = $1 AND sr.word_id = w.id
				) THEN true ELSE false END as is_learned,
				w.frequency,
				CASE WHEN EXISTS (
					SELECT 1 FROM sr WHERE sr.user_id = $1 AND sr.word_id = w.id AND sr.suspended = true
				) THEN true ELSE false END as is_suspended
			FROM words w
			WHERE w.frequency IS NOT NULL
			ORDER BY w.frequency ASC
			LIMIT $2 OFFSET $3
		`
		rows, err = db.DB.Query(query, userID, limit, offset)
	} else {
		query = `
			SELECT 
				w.id, w.word, w.furigana, w.romaji, w.level, 
				w.definitions, w.parts_of_speech, w.hiragana_only, 
				COALESCE(w.katakana_only, false) as katakana_only,
				CASE WHEN EXISTS (
					SELECT 1 FROM sr WHERE sr.user_id = $1 AND sr.word_id = w.id
				) THEN true ELSE false END as is_learned,
				w.frequency,
				CASE WHEN EXISTS (
					SELECT 1 FROM sr WHERE sr.user_id = $1 AND sr.word_id = w.id AND sr.suspended = true
				) THEN true ELSE false END as is_suspended
			FROM words w
			WHERE w.level = $2
			ORDER BY w.frequency ASC NULLS LAST, w.id ASC
			LIMIT $3 OFFSET $4
		`
		rows, err = db.DB.Query(query, userID, level, limit, offset)
	}
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get words for learning: %w", err)
	}
	defer rows.Close()

	var words []LearnWord
	for rows.Next() {
		var word LearnWord
		var furigana, romaji, definitions, partsOfSpeech sql.NullString
		var frequency sql.NullInt64
		err := rows.Scan(
			&word.ID, &word.Word, &furigana, &romaji, &word.Level,
			&definitions, &partsOfSpeech, &word.HiraganaOnly, &word.KatakanaOnly, &word.IsLearned,
			&frequency, &word.IsSuspended,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan word: %w", err)
		}
		word.Furigana = furigana.String
		word.Romaji = romaji.String
		word.Definitions = definitions.String
		word.PartsOfSpeech = partsOfSpeech.String
		if frequency.Valid {
			freq := int(frequency.Int64)
			word.Frequency = &freq
		}
		words = append(words, word)
	}

	return words, totalCount, nil
}

// AddWordToSR adds a specific word to a user's SR deck
// Creates both english meaning and japanese pronunciation entries (if not katakana_only)
func (db *Database) AddWordToSR(userID int, wordID int) error {
	// First check if word exists and get its katakana_only status
	var katakanaOnly bool
	checkQuery := `SELECT COALESCE(katakana_only, false) FROM words WHERE id = $1`
	err := db.DB.QueryRow(checkQuery, wordID).Scan(&katakanaOnly)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("word not found")
		}
		return fmt.Errorf("failed to check word: %w", err)
	}

	// Insert meaning entry
	insertMeaning := `
		INSERT INTO sr (user_id, word_id, repetitions, ef, interval, type, last_reviewed, next_review)
		VALUES ($1, $2, 0, 2.5, 0, 'english meaning', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		ON CONFLICT DO NOTHING
	`
	_, err = db.DB.Exec(insertMeaning, userID, wordID)
	if err != nil {
		return fmt.Errorf("failed to add meaning entry: %w", err)
	}

	// Insert pronunciation entry only for non-katakana words
	if !katakanaOnly {
		insertPronunciation := `
			INSERT INTO sr (user_id, word_id, repetitions, ef, interval, type, last_reviewed, next_review)
			VALUES ($1, $2, 0, 2.5, 0, 'japanese pronunciation', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
			ON CONFLICT DO NOTHING
		`
		_, err = db.DB.Exec(insertPronunciation, userID, wordID)
		if err != nil {
			return fmt.Errorf("failed to add pronunciation entry: %w", err)
		}
	}

	log.Printf("✅ Added word %d to SR deck for user %d", wordID, userID)
	return nil
}

// AddMultipleWordsToSR adds multiple words to a user's SR deck
func (db *Database) AddMultipleWordsToSR(userID int, wordIDs []int) error {
	for _, wordID := range wordIDs {
		err := db.AddWordToSR(userID, wordID)
		if err != nil {
			return fmt.Errorf("failed to add word %d: %w", wordID, err)
		}
	}
	return nil
}

// ToggleWordSuspended toggles the suspended status for all SR entries of a word for a user
// Returns the new suspended state
func (db *Database) ToggleWordSuspended(userID int, wordID int) (bool, error) {
	// First check if any SR entry exists for this word
	var exists bool
	checkQuery := `SELECT EXISTS(SELECT 1 FROM sr WHERE user_id = $1 AND word_id = $2)`
	err := db.DB.QueryRow(checkQuery, userID, wordID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check SR entry: %w", err)
	}
	if !exists {
		return false, fmt.Errorf("word not in SR deck")
	}

	// Get current suspended state (check if ANY entry is suspended)
	var currentSuspended bool
	stateQuery := `SELECT COALESCE(bool_or(suspended), false) FROM sr WHERE user_id = $1 AND word_id = $2`
	err = db.DB.QueryRow(stateQuery, userID, wordID).Scan(&currentSuspended)
	if err != nil {
		return false, fmt.Errorf("failed to get suspended state: %w", err)
	}

	// Toggle to the opposite state
	newSuspended := !currentSuspended

	// Update all SR entries for this word
	updateQuery := `UPDATE sr SET suspended = $1 WHERE user_id = $2 AND word_id = $3`
	_, err = db.DB.Exec(updateQuery, newSuspended, userID, wordID)
	if err != nil {
		return false, fmt.Errorf("failed to update suspended state: %w", err)
	}

	log.Printf("✅ Toggled suspended state for word %d, user %d: %v → %v", wordID, userID, currentSuspended, newSuspended)
	return newSuspended, nil
}

// SuspendWord sets the suspended status to true for all SR entries of a word for a user
func (db *Database) SuspendWord(userID int, wordID int) (bool, error) {
	// Update all SR entries for this word to suspended
	updateQuery := `UPDATE sr SET suspended = true WHERE user_id = $1 AND word_id = $2`
	_, err := db.DB.Exec(updateQuery, userID, wordID)
	if err != nil {
		return false, fmt.Errorf("failed to suspend word: %w", err)
	}
	return true, nil
}

// GetLevelWordCounts returns the count of words at each JLPT level
func (db *Database) GetLevelWordCounts() (map[int]int, error) {
	query := `SELECT level, COUNT(*) FROM words GROUP BY level ORDER BY level DESC`
	rows, err := db.DB.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get word counts: %w", err)
	}
	defer rows.Close()

	counts := make(map[int]int)
	for rows.Next() {
		var level, count int
		if err := rows.Scan(&level, &count); err != nil {
			return nil, fmt.Errorf("failed to scan count: %w", err)
		}
		counts[level] = count
	}
	return counts, nil
}

// GetUserLearnedCountByLevel returns how many words the user has learned at each level
func (db *Database) GetUserLearnedCountByLevel(userID int) (map[int]int, error) {
	query := `
		SELECT w.level, COUNT(DISTINCT sr.word_id) 
		FROM sr 
		JOIN words w ON sr.word_id = w.id 
		WHERE sr.user_id = $1 
		GROUP BY w.level
	`
	rows, err := db.DB.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get learned counts: %w", err)
	}
	defer rows.Close()

	counts := make(map[int]int)
	for rows.Next() {
		var level, count int
		if err := rows.Scan(&level, &count); err != nil {
			return nil, fmt.Errorf("failed to scan count: %w", err)
		}
		counts[level] = count
	}
	return counts, nil
}

// KanjiWord represents a word containing a specific kanji for the kanji lookup page
type KanjiWord struct {
	ID            int
	Word          string
	Furigana      string
	Level         int
	Definitions   string
	PartsOfSpeech string
}

// GetWordsByKanji searches for all words containing a specific kanji character
// Results are ordered by JLPT level (N5 first, then N4, etc.)
func (db *Database) GetWordsByKanji(kanji string) ([]KanjiWord, error) {
	query := `
		SELECT id, word, furigana, level, definitions, parts_of_speech
		FROM words
		WHERE word LIKE '%' || $1 || '%'
		ORDER BY level DESC, word ASC
	`

	rows, err := db.DB.Query(query, kanji)
	if err != nil {
		return nil, fmt.Errorf("failed to search words by kanji: %w", err)
	}
	defer rows.Close()

	var words []KanjiWord
	for rows.Next() {
		var word KanjiWord
		var furigana, definitions, partsOfSpeech sql.NullString
		err := rows.Scan(&word.ID, &word.Word, &furigana, &word.Level, &definitions, &partsOfSpeech)
		if err != nil {
			return nil, fmt.Errorf("failed to scan word: %w", err)
		}
		word.Furigana = furigana.String
		word.Definitions = definitions.String
		word.PartsOfSpeech = partsOfSpeech.String
		words = append(words, word)
	}

	return words, nil
}

// SearchWords searches for words based on the query string
// If the query contains Japanese characters (hiragana, katakana, kanji), it searches word and furigana
// Otherwise, it searches the English definitions
// Results are ordered by JLPT level (N5 first, then N4, etc.)
func (db *Database) SearchWords(query string) ([]KanjiWord, string, error) {
	// Detect if query contains Japanese characters
	isJapanese := containsJapanese(query)

	var sqlQuery string
	var searchType string

	if isJapanese {
		// Search in word and furigana fields
		sqlQuery = `
			SELECT id, word, furigana, level, definitions, parts_of_speech
			FROM words
			WHERE word LIKE '%' || $1 || '%' OR furigana LIKE '%' || $1 || '%'
			ORDER BY level DESC, word ASC
			LIMIT 100
		`
		searchType = "japanese"
	} else {
		// Search in English definitions (case-insensitive)
		sqlQuery = `
			SELECT id, word, furigana, level, definitions, parts_of_speech
			FROM words
			WHERE LOWER(definitions) LIKE '%' || LOWER($1) || '%'
			ORDER BY level DESC, word ASC
			LIMIT 100
		`
		searchType = "english"
	}

	rows, err := db.DB.Query(sqlQuery, query)
	if err != nil {
		return nil, searchType, fmt.Errorf("failed to search words: %w", err)
	}
	defer rows.Close()

	var words []KanjiWord
	for rows.Next() {
		var word KanjiWord
		var furigana, definitions, partsOfSpeech sql.NullString
		err := rows.Scan(&word.ID, &word.Word, &furigana, &word.Level, &definitions, &partsOfSpeech)
		if err != nil {
			return nil, searchType, fmt.Errorf("failed to scan word: %w", err)
		}
		word.Furigana = furigana.String
		word.Definitions = definitions.String
		word.PartsOfSpeech = partsOfSpeech.String
		words = append(words, word)
	}

	return words, searchType, nil
}

// containsJapanese checks if a string contains Japanese characters
func containsJapanese(s string) bool {
	for _, r := range s {
		// Hiragana: U+3040 - U+309F
		// Katakana: U+30A0 - U+30FF
		// CJK Unified Ideographs (Kanji): U+4E00 - U+9FAF
		// CJK Extension A: U+3400 - U+4DBF
		if (r >= 0x3040 && r <= 0x309F) || // Hiragana
			(r >= 0x30A0 && r <= 0x30FF) || // Katakana
			(r >= 0x4E00 && r <= 0x9FAF) || // CJK Unified Ideographs
			(r >= 0x3400 && r <= 0x4DBF) { // CJK Extension A
			return true
		}
	}
	return false
}
