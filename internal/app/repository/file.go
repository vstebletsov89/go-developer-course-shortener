package repository

import (
	"context"
	"encoding/json"
	"errors"
	"go-developer-course-shortener/internal/app/types"
	"io"
	"io/fs"
	"log"
	"os"
)

// FileRepository implements Repository interface
type FileRepository struct {
	fileStoragePath string
	file            *os.File
}

type fileRecord struct {
	UserID      string `json:"user_id"`
	ID          string `json:"id"`
	OriginalURL string `json:"original_url"`
}

func (r *FileRepository) SaveURL(userID string, shortURL string, originalURL string) error {
	var err error
	r.file, err = os.OpenFile(r.fileStoragePath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0777)
	if err != nil {
		return err
	}
	defer r.ReleaseStorage()

	encoder := json.NewEncoder(r.file)
	err = encoder.Encode(&fileRecord{UserID: userID, ID: shortURL, OriginalURL: originalURL})
	if err != nil {
		return err
	}
	return nil
}

func (r *FileRepository) GetShortURLByOriginalURL(originalURL string) (string, error) {
	return "", nil
}

func (r *FileRepository) SaveBatchURLS(userID string, links types.BatchLinks) (types.ResponseBatch, error) {
	response := make(types.ResponseBatch, len(links)) // allocate required capacity for the links
	for i, v := range links {
		response[i] = types.ResponseBatchJSON{CorrelationID: v.CorrelationID, ShortURL: v.ShortURL}
	}
	return response, nil
}

func (r *FileRepository) DeleteURLS(ctx context.Context, userID string, shortURLS []string) error {
	return nil
}

func (r *FileRepository) GetURL(shortURL string) (types.OriginalLink, error) {
	var err error
	r.file, err = os.OpenFile(r.fileStoragePath, os.O_RDONLY|os.O_CREATE, 0777)
	if err != nil {
		return types.OriginalLink{}, err
	}
	defer r.ReleaseStorage()

	decoder := json.NewDecoder(r.file)
	for {
		record := &fileRecord{}
		if err := decoder.Decode(&record); err == io.EOF {
			break
		} else if err != nil {
			return types.OriginalLink{}, err
		}

		log.Printf("Record from file (get): %+v", record)
		if record.ID == shortURL {
			return types.OriginalLink{OriginalURL: record.OriginalURL, Deleted: false}, nil
		}
	}
	return types.OriginalLink{}, errors.New("ID not found")
}

func (r *FileRepository) GetUserStorage(userID string) ([]types.Link, error) {
	var links []types.Link
	var err error
	r.file, err = os.OpenFile(r.fileStoragePath, os.O_RDONLY|os.O_CREATE, 0777)
	if err != nil {
		return links, err
	}
	defer r.ReleaseStorage()

	decoder := json.NewDecoder(r.file)
	for {
		record := &fileRecord{}
		if err := decoder.Decode(&record); err == io.EOF {
			break
		} else if err != nil {
			return links, err
		}

		log.Printf("Record from file (getUserStorage): %+v", record)
		if record.UserID == userID {
			links = append(links, types.Link{ShortURL: record.ID, OriginalURL: record.OriginalURL})
		}
	}
	return links, nil
}

func (r *FileRepository) Ping() bool {
	return true
}

func (r *FileRepository) ReleaseStorage() {
	log.Println("Storage released")
	err := r.file.Close()
	if err != nil && !errors.Is(err, fs.ErrClosed) {
		log.Fatalf("Failed to release file storage. Error: %v", err.Error())
	}
}

// NewFileRepository returns a new FileRepository.
func NewFileRepository(fileStoragePath string) *FileRepository {
	log.Print("File storage is used")
	return &FileRepository{fileStoragePath: fileStoragePath}
}
