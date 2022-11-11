package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"go-developer-course-shortener/internal/app/middleware"
	"go-developer-course-shortener/internal/app/repository"
	"go-developer-course-shortener/internal/app/types"
	"go-developer-course-shortener/internal/configs"
	"go-developer-course-shortener/internal/worker"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
)

func testRequest(t *testing.T, ts *httptest.Server, method, path string, body io.Reader) (*http.Response, string) {
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

func createFileRepository(t *testing.T) (string, string) {
	dir, err := os.Getwd()
	assert.NoError(t, err)

	temp, err := os.MkdirTemp(dir, "test")
	assert.NoError(t, err)

	// create file repository
	file := filepath.Join(temp, "file.db")
	err = os.WriteFile(file, []byte(""), 0666)
	assert.NoError(t, err)

	return temp, file
}

func AuthHandleMock(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), middleware.UserCtx, "4b003ed0-4d8f-46eb-8322-e90174110517")
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func NewRouter(config *configs.Config, auth ...bool) chi.Router {
	var storage repository.Repository
	if config.FileStoragePath != "" {
		storage = repository.NewFileRepository(config.FileStoragePath)
	} else {
		storage = repository.NewInMemoryRepository()
	}
	handler := NewHTTPHandler(config, storage)

	log.Printf("Server started on %v", config.ServerAddress)
	r := chi.NewRouter()
	if len(auth) > 0 {
		r.Use(middleware.GzipHandle, middleware.AuthHandle)
	} else {
		r.Use(AuthHandleMock)
	}

	// setup worker pool to handle delete requests
	jobs := make(chan worker.Job, worker.MaxWorkerPoolSize)
	workerPool := worker.NewWorkerPool(storage, jobs)
	go workerPool.Run(context.Background())

	// routing
	r.Post("/", handler.HandlerPOST)
	r.Post("/api/shorten", handler.HandlerJSONPOST)
	r.Post("/api/shorten/batch", handler.HandlerBatchPOST)
	r.Get("/{ID}", handler.HandlerGET)
	r.Get("/api/user/urls", handler.HandlerUserStorageGET)
	r.Get("/ping", handler.HandlerPing)
	r.Delete("/api/user/urls", handler.HandlerUseStorageDELETE(jobs))

	return r
}

func TestMiddlewareHandlersMemoryStorage(t *testing.T) {
	config := &configs.Config{
		ServerAddress:   "localhost:8080",
		BaseURL:         "http://localhost:8080",
		FileStoragePath: "",
	}
	r := NewRouter(config, true)
	ts := httptest.NewServer(r)
	defer ts.Close()

	// prepare short url
	originalURL := "https://github.com/test_repo1"
	resp, body := testRequest(t, ts, http.MethodPost, "/", bytes.NewBufferString(originalURL))
	err := resp.Body.Close()
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	shortURL, err := url.Parse(body)
	assert.NoError(t, err)

	// get original url
	resp, _ = testRequest(t, ts, http.MethodGet, shortURL.Path, nil)
	err = resp.Body.Close()
	assert.NoError(t, err)
	assert.Equal(t, "https://github.com/test_repo1", resp.Header.Get("Location"))
	assert.Equal(t, http.StatusTemporaryRedirect, resp.StatusCode)
}

func TestHandlerUserStorageGETNoUrlsMemoryStorage(t *testing.T) {
	config := &configs.Config{
		ServerAddress:   "localhost:8080",
		BaseURL:         "http://localhost:8080",
		FileStoragePath: "",
	}
	r := NewRouter(config)
	ts := httptest.NewServer(r)
	defer ts.Close()

	// get original url
	resp, _ := testRequest(t, ts, http.MethodGet, "/api/user/urls", nil)
	err := resp.Body.Close()
	assert.NoError(t, err)

	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
}

