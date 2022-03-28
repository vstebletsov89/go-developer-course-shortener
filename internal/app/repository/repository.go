package repository

import (
	"context"
	"go-developer-course-shortener/internal/app/types"
)

type Repository interface {
	SaveURL(userID string, shortURL string, originalURL string) error
	SaveBatchURLS(userID string, links types.BatchLinks) (types.ResponseBatch, error)
	GetURL(shortURL string) (types.OriginalLink, error)
	GetShortURLByOriginalURL(originalURL string) (string, error)
	GetUserStorage(userID string) ([]types.Link, error)
	Ping() bool
	DeleteURLS(ctx context.Context, userID string, shortURLS []string) error
}
