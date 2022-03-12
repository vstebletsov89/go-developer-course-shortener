package utils

import (
	"errors"
	"fmt"
	"math/rand"
	"net/url"
	"time"
)

func GenerateRandom(size int) []byte {
	rand.Seed(time.Now().UnixNano())
	b := make([]byte, size)
	for i := range b {
		b[i] = Letters[rand.Intn(len(Letters))]
	}
	return b
}

func MakeShortURL(baseURL string, id string) string {
	shortURL := fmt.Sprintf("%v/%s", baseURL, id)
	return shortURL
}

func ParseURL(strURL string) (string, error) {
	longURL, err := url.Parse(strURL)
	if err != nil {
		return "", err
	}
	if longURL.String() == "" {
		return "", errors.New("URL must not be empty")
	}
	return longURL.String(), nil
}
