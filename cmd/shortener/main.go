package main

import (
	"github.com/go-chi/chi/v5"
	"go-developer-course-shortener/internal/app/handlers"
	"go-developer-course-shortener/internal/app/repository"
	"go-developer-course-shortener/internal/configs"
	"log"
	"net/http"
	"os"
)

func main() {
	log.SetOutput(os.Stdout)
	config := configs.ReadConfig()

	var storage repository.Repository
	if config.FileStoragePath != configs.FileStorageDefault {
		storage = repository.NewFileRepository(config.FileStoragePath)
	} else {
		storage = repository.NewInMemoryRepository()
	}
	handler := handlers.NewHTTPHandler(config, storage)

	log.Printf("Server started on %v", config.ServerAddress)
	r := chi.NewRouter()

	// маршрутизация запросов обработчику
	r.Post("/", handler.HandlerPOST)
	r.Post("/api/shorten", handler.HandlerJSONPOST)
	r.Get("/{ID}", handler.HandlerGET)

	log.Fatal(http.ListenAndServe(config.ServerAddress, r))
}
