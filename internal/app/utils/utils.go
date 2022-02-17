package utils

import (
	"errors"
	"fmt"
	"go-developer-course-shortener/internal/app/repository"
	"go-developer-course-shortener/internal/configs"
	"log"
	"net/url"
)

func MakeShortURL(strURL string) (string, error) {
	log.Printf("Long URL: %v", strURL)
	longURL, err := url.Parse(strURL)
	if err != nil {
		return "", err
	}
	if longURL.String() == "" {
		return "", errors.New("URL must not be empty")
	}
	id := repository.SaveURL(longURL.String())
	shortURL := fmt.Sprintf("http://%v/%d", configs.ServerAddress, id)
	log.Printf("Short URL: %v", shortURL)
	return shortURL, nil
}
