package handlers

import (
	"bytes"
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"go-developer-course-shortener/internal/app/repository"
	"go-developer-course-shortener/internal/app/utils"
	"go-developer-course-shortener/internal/configs"
	"io"
	"log"
	"net/http"
	"strconv"
)

type Handler struct {
	config  *configs.Config
	storage repository.Repository
}

type RequestJSON struct {
	URL string `json:"url"`
}

type ResponseJSON struct {
	Result string `json:"result"`
}

func NewHTTPHandler(cfg *configs.Config, s repository.Repository) *Handler {
	return &Handler{config: cfg, storage: s}
}

func (h *Handler) HandlerJSONPOST(w http.ResponseWriter, r *http.Request) {
	var request RequestJSON

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if request.URL == "" {
		http.Error(w, `Invalid JSON format in request body. Expected: {"url": "<some_url>"}`, http.StatusBadRequest)
		return
	}
	log.Printf("Request JSON: %+v", request)

	log.Printf("Long URL: %v", request.URL)
	longURL, err := utils.ParseURL(request.URL)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	id, err := h.storage.SaveURL(longURL)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	shortURL := utils.MakeShortURL(h.config.BaseURL, id)
	log.Printf("Short URL: %v", shortURL)

	response := ResponseJSON{Result: shortURL}
	log.Printf("Response JSON: %+v", response)

	buf := bytes.NewBuffer([]byte{})
	encoder := json.NewEncoder(buf)
	encoder.SetEscapeHTML(false)
	err = encoder.Encode(response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	log.Printf("Encoded JSON: %s", buf.String())

	w.Header().Set(ContentType, ContentValueJSON)
	w.WriteHeader(http.StatusCreated)

	_, err = w.Write(buf.Bytes())
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
}

func (h *Handler) HandlerPOST(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	log.Printf("Long URL: %v", string(body))
	longURL, err := utils.ParseURL(string(body))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	id, err := h.storage.SaveURL(longURL)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	shortURL := utils.MakeShortURL(h.config.BaseURL, id)
	log.Printf("Short URL: %v", shortURL)

	w.Header().Set(ContentType, ContentValuePlainText)
	w.WriteHeader(http.StatusCreated)
	_, err = w.Write([]byte(shortURL))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
}

func (h *Handler) HandlerUserStorageGET(w http.ResponseWriter, r *http.Request) {
	//TODO: get all
}

func (h *Handler) HandlerGET(w http.ResponseWriter, r *http.Request) {
	strID := chi.URLParam(r, "ID")
	log.Printf("strID: `%s`", strID)
	id, err := strconv.Atoi(strID)
	if err != nil || id < 1 {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}
	log.Printf("ID: %d", id)
	originalURL, err := h.storage.GetURL(id)
	if err != nil {
		http.Error(w, "ID not found", http.StatusBadRequest)
		return
	}
	log.Printf("Original URL: %s", originalURL)
	w.Header().Set(ContentType, ContentValuePlainText)
	w.Header().Set("Location", originalURL)
	w.WriteHeader(http.StatusTemporaryRedirect)
}
