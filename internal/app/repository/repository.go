package repository

import (
	"errors"
	"fmt"
	"log"
	"net/url"
)

type Repository interface {
	getNextID() (int, error)
	SaveURL(URL string) (int, error)
	GetURL(id int) (string, error)
}

func SaveShortURL(storage Repository, strURL string, baseURL string) (string, error) {
	log.Printf("Long URL: %v", strURL)
	longURL, err := url.Parse(strURL)
	if err != nil {
		return "", err
	}
	if longURL.String() == "" {
		return "", errors.New("URL must not be empty")
	}
	id, err := storage.SaveURL(longURL.String())
	if err != nil {
		return "", err
	}
	shortURL := fmt.Sprintf("%v/%d", baseURL, id)
	log.Printf("Short URL: %v", shortURL)
	return shortURL, nil
}
