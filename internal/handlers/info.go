package handlers

import (
	"encoding/json"
	"net/http"

	"avito/internal/services"
)

func GetInfo(w http.ResponseWriter, r *http.Request) {
	userId, ok := r.Context().Value("user_id").(int)
	if !ok {
		http.Error(w, "Invalid user ID in context", http.StatusInternalServerError)
		return
	}

	info, err := services.GetUserInfo(userId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(info)
}
