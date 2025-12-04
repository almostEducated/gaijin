@echo off
REM Set console to UTF-8 encoding
chcp 65001 >nul

REM Set environment variable for PostgreSQL client encoding
set PGCLIENTENCODING=UTF8

REM Login to PostgreSQL
REM Replace these with your actual credentials:
REM   - hostname (default: localhost)
REM   - port (default: 5432)
REM   - database name
REM   - username

psql -h localhost -p 5432 -U postgres -d gaijin

REM Alternative: If you want to set encoding after connection, uncomment below
REM and comment out the PGCLIENTENCODING line above
REM psql -h localhost -p 5432 -U your_username -d your_database -c "SET CLIENT_ENCODING TO 'UTF8';" -c "\encoding UTF8"