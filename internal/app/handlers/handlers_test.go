package handlers

import (
	"bytes"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"go-developer-course-shortener/internal/app/repository"
	"go-developer-course-shortener/internal/configs"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
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

	respBody, err := ioutil.ReadAll(resp.Body)
	assert.NoError(t, err)
	defer resp.Body.Close()

	return resp, string(respBody)
}

func NewRouter(config *configs.Config) chi.Router {
	var storage repository.Repository
	if config.FileStoragePath != configs.FileStorageDefault {
		storage = repository.NewFileRepository(config.FileStoragePath)
	} else {
		storage = repository.NewInMemoryRepository()
	}
	handler := NewHTTPHandler(config, storage)

	log.Printf("Server started on %v", config.ServerAddress)
	r := chi.NewRouter()

	// маршрутизация запросов обработчику
	r.Post("/", handler.HandlerPOST)
	r.Post("/api/shorten", handler.HandlerJSONPOST)
	r.Get("/{ID}", handler.HandlerGET)

	return r
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

	//сначала подготавливаем сокращенную ссылку через POST
	resp, body := testRequest(t, ts, http.MethodPost, "/", bytes.NewBufferString("https://github.com/test_repo1"))
	defer resp.Body.Close()
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	assert.Equal(t, body, "http://localhost:8080/1")

	//получаем оригинальную ссылку через GET запрос
	resp, _ = testRequest(t, ts, http.MethodGet, fmt.Sprintf("/%d", 999), nil)
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

	//сначала подготавливаем сокращенную ссылку через POST
	resp, body := testRequest(t, ts, http.MethodPost, "/", bytes.NewBufferString("https://github.com/test_repo1"))
	defer resp.Body.Close()
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	assert.Equal(t, body, "http://localhost:8080/1")

	//получаем оригинальную ссылку через GET запрос
	resp, _ = testRequest(t, ts, http.MethodGet, fmt.Sprintf("/%d", 1), nil)
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

	//сначала подготавливаем сокращенную ссылку через POST
	resp, body := testRequest(t, ts, http.MethodPost, "/", bytes.NewBufferString("https://github.com/test_repo1"))
	defer resp.Body.Close()
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	assert.Equal(t, body, "http://localhost:8080/1")

	//сначала подготавливаем сокращенную ссылку через POST
	resp, body = testRequest(t, ts, http.MethodPost, "/", bytes.NewBufferString("https://github.com/test_repo2"))
	defer resp.Body.Close()
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	assert.Equal(t, body, "http://localhost:8080/2")

	//получаем оригинальную ссылку через GET запрос
	resp, _ = testRequest(t, ts, http.MethodGet, fmt.Sprintf("/%d", 2), nil)
	defer resp.Body.Close()
	assert.Equal(t, "https://github.com/test_repo2", resp.Header.Get("Location"))
	assert.Equal(t, http.StatusTemporaryRedirect, resp.StatusCode)

	//получаем оригинальную ссылку через GET запрос
	resp, _ = testRequest(t, ts, http.MethodGet, fmt.Sprintf("/%d", 1), nil)
	defer resp.Body.Close()
	assert.Equal(t, "https://github.com/test_repo1", resp.Header.Get("Location"))
	assert.Equal(t, http.StatusTemporaryRedirect, resp.StatusCode)
}

func TestBothHandlersMemoryStorage(t *testing.T) {
	config := &configs.Config{
		ServerAddress:   "localhost:8080",
		BaseURL:         "http://localhost:8080",
		FileStoragePath: configs.FileStorageDefault,
	}
	r := NewRouter(config)
	ts := httptest.NewServer(r)
	defer ts.Close()

	//сначала подготавливаем сокращенную ссылку через POST
	resp, body := testRequest(t, ts, http.MethodPost, "/", bytes.NewBufferString("https://github.com/test_repo1"))
	defer resp.Body.Close()
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	assert.Equal(t, body, "http://localhost:8080/1")

	//получаем оригинальную ссылку через GET запрос
	resp, _ = testRequest(t, ts, http.MethodGet, fmt.Sprintf("/%d", 1), nil)
	defer resp.Body.Close()
	assert.Equal(t, "https://github.com/test_repo1", resp.Header.Get("Location"))
	assert.Equal(t, http.StatusTemporaryRedirect, resp.StatusCode)
}

