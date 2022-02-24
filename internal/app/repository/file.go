package repository

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"os"
)

// FileRepository implements Repository interface
type FileRepository struct {
	fileStoragePath string
}

type fileRecord struct {
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

func (r *FileRepository) SaveURL(URL string) (int, error) {
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
	err = encoder.Encode(&fileRecord{ID: id, OriginalURL: URL})
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (r *FileRepository) GetURL(id int) (string, error) {
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

func NewFileRepository(fileStoragePath string) *FileRepository {
	log.Print("File storage is used")
	return &FileRepository{fileStoragePath: fileStoragePath}
}
