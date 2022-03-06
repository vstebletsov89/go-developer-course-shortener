package main

import (
	"github.com/go-chi/chi/v5"
	"go-developer-course-shortener/internal/app/handlers"
	"go-developer-course-shortener/internal/app/middleware"
	"go-developer-course-shortener/internal/app/repository"
	"go-developer-course-shortener/internal/configs"
	"log"
	"net/http"
	"os"
)

func main() {
	//,,,
	log.SetOutput(os.Stdout)
	config, err := configs.ReadConfig()
	if err != nil {
		log.Fatal("Failed to read server configuration")
	}

	var storage repository.Repository
	if config.FileStoragePath != "" {
		storage = repository.NewFileRepository(config.FileStoragePath)
	} else {
		storage = repository.NewInMemoryRepository()
	}
	handler := handlers.NewHTTPHandler(config, storage)

	log.Printf("Server started on %v", config.ServerAddress)
	r := chi.NewRouter()
	r.Use(middleware.GzipHandle, middleware.AuthHandle)

	// маршрутизация запросов обработчику
	r.Post("/", handler.HandlerPOST)
	r.Post("/api/shorten", handler.HandlerJSONPOST)
	r.Get("/{ID}", handler.HandlerGET)
	r.Get("/api/user/urls", handler.HandlerUserStorageGET)

	log.Fatal(http.ListenAndServe(config.ServerAddress, r))
}
