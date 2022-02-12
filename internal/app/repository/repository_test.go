package repository

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetURL(t *testing.T) {
	tests := []struct {
		name string
		id   int
		want string
	}{
		{name: "Test #1", id: 1, want: "http://localhost:8080/test_long_url1"},
		{name: "Test #2", id: 2, want: "http://localhost:8080/test_long_url2"},
		{name: "Test #3", id: 3, want: "http://localhost:8080/test_long_url3"},
		{name: "Test #4", id: 777, want: ""},
	}
	SaveURL("http://localhost:8080/test_long_url1")
	SaveURL("http://localhost:8080/test_long_url2")
	SaveURL("http://localhost:8080/test_long_url3")
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			URL, _ := GetURL(tt.id)
			assert.Equal(t, tt.want, URL)
		})
	}
	//очищаем репозиторий
	InitRepository()
}

func TestSaveURL(t *testing.T) {
	tests := []struct {
		name string
		URL  string
		want int
	}{
		{name: "Test #1", URL: "http://localhost:8080/test_long_url1", want: 1},
		{name: "Test #2", URL: "http://localhost:8080/test_long_url2", want: 2},
		{name: "Test #3", URL: "http://localhost:8080/test_long_url3", want: 3},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, SaveURL(tt.URL))
		})
	}
	//очищаем репозиторий
	InitRepository()
}
