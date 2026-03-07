package main

import (
	"database/sql"
	_ "modernc.org/sqlite"
	"log"
)

var db *sql.DB

func initDB() {
	var err error
	db, err = sql.Open("sqlite", "./gitpatrol.db")
	if err != nil {
		log.Fatal(err)
	}

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
}
