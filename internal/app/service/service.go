package service

import (
	"errors"
	"go-developer-course-shortener/internal/app/repository"
	"go-developer-course-shortener/internal/app/types"
	"go-developer-course-shortener/internal/worker"
	"log"
	"net"
)

// Service represents struct for http/https and grpc servers.
type Service struct {
	storage repository.Repository
	job     chan worker.Job
	network *net.IPNet
	BaseURL string
}

// ShortenerStorage is the interface that must be implemented by the service.
type ShortenerStorage interface {
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
	DeleteURLS(userID string, shortURLS []string) error
	// GetInternalStats returns internal stats for repository.
	GetInternalStats(userIP net.IP) (types.ResponseStatsJSON, error)
}

// check that Service implements all required methods
var _ ShortenerStorage = (*Service)(nil)

func NewService(storage repository.Repository, job chan worker.Job, network *net.IPNet, baseURL string) *Service {
	return &Service{
		storage: storage,
		job:     job,
		network: network,
		BaseURL: baseURL,
	}
}

func (s *Service) SaveURL(userID string, shortURL string, originalURL string) error {
	return s.storage.SaveURL(userID, shortURL, originalURL)
}

func (s *Service) SaveBatchURLS(userID string, links types.BatchLinks) (types.ResponseBatch, error) {
	return s.storage.SaveBatchURLS(userID, links)
}

func (s *Service) GetURL(shortURL string) (types.OriginalLink, error) {
	return s.storage.GetURL(shortURL)
}

func (s *Service) GetShortURLByOriginalURL(originalURL string) (string, error) {
	return s.storage.GetShortURLByOriginalURL(originalURL)
}

func (s *Service) GetUserStorage(userID string) ([]types.Link, error) {
	return s.storage.GetUserStorage(userID)
}

func (s *Service) Ping() bool {
	return s.storage.Ping()
}

func (s *Service) DeleteURLS(userID string, shortURLS []string) error {
	j := worker.Job{UserID: userID, ShortURLS: shortURLS}
	s.job <- j
	return nil
}

func (s *Service) GetInternalStats(userIP net.IP) (types.ResponseStatsJSON, error) {
	if s.network == nil || !s.network.Contains(userIP) {
		return types.ResponseStatsJSON{}, errors.New("access forbidden")
	}

	urls, users, err := s.storage.GetInternalStats()
	if err != nil {
		return types.ResponseStatsJSON{}, err
	}

	response := types.ResponseStatsJSON{
		URLs:  urls,
		Users: users,
	}
	log.Printf("GetInternalStats ResponseStatsJSON: %+v", response)

	return response, nil
}
