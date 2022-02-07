package handlers

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

//TODO: check 6 bad requests + one incorrect (not GET, not POST request)
//TODO: add tests for POST (all error and edge cases)
//TODO: add test for GET (all error and edge cases)
//TODO: add test for POST -> GET (positive + positive)
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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(tt.longURL))

			w := httptest.NewRecorder()
			h := http.HandlerFunc(HandlerShortener)
			h.ServeHTTP(w, request)
			res := w.Result()

			if res.StatusCode != tt.want.statusCode {
				t.Errorf("Expected status code %d, got %d", tt.want.statusCode, w.Code)
			}

			defer res.Body.Close()
			resBody, err := io.ReadAll(res.Body)
			if err != nil {
				t.Fatal(err)
			}
			if string(resBody) != tt.want.responseBody {
				t.Errorf("Expected body %s, got %s", tt.want.responseBody, w.Body.String())
			}

			if res.Header.Get("Content-Type") != tt.want.contentType {
				t.Errorf("Expected Content-Type %s, got %s", tt.want.contentType, res.Header.Get("Content-Type"))
			}
		})
	}
}
