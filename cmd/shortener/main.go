package main

import (
	"context"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v4"
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

	// routing
	r.Post("/", handler.HandlerPOST)
	r.Post("/api/shorten", handler.HandlerJSONPOST)
	r.Post("/api/shorten/batch", handler.HandlerBatchPOST)
	r.Get("/{ID}", handler.HandlerGET)
	r.Get("/api/user/urls", handler.HandlerUserStorageGET)
	r.Get("/ping", handler.HandlerPing)

	//TODO:
	// Задание для трека «Сервис сокращения URL»
	// +Сделайте в таблице базы данных с сокращёнными URL уникальный индекс для поля с исходным URL.
	// Это позволит избавиться от дублирующих записей в базе данных.
	//	При попытке пользователя сократить уже имеющийся в базе URL через хендлеры
	//	POST / и POST /api/shorten сервис должен вернуть HTTP-статус 409 Conflict,
	//	а в теле ответа — уже имеющийся сокращённый URL в правильном для хендлера формате.
	//	Стратегии реализации:
	//  Чтобы не проверять наличие оригинального URL в базе данных отдельным запросом,
	//  можно воспользоваться конструкцией INSERT ... ON CONFLICT в PostgreSQL.
	//  Однако в таком случае придётся самостоятельно возвращать и проверять собственную ошибку.
	//	Чтобы определить тип ошибки PostgreSQL, с которой завершился запрос,
	//	можно воспользоваться библиотекой github.com/jackc/pgerrcode,
	//	в частности pgerrcode.UniqueViolation.
	//	В таком случае придётся делать дополнительный запрос к хранилищу, чтобы определить сокращённый вариант URL.

	log.Fatal(http.ListenAndServe(config.ServerAddress, r))
}
