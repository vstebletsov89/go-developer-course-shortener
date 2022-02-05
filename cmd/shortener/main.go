package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
)

const (
	serverHost    = "localhost"
	serverPort    = "8080"
	serverAddress = serverHost + ":" + serverPort
)

var Repository = make(map[int]string)

func GetNextID() int {
	return len(Repository) + 1
}

func HandlerShortener(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		strID := r.URL.Path
		id, err := strconv.Atoi(strID[1:])
		if err != nil || id < 1 {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		log.Printf("ID: %d", id)
		originalURL := Repository[id]
		if originalURL == "" {
			http.Error(w, "ID not found", http.StatusBadRequest)
			return
		}
		log.Printf("Original URL: %s", originalURL)
		w.Header().Add("Location", originalURL)
		w.WriteHeader(http.StatusTemporaryRedirect)

	case "POST":
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		log.Printf("Long URL: %v", string(body))
		longURL, err := url.Parse(string(body))
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		id := GetNextID()
		Repository[id] = longURL.String()
		shortURL := fmt.Sprintf("http://%v/%d", serverAddress, id)
		log.Printf("Short URL: %v", shortURL)

		w.WriteHeader(http.StatusCreated)
		_, err = w.Write([]byte(shortURL))
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	default:
		w.WriteHeader(http.StatusBadRequest)
		log.Printf("Unexpected error")
	}
}

func main() {
	log.Printf("Server started on %v\n", serverAddress)
	// маршрутизация запросов обработчику
	http.HandleFunc("/", HandlerShortener)
	// запуск сервера с адресом localhost, порт 8080
	log.Fatal(http.ListenAndServe(serverAddress, nil))
}
