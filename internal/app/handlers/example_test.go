package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"go-developer-course-shortener/internal/app/repository"
	"go-developer-course-shortener/internal/app/service"
	"go-developer-course-shortener/internal/app/types"
	"go-developer-course-shortener/internal/configs"
	"go-developer-course-shortener/internal/worker"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/go-chi/chi/v5"
)

func exampleRequest(t *testing.T, ts *httptest.Server, method, path string, body io.Reader) (*http.Response, string) {
	req, err := http.NewRequest(method, ts.URL+path, body)
	assert.NoError(t, err)

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	resp, err := client.Do(req)
	assert.NoError(t, err)

	respBody, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)
	err = resp.Body.Close()
	assert.NoError(t, err)

	return resp, string(respBody)
}

func NewExampleRouter(config *configs.Config) chi.Router {
	var storage repository.Repository
	if config.FileStoragePath != "" {
		storage = repository.NewFileRepository(config.FileStoragePath)
	} else {
		storage = repository.NewInMemoryRepository()
	}
	// setup worker pool to handle delete requests
	jobs := make(chan worker.Job, worker.MaxWorkerPoolSize)
	workerPool := worker.NewWorkerPool(storage, jobs)
	go workerPool.Run(context.Background())

	svc := service.NewService(storage, jobs, nil, config.BaseURL)
	handler := NewHTTPHandler(svc)

	log.Printf("Server started on %v", config.ServerAddress)
	r := chi.NewRouter()
	r.Use(AuthHandleMock)

	// routing
	r.Post("/", handler.HandlerPOST)
	r.Post("/api/shorten", handler.HandlerJSONPOST)
	r.Post("/api/shorten/batch", handler.HandlerBatchPOST)
	r.Get("/{ID}", handler.HandlerGET)
	r.Get("/api/user/urls", handler.HandlerUserStorageGET)
	r.Get("/ping", handler.HandlerPing)

	return r
}

func ExampleHandler_HandlerPing() {
	config := &configs.Config{
		ServerAddress:   "localhost:8080",
		BaseURL:         "http://localhost:8080",
		FileStoragePath: "",
	}
	r := NewExampleRouter(config)
	ts := httptest.NewServer(r)
	defer ts.Close()

	// try to ping connection
	resp, _ := exampleRequest(nil, ts, http.MethodGet, "/ping", nil)
	err := resp.Body.Close()
	if err != nil {
		return
	}

	fmt.Println(resp.StatusCode)

	// Output:
	// 200
}

func ExampleHandler_HandlerUserStorageGET() {
	config := &configs.Config{
		ServerAddress:   "localhost:8080",
		BaseURL:         "http://localhost:8080",
		FileStoragePath: "",
	}
	r := NewExampleRouter(config)
	ts := httptest.NewServer(r)
	defer ts.Close()

	counter := 10
	for i := 0; i < counter; i++ {
		// prepare short url
		originalURL := "https://github.com/test_repo" + strconv.Itoa(i)
		resp, _ := exampleRequest(nil, ts, http.MethodPost, "/", bytes.NewBufferString(originalURL))
		err := resp.Body.Close()
		if err != nil {
			return
		}
	}

	// get original urls
	resp, body := exampleRequest(nil, ts, http.MethodGet, "/api/user/urls", nil)
	err := resp.Body.Close()
	if err != nil {
		return
	}

	var response []types.Link
	err = json.Unmarshal([]byte(body), &response)
	if err != nil {
		return
	}

	fmt.Println(resp.StatusCode)

	// Output:
	// 200
}

func ExampleHandler_HandlerPOST() {
	config := &configs.Config{
		ServerAddress:   "localhost:8080",
		BaseURL:         "http://localhost:8080",
		FileStoragePath: "",
	}
	r := NewExampleRouter(config)
	ts := httptest.NewServer(r)
	defer ts.Close()

	// prepare short url
	originalURL := "https://github.com/test_repo1"
	resp, _ := exampleRequest(nil, ts, http.MethodPost, "/", bytes.NewBufferString(originalURL))
	err := resp.Body.Close()
	if err != nil {
		return
	}

	fmt.Println(resp.StatusCode)

	// Output:
	// 201
}

func ExampleHandler_HandlerGET() {
	config := &configs.Config{
		ServerAddress:   "localhost:8080",
		BaseURL:         "http://localhost:8080",
		FileStoragePath: "",
	}
	r := NewExampleRouter(config)
	ts := httptest.NewServer(r)
	defer ts.Close()

	// prepare short url
	originalURL := "https://github.com/test_repo1"
	resp, body := exampleRequest(nil, ts, http.MethodPost, "/", bytes.NewBufferString(originalURL))
	err := resp.Body.Close()
	if err != nil {
		return
	}
	shortURL, _ := url.Parse(body)

	// get original url
	resp, _ = exampleRequest(nil, ts, http.MethodGet, shortURL.Path, nil)
	err = resp.Body.Close()
	if err != nil {
		return
	}

	fmt.Println(resp.Header.Get("Location"))

	// Output:
	// https://github.com/test_repo1
}

func ExampleHandler_HandlerJSONPOST() {
	config := &configs.Config{
		ServerAddress:   "localhost:8080",
		BaseURL:         "http://localhost:8080",
		FileStoragePath: "",
	}
	r := NewExampleRouter(config)
	ts := httptest.NewServer(r)
	defer ts.Close()

	// prepare short url from json
	resp, body := exampleRequest(nil, ts, http.MethodPost, "/api/shorten", bytes.NewBufferString(`{"url": "https://github.com/test_repo1"}`))
	err := resp.Body.Close()
	if err != nil {
		return
	}

	var response types.ResponseJSON
	err = json.Unmarshal([]byte(body), &response)
	if err != nil {
		return
	}

	fmt.Println(resp.StatusCode)

	// Output:
	// 201
}

func ExampleHandler_HandlerBatchPOST() {
	config := &configs.Config{
		ServerAddress:   "localhost:8080",
		BaseURL:         "http://localhost:8080",
		FileStoragePath: "",
	}
	r := NewExampleRouter(config)
	ts := httptest.NewServer(r)
	defer ts.Close()

	// prepare short url from json
	resp, body := exampleRequest(nil, ts, http.MethodPost, "/api/shorten/batch",
		bytes.NewBufferString(`[{"url": "https://github.com/test_repo1"},{"url": "https://github.com/test_repo2"},{"url": "https://github.com/test_repo3"}]`))
	err := resp.Body.Close()
	if err != nil {
		return
	}

	var response types.ResponseBatch
	err = json.Unmarshal([]byte(body), &response)
	if err != nil {
		return
	}

	fmt.Println(resp.StatusCode)

	// Output:
	// 201
}
