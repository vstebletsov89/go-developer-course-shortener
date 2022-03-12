package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgerrcode"
	"go-developer-course-shortener/internal/app/middleware"
	"go-developer-course-shortener/internal/app/repository"
	"go-developer-course-shortener/internal/app/types"
	"go-developer-course-shortener/internal/app/utils"
	"go-developer-course-shortener/internal/configs"
	"io"
	"log"
	"net/http"
)

type Handler struct {
	config  *configs.Config
	storage repository.Repository
}

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

func (h *Handler) HandlerBatchPOST(w http.ResponseWriter, r *http.Request) {
	var request types.RequestBatch
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	log.Printf("Request batch JSON: %+v", request)

	userID := extractUserID(r)

	var batchLinks types.BatchLinks
	for _, v := range request {
		id := string(utils.GenerateRandom(utils.ShortLinkLength))
		shortURL := utils.MakeShortURL(h.config.BaseURL, id)
		batchLinks = append(batchLinks, types.BatchLink{CorrelationID: v.CorrelationID, ShortURL: shortURL, OriginalURL: v.OriginalURL})
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
	longURL, err := utils.ParseURL(request.URL)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	userID := extractUserID(r)

	id := string(utils.GenerateRandom(utils.ShortLinkLength))
	shortURL := utils.MakeShortURL(h.config.BaseURL, id)
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

	userID := extractUserID(r)

	id := string(utils.GenerateRandom(utils.ShortLinkLength))
	shortURL := utils.MakeShortURL(h.config.BaseURL, id)
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

func (h *Handler) HandlerUserStorageGET(w http.ResponseWriter, r *http.Request) {
	userID := extractUserID(r)
	log.Printf("Get all links for userID: %s", userID)
	links, err := h.storage.GetUserStorage(userID)
	if err != nil {
		w.WriteHeader(http.StatusNoContent)
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

func (h *Handler) HandlerGET(w http.ResponseWriter, r *http.Request) {
	strID := chi.URLParam(r, "ID")
	log.Printf("strID: `%s`", strID)

	originalURL, err := h.storage.GetURL(utils.MakeShortURL(h.config.BaseURL, strID))
	if err != nil {
		http.Error(w, "ID not found", http.StatusBadRequest)
		return
	}
	log.Printf("Original URL: %s", originalURL)
	w.Header().Set(ContentType, ContentValuePlainText)
	w.Header().Set("Location", originalURL)
	w.WriteHeader(http.StatusTemporaryRedirect)
}

func (h *Handler) HandlerPing(w http.ResponseWriter, r *http.Request) {
	if !h.storage.Ping() {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
