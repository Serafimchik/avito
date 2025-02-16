package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"avito/internal/services"

	"github.com/Masterminds/squirrel"
	"github.com/golang-jwt/jwt/v4"
	"github.com/jackc/pgx/v5"
)

func Auth(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Username == "" || req.Password == "" {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	hashedPassword, err := services.HashPassword(req.Password)
	if err != nil {
		http.Error(w, "Failed to hash password", http.StatusInternalServerError)
		return
	}

	err = services.RegisterUser(req.Username, hashedPassword)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) || errors.Is(err, errors.New("user already exists")) {

		} else {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
	}

	var userHash string
	query := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar).
		Select("password_hash").
		From("users").
		Where(squirrel.Eq{"username": req.Username})

	sqlStr, args, err := query.ToSql()
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	err = services.Pool.QueryRow(context.Background(), sqlStr, args...).Scan(&userHash)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			http.Error(w, "User not found", http.StatusUnauthorized)
			return
		}
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if err := services.VerifyPassword(req.Password, userHash); err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	var userId int
	query = squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar).
		Select("id").
		From("users").
		Where(squirrel.Eq{"username": req.Username})

	sqlStr, args, err = query.ToSql()
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	err = services.Pool.QueryRow(context.Background(), sqlStr, args...).Scan(&userId)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userId,
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
	})
	tokenString, err := token.SignedString([]byte("secret-key"))
	if err != nil {
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	ctx := context.WithValue(r.Context(), "user_id", userId)
	r = r.WithContext(ctx)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"token": tokenString})
}
