package main

import (
	"context"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v4"
	"go-developer-course-shortener/internal/app/handlers"
	"go-developer-course-shortener/internal/app/middleware"
	"go-developer-course-shortener/internal/app/repository"
	"go-developer-course-shortener/internal/configs"
	"go-developer-course-shortener/internal/worker"
	"log"
	"net/http"
	"os"
	"time"
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

	//setup worker pool to handle delete requests
	jobs := make(chan worker.Job, worker.MaxWorkerPoolSize)
	workerPool := worker.NewWorkerPool(storage, jobs)
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(time.Second*20))
	defer cancel()
	go workerPool.Run(ctx)

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

	//TODO:
	//	Задание для трека «Сервис сокращения URL»
	//	Сделайте в таблице базы данных с сокращёнными URL дополнительное поле с флагом,
	//	указывающим на то, что URL должен считаться удалённым.
	//		Далее добавьте в сервис новый асинхронный хендлер DELETE /api/user/urls,
	//		который принимает список идентификаторов сокращённых URL для удаления в формате:
	//		[ "a", "b", "c", "d", ...]
	//	В случае успешного приёма запроса хендлер должен возвращать HTTP-статус 202 Accepted.
	//	Фактический результат удаления может происходить позже — каким-либо образом оповещать пользователя
	//	об успешности или неуспешности не нужно.
	//	Успешно удалить URL может пользователь, его создавший.
	//	При запросе удалённого URL с помощью хендлера GET /{id} нужно вернуть статус 410 Gone.
	//	Совет:
	//	Для эффективного проставления флага удаления в базе данных используйте множественное обновление (batch update).
	//	Используйте паттерн fanIn для максимального наполнения буфера объектов обновления.
}
