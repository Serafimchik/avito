package handlers

import (
	"encoding/json"
	"net/http"

	"avito-shop/internal/services"
)

func GetInfo(w http.ResponseWriter, r *http.Request) {
	username := r.Context().Value("username").(string)
	info, err := services.GetUserInfo(username)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(info)
}
