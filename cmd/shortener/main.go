package main

import (
	"context"
	"fmt"
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

var (
	BuildVersion = "N/A"
	BuildDate    = "N/A"
	BuildCommit  = "N/A"
)

func main() {
	fmt.Printf("Build version: %s\n", BuildVersion)
	fmt.Printf("Build date: %s\n", BuildDate)
	fmt.Printf("Build commit: %s\n", BuildCommit)

	log.SetOutput(os.Stdout)
	config, err := configs.ReadConfig()
	if err != nil {
		log.Panicln("Failed to read server configuration. Error: " + err.Error())
	}

	var storage repository.Repository
	switch {
	case config.DatabaseDsn != "":
		conn, err := pgx.Connect(context.Background(), config.DatabaseDsn)
		if err != nil {
			log.Panicln("Failed to connect to database. Error: " + err.Error())
		}
		defer conn.Close(context.Background())

		storage, err = repository.NewDBRepository(conn)
		if err != nil {
			log.Panicln("Failed to create DB repository. Error: " + err.Error())
		}
	case config.FileStoragePath != "":
		storage = repository.NewFileRepository(config.FileStoragePath)
	default:
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

	log.Panicln(http.ListenAndServe(config.ServerAddress, r))
}