func TestBothHandlersWithJSON(t *testing.T) {
	config := &configs.Config{
		ServerAddress:   "localhost:8080",
		BaseURL:         "http://localhost:8080",
		FileStoragePath: configs.FileStorageDefault,
	}
	r := NewRouter(config)
	ts := httptest.NewServer(r)
	defer ts.Close()

	//сначала подготавливаем сокращенную ссылку через POST /api/shorten
	resp, body := testRequest(t, ts, http.MethodPost, "/api/shorten", bytes.NewBufferString(`{"url": "https://github.com/test_repo1"}`))
	defer resp.Body.Close()
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	assert.JSONEq(t, body, "{\"result\":\"http://localhost:8080/1\"}")

	//получаем оригинальную ссылку через GET запрос
	resp, _ = testRequest(t, ts, http.MethodGet, fmt.Sprintf("/%d", 1), nil)
	defer resp.Body.Close()
	assert.Equal(t, "https://github.com/test_repo1", resp.Header.Get("Location"))
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
		{
			name: "Test #2",
			id:   -555,
			want: want{
				headerLocation: "",
				statusCode:     http.StatusBadRequest,
				responseBody:   "Invalid ID\n",
			},
		},
	}

	config := &configs.Config{
		ServerAddress:   "localhost:8080",
		BaseURL:         "http://localhost:8080",
		FileStoragePath: configs.FileStorageDefault,
	}
	r := NewRouter(config)
	ts := httptest.NewServer(r)
	defer ts.Close()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, body := testRequest(t, ts, http.MethodGet, fmt.Sprintf("/%d", tt.id), nil)
			defer resp.Body.Close()
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
			longURL: "https://practicum.yandex.ru/learn/go-developer/courses/",
			want: want{
				contentType:  configs.ContentValuePlainText,
				statusCode:   http.StatusCreated,
				responseBody: "http://localhost:8080/1",
			},
		},
		{
			name:    "Test #2",
			longURL: "",
			want: want{
				contentType:  configs.ContentValuePlainText,
				statusCode:   http.StatusBadRequest,
				responseBody: "URL must not be empty\n",
			},
		},
		{
			name:    "Test #3",
			longURL: "htt p://incorrect_url_here",
			want: want{
				contentType:  configs.ContentValuePlainText,
				statusCode:   http.StatusBadRequest,
				responseBody: "parse \"htt p://incorrect_url_here\": first path segment in URL cannot contain colon\n",
			},
		},
	}
	config := &configs.Config{
		ServerAddress:   "localhost:8080",
		BaseURL:         "http://localhost:8080",
		FileStoragePath: configs.FileStorageDefault,
	}
	r := NewRouter(config)
	ts := httptest.NewServer(r)
	defer ts.Close()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, body := testRequest(t, ts, http.MethodPost, "/", bytes.NewBufferString(tt.longURL))
			defer resp.Body.Close()
			assert.Equal(t, tt.want.statusCode, resp.StatusCode)
			assert.Equal(t, tt.want.responseBody, body)
			assert.Equal(t, tt.want.contentType, resp.Header.Get(configs.ContentType))
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
			name:     "test valid JSON request",
			jsonBody: `{"url": "<some_url>"}`,
			want: want{
				contentType:  configs.ContentValueJSON,
				statusCode:   http.StatusCreated,
				responseBody: `{"result": "http://localhost:8080/1"}`,
				checkJSON:    true,
			},
		},
		{
			name:     "test invalid format for JSON request",
			jsonBody: `{"invalid": "<some_url>"}`,
			want: want{
				contentType:  configs.ContentValuePlainText,
				statusCode:   http.StatusBadRequest,
				responseBody: "Invalid JSON format in request body. Expected: {\"url\": \"<some_url>\"}\n",
				checkJSON:    false,
			},
		},
		{
			name:     "test invalid input for decoder",
			jsonBody: `{"url": "invalid\github.com/test_repo"}`,
			want: want{
				contentType:  configs.ContentValuePlainText,
				statusCode:   http.StatusBadRequest,
				responseBody: "invalid character 'g' in string escape code\n",
				checkJSON:    false,
			},
		},
		{
			name:     "test empty input URL",
			jsonBody: `{"url": ""}`,
			want: want{
				contentType:  configs.ContentValuePlainText,
				statusCode:   http.StatusBadRequest,
				responseBody: "Invalid JSON format in request body. Expected: {\"url\": \"<some_url>\"}\n",
				checkJSON:    false,
			},
		},
	}
	config := &configs.Config{
		ServerAddress:   "localhost:8080",
		BaseURL:         "http://localhost:8080",
		FileStoragePath: configs.FileStorageDefault,
	}
	r := NewRouter(config)
	ts := httptest.NewServer(r)
	defer ts.Close()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, body := testRequest(t, ts, http.MethodPost, "/api/shorten", bytes.NewBufferString(tt.jsonBody))
			defer resp.Body.Close()
			assert.Equal(t, tt.want.statusCode, resp.StatusCode)

			if tt.want.checkJSON {
				assert.JSONEq(t, tt.want.responseBody, body)
			} else {
				assert.Equal(t, tt.want.responseBody, body)
			}

			assert.Equal(t, tt.want.contentType, resp.Header.Get(configs.ContentType))
		})
	}
}
