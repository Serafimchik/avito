package handlers

import (
	"encoding/json"
	"net/http"

	"avito-shop/internal/services"

	"github.com/go-chi/chi/v5"
)

func BuyItem(w http.ResponseWriter, r *http.Request) {
	item := chi.URLParam(r, "item")
	username := r.Context().Value("username").(string)

	if err := services.BuyItem(username, item); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Item bought successfully"})
}
