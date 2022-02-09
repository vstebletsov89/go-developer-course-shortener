package main

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"go-developer-course-shortener/configs"
	"go-developer-course-shortener/internal/app/handlers"
	"log"
	"net/http"
	"os"
)

func main() {
	log.SetOutput(os.Stdout)
	log.Printf("Server started on %v\n", configs.ServerAddress)
	r := chi.NewRouter()

	// зададим встроенные middleware, чтобы улучшить стабильность приложения
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// маршрутизация запросов обработчику
	r.Get("/", handlers.HandlerShortener)
	r.Post("/", handlers.HandlerShortener)

	// запуск сервера с адресом localhost, порт 8080
	log.Fatal(http.ListenAndServe(configs.ServerAddress, r))
}
