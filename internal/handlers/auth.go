package handlers

import (
	"avito/internal/services"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

func Auth(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	// Декодируем запрос
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Username == "" || req.Password == "" {
		log.Printf("Invalid request: %v", err)
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	log.Printf("Checking if user exists: %s", req.Username)

	// Проверяем, существует ли пользователь
	userExists, userId, userHash, err := services.CheckUserExists(req.Username)
	if err != nil {
		log.Printf("Error checking user existence: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if userExists {
		log.Printf("User exists: %s", req.Username)

		// Проверяем пароль
		if err := services.VerifyPassword(req.Password, userHash); err != nil {
			log.Printf("Invalid credentials for user: %s", req.Username)
			http.Error(w, "Invalid credentials", http.StatusUnauthorized)
			return
		}
	} else {
		log.Printf("User does not exist, registering: %s", req.Username)

		// Хэшируем пароль
		hashedPassword, err := services.HashPassword(req.Password)
		if err != nil {
			log.Printf("Failed to hash password: %v", err)
			http.Error(w, "Failed to hash password", http.StatusInternalServerError)
			return
		}

		// Регистрируем нового пользователя
		userId, err = services.RegisterUser(req.Username, hashedPassword)
		if err != nil {
			log.Printf("Failed to register user: %v", err)
			http.Error(w, "Failed to register user", http.StatusInternalServerError)
			return
		}
	}

	log.Printf("Generating token for user: %s", req.Username)

	// Генерируем токен
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userId,
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
	})

	tokenString, err := token.SignedString([]byte("secret-key"))
	if err != nil {
		log.Printf("Failed to generate token: %v", err)
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	// Возвращаем токен
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"token": tokenString})
}
