package main

import (
	"context"
	"go-developer-course-shortener/internal/app/handlers"
	"go-developer-course-shortener/internal/app/middleware"
	"go-developer-course-shortener/internal/app/repository"
	"go-developer-course-shortener/internal/configs"
	"go-developer-course-shortener/internal/worker"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v4"
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

	// setup worker pool to handle delete requests
	jobs := make(chan worker.Job, worker.MaxWorkerPoolSize)
	workerPool := worker.NewWorkerPool(storage, jobs)
	go workerPool.Run(context.Background())

	log.Printf("Server started on %v", config.ServerAddress)
	r := chi.NewRouter()
	r.Use(middleware.GzipHandle, middleware.AuthHandle)

	// routing
	r.Post("/", handler.HandlerPOST)
	r.Post("/api/shorten", handler.HandlerJSONPOST)
	r.Post("/api/shorten/batch", handler.HandlerBatchPOST)
	r.Get("/{ID}", handler.HandlerGET)
	r.Get("/api/user/urls", handler.HandlerUserStorageGET)
	r.Get("/ping", handler.HandlerPing)
	r.Delete("/api/user/urls", handler.HandlerUseStorageDELETE(jobs))

	log.Fatal(http.ListenAndServe(config.ServerAddress, r))
}
