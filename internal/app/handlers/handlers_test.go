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
	"net/http"
	"net/http/httptest"
	"testing"
)

func testRequest(t *testing.T, ts *httptest.Server, method, path string, body io.Reader) (*http.Response, string) {
	req, err := http.NewRequest(method, ts.URL+path, body)
	if err != nil {
		t.Fatal(err)
		return nil, ""
	}

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
		return nil, ""
	}

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
		return nil, ""
	}
	defer resp.Body.Close()

	return resp, string(respBody)
}

func NewRouter() chi.Router {
	r := chi.NewRouter()

	r.Route("/",
		func(r chi.Router) {
			r.Get("/{ID}", HandlerGET)
			r.Post("/", HandlerPOST)
			r.Post("/api/shorten", HandlerJSONPOST)
		})

	return r
}

func TestBothHandlers(t *testing.T) {
	r := NewRouter()
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
	repository.InitRepository()
}

func TestBothHandlersWithJSON(t *testing.T) {
	r := NewRouter()
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
	repository.InitRepository()
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

	r := NewRouter()
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
	r := NewRouter()
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
	repository.InitRepository()
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
	r := NewRouter()
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
	repository.InitRepository()
}
