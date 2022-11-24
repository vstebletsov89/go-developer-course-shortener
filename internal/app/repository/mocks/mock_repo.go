package mocks

import (
	"context"
	"errors"
	"go-developer-course-shortener/internal/app/repository"
	"go-developer-course-shortener/internal/app/types"
	"log"
)

// MockRepository implements Repository interface to check negative scenarios
type MockRepository struct {
}

// check that MockRepository implements all required methods
var _ repository.Repository = (*MockRepository)(nil)

func (r *MockRepository) GetInternalStats() (int, int, error) {
	return 0, 0, errors.New("GetInternalStats error")
}

func (r *MockRepository) SaveURL(userID string, shortURL string, originalURL string) error {
	return errors.New("SaveURL error")
}

func (r *MockRepository) GetShortURLByOriginalURL(originalURL string) (string, error) {
	return "", errors.New("GetShortURLByOriginalURL error")
}

func (r *MockRepository) DeleteURLS(ctx context.Context, userID string, shortURLS []string) error {
	return errors.New("DeleteURLS error")
}

func (r *MockRepository) SaveBatchURLS(userID string, links types.BatchLinks) (types.ResponseBatch, error) {
	var resp types.ResponseBatch
	return resp, errors.New("SaveBatchURLS error")
}

func (r *MockRepository) GetURL(shortURL string) (types.OriginalLink, error) {
	var link types.OriginalLink
	return link, errors.New("GetURL error")
}

func (r *MockRepository) GetUserStorage(userID string) ([]types.Link, error) {
	return nil, errors.New("GetUserStorage error")
}

func (r *MockRepository) Ping() bool {
	return false
}

func (r *MockRepository) ReleaseStorage() {
}

// NewMockRepository returns a new MockRepository.
func NewMockRepository() *MockRepository {
	log.Print("MockRepository storage is used")
	return &MockRepository{}
}
