package main

type User struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Password string `json:"password"`
	Email    string `json:"email"`
	Role     bool   `json:"role"` // по умолчанию false (0)
}
