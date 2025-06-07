package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

func RegisterHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST allowed", http.StatusMethodNotAllowed)
		return
	}

	var u User
	if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	u.Username = strings.TrimSpace(u.Username)
	u.Email = strings.TrimSpace(u.Email)

	if len(u.Password) < 8 {
		http.Error(w, "Password must be at least 8 characters", http.StatusBadRequest)
		return
	}
	if u.Username == "" || u.Email == "" {
		http.Error(w, "Username and email are required", http.StatusBadRequest)
		return
	}

	// Проверка, существует ли уже пользователь с таким username или email
	var exists int
	err := db.QueryRow("SELECT COUNT(*) FROM users WHERE username = ? OR email = ?", u.Username, u.Email).Scan(&exists)
	if err != nil {
		log.Println("Check user error:", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	if exists > 0 {
		http.Error(w, "Username or email already exists", http.StatusConflict)
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Println("Password hash error:", err)
		http.Error(w, "Failed to hash password", http.StatusInternalServerError)
		return
	}

	_, err = db.Exec("INSERT INTO users (username, password, email, role) VALUES (?, ?, ?, ?)",
		u.Username, string(hash), u.Email, u.Role)
	if err != nil {
		log.Println("Registration error:", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("User registered"))
}

func LoginHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST allowed", http.StatusMethodNotAllowed)
		return
	}

	var u User
	if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	var stored User
	err := db.QueryRow("SELECT id, password FROM users WHERE username = ?", u.Username).Scan(&stored.ID, &stored.Password)
	if err == sql.ErrNoRows {
		http.Error(w, "Invalid username or password", http.StatusUnauthorized)
		return
	} else if err != nil {
		log.Println("Login DB error:", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(stored.Password), []byte(u.Password))
	if err != nil {
		http.Error(w, "Invalid username or password", http.StatusUnauthorized)
		return
	}

	w.Write([]byte("Login successful"))
}
