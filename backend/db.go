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

	// Enable WAL mode and set busy timeout for better concurrency
	db.Exec("PRAGMA journal_mode=WAL;")
	db.Exec("PRAGMA busy_timeout=5000;")

	createUsersTable := `
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT UNIQUE,
		password_hash TEXT,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);`

	_, err = db.Exec(createUsersTable)
	if err != nil { log.Fatal(err) }

	createIncidentsTable := `
	CREATE TABLE IF NOT EXISTS incidents (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		repo_id INTEGER,
		repo_name TEXT,
		message TEXT,
		created_at TIMESTAMP,
		resolved INTEGER DEFAULT 0,
		FOREIGN KEY(repo_id) REFERENCES repositories(id)
	);`

	_, err = db.Exec(createIncidentsTable)
	if err != nil { log.Fatal(err) }

	createTable := `
	CREATE TABLE IF NOT EXISTS repositories (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT,
		url TEXT UNIQUE,
		interval_minutes INTEGER DEFAULT 60,
		last_sync TIMESTAMP,
		status TEXT DEFAULT 'pending',
		last_commit TEXT,
		error_message TEXT,
		stars INTEGER DEFAULT 0,
		forks INTEGER DEFAULT 0,
		open_issues INTEGER DEFAULT 0,
		commit_history TEXT,
		health_score INTEGER DEFAULT 0,
		default_branch TEXT DEFAULT 'main',
		auto_patrol INTEGER DEFAULT 1
	);`

	_, err = db.Exec(createTable)
	if err != nil {
		log.Fatal(err)
	}

	// Migrations
	db.Exec("ALTER TABLE repositories ADD COLUMN stars INTEGER DEFAULT 0")
	db.Exec("ALTER TABLE repositories ADD COLUMN forks INTEGER DEFAULT 0")
	db.Exec("ALTER TABLE repositories ADD COLUMN open_issues INTEGER DEFAULT 0")
	db.Exec("ALTER TABLE repositories ADD COLUMN commit_history TEXT")
	db.Exec("ALTER TABLE repositories ADD COLUMN health_score INTEGER DEFAULT 0")
	db.Exec("ALTER TABLE repositories ADD COLUMN default_branch TEXT DEFAULT 'main'")
	db.Exec("ALTER TABLE repositories ADD COLUMN auto_patrol INTEGER DEFAULT 1")
}
