package handlers

import (
	"avito/internal/services"
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx"
)

func SendCoin(w http.ResponseWriter, r *http.Request) {
	fromUserId, ok := r.Context().Value("user_id").(int)
	if !ok {
		http.Error(w, "Invalid user ID in context", http.StatusInternalServerError)
		return
	}

	var req struct {
		ToUser string `json:"toUser"`
		Amount int64  `json:"amount"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.ToUser == "" || req.Amount <= 0 {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	var toUserId int
	query := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar).
		Select("id").
		From("users").
		Where(squirrel.Eq{"username": req.ToUser})

	sqlStr, args, err := query.ToSql()
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	err = services.Pool.QueryRow(context.Background(), sqlStr, args...).Scan(&toUserId)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			http.Error(w, "Recipient not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if err := services.SendCoins(fromUserId, toUserId, req.Amount); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Coins sent successfully"})
}
