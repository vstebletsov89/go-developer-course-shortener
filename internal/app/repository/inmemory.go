package repository

import (
	"errors"
	"log"
)

// InMemoryRepository implements Repository interface
type InMemoryRepository struct {
	inMemoryMap map[int]string
}

func (r *InMemoryRepository) getNextID() (int, error) {
	return len(r.inMemoryMap) + 1, nil
}

func (r *InMemoryRepository) SaveURL(URL string) (int, error) {
	id, err := r.getNextID()
	if err != nil {
		return 0, err
	}
	r.inMemoryMap[id] = URL
	return id, nil
}

func (r *InMemoryRepository) GetURL(id int) (string, error) {
	URL, ok := r.inMemoryMap[id]
	if !ok {
		return "", errors.New("ID not found")
	}
	return URL, nil
}

func NewInMemoryRepository() *InMemoryRepository {
	log.Print("Memory storage is used")
	return &InMemoryRepository{inMemoryMap: make(map[int]string)}
}
