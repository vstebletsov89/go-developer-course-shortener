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
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
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
	defer resp.Body.Close()

	return resp, string(respBody)
}

func AuthHandleMock(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), middleware.UserCtx, "4b003ed0-4d8f-46eb-8322-e90174110517")
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func NewRouter(config *configs.Config) chi.Router {
	var storage repository.Repository
	if config.FileStoragePath != "" {
		storage = repository.NewFileRepository(config.FileStoragePath)
	} else {
		storage = repository.NewInMemoryRepository()
	}
	handler := NewHTTPHandler(config, storage)

	log.Printf("Server started on %v", config.ServerAddress)
	r := chi.NewRouter()
	r.Use(AuthHandleMock)

	// routing
	r.Post("/", handler.HandlerPOST)
	r.Post("/api/shorten", handler.HandlerJSONPOST)
	r.Get("/{ID}", handler.HandlerGET)
	r.Get("/ping", handler.HandlerPing)
	r.Get("/api/user/urls", handler.HandlerUserStorageGET)

	return r
}

func TestHandlerUserStorageGETNoUrls(t *testing.T) {
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
	defer resp.Body.Close()

	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
}

func TestHandlerPing(t *testing.T) {
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
	defer resp.Body.Close()
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
		resp.Body.Close()
		assert.Equal(t, http.StatusCreated, resp.StatusCode)
	}

	// get original urls
	resp, body := testRequest(t, ts, http.MethodGet, "/api/user/urls", nil)
	defer resp.Body.Close()

	var response []types.Link
	err := json.Unmarshal([]byte(body), &response)
	assert.NoError(t, err)

	for i, v := range response {
		originalURL := "https://github.com/test_repo" + strconv.Itoa(i)
		assert.Equal(t, originalURL, v.OriginalURL)
	}

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestBothHandlersFileStorageInvalidRecord(t *testing.T) {
	homeDir, err := os.UserHomeDir()
	assert.NoError(t, err)

	file, err := os.CreateTemp(homeDir, "test")
	assert.NoError(t, err)

	defer os.RemoveAll(file.Name())

	log.Printf("Temporary file name: %s", file.Name())

	config := &configs.Config{
		ServerAddress:   "localhost:8080",
		BaseURL:         "http://localhost:8080",
		FileStoragePath: file.Name(),
	}
	r := NewRouter(config)
	ts := httptest.NewServer(r)
	defer ts.Close()

	// prepare short url
	originalURL := "https://github.com/test_repo1"
	resp, _ := testRequest(t, ts, http.MethodPost, "/", bytes.NewBufferString(originalURL))
	defer resp.Body.Close()
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	// get original url
	resp, _ = testRequest(t, ts, http.MethodGet, "/AAAAA", nil)
	defer resp.Body.Close()
	assert.Equal(t, "", resp.Header.Get("Location"))
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestBothHandlersFileStorageOneRecord(t *testing.T) {
	homeDir, err := os.UserHomeDir()
	assert.NoError(t, err)

	file, err := os.CreateTemp(homeDir, "test")
	assert.NoError(t, err)
	defer os.RemoveAll(file.Name())

	log.Printf("Temporary file name: %s", file.Name())

	config := &configs.Config{
		ServerAddress:   "localhost:8080",
		BaseURL:         "http://localhost:8080",
		FileStoragePath: file.Name(),
	}
	r := NewRouter(config)
	ts := httptest.NewServer(r)
	defer ts.Close()

	// prepare short url
	originalURL := "https://github.com/test_repo1"
	resp, body := testRequest(t, ts, http.MethodPost, "/", bytes.NewBufferString(originalURL))
	defer resp.Body.Close()
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	shortURL, err := url.Parse(body)
	assert.NoError(t, err)

	// get original url
	resp, _ = testRequest(t, ts, http.MethodGet, shortURL.Path, nil)
	defer resp.Body.Close()
	assert.Equal(t, "https://github.com/test_repo1", resp.Header.Get("Location"))
	assert.Equal(t, http.StatusTemporaryRedirect, resp.StatusCode)
}

func TestBothHandlersFileStorageTwoRecords(t *testing.T) {
	homeDir, err := os.UserHomeDir()
	assert.NoError(t, err)

	file, err := os.CreateTemp(homeDir, "test")
	assert.NoError(t, err)
	defer os.RemoveAll(file.Name())

	log.Printf("Temporary file name: %s", file.Name())

	config := &configs.Config{
		ServerAddress:   "localhost:8080",
		BaseURL:         "http://localhost:8080",
		FileStoragePath: file.Name(),
	}
	r := NewRouter(config)
	ts := httptest.NewServer(r)
	defer ts.Close()

	// prepare short url
	originalURL := "https://github.com/test_repo1"
	resp, body := testRequest(t, ts, http.MethodPost, "/", bytes.NewBufferString(originalURL))
	defer resp.Body.Close()
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	shortURL1, err := url.Parse(body)
	assert.NoError(t, err)

	// prepare short url
	resp, body = testRequest(t, ts, http.MethodPost, "/", bytes.NewBufferString("https://github.com/test_repo2"))
	defer resp.Body.Close()
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	shortURL2, err := url.Parse(body)
	assert.NoError(t, err)

	// get original url
	resp, _ = testRequest(t, ts, http.MethodGet, shortURL2.Path, nil)
	defer resp.Body.Close()
	assert.Equal(t, "https://github.com/test_repo2", resp.Header.Get("Location"))
	assert.Equal(t, http.StatusTemporaryRedirect, resp.StatusCode)

	// get original url
	resp, _ = testRequest(t, ts, http.MethodGet, shortURL1.Path, nil)
	defer resp.Body.Close()
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
	defer resp.Body.Close()
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	shortURL, err := url.Parse(body)
	assert.NoError(t, err)

	// get original url
	resp, _ = testRequest(t, ts, http.MethodGet, shortURL.Path, nil)
	defer resp.Body.Close()
	assert.Equal(t, "https://github.com/test_repo1", resp.Header.Get("Location"))
	assert.Equal(t, http.StatusTemporaryRedirect, resp.StatusCode)
}

func TestBothHandlersWithJSON(t *testing.T) {
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
	defer resp.Body.Close()
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	var response types.ResponseJSON
	err := json.Unmarshal([]byte(body), &response)
	assert.NoError(t, err)
	shortURL, err := url.Parse(response.Result)
	assert.NoError(t, err)

	// get original url
	resp, _ = testRequest(t, ts, http.MethodGet, shortURL.Path, nil)
	defer resp.Body.Close()
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
			resp.Body.Close()
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
			resp.Body.Close()
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
			resp.Body.Close()
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
