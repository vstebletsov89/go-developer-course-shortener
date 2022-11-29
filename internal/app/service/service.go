package service

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgerrcode"
	"go-developer-course-shortener/internal/app/rand"
	"go-developer-course-shortener/internal/app/repository"
	"go-developer-course-shortener/internal/app/types"
	"go-developer-course-shortener/internal/worker"
	"google.golang.org/grpc/metadata"
	"log"
	"net"
	"net/http"
	"net/url"
	"sync"
)

// Service represents struct for http/https and grpc servers.
type Service struct {
	storage repository.Repository
	job     chan worker.Job
	network *net.IPNet
	BaseURL string
}

// UserContextType user context type.
type UserContextType string

const (
	// AccessToken defines cookie name for current user.
	AccessToken = "uniqueAuthToken"
	// UserCtx defines user context name.
	UserCtx UserContextType = "UserCtx"
)

type cipherData struct {
	key    []byte
	nonce  []byte
	aesGCM cipher.AEAD
}

var cipherInstance *cipherData
var once sync.Once

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
	// CreateUser creates new uuid user.
	CreateUser() string
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

func (s *Service) CreateUser() string {
	return uuid.NewString()
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

func cipherInit() error {
	var e error
	once.Do(func() {
		key := rand.GenerateRandom(2 * aes.BlockSize)

		aesblock, err := aes.NewCipher(key)
		if err != nil {
			e = err
		}

		aesgcm, err := cipher.NewGCM(aesblock)
		if err != nil {
			e = err
		}

		nonce := rand.GenerateRandom(aesgcm.NonceSize())
		cipherInstance = &cipherData{key: key, aesGCM: aesgcm, nonce: nonce}
	})
	return e
}

func Encrypt(userID string) (string, error) {
	if err := cipherInit(); err != nil {
		return "", err
	}
	encrypted := cipherInstance.aesGCM.Seal(nil, cipherInstance.nonce, []byte(userID), nil)
	return hex.EncodeToString(encrypted), nil
}

func Decrypt(token string) (string, error) {
	if err := cipherInit(); err != nil {
		return "", err
	}
	b, err := hex.DecodeString(token)
	if err != nil {
		return "", err
	}
	userID, err := cipherInstance.aesGCM.Open(nil, cipherInstance.nonce, b, nil)
	if err != nil {
		return "", err
	}
	return string(userID), nil
}

func ExtractUserIDFromContext(ctx context.Context) string {
	// try to get userID from metadata
	md, ok := metadata.FromIncomingContext(ctx)
	if ok {
		values := md.Get(AccessToken)
		if len(values) > 0 {
			userID := values[0]
			log.Printf("ExtractUserIDFromContext (GRPC): '%s'", userID)
			return userID
		}
		return ""
	}

	// if not in metadata try to find in context value
	userID, ok := ctx.Value(UserCtx).(string)
	if ok {
		log.Printf("ExtractUserIDFromContext (HTTP): '%s'", userID)
		return userID
	}
	return ""
}

func MakeShortURL(baseURL string, id string) string {
	shortURL := fmt.Sprintf("%v/%s", baseURL, id)
	return shortURL
}

func ParseURL(strURL string) (string, error) {
	longURL, err := url.Parse(strURL)
	if err != nil {
		return "", err
	}
	if longURL.String() == "" {
		return "", errors.New("URL must not be empty")
	}
	return longURL.String(), nil
}

func CheckDBViolation(err error) (int, error) {
	var pgError *pgconn.PgError
	if errors.As(err, &pgError) && pgError.Code == pgerrcode.UniqueViolation {
		return http.StatusConflict, nil
	} else if err != nil {
		return http.StatusInternalServerError, err
	}
	return http.StatusCreated, nil
}
