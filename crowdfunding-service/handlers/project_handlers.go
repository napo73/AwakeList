package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func CreateProjectHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		http.Error(w, "Could not parse form", http.StatusBadRequest)
		return
	}

	name := r.FormValue("name")
	description := r.FormValue("description")
	hashtag := r.FormValue("hashtag")
	targetAmountStr := r.FormValue("target_amount")
	targetAmount, err := strconv.Atoi(targetAmountStr)
	if err != nil {
		http.Error(w, "Invalid target amount", http.StatusBadRequest)
		return
	}

	// Сохраняем изображение
	file, header, err := r.FormFile("image")
	if err != nil {
		http.Error(w, "Image upload failed", http.StatusBadRequest)
		return
	}
	defer file.Close()

	imagePath := filepath.Join("uploads", header.Filename)
	dst, err := os.Create(imagePath)
	if err != nil {
		http.Error(w, "Could not save image", http.StatusInternalServerError)
		return
	}
	defer dst.Close()
	io.Copy(dst, file)

	// Сохраняем в базу
	tx, err := db.Begin()
	if err != nil {
		http.Error(w, "DB transaction error", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	res, err := tx.Exec(`INSERT INTO projects (name, description, image_path) VALUES (?, ?, ?)`, name, description, imagePath)
	if err != nil {
		http.Error(w, "Could not insert project", http.StatusInternalServerError)
		return
	}
	projectID, err := res.LastInsertId()
	if err != nil {
		http.Error(w, "Could not get project ID", http.StatusInternalServerError)
		return
	}

	_, err = tx.Exec(`INSERT INTO funding (project_id, target_amount) VALUES (?, ?)`, projectID, targetAmount)
	if err != nil {
		http.Error(w, "Could not insert funding info", http.StatusInternalServerError)
		return
	}

	_, err = tx.Exec(`INSERT INTO project_meta (project_id, hashtag, published) VALUES (?, ?, ?)`, projectID, hashtag, false)
	if err != nil {
		http.Error(w, "Could not insert meta info", http.StatusInternalServerError)
		return
	}

	if err := tx.Commit(); err != nil {
		http.Error(w, "Could not commit transaction", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, "Project created with ID %d", projectID)
}

func GetAllProjectsHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	rows, err := db.Query(`
		SELECT p.id, p.name, p.description, p.image_path,
		       f.current_amount, f.target_amount,
		       m.hashtag, m.published
		FROM projects p
		JOIN funding f ON p.id = f.project_id
		JOIN project_meta m ON p.id = m.project_id`)
	if err != nil {
		http.Error(w, "DB query failed", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type Project struct {
		ID            int    `json:"id"`
		Name          string `json:"name"`
		Description   string `json:"description"`
		ImagePath     string `json:"image_path"`
		CurrentAmount int    `json:"current_amount"`
		TargetAmount  int    `json:"target_amount"`
		Hashtag       string `json:"hashtag"`
		Published     bool   `json:"published"`
	}

	var projects []Project
	for rows.Next() {
		var p Project
		if err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.ImagePath, &p.CurrentAmount, &p.TargetAmount, &p.Hashtag, &p.Published); err != nil {
			http.Error(w, "Row scan failed", http.StatusInternalServerError)
			return
		}
		projects = append(projects, p)
	}

	writeJSON(w, projects)
}

func writeJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

func GetProjectByIDHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 3 {
		http.Error(w, "Invalid project ID", http.StatusBadRequest)
		return
	}
	idStr := parts[2]
	projectID, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid project ID", http.StatusBadRequest)
		return
	}

	row := db.QueryRow(`
		SELECT p.id, p.name, p.description, p.image_path,
		       f.current_amount, f.target_amount,
		       m.hashtag, m.published
		FROM projects p
		JOIN funding f ON p.id = f.project_id
		JOIN project_meta m ON p.id = m.project_id
		WHERE p.id = ?
	`, projectID)

	var p struct {
		ID            int    `json:"id"`
		Name          string `json:"name"`
		Description   string `json:"description"`
		ImagePath     string `json:"image_path"`
		CurrentAmount int    `json:"current_amount"`
		TargetAmount  int    `json:"target_amount"`
		Hashtag       string `json:"hashtag"`
		Published     bool   `json:"published"`
	}

	err = row.Scan(&p.ID, &p.Name, &p.Description, &p.ImagePath, &p.CurrentAmount, &p.TargetAmount, &p.Hashtag, &p.Published)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Project not found", http.StatusNotFound)
		} else {
			http.Error(w, "DB error", http.StatusInternalServerError)
		}
		return
	}

	writeJSON(w, p)
}

func FundProjectHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 3 {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	}
	idStr := parts[2]
	projectID, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid project ID", http.StatusBadRequest)
		return
	}

	amountStr := r.FormValue("amount")
	amount, err := strconv.Atoi(amountStr)
	if err != nil || amount <= 0 {
		http.Error(w, "Invalid amount", http.StatusBadRequest)
		return
	}

	res, err := db.Exec(`UPDATE funding SET current_amount = current_amount + ? WHERE project_id = ?`, amount, projectID)
	if err != nil {
		http.Error(w, "Could not update funding", http.StatusInternalServerError)
		return
	}

	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		http.Error(w, "Project not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Funding updated successfully"))
}
