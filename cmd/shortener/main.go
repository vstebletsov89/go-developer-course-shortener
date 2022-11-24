package main

import (
	"context"
	"fmt"
	"go-developer-course-shortener/internal/app/handlers"
	"go-developer-course-shortener/internal/app/middleware"
	"go-developer-course-shortener/internal/app/repository"
	"go-developer-course-shortener/internal/app/repository/postgres"
	"go-developer-course-shortener/internal/configs"
	"go-developer-course-shortener/internal/worker"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"

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
		log.Fatalf("Failed to read server configuration. Error: %v", err.Error())
	}

	// Server run context
	ctx, cancel := context.WithCancel(context.Background())

	var storage repository.Repository
	switch {
	case config.DatabaseDsn != "":
		conn, err := pgx.Connect(ctx, config.DatabaseDsn)
		if err != nil {
			log.Fatalf("Failed to connect to database. Error: %v", err.Error())
		}

		storage, err = postgres.NewDBRepository(conn)
		if err != nil {
			log.Fatalf("Failed to create DB repository. Error: %v", err.Error())
		}
	case config.FileStoragePath != "":
		storage = repository.NewFileRepository(config.FileStoragePath)
	default:
		storage = repository.NewInMemoryRepository()
	}

	// setup worker pool to handle delete requests
	jobs := make(chan worker.Job, worker.MaxWorkerPoolSize)
	workerPool := worker.NewWorkerPool(storage, jobs)
	go workerPool.Run(ctx)

	srv := http.Server{Addr: config.ServerAddress, Handler: service(jobs, config, storage)}

	connClosed := make(chan struct{})
	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)

	go func() {
		<-sigint

		// graceful shutdown
		if err := srv.Shutdown(ctx); err != nil {
			log.Fatalf("HTTP/HTTPS server Shutdown: %v", err)
		}

		// connection closed
		close(connClosed)

		// stop server context and release resources
		cancel()

		// close worker pool
		workerPool.ClosePool()
	}()

	if config.EnableHTTPS {
		// start https server
		log.Printf("HTTPS server started on %v", config.ServerAddress)
		// certificate and key were created by crypto/x509 package
		if err := srv.ListenAndServeTLS("cert.pem", "key.pem"); err != http.ErrServerClosed {
			log.Fatalf("HTTPS server ListenAndServe: %v", err)
		}
	} else {
		// start http server
		log.Printf("HTTP server started on %v", config.ServerAddress)
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatalf("HTTP server ListenAndServe: %v", err)
		}
	}

	// wait for graceful shutdown
	<-connClosed
	// release resources
	storage.ReleaseStorage()
	log.Println("Server Shutdown gracefully")
}

func service(job chan worker.Job, config *configs.Config, storage repository.Repository) http.Handler {
	handler := handlers.NewHTTPHandler(config, storage)

	r := chi.NewRouter()
	r.Use(middleware.GzipHandle, middleware.AuthHandle, middleware.TrustedSubnetHandle(config.TrustedSubnet))

	// routing
	r.Post("/", handler.HandlerPOST)
	r.Post("/api/shorten", handler.HandlerJSONPOST)
	r.Post("/api/shorten/batch", handler.HandlerBatchPOST)
	r.Get("/{ID}", handler.HandlerGET)
	r.Get("/api/user/urls", handler.HandlerUserStorageGET)
	r.Get("/ping", handler.HandlerPing)
	r.Delete("/api/user/urls", handler.HandlerUseStorageDELETE(job))
	r.Get("/api/internal/stats", handler.HandlerStats)

	return r
}
