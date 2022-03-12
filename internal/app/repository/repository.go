package repository

import "go-developer-course-shortener/internal/app/types"

type Repository interface {
	SaveURL(userID string, shortURL string, originalURL string) error
	SaveBatchURLS(userID string, links types.BatchLinks) (types.ResponseBatch, error)
	GetURL(shortURL string) (string, error)
	GetShortURLByOriginalURL(originalURL string) (string, error)
	GetUserStorage(userID string) ([]types.Link, error)
	Ping() bool
}
