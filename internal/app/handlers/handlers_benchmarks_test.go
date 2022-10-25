package handlers

import (
	"bytes"
	"context"
	"go-developer-course-shortener/internal/app/middleware"
	"go-developer-course-shortener/internal/app/repository"
	"go-developer-course-shortener/internal/configs"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"

	"github.com/go-chi/chi/v5"
)

func testBenchmarkRequest(b *testing.B, ts *httptest.Server, method, path string, body io.Reader) (*http.Response, string) {
	req, _ := http.NewRequest(method, ts.URL+path, body)

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	b.StartTimer() // start timer
	resp, _ := client.Do(req)
	b.StopTimer() // stop all timers

	return resp, ""
}

func AuthHandleMockBenchmark(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), middleware.UserCtx, "4b003ed0-4d8f-46eb-8322-e90174110517")
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func NewRouterBenchmark(config *configs.Config) chi.Router {
	var storage repository.Repository
	if config.FileStoragePath != "" {
		storage = repository.NewFileRepository(config.FileStoragePath)
	} else {
		storage = repository.NewInMemoryRepository()
	}
	handler := NewHTTPHandler(config, storage)

	log.Printf("Server started on %v", config.ServerAddress)
	r := chi.NewRouter()
	r.Use(AuthHandleMockBenchmark)

	// routing
	r.Post("/", handler.HandlerPOST)
	r.Post("/api/shorten", handler.HandlerJSONPOST)
	r.Get("/{ID}", handler.HandlerGET)
	r.Get("/api/user/urls", handler.HandlerUserStorageGET)

	return r
}

func BenchmarkSaveGetMemoryStorage(b *testing.B) {
	config := &configs.Config{
		ServerAddress:   "localhost:8080",
		BaseURL:         "http://localhost:8080",
		FileStoragePath: "",
	}
	r := NewRouterBenchmark(config)
	ts := httptest.NewServer(r)
	defer ts.Close()

	b.ResetTimer() // reset timer

	counter := 1 // counter for unique urls
	for i := 0; i < b.N; i++ {
		b.StopTimer() // stop timer
		originalURL := "https://github.com/test_repo" + strconv.Itoa(counter)

		// prepare short url
		resp, body := testBenchmarkRequest(b, ts, http.MethodPost, "/", bytes.NewBufferString(originalURL))

		shortURL, _ := url.Parse(body)
		resp.Body.Close()

		// get original url
		resp, _ = testBenchmarkRequest(b, ts, http.MethodGet, shortURL.Path, nil)

		resp.Body.Close()
		counter++
	}
}

func BenchmarkSaveGetAllMemoryStorage(b *testing.B) {
	config := &configs.Config{
		ServerAddress:   "localhost:8080",
		BaseURL:         "http://localhost:8080",
		FileStoragePath: "",
	}
	r := NewRouterBenchmark(config)
	ts := httptest.NewServer(r)
	defer ts.Close()

	b.ResetTimer() // reset timer

	counter := 1 // counter for unique urls
	for i := 0; i < b.N; i++ {
		b.StopTimer() // stop timer
		originalURL := "https://github.com/test_repo" + strconv.Itoa(counter)

		// prepare short url
		resp, _ := testBenchmarkRequest(b, ts, http.MethodPost, "/", bytes.NewBufferString(originalURL))

		resp.Body.Close()

		// get original url
		resp, _ = testBenchmarkRequest(b, ts, http.MethodGet, "/api/user/urls", nil)

		resp.Body.Close()
		counter++
	}
}
