// Package handlers is a collection of handlers for the shortener service.
package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"go-developer-course-shortener/internal/app/middleware"
	"go-developer-course-shortener/internal/app/rand"
	"go-developer-course-shortener/internal/app/repository"
	"go-developer-course-shortener/internal/app/types"
	"go-developer-course-shortener/internal/configs"
	"go-developer-course-shortener/internal/worker"
	"io"
	"log"
	"net/http"
	"net/url"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgerrcode"
)

const (
	ContentType           = "Content-Type"
	ContentValuePlainText = "text/plain; charset=utf-8"
	ContentValueJSON      = "application/json"
	shortLinkLength       = 5
)

// Handler contains settings and storage for current Repository.
type Handler struct {
	config  *configs.Config
	storage repository.Repository
}

// NewHTTPHandler returns a new Handler for the Repository.
func NewHTTPHandler(cfg *configs.Config, s repository.Repository) *Handler {
	return &Handler{config: cfg, storage: s}
}

func extractUserID(r *http.Request) string {
	userID, ok := r.Context().Value(middleware.UserCtx).(string)
	if ok {
		log.Printf("userID: %s", userID)
		return userID
	}
	return ""
}

// HandlerBatchPOST implements saving list of urls to the repository.
func (h *Handler) HandlerBatchPOST(w http.ResponseWriter, r *http.Request) {
	var request types.RequestBatch
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	log.Printf("Request batch JSON: %+v", request)

	userID := extractUserID(r)

	batchLinks := make(types.BatchLinks, len(request), len(request)) // allocate required capacity for the links
	for i, v := range request {
		id := string(rand.GenerateRandom(shortLinkLength))
		shortURL := makeShortURL(h.config.BaseURL, id)
		batchLinks[i] = types.BatchLink{CorrelationID: v.CorrelationID, ShortURL: shortURL, OriginalURL: v.OriginalURL}
	}

	var response types.ResponseBatch
	response, err := h.storage.SaveBatchURLS(userID, batchLinks)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	log.Printf("Response batch JSON: %+v", response)

	buf := bytes.NewBuffer([]byte{})
	encoder := json.NewEncoder(buf)
	encoder.SetEscapeHTML(false)
	err = encoder.Encode(response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	log.Printf("Encoded JSON: %s", buf.String())

	w.Header().Set(ContentType, ContentValueJSON)
	w.WriteHeader(http.StatusCreated)

	_, err = w.Write(buf.Bytes())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// HandlerJSONPOST implements getting short url by original url for json request.
func (h *Handler) HandlerJSONPOST(w http.ResponseWriter, r *http.Request) {
	var request types.RequestJSON

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
	longURL, err := parseURL(request.URL)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	userID := extractUserID(r)

	id := string(rand.GenerateRandom(shortLinkLength))
	shortURL := makeShortURL(h.config.BaseURL, id)
	log.Printf("Short URL: %v", shortURL)

	err = h.storage.SaveURL(userID, shortURL, longURL)
	status, err := checkDBViolation(err)
	if err != nil {
		http.Error(w, err.Error(), status)
		return
	}

	if status == http.StatusConflict {
		shortURL, err = h.storage.GetShortURLByOriginalURL(longURL)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	response := types.ResponseJSON{Result: shortURL}
	log.Printf("Response JSON: %+v", response)

	buf := bytes.NewBuffer([]byte{})
	encoder := json.NewEncoder(buf)
	encoder.SetEscapeHTML(false)
	err = encoder.Encode(response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	log.Printf("Encoded JSON: %s", buf.String())

	w.Header().Set(ContentType, ContentValueJSON)
	w.WriteHeader(status)

	_, err = w.Write(buf.Bytes())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func checkDBViolation(err error) (int, error) {
	var pgError *pgconn.PgError
	if errors.As(err, &pgError) && pgError.Code == pgerrcode.UniqueViolation {
		return http.StatusConflict, nil
	} else if err != nil {
		return http.StatusInternalServerError, err
	}
	return http.StatusCreated, nil
}

// HandlerPOST implements saving short and original url to the repository.
func (h *Handler) HandlerPOST(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	log.Printf("Long URL: %v", string(body))
	longURL, err := parseURL(string(body))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	userID := extractUserID(r)

	id := string(rand.GenerateRandom(shortLinkLength))
	shortURL := makeShortURL(h.config.BaseURL, id)
	log.Printf("Short URL: %v", shortURL)

	err = h.storage.SaveURL(userID, shortURL, longURL)
	status, err := checkDBViolation(err)
	if err != nil {
		http.Error(w, err.Error(), status)
		return
	}

	if status == http.StatusConflict {
		shortURL, err = h.storage.GetShortURLByOriginalURL(longURL)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set(ContentType, ContentValuePlainText)
	w.WriteHeader(status)
	_, err = w.Write([]byte(shortURL))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// HandlerUseStorageDELETE implements deleting short urls for current user id.
func (h *Handler) HandlerUseStorageDELETE(job chan worker.Job) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := extractUserID(r)
		log.Printf("Delete all links for userID: %s", userID)

		var deleteURLS []string
		if err := json.NewDecoder(r.Body).Decode(&deleteURLS); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		log.Printf("Request deleteURLS: %+v", deleteURLS)

		shortURLS := make([]string, len(deleteURLS), len(deleteURLS)) // allocate required capacity for the links
		for i, id := range deleteURLS {
			shortURLS[i] = makeShortURL(h.config.BaseURL, id)
		}
		log.Printf("Request shortURLS: %+v", shortURLS)

		j := worker.Job{UserID: userID, ShortURLS: shortURLS}
		job <- j

		w.Header().Set(ContentType, ContentValueJSON)
		w.WriteHeader(http.StatusAccepted)
		log.Printf("New worker created: %+v", j)
	}
}

// HandlerUserStorageGET implements getting list of urls for current user id.
func (h *Handler) HandlerUserStorageGET(w http.ResponseWriter, r *http.Request) {
	userID := extractUserID(r)
	log.Printf("Get all links for userID: %s", userID)
	links, err := h.storage.GetUserStorage(userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if len(links) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	body, err := json.Marshal(links)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set(ContentType, ContentValueJSON)
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// HandlerGET implements getting original url by short url.
func (h *Handler) HandlerGET(w http.ResponseWriter, r *http.Request) {
	strID := chi.URLParam(r, "ID")
	log.Printf("strID: `%s`", strID)

	originalLink, err := h.storage.GetURL(makeShortURL(h.config.BaseURL, strID))
	if err != nil {
		http.Error(w, "ID not found", http.StatusBadRequest)
		return
	}
	log.Printf("Original URL: %s deleted: %v", originalLink.OriginalURL, originalLink.Deleted)

	w.Header().Set(ContentType, ContentValuePlainText)
	w.Header().Set("Location", originalLink.OriginalURL)
	if !originalLink.Deleted {
		w.WriteHeader(http.StatusTemporaryRedirect)
	} else {
		w.WriteHeader(http.StatusGone)
	}
}

// HandlerPing verifies current status of repository.
func (h *Handler) HandlerPing(w http.ResponseWriter, r *http.Request) {
	if !h.storage.Ping() {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func makeShortURL(baseURL string, id string) string {
	shortURL := fmt.Sprintf("%v/%s", baseURL, id)
	return shortURL
}

func parseURL(strURL string) (string, error) {
	longURL, err := url.Parse(strURL)
	if err != nil {
		return "", err
	}
	if longURL.String() == "" {
		return "", errors.New("URL must not be empty")
	}
	return longURL.String(), nil
}
