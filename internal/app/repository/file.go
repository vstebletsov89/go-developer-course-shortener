package repository

import (
	"encoding/json"
	"errors"
	"go-developer-course-shortener/internal/app/types"
	"go-developer-course-shortener/internal/app/utils"
	"io"
	"log"
	"os"
)

// FileRepository implements Repository interface
type FileRepository struct {
	fileStoragePath string
}

type fileRecord struct {
	UserID      string `json:"user_id"`
	ID          int    `json:"id"`
	OriginalURL string `json:"original_url"`
}

func (r *FileRepository) getNextID() (int, error) {
	file, err := os.OpenFile(r.fileStoragePath, os.O_RDONLY|os.O_CREATE, 0777)
	if err != nil {
		return 0, err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	var counter int
	for {
		record := &fileRecord{}
		if err := decoder.Decode(&record); err == io.EOF {
			break
		} else if err != nil {
			return 0, err
		}
		log.Printf("Record from file (count): %+v", record)
		counter += 1
	}
	return counter + 1, nil
}

func (r *FileRepository) SaveURL(userID string, URL string) (int, error) {
	id, err := r.getNextID()
	if err != nil {
		return 0, err
	}

	file, err := os.OpenFile(r.fileStoragePath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0777)
	if err != nil {
		return 0, err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	err = encoder.Encode(&fileRecord{UserID: userID, ID: id, OriginalURL: URL})
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (r *FileRepository) GetURL(userID string, id int) (string, error) {
	file, err := os.OpenFile(r.fileStoragePath, os.O_RDONLY|os.O_CREATE, 0777)
	if err != nil {
		return "", err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	for {
		record := &fileRecord{}
		if err := decoder.Decode(&record); err == io.EOF {
			break
		} else if err != nil {
			return "", err
		}

		log.Printf("Record from file (get): %+v", record)
		if record.ID == id {
			return record.OriginalURL, nil
		}
	}
	return "", errors.New("ID not found")
}

func (r *FileRepository) GetUserStorage(userID string, baseURL string) ([]types.Link, error) {
	var links []types.Link
	file, err := os.OpenFile(r.fileStoragePath, os.O_RDONLY|os.O_CREATE, 0777)
	if err != nil {
		return links, err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	for {
		record := &fileRecord{}
		if err := decoder.Decode(&record); err == io.EOF {
			break
		} else if err != nil {
			return links, err
		}

		log.Printf("Record from file (getUserStorage): %+v", record)
		if record.UserID == userID {
			links = append(links, types.Link{ShortURL: utils.MakeShortURL(baseURL, record.ID), OriginalURL: record.OriginalURL})
		}
	}
	return links, nil
}

func NewFileRepository(fileStoragePath string) *FileRepository {
	log.Print("File storage is used")
	return &FileRepository{fileStoragePath: fileStoragePath}
}
