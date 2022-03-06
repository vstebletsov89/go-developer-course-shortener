package repository

import (
	"errors"
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
	ids := r.inMemoryUserStorage[userID]
	var idExists = false
	for i := range ids {
		if ids[i] == id {
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

func NewInMemoryRepository() *InMemoryRepository {
	log.Print("Memory storage is used")
	return &InMemoryRepository{inMemoryMap: make(map[int]string), inMemoryUserStorage: make(map[string][]int)}
}
