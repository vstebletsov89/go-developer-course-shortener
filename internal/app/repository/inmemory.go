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

// check that InMemoryRepository implements all required methods
var _ Repository = (*InMemoryRepository)(nil)

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
	response := make(types.ResponseBatch, len(links)) // allocate required capacity for the links
	for i, v := range links {
		response[i] = types.ResponseBatchJSON{CorrelationID: v.CorrelationID, ShortURL: v.ShortURL}
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
	ids, ok := r.inMemoryUserStorage[userID]
	if !ok {
		return nil, errors.New("UserID not found")
	}

	links := make([]types.Link, len(ids)) // allocate required capacity for the links
	for i, v := range ids {
		URL, ok := r.inMemoryMap[v]
		if !ok {
			return links, errors.New("ID not found")
		}
		links[i] = types.Link{ShortURL: v, OriginalURL: URL}
	}
	return links, nil
}

func (r *InMemoryRepository) Ping() bool {
	return true
}

func (r *InMemoryRepository) ReleaseStorage() {
	log.Println("Storage released")
	// no need to release in memory storage
}

// NewInMemoryRepository returns a new InMemoryRepository.
func NewInMemoryRepository() *InMemoryRepository {
	log.Print("Memory storage is used")
	return &InMemoryRepository{inMemoryMap: make(map[string]string), inMemoryUserStorage: make(map[string][]string)}
}
