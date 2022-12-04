package main

import (
	"context"
	"fmt"
	"go-developer-course-shortener/internal/app/handlers"
	"go-developer-course-shortener/internal/app/middleware"
	"go-developer-course-shortener/internal/app/repository"
	"go-developer-course-shortener/internal/app/repository/postgres"
	"go-developer-course-shortener/internal/app/service"
	"go-developer-course-shortener/internal/configs"
	"go-developer-course-shortener/internal/worker"
	pb "go-developer-course-shortener/proto"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"log"
	"net"
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

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// error group to control server instances
	g, ctx := errgroup.WithContext(ctx)

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

	_, subnet, err := net.ParseCIDR(config.TrustedSubnet)
	if err != nil {
		log.Printf("Failed to read trusted subnet parameter. Error: %v", err.Error())
	}

	// create new service for all servers
	svc := service.NewService(storage, jobs, subnet, config.BaseURL)
	var httpSrv http.Server
	var grpcSrv *grpc.Server

	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)

	// start http/https server
	g.Go(func() error {
		httpSrv = http.Server{Addr: config.ServerAddress, Handler: NewHTTPHandler(svc)}
		if config.EnableHTTPS {
			// start https server
			log.Printf("HTTPS server started on %v", config.ServerAddress)
			// certificate and key were created by crypto/x509 package
			if err := httpSrv.ListenAndServeTLS("cert.pem", "key.pem"); err != http.ErrServerClosed {
				log.Fatalf("HTTPS server ListenAndServe: %v", err)
			}
		} else {
			// start http server
			log.Printf("HTTP server started on %v", config.ServerAddress)
			if err := httpSrv.ListenAndServe(); err != http.ErrServerClosed {
				log.Fatalf("HTTP server ListenAndServe: %v", err)
			}
		}
		return nil
	})

	// start grpc server
	g.Go(func() error {
		listen, err := net.Listen("tcp", fmt.Sprintf(":%d", config.GrpcPort))
		if err != nil {
			log.Fatalf("GRPC server net.Listen: %v", err)
		}

		grpcSrv = grpc.NewServer(grpc.UnaryInterceptor(handlers.UnaryInterceptor))
		pb.RegisterShortenerServer(grpcSrv, handlers.NewGrpcHandler(svc))

		log.Printf("GRPC server started on %v", config.GrpcPort)
		// start grc server
		if err := grpcSrv.Serve(listen); err != nil {
			log.Fatal(err)
		}
		return nil
	})

	<-sigint

	// graceful shutdown
	if err := httpSrv.Shutdown(ctx); err != nil {
		log.Fatalf("HTTP/HTTPS server Shutdown: %v", err)
	}

	grpcSrv.GracefulStop()

	// stop server context and release resources
	cancel()

	// close worker pool
	workerPool.ClosePool()

	// release resources
	storage.ReleaseStorage()
	log.Println("Server Shutdown gracefully")

	err = g.Wait()
	if err != nil {
		log.Fatalf("error group: %v", err)
	}
}

func NewHTTPHandler(svc *service.Service) http.Handler {
	handler := handlers.NewHTTPHandler(svc)

	r := chi.NewRouter()
	r.Use(middleware.GzipHandle, middleware.AuthHandle)

	// routing
	r.Post("/", handler.HandlerPOST)
	r.Post("/api/shorten", handler.HandlerJSONPOST)
	r.Post("/api/shorten/batch", handler.HandlerBatchPOST)
	r.Get("/{ID}", handler.HandlerGET)
	r.Get("/api/user/urls", handler.HandlerUserStorageGET)
	r.Get("/ping", handler.HandlerPing)
	r.Delete("/api/user/urls", handler.HandlerUseStorageDELETE())
	r.Get("/api/internal/stats", handler.HandlerStats)

	return r
}
