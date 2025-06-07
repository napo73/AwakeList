package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"

	"crowdfunding-service/db"
	"crowdfunding-service/handlers"
)

func main() {
	// Создание директории для загрузок, если её нет
	err := os.MkdirAll("uploads", os.ModePerm)
	if err != nil {
		log.Fatal("Failed to create uploads directory:", err)
	}

	// Инициализация базы данных
	database, err := db.InitDB("projects.db")
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer database.Close()

	// Маршруты
	http.HandleFunc("/projects", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			handlers.CreateProjectHandler(w, r, database)
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
			handlers.FundProjectHandler(w, r, database)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Сервер
	log.Println("Crowdfunding service started on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal("Server error:", err)
	}
}
