package main

import (
	"database/sql"
	_ "modernc.org/sqlite"
	"log"
	"os"
)

var db *sql.DB

func initDB() {
	os.MkdirAll("./db", 0755)
	var err error
	db, err = sql.Open("sqlite", "./db/gitpatrol.db")
	if err != nil {
		log.Fatal(err)
	}

	// Create table if not exists
	createTable := `
	CREATE TABLE IF NOT EXISTS repositories (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT,
		url TEXT UNIQUE,
		interval_minutes INTEGER DEFAULT 60,
		last_sync TIMESTAMP,
		status TEXT DEFAULT 'pending',
		last_commit TEXT,
		error_message TEXT
	);`

	_, err = db.Exec(createTable)
	if err != nil {
		log.Fatal(err)
	}

	// Migrations for new columns (ignoring errors if they already exist)
	db.Exec("ALTER TABLE repositories ADD COLUMN stars INTEGER DEFAULT 0")
	db.Exec("ALTER TABLE repositories ADD COLUMN forks INTEGER DEFAULT 0")
	db.Exec("ALTER TABLE repositories ADD COLUMN open_issues INTEGER DEFAULT 0")
	db.Exec("ALTER TABLE repositories ADD COLUMN commit_history TEXT")
	db.Exec("ALTER TABLE repositories ADD COLUMN health_score INTEGER DEFAULT 0")
}
