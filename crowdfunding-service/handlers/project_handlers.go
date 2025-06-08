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

	"crowdfunding-service/pkg"
)

// Project выводится клиенту
type Project struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	ImagePath   string `json:"image_path"`
	Target      int    `json:"target_amount"`
	Current     int    `json:"current_amount"`
	Hashtag     string `json:"hashtag"`
	Published   bool   `json:"published"`
}

// POST /projects
func CreateProjectHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	userIDVal := r.Context().Value(pkg.UserIDKey)
	if userIDVal == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	userID := userIDVal.(int)

	if err := r.ParseMultipartForm(10 << 20); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	name := r.FormValue("name")
	description := r.FormValue("description")
	targetStr := r.FormValue("target")
	hashtag := r.FormValue("hashtag")

	if name == "" || targetStr == "" {
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return
	}

	target, err := strconv.Atoi(targetStr)
	if err != nil {
		http.Error(w, "Invalid target amount", http.StatusBadRequest)
		return
	}

	// Обработка изображения
	file, handler, err := r.FormFile("image")
	if err != nil {
		http.Error(w, "Image upload failed", http.StatusBadRequest)
		return
	}
	defer file.Close()

	imagePath := filepath.Join("uploads", handler.Filename)
	dst, err := os.Create(imagePath)
	if err != nil {
		http.Error(w, "Failed to save image", http.StatusInternalServerError)
		return
	}
	defer dst.Close()
	io.Copy(dst, file)

	// Вставка в projects
	res, err := db.Exec(
		`INSERT INTO projects (name, description, image_path, created_by) VALUES (?, ?, ?, ?)`,
		name, description, imagePath, userID,
	)
	if err != nil {
		http.Error(w, "DB error on project insert", http.StatusInternalServerError)
		return
	}
	projectID, _ := res.LastInsertId()

	// Вставка в funding
	_, err = db.Exec(`INSERT INTO funding (project_id, target_amount) VALUES (?, ?)`, projectID, target)
	if err != nil {
		http.Error(w, "DB error on funding insert", http.StatusInternalServerError)
		return
	}

	// Вставка в project_meta
	_, err = db.Exec(`INSERT INTO project_meta (project_id, hashtag, published) VALUES (?, ?, 0)`, projectID, hashtag)
	if err != nil {
		http.Error(w, "DB error on metadata insert", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, "Project %s created with ID %d", name, projectID)
}

// GET /projects
func GetAllProjectsHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	rows, err := db.Query(`
		SELECT p.id, p.name, p.description, p.image_path, f.target_amount, f.current_amount, m.hashtag, m.published
		FROM projects p
		JOIN funding f ON p.id = f.project_id
		JOIN project_meta m ON p.id = m.project_id
	`)
	if err != nil {
		http.Error(w, "DB error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var projects []Project
	for rows.Next() {
		var p Project
		err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.ImagePath, &p.Target, &p.Current, &p.Hashtag, &p.Published)
		if err != nil {
			http.Error(w, "Row scan error", http.StatusInternalServerError)
			return
		}
		projects = append(projects, p)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(projects)
}

// GET /projects/{id}
func GetProjectByIDHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	idStr := strings.TrimPrefix(r.URL.Path, "/projects/")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	var p Project
	err = db.QueryRow(`
		SELECT p.id, p.name, p.description, p.image_path, f.target_amount, f.current_amount, m.hashtag, m.published
		FROM projects p
		JOIN funding f ON p.id = f.project_id
		JOIN project_meta m ON p.id = m.project_id
		WHERE p.id = ?
	`, id).Scan(&p.ID, &p.Name, &p.Description, &p.ImagePath, &p.Target, &p.Current, &p.Hashtag, &p.Published)

	if err == sql.ErrNoRows {
		http.Error(w, "Project not found", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, "DB error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(p)
}

// POST /projects/{id}/fund
func FundProjectHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	idStr := strings.TrimPrefix(r.URL.Path, "/projects/")
	idStr = strings.TrimSuffix(idStr, "/fund")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid project ID", http.StatusBadRequest)
		return
	}

	amountStr := r.URL.Query().Get("amount")
	amount, err := strconv.Atoi(amountStr)
	if err != nil || amount <= 0 {
		http.Error(w, "Invalid amount", http.StatusBadRequest)
		return
	}

	_, err = db.Exec(`UPDATE funding SET current_amount = current_amount + ? WHERE project_id = ?`, amount, id)
	if err != nil {
		http.Error(w, "Failed to fund project", http.StatusInternalServerError)
		return
	}

	w.Write([]byte("Project funded"))
}
