package middleware

import (
	"context"
	"net/http"
	"strings"

	"crowdfunding-service/pkg"

	"github.com/golang-jwt/jwt/v5"
)

var JwtKey = []byte("my_super_secret_key")

func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Missing Authorization header", http.StatusUnauthorized)
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			http.Error(w, "Invalid token format", http.StatusUnauthorized)
			return
		}

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return JwtKey, nil
		})

		if err != nil || !token.Valid {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		// Извлекаем user_id из claims
		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			if userIDFloat, ok := claims["user_id"].(float64); ok {
				userID := int(userIDFloat)
				ctx := context.WithValue(r.Context(), pkg.UserIDKey, userID)
				next(w, r.WithContext(ctx))
				return
			}
		}

		http.Error(w, "Invalid token claims", http.StatusUnauthorized)
	}
}
