package main

import (
	"database/sql"

	_ "modernc.org/sqlite"
)

func InitDB(path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}

	schema := `
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT NOT NULL UNIQUE,
		password TEXT NOT NULL,
		email TEXT NOT NULL UNIQUE,
		role BOOLEAN NOT NULL DEFAULT 0
	);
	`
	_, err = db.Exec(schema)
	if err != nil {
		return nil, err
	}

	return db, nil
}
