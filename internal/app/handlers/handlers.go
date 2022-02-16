package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi/v5"
	"go-developer-course-shortener/internal/app/repository"
	"go-developer-course-shortener/internal/configs"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
)

type RequestJSON struct {
	URL string `json:"url"`
}

type ResponseJSON struct {
	Result string `json:"result"`
}

//TODO: refactor code to extract common parts for POST handlers (block Long URL -> Short URL)
//TODO: add unit tests
//TODO: push to github to check tests
func HandlerJsonPOST(w http.ResponseWriter, r *http.Request) {
	var request RequestJSON

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	log.Printf("Request JSON: %+v", request)

	longURL, err := url.Parse(string(request.URL))
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
	response := ResponseJSON{Result: shortURL}
	log.Printf("Response JSON: %+v", response)

	buf := bytes.NewBuffer([]byte{})
	encoder := json.NewEncoder(buf)
	encoder.SetEscapeHTML(false)
	encoder.Encode(response)
	log.Printf("Encoded JSON: %s", buf.String())

	w.Header().Set(configs.ContentType, configs.ContentValueJSON)
	w.WriteHeader(http.StatusCreated)

	_, err = w.Write(buf.Bytes())
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
}

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

	w.Header().Set(configs.ContentType, configs.ContentValue)
	w.WriteHeader(http.StatusCreated)
	_, err = w.Write([]byte(shortURL))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
}

func HandlerGET(w http.ResponseWriter, r *http.Request) {
	strID := chi.URLParam(r, "ID")
	log.Printf("strID: `%s`", strID)
	id, err := strconv.Atoi(strID)
	if err != nil || id < 1 {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}
	log.Printf("ID: %d", id)
	originalURL, err := repository.GetURL(id)
	if err != nil {
		http.Error(w, "ID not found", http.StatusBadRequest)
		return
	}
	log.Printf("Original URL: %s", originalURL)
	w.Header().Set(configs.ContentType, configs.ContentValue)
	w.Header().Set("Location", originalURL)
	w.WriteHeader(http.StatusTemporaryRedirect)
}
