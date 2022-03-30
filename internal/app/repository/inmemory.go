package repository

import (
	"context"
	"errors"
	"go-developer-course-shortener/internal/app/types"
	"log"
)

// InMemoryRepository implements Repository interface
type InMemoryRepository struct {
	inMemoryMap         map[string]string
	inMemoryUserStorage map[string][]string
}

func (r *InMemoryRepository) SaveURL(userID string, shortURL string, originalURL string) error {
	r.inMemoryMap[shortURL] = originalURL
	r.inMemoryUserStorage[userID] = append(r.inMemoryUserStorage[userID], shortURL)
	return nil
}

func (r *InMemoryRepository) GetShortURLByOriginalURL(originalURL string) (string, error) {
	return "", nil
}

func (r *InMemoryRepository) DeleteURLS(ctx context.Context, userID string, shortURLS []string) error {
	return nil
}

func (r *InMemoryRepository) SaveBatchURLS(userID string, links types.BatchLinks) (types.ResponseBatch, error) {
	var response types.ResponseBatch
	for _, v := range links {
		response = append(response, types.ResponseBatchJSON{CorrelationID: v.CorrelationID, ShortURL: v.ShortURL})
	}
	return response, nil
}

func (r *InMemoryRepository) GetURL(shortURL string) (types.OriginalLink, error) {
	URL, ok := r.inMemoryMap[shortURL]
	if !ok {
		return types.OriginalLink{}, errors.New("ID not found")
	}
	return types.OriginalLink{OriginalURL: URL, Deleted: false}, nil
}

func (r *InMemoryRepository) GetUserStorage(userID string) ([]types.Link, error) {
	var links []types.Link
	ids, ok := r.inMemoryUserStorage[userID]
	if !ok {
		return links, errors.New("UserID not found")
	}
	for _, v := range ids {
		URL, ok := r.inMemoryMap[v]
		if !ok {
			return links, errors.New("ID not found")
		}
		links = append(links, types.Link{ShortURL: v, OriginalURL: URL})
	}
	return links, nil
}

func (r *InMemoryRepository) Ping() bool {
	return true
}

func NewInMemoryRepository() *InMemoryRepository {
	log.Print("Memory storage is used")
	return &InMemoryRepository{inMemoryMap: make(map[string]string), inMemoryUserStorage: make(map[string][]string)}
}
