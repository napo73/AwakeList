package main

import (
	"log"
	"net/http"
)

func main() {
	db, err := InitDB("auth.db")
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// Регистрируем CORS-обёрнутые обработчики
	http.HandleFunc("/register", WithCORS(func(w http.ResponseWriter, r *http.Request) {
		RegisterHandler(w, r, db)
	}))
	http.HandleFunc("/login", WithCORS(func(w http.ResponseWriter, r *http.Request) {
		LoginHandler(w, r, db)
	}))

	log.Println("Auth service started on :8000")
	// Запускаем сервер после регистрации маршрутов
	if err := http.ListenAndServe(":8000", nil); err != nil {
		log.Fatal("Server error:", err)
	}
}
