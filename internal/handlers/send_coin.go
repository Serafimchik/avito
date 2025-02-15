package handlers

import (
	"encoding/json"
	"net/http"

	"avito-shop/internal/services"
)

func SendCoin(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ToUser string `json:"toUser"`
		Amount int64  `json:"amount"`
	}

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil || req.ToUser == "" || req.Amount <= 0 {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	username := r.Context().Value("username").(string)
	if err := services.SendCoins(username, req.ToUser, req.Amount); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Coins sent successfully"})
}
