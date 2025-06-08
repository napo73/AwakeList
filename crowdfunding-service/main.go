package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"

	"crowdfunding-service/db"
	"crowdfunding-service/handlers"
	"crowdfunding-service/middleware" // добавь
)

func main() {
	err := os.MkdirAll("uploads", os.ModePerm)
	if err != nil {
		log.Fatal("Failed to create uploads directory:", err)
	}

	database, err := db.InitDB("projects.db")
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer database.Close()

	http.HandleFunc("/projects", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			// Добавляем проверку авторизации
			middleware.AuthMiddleware(func(w http.ResponseWriter, r *http.Request) {
				handlers.CreateProjectHandler(w, r, database)
			})(w, r)
		case http.MethodGet:
			handlers.GetAllProjectsHandler(w, r, database)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	http.HandleFunc("/projects/", func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet:
			handlers.GetProjectByIDHandler(w, r, database)
		case r.Method == http.MethodPost && filepath.Base(r.URL.Path) == "fund":
			middleware.AuthMiddleware(func(w http.ResponseWriter, r *http.Request) {
				handlers.FundProjectHandler(w, r, database)
			})(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	log.Println("Crowdfunding service started on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal("Server error:", err)
	}
}
