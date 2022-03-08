package main

import (
	"context"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v4"
	_ "github.com/jackc/pgx/v4"
	"go-developer-course-shortener/internal/app/handlers"
	"go-developer-course-shortener/internal/app/middleware"
	"go-developer-course-shortener/internal/app/repository"
	"go-developer-course-shortener/internal/configs"
	"log"
	"net/http"
	"os"
)

func main() {
	log.SetOutput(os.Stdout)
	config, err := configs.ReadConfig()
	if err != nil {
		log.Fatal("Failed to read server configuration")
	}

	var storage repository.Repository
	if config.DatabaseDsn != "" {
		conn, err := pgx.Connect(context.Background(), config.DatabaseDsn)
		if err != nil {
			log.Fatal("Failed to connect to database")
		}
		defer conn.Close(context.Background())

		storage, err = repository.NewDBRepository(conn)
		if err != nil {
			log.Fatal("Failed to create DB repository")
		}
	} else if config.FileStoragePath != "" {
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
	r.Get("/ping", handler.HandlerPing)

	log.Fatal(http.ListenAndServe(config.ServerAddress, r))
}
