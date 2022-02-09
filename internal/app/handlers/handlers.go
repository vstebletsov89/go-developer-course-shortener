package handlers

import (
	"fmt"
	"go-developer-course-shortener/configs"
	"go-developer-course-shortener/internal/app/repository"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
)

func HandlerPOST(w http.ResponseWriter, r *http.Request) {
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
	if longURL.String() == "" {
		http.Error(w, "URL must not be empty", http.StatusBadRequest)
		return
	}
	id := repository.SaveURL(longURL.String())
	shortURL := fmt.Sprintf("http://%v/%d", configs.ServerAddress, id)
	log.Printf("Short URL: %v", shortURL)

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusCreated)
	_, err = w.Write([]byte(shortURL))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	return
}

func HandlerGET(w http.ResponseWriter, r *http.Request) {
	strID := r.URL.Path
	id, err := strconv.Atoi(strID[1:])
	if err != nil || id < 1 {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}
	log.Printf("ID: %d", id)
	originalURL := repository.GetURL(id)
	if originalURL == "" {
		http.Error(w, "ID not found", http.StatusBadRequest)
		return
	}
	log.Printf("Original URL: %s", originalURL)
	w.Header().Add("Location", originalURL)
	w.WriteHeader(http.StatusTemporaryRedirect)
	return
}
