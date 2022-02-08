package handlers

import (
	"bytes"
	"fmt"
	"github.com/stretchr/testify/assert"
	"go-developer-course-shortener/internal/app/storage"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandlerShortenerInvalidMethod(t *testing.T) {
	request := httptest.NewRequest(http.MethodPatch, "/", bytes.NewBufferString("https://practicum.yandex.ru/learn/go-developer/courses/"))

	w := httptest.NewRecorder()
	h := http.HandlerFunc(HandlerShortener)
	h.ServeHTTP(w, request)
	res := w.Result()

	assert.Equal(t, http.StatusBadRequest, w.Code)

	defer res.Body.Close()
	_, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, "Unexpected error\n", w.Body.String())
	assert.Equal(t, "text/plain; charset=utf-8", res.Header.Get("Content-Type"))
}

func TestHandlerShortenerGetSuccess(t *testing.T) {
	//сначала подготавливаем сокращенную ссылку через POST
	r := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString("https://github.com/test_repo1"))
	w := httptest.NewRecorder()
	h := http.HandlerFunc(HandlerShortener)
	h.ServeHTTP(w, r)
	assert.Equal(t, http.StatusCreated, w.Code)

	r2 := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/%d", 1), nil)

	w2 := httptest.NewRecorder()
	h2 := http.HandlerFunc(HandlerShortener)
	h2.ServeHTTP(w2, r2)
	res := w2.Result()

	assert.Equal(t, "https://github.com/test_repo1", res.Header.Get("Location"))
	storage.InitRepository()
}

func TestHandlerShortenerGetError(t *testing.T) {
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
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/%d", tt.id), nil)

			w := httptest.NewRecorder()
			h := http.HandlerFunc(HandlerShortener)
			h.ServeHTTP(w, request)
			res := w.Result()

			assert.Equal(t, tt.want.statusCode, w.Code)

			defer res.Body.Close()
			_, err := io.ReadAll(res.Body)
			if err != nil {
				t.Fatal(err)
			}

			assert.Equal(t, tt.want.responseBody, w.Body.String())
			assert.Equal(t, tt.want.headerLocation, res.Header.Get("Location"))
		})
	}
}

func TestHandlerShortenerPost(t *testing.T) {
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
				contentType:  "text/plain; charset=utf-8",
				statusCode:   http.StatusCreated,
				responseBody: "http://localhost:8080/1",
			},
		},
		{
			name:    "Test #2",
			longURL: "",
			want: want{
				contentType:  "text/plain; charset=utf-8",
				statusCode:   http.StatusBadRequest,
				responseBody: "URL must not be empty\n",
			},
		},
		{
			name:    "Test #3",
			longURL: "htt p://incorrect_url_here",
			want: want{
				contentType:  "text/plain; charset=utf-8",
				statusCode:   http.StatusBadRequest,
				responseBody: "parse \"htt p://incorrect_url_here\": first path segment in URL cannot contain colon\n",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(tt.longURL))

			w := httptest.NewRecorder()
			h := http.HandlerFunc(HandlerShortener)
			h.ServeHTTP(w, request)
			res := w.Result()

			assert.Equal(t, tt.want.statusCode, w.Code)

			defer res.Body.Close()
			_, err := io.ReadAll(res.Body)
			if err != nil {
				t.Fatal(err)
			}

			assert.Equal(t, tt.want.responseBody, w.Body.String())
			assert.Equal(t, tt.want.contentType, res.Header.Get("Content-Type"))
		})
	}
}
