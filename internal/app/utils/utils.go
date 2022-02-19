package utils

import (
	"errors"
	"fmt"
	"go-developer-course-shortener/internal/app/repository"
	"log"
	"net/url"
)

func SaveShortURL(storage *repository.Repository, strURL string, baseURL string) (string, error) {
	log.Printf("Long URL: %v", strURL)
	longURL, err := url.Parse(strURL)
	if err != nil {
		return "", err
	}
	if longURL.String() == "" {
		return "", errors.New("URL must not be empty")
	}
	id := storage.SaveURL(longURL.String())

	shortURL := fmt.Sprintf("%v/%d", baseURL, id)
	log.Printf("Short URL: %v", shortURL)
	return shortURL, nil
}
