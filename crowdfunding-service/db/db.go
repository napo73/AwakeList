package db

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
	CREATE TABLE IF NOT EXISTS projects (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		description TEXT,
		image_path TEXT,
		created_by INTEGER
	);

	CREATE TABLE IF NOT EXISTS funding (
		project_id INTEGER PRIMARY KEY,
		target_amount INTEGER NOT NULL,
		current_amount INTEGER NOT NULL DEFAULT 0,
		FOREIGN KEY(project_id) REFERENCES projects(id)
	);

	CREATE TABLE IF NOT EXISTS project_meta (
		project_id INTEGER PRIMARY KEY,
		hashtag TEXT,
		published BOOLEAN NOT NULL DEFAULT 0,
		FOREIGN KEY(project_id) REFERENCES projects(id)
	);
	`

	if _, err := db.Exec(schema); err != nil {
		return nil, err
	}

	return db, nil
}
