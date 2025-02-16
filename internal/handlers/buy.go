package handlers

import (
	"encoding/json"
	"net/http"

	"avito/internal/services"

	"github.com/go-chi/chi/v5"
)

func BuyItem(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("user_id").(int)
	if !ok {
		http.Error(w, "Invalid user ID in context", http.StatusInternalServerError)
		return
	}

	itemType := chi.URLParam(r, "item")
	if itemType == "" {
		http.Error(w, "Item type is required", http.StatusBadRequest)
		return
	}

	if err := services.BuyItem(userID, itemType); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Item bought successfully"})
}