func TestHandlerUseStorageDELETENoUrlsMemoryStorage(t *testing.T) {
	config := &configs.Config{
		ServerAddress:   "localhost:8080",
		BaseURL:         "http://localhost:8080",
		FileStoragePath: "",
	}
	r := NewRouter(config)
	ts := httptest.NewServer(r)
	defer ts.Close()

	resp, _ := testRequest(t, ts, http.MethodDelete, "/api/user/urls", nil)
	err := resp.Body.Close()
	assert.NoError(t, err)

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestHandlerUseStorageDELETENoUrlsFileStorage(t *testing.T) {
	temp, storagePath := createFileRepository(t)
	defer func() {
		err := os.RemoveAll(temp)
		if err != nil {
			panic(err)
		}
	}()

	config := &configs.Config{
		ServerAddress:   "localhost:8080",
		BaseURL:         "http://localhost:8080",
		FileStoragePath: storagePath,
	}
	r := NewRouter(config)
	ts := httptest.NewServer(r)
	defer ts.Close()

	resp, _ := testRequest(t, ts, http.MethodDelete, "/api/user/urls", nil)
	err := resp.Body.Close()
	assert.NoError(t, err)

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestHandlerUseStorageDELETEMemoryStorage(t *testing.T) {
	config := &configs.Config{
		ServerAddress:   "localhost:8080",
		BaseURL:         "http://localhost:8080",
		FileStoragePath: "",
	}
	r := NewRouter(config)
	ts := httptest.NewServer(r)
	defer ts.Close()

	// prepare short url
	originalURL := "https://github.com/test_repo1"
	resp, body := testRequest(t, ts, http.MethodPost, "/", bytes.NewBufferString(originalURL))
	err := resp.Body.Close()
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	shortURL, err := url.Parse(body)
	assert.NoError(t, err)

	var deleteURLS []string
	deleteURLS = append(deleteURLS, shortURL.Path)

	j, _ := json.Marshal(deleteURLS)

	// delete user urls
	resp, _ = testRequest(t, ts, http.MethodDelete, "/api/user/urls", bytes.NewBufferString(string(j)))
	err = resp.Body.Close()
	assert.NoError(t, err)

	assert.Equal(t, http.StatusAccepted, resp.StatusCode)
}

func TestHandlerUseStorageDELETEFileStorage(t *testing.T) {
	temp, storagePath := createFileRepository(t)
	defer func() {
		err := os.RemoveAll(temp)
		if err != nil {
			panic(err)
		}
	}()

	config := &configs.Config{
		ServerAddress:   "localhost:8080",
		BaseURL:         "http://localhost:8080",
		FileStoragePath: storagePath,
	}
	r := NewRouter(config)
	ts := httptest.NewServer(r)
	defer ts.Close()

	// prepare short url
	originalURL := "https://github.com/test_repo1"
	resp, body := testRequest(t, ts, http.MethodPost, "/", bytes.NewBufferString(originalURL))
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	shortURL, err := url.Parse(body)
	assert.NoError(t, err)
	err = resp.Body.Close()
	assert.NoError(t, err)

	var deleteURLS []string
	deleteURLS = append(deleteURLS, shortURL.Path)

	j, _ := json.Marshal(deleteURLS)

	// delete user urls
	resp, _ = testRequest(t, ts, http.MethodDelete, "/api/user/urls", bytes.NewBufferString(string(j)))
	err = resp.Body.Close()
	assert.NoError(t, err)

	assert.Equal(t, http.StatusAccepted, resp.StatusCode)
}

func TestHandlerPingMemoryStorage(t *testing.T) {
	config := &configs.Config{
		ServerAddress:   "localhost:8080",
		BaseURL:         "http://localhost:8080",
		FileStoragePath: "",
	}
	r := NewRouter(config)
	ts := httptest.NewServer(r)
	defer ts.Close()

	// try to ping connection
	resp, _ := testRequest(t, ts, http.MethodGet, "/ping", nil)
	err := resp.Body.Close()
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

}

func TestHandlerPingFileStorage(t *testing.T) {
	temp, storagePath := createFileRepository(t)
	defer func() {
		err := os.RemoveAll(temp)
		if err != nil {
			panic(err)
		}
	}()

	config := &configs.Config{
		ServerAddress:   "localhost:8080",
		BaseURL:         "http://localhost:8080",
		FileStoragePath: storagePath,
	}
	r := NewRouter(config)
	ts := httptest.NewServer(r)
	defer ts.Close()

	// try to ping connection
	resp, _ := testRequest(t, ts, http.MethodGet, "/ping", nil)
	err := resp.Body.Close()
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetUserLinksMemoryStorage(t *testing.T) {
	config := &configs.Config{
		ServerAddress:   "localhost:8080",
		BaseURL:         "http://localhost:8080",
		FileStoragePath: "",
	}
	r := NewRouter(config)
	ts := httptest.NewServer(r)
	defer ts.Close()

	counter := 10
	for i := 0; i < counter; i++ {
		// prepare short url
		originalURL := "https://github.com/test_repo" + strconv.Itoa(i)
		resp, _ := testRequest(t, ts, http.MethodPost, "/", bytes.NewBufferString(originalURL))
		err := resp.Body.Close()
		assert.NoError(t, err)
		assert.Equal(t, http.StatusCreated, resp.StatusCode)
	}

	// get original urls
	resp, body := testRequest(t, ts, http.MethodGet, "/api/user/urls", nil)
	err := resp.Body.Close()
	assert.NoError(t, err)

	var response []types.Link
	err = json.Unmarshal([]byte(body), &response)
	assert.NoError(t, err)

	for i, v := range response {
		originalURL := "https://github.com/test_repo" + strconv.Itoa(i)
		assert.Equal(t, originalURL, v.OriginalURL)
	}

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetUserLinksFileStorage(t *testing.T) {
	temp, storagePath := createFileRepository(t)
	defer func() {
		err := os.RemoveAll(temp)
		if err != nil {
			panic(err)
		}
	}()

	config := &configs.Config{
		ServerAddress:   "localhost:8080",
		BaseURL:         "http://localhost:8080",
		FileStoragePath: storagePath,
	}
	r := NewRouter(config)
	ts := httptest.NewServer(r)
	defer ts.Close()

	counter := 10
	for i := 0; i < counter; i++ {
		// prepare short url
		originalURL := "https://github.com/test_repo" + strconv.Itoa(i)
		resp, _ := testRequest(t, ts, http.MethodPost, "/", bytes.NewBufferString(originalURL))
		err := resp.Body.Close()
		assert.NoError(t, err)
		assert.Equal(t, http.StatusCreated, resp.StatusCode)
	}

	// get original urls
	resp, body := testRequest(t, ts, http.MethodGet, "/api/user/urls", nil)
	err := resp.Body.Close()
	assert.NoError(t, err)

	var response []types.Link
	err = json.Unmarshal([]byte(body), &response)
	assert.NoError(t, err)

	for i, v := range response {
		originalURL := "https://github.com/test_repo" + strconv.Itoa(i)
		assert.Equal(t, originalURL, v.OriginalURL)
	}

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestBothHandlersFileStorageInvalidRecord(t *testing.T) {
	temp, storagePath := createFileRepository(t)
	defer func() {
		err := os.RemoveAll(temp)
		if err != nil {
			panic(err)
		}
	}()

	config := &configs.Config{
		ServerAddress:   "localhost:8080",
		BaseURL:         "http://localhost:8080",
		FileStoragePath: storagePath,
	}
	r := NewRouter(config)
	ts := httptest.NewServer(r)
	defer ts.Close()

	// prepare short url
	originalURL := "https://github.com/test_repo1"
	resp, _ := testRequest(t, ts, http.MethodPost, "/", bytes.NewBufferString(originalURL))
	err := resp.Body.Close()
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	// get original url
	resp, _ = testRequest(t, ts, http.MethodGet, "/AAAAA", nil)
	err = resp.Body.Close()
	assert.NoError(t, err)
	assert.Equal(t, "", resp.Header.Get("Location"))
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestBothHandlersFileStorageOneRecord(t *testing.T) {
	temp, storagePath := createFileRepository(t)
	defer func() {
		err := os.RemoveAll(temp)
		if err != nil {
			panic(err)
		}
	}()

	config := &configs.Config{
		ServerAddress:   "localhost:8080",
		BaseURL:         "http://localhost:8080",
		FileStoragePath: storagePath,
	}
	r := NewRouter(config)
	ts := httptest.NewServer(r)
	defer ts.Close()

	// prepare short url
	originalURL := "https://github.com/test_repo1"
	resp, body := testRequest(t, ts, http.MethodPost, "/", bytes.NewBufferString(originalURL))
	err := resp.Body.Close()
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	shortURL, err := url.Parse(body)
	assert.NoError(t, err)

	// get original url
	resp, _ = testRequest(t, ts, http.MethodGet, shortURL.Path, nil)
	err = resp.Body.Close()
	assert.NoError(t, err)
	assert.Equal(t, "https://github.com/test_repo1", resp.Header.Get("Location"))
	assert.Equal(t, http.StatusTemporaryRedirect, resp.StatusCode)
}

func TestBothHandlersFileStorageTwoRecords(t *testing.T) {
	temp, storagePath := createFileRepository(t)
	defer func() {
		err := os.RemoveAll(temp)
		if err != nil {
			panic(err)
		}
	}()

	config := &configs.Config{
		ServerAddress:   "localhost:8080",
		BaseURL:         "http://localhost:8080",
		FileStoragePath: storagePath,
	}
	r := NewRouter(config)
	ts := httptest.NewServer(r)
	defer ts.Close()

	// prepare short url
	originalURL := "https://github.com/test_repo1"
	resp, body := testRequest(t, ts, http.MethodPost, "/", bytes.NewBufferString(originalURL))
	err := resp.Body.Close()
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	shortURL1, err := url.Parse(body)
	assert.NoError(t, err)

	// prepare short url
	resp, body = testRequest(t, ts, http.MethodPost, "/", bytes.NewBufferString("https://github.com/test_repo2"))
	err = resp.Body.Close()
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	shortURL2, err := url.Parse(body)
	assert.NoError(t, err)

	// get original url
	resp, _ = testRequest(t, ts, http.MethodGet, shortURL2.Path, nil)
	err = resp.Body.Close()
	assert.NoError(t, err)
	assert.Equal(t, "https://github.com/test_repo2", resp.Header.Get("Location"))
	assert.Equal(t, http.StatusTemporaryRedirect, resp.StatusCode)

	// get original url
	resp, _ = testRequest(t, ts, http.MethodGet, shortURL1.Path, nil)
	err = resp.Body.Close()
	assert.NoError(t, err)
	assert.Equal(t, "https://github.com/test_repo1", resp.Header.Get("Location"))
	assert.Equal(t, http.StatusTemporaryRedirect, resp.StatusCode)
}

func TestBothHandlersMemoryStorage(t *testing.T) {
	config := &configs.Config{
		ServerAddress:   "localhost:8080",
		BaseURL:         "http://localhost:8080",
		FileStoragePath: "",
	}
	r := NewRouter(config)
	ts := httptest.NewServer(r)
	defer ts.Close()

	// prepare short url
	originalURL := "https://github.com/test_repo1"
	resp, body := testRequest(t, ts, http.MethodPost, "/", bytes.NewBufferString(originalURL))
	err := resp.Body.Close()
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	shortURL, err := url.Parse(body)
	assert.NoError(t, err)

	// get original url
	resp, _ = testRequest(t, ts, http.MethodGet, shortURL.Path, nil)
	err = resp.Body.Close()
	assert.NoError(t, err)
	assert.Equal(t, "https://github.com/test_repo1", resp.Header.Get("Location"))
	assert.Equal(t, http.StatusTemporaryRedirect, resp.StatusCode)
}

func TestBothHandlersFileStorage(t *testing.T) {
	temp, storagePath := createFileRepository(t)
	defer func() {
		err := os.RemoveAll(temp)
		if err != nil {
			panic(err)
		}
	}()

	config := &configs.Config{
		ServerAddress:   "localhost:8080",
		BaseURL:         "http://localhost:8080",
		FileStoragePath: storagePath,
	}
	r := NewRouter(config)
	ts := httptest.NewServer(r)
	defer ts.Close()

	// prepare short url
	originalURL := "https://github.com/test_repo1"
	resp, body := testRequest(t, ts, http.MethodPost, "/", bytes.NewBufferString(originalURL))
	err := resp.Body.Close()
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	shortURL, err := url.Parse(body)
	assert.NoError(t, err)

	// get original url
	resp, _ = testRequest(t, ts, http.MethodGet, shortURL.Path, nil)
	err = resp.Body.Close()
	assert.NoError(t, err)
	assert.Equal(t, "https://github.com/test_repo1", resp.Header.Get("Location"))
	assert.Equal(t, http.StatusTemporaryRedirect, resp.StatusCode)
}

func TestBothHandlersWithJSONMemoryStorage(t *testing.T) {
	config := &configs.Config{
		ServerAddress:   "localhost:8080",
		BaseURL:         "http://localhost:8080",
		FileStoragePath: "",
	}
	r := NewRouter(config)
	ts := httptest.NewServer(r)
	defer ts.Close()

	// prepare short url
	originalURL := "https://github.com/test_repo1"
	resp, body := testRequest(t, ts, http.MethodPost, "/api/shorten", bytes.NewBufferString(`{"url": "https://github.com/test_repo1"}`))
	err := resp.Body.Close()
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	var response types.ResponseJSON
	err = json.Unmarshal([]byte(body), &response)
	assert.NoError(t, err)
	shortURL, err := url.Parse(response.Result)
	assert.NoError(t, err)

	// get original url
	resp, _ = testRequest(t, ts, http.MethodGet, shortURL.Path, nil)
	err = resp.Body.Close()
	assert.NoError(t, err)
	assert.Equal(t, originalURL, resp.Header.Get("Location"))
	assert.Equal(t, http.StatusTemporaryRedirect, resp.StatusCode)
}

func TestBothHandlersWithJSONFileStorage(t *testing.T) {
	temp, storagePath := createFileRepository(t)
	defer func() {
		err := os.RemoveAll(temp)
		if err != nil {
			panic(err)
		}
	}()

	config := &configs.Config{
		ServerAddress:   "localhost:8080",
		BaseURL:         "http://localhost:8080",
		FileStoragePath: storagePath,
	}
	r := NewRouter(config)
	ts := httptest.NewServer(r)
	defer ts.Close()

	// prepare short url
	originalURL := "https://github.com/test_repo1"
	resp, body := testRequest(t, ts, http.MethodPost, "/api/shorten", bytes.NewBufferString(`{"url": "https://github.com/test_repo1"}`))
	err := resp.Body.Close()
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	var response types.ResponseJSON
	err = json.Unmarshal([]byte(body), &response)
	assert.NoError(t, err)
	shortURL, err := url.Parse(response.Result)
	assert.NoError(t, err)

	// get original url
	resp, _ = testRequest(t, ts, http.MethodGet, shortURL.Path, nil)
	err = resp.Body.Close()
	assert.NoError(t, err)
	assert.Equal(t, originalURL, resp.Header.Get("Location"))
	assert.Equal(t, http.StatusTemporaryRedirect, resp.StatusCode)
}

func TestHandlerGetErrors(t *testing.T) {
	type want struct {
		headerLocation string
		statusCode     int
		responseBody   string
	}
	tests := []struct {
		name string
		id   int
		want want
	}{
		{
			name: "Test #1",
			id:   999,
			want: want{
				headerLocation: "",
				statusCode:     http.StatusBadRequest,
				responseBody:   "ID not found\n",
			},
		},
	}

	config := &configs.Config{
		ServerAddress:   "localhost:8080",
		BaseURL:         "http://localhost:8080",
		FileStoragePath: "",
	}
	r := NewRouter(config)
	ts := httptest.NewServer(r)
	defer ts.Close()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, body := testRequest(t, ts, http.MethodGet, fmt.Sprintf("/%d", tt.id), nil)
			err := resp.Body.Close()
			assert.NoError(t, err)
			assert.Equal(t, tt.want.statusCode, resp.StatusCode)
			assert.Equal(t, tt.want.responseBody, body)
			assert.Equal(t, tt.want.headerLocation, resp.Header.Get("Location"))
		})
	}
}

func TestHandlerPost(t *testing.T) {
	type want struct {
		contentType  string
		statusCode   int
		responseBody string
	}
	tests := []struct {
		name    string
		longURL string
		want    want
	}{
		{
			name:    "Test #1",
			longURL: "",
			want: want{
				contentType:  ContentValuePlainText,
				statusCode:   http.StatusBadRequest,
				responseBody: "URL must not be empty\n",
			},
		},
		{
			name:    "Test #2",
			longURL: "htt p://incorrect_url_here",
			want: want{
				contentType:  ContentValuePlainText,
				statusCode:   http.StatusBadRequest,
				responseBody: "parse \"htt p://incorrect_url_here\": first path segment in URL cannot contain colon\n",
			},
		},
	}
	config := &configs.Config{
		ServerAddress:   "localhost:8080",
		BaseURL:         "http://localhost:8080",
		FileStoragePath: "",
	}
	r := NewRouter(config)
	ts := httptest.NewServer(r)
	defer ts.Close()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, body := testRequest(t, ts, http.MethodPost, "/", bytes.NewBufferString(tt.longURL))
			err := resp.Body.Close()
			assert.NoError(t, err)
			assert.Equal(t, tt.want.statusCode, resp.StatusCode)
			assert.Equal(t, tt.want.responseBody, body)
			assert.Equal(t, tt.want.contentType, resp.Header.Get(ContentType))
		})
	}
}

