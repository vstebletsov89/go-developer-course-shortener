package main

import (
	"go-developer-course-shortener/configs"
	"go-developer-course-shortener/internal/app/handlers"
	"log"
	"net/http"
)

func main() {
	log.Printf("Server started on %v\n", configs.ServerAddress)
	// маршрутизация запросов обработчику
	http.HandleFunc("/", handlers.HandlerShortener)
	// запуск сервера с адресом localhost, порт 8080
	log.Fatal(http.ListenAndServe(configs.ServerAddress, nil))
}
