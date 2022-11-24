// Package repository defines and implements interface for Repository.
package repository

import (
	"context"
	"go-developer-course-shortener/internal/app/types"
)

// Repository is the interface that must be implemented by specific repository.
type Repository interface {
	// SaveURL saves url to the current repository.
	SaveURL(userID string, shortURL string, originalURL string) error
	// SaveBatchURLS saves list of urls to the current repository.
	SaveBatchURLS(userID string, links types.BatchLinks) (types.ResponseBatch, error)
	// GetURL returns original url by short url.
	GetURL(shortURL string) (types.OriginalLink, error)
	// GetShortURLByOriginalURL returns short url by original url.
	GetShortURLByOriginalURL(originalURL string) (string, error)
	// GetUserStorage returns list of urls for current user id.
	GetUserStorage(userID string) ([]types.Link, error)
	// Ping verifies that current repository can accept requests.
	Ping() bool
	// DeleteURLS deletes list of short urls for current user id.
	DeleteURLS(ctx context.Context, userID string, shortURLS []string) error
	// GetInternalStats returns internal stats for repository.
	GetInternalStats() (int, int, error)
	// ReleaseStorage releases current storage.
	ReleaseStorage()
}