func TestHandlerJsonPost(t *testing.T) {
	type want struct {
		contentType  string
		statusCode   int
		responseBody string
		checkJSON    bool
	}
	tests := []struct {
		name     string
		jsonBody string
		want     want
	}{
		{
			name:     "test invalid format for JSON request",
			jsonBody: `{"invalid": "<some_url>"}`,
			want: want{
				contentType:  ContentValuePlainText,
				statusCode:   http.StatusBadRequest,
				responseBody: "Invalid JSON format in request body. Expected: {\"url\": \"<some_url>\"}\n",
				checkJSON:    false,
			},
		},
		{
			name:     "test invalid input for decoder",
			jsonBody: `{"url": "invalid\github.com/test_repo"}`,
			want: want{
				contentType:  ContentValuePlainText,
				statusCode:   http.StatusBadRequest,
				responseBody: "invalid character 'g' in string escape code\n",
				checkJSON:    false,
			},
		},
		{
			name:     "test empty input URL",
			jsonBody: `{"url": ""}`,
			want: want{
				contentType:  ContentValuePlainText,
				statusCode:   http.StatusBadRequest,
				responseBody: "Invalid JSON format in request body. Expected: {\"url\": \"<some_url>\"}\n",
				checkJSON:    false,
			},
		},
	}
	config := &configs.Config{
		ServerAddress:   "localhost:8080",
		BaseURL:         "http://localhost:8080",
		FileStoragePath: "",
	}
	r := NewRouter(config)
	ts := httptest.NewServer(r)
	defer ts.Close()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, body := testRequest(t, ts, http.MethodPost, "/api/shorten", bytes.NewBufferString(tt.jsonBody))
			err := resp.Body.Close()
			assert.NoError(t, err)
			assert.Equal(t, tt.want.statusCode, resp.StatusCode)

			if tt.want.checkJSON {
				assert.JSONEq(t, tt.want.responseBody, body)
			} else {
				assert.Equal(t, tt.want.responseBody, body)
			}

			assert.Equal(t, tt.want.contentType, resp.Header.Get(ContentType))
		})
	}
}
