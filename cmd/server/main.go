package main

import (
	"log"
	"net/http"

	"avito-shop/internal/appMiddleware"
	"avito-shop/internal/handlers"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	r := chi.NewRouter()

	// Подключение middleware
	r.Use(middleware.Recoverer)
	r.Use(middleware.Logger)

	// Обработчики без авторизации
	r.Post("/api/auth", handlers.Auth)

	// Обработчики с авторизацией
	r.Group(func(r chi.Router) {
		r.Use(appMiddleware.AuthMiddleware)

		r.Get("/api/info", handlers.GetInfo)
		r.Post("/api/sendCoin", handlers.SendCoin)
		r.Get("/api/buy/{item}", handlers.BuyItem)
	})

	log.Println("Server started on port 8080")
	if err := http.ListenAndServe(":8080", r); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
