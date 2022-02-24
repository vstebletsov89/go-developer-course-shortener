package utils

import (
	"errors"
	"fmt"
	"net/url"
)

func MakeShortURL(baseURL string, id int) string {
	shortURL := fmt.Sprintf("%v/%d", baseURL, id)
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
