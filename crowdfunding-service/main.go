package main

import (
	"log"
	"net/http"
	"os"
	"strings"

	"crowdfunding-service/db"
	"crowdfunding-service/handlers"
	"crowdfunding-service/middleware"
)

func main() {
	// Создание папки для изображений
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

	// Создаем маршрутизатор
	mux := http.NewServeMux()

	// /projects — GET и POST
	mux.HandleFunc("/projects", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			// Защищённый маршрут
			middleware.AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				handlers.CreateProjectHandler(w, r, database)
			})).ServeHTTP(w, r)
		case http.MethodGet:
			handlers.GetAllProjectsHandler(w, r, database)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// /projects/{id} и /projects/{id}/fund
	mux.HandleFunc("/projects/", func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet:
			handlers.GetProjectByIDHandler(w, r, database)
		case r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/fund"):
			middleware.AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				handlers.FundProjectHandler(w, r, database)
			})).ServeHTTP(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Оборачиваем весь mux в CORS middleware
	handlerWithCORS := middleware.CORS(mux)

	log.Println("Crowdfunding service started on :8080")
	if err := http.ListenAndServe(":8080", handlerWithCORS); err != nil {
		log.Fatal("Server error:", err)
	}
}
