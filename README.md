# Gaijin - Japanese Study App

A web application for studying Japanese Kanji and Grammar, built with Go backend, PostgreSQL database, and HTMX frontend.

## Features

- **漢字 (Kanji)**: Learn and practice Japanese characters with stroke order and meanings
- **文法 (Grammar)**: Master Japanese grammar patterns and sentence structures  
- **練習 (Practice)**: Interactive exercises and quizzes to reinforce your learning

## Tech Stack

- **Backend**: Go with Gorilla Mux router
- **Database**: PostgreSQL
- **Frontend**: HTMX for dynamic interactions
- **Templates**: Go HTML templates

## Prerequisites

- Go 1.21 or later
- PostgreSQL 12 or later

## Setup

1. **Clone and navigate to the project**:
   ```bash
   cd gaijin
   ```

2. **Install dependencies**:
   ```bash
   go mod tidy
   ```

3. **Set up PostgreSQL database**:
   ```sql
   CREATE DATABASE gaijin;
   ```

4. **Set environment variables** (optional, defaults provided):
   ```bash
   export DB_HOST=localhost
   export DB_PORT=5432
   export DB_USER=postgres
   export DB_PASSWORD=password
   export DB_NAME=gaijin
   export PORT=8080
   ```

5. **Run the application**:
   ```bash
   go run main.go
   ```

6. **Visit the application**:
   Open your browser and go to `http://localhost:8080`

## Project Structure

```
gaijin/
├── main.go          # Main application file
├── go.mod           # Go module dependencies
├── static/          # Static assets directory
└── README.md        # This file
```

## Development

The application uses HTMX for dynamic frontend interactions. The landing page displays "よこそう" (welcome) and includes:

- Responsive design with modern CSS
- Japanese text support with proper UTF-8 encoding
- HTMX integration for dynamic content loading
- Database connection setup (ready for future features)

## Next Steps

- Add Kanji study modules
- Implement grammar lessons
- Create user authentication
- Add progress tracking
- Build practice exercises
