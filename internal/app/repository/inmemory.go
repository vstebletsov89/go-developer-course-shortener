package repository

import (
	"errors"
	"go-developer-course-shortener/internal/app/types"
	"go-developer-course-shortener/internal/app/utils"
	"log"
)

// InMemoryRepository implements Repository interface
type InMemoryRepository struct {
	inMemoryMap         map[int]string
	inMemoryUserStorage map[string][]int
}

func (r *InMemoryRepository) getNextID() (int, error) {
	return len(r.inMemoryMap) + 1, nil
}

func (r *InMemoryRepository) SaveURL(userID string, URL string) (int, error) {
	id, err := r.getNextID()
	if err != nil {
		return 0, err
	}
	r.inMemoryMap[id] = URL
	r.inMemoryUserStorage[userID] = append(r.inMemoryUserStorage[userID], id)
	return id, nil
}

func (r *InMemoryRepository) GetURL(userID string, id int) (string, error) {
	ids, ok := r.inMemoryUserStorage[userID]
	if !ok {
		return "", errors.New("UserID not found")
	}
	var idExists = false
	for _, v := range ids {
		if v == id {
			idExists = true
		}
	}
	if !idExists {
		return "", errors.New("ID not found")
	}

	URL, ok := r.inMemoryMap[id]
	if !ok {
		return "", errors.New("ID not found")
	}
	return URL, nil
}

func (r *InMemoryRepository) GetUserStorage(userID string, baseURL string) ([]types.Link, error) {
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
		links = append(links, types.Link{ShortURL: utils.MakeShortURL(baseURL, v), OriginalURL: URL})
	}
	return links, nil
}

func NewInMemoryRepository() *InMemoryRepository {
	log.Print("Memory storage is used")
	return &InMemoryRepository{inMemoryMap: make(map[int]string), inMemoryUserStorage: make(map[string][]int)}
}
