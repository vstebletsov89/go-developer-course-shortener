package main

import (
	"github.com/go-chi/chi/v5"
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

	// маршрутизация запросов обработчику
	r.Post("/", handlers.HandlerPOST)
	r.Get("/", handlers.HandlerGET)

	// запуск сервера с адресом localhost, порт 8080
	log.Fatal(http.ListenAndServe(configs.ServerAddress, r))
}
