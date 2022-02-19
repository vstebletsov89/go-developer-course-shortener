package repository

import (
	"encoding/json"
	"fmt"
	"go-developer-course-shortener/internal/configs"
	"io"
	"log"
	"os"
)

type Repository struct {
	repositoryType  string
	fileStoragePath string
	inMemoryMap     map[int]string
	counterURL      int
}

type FileRecord struct {
	ID          int    `json:"id"`
	OriginalURL string `json:"original_url"`
}

func NewRepository(fileStoragePath string) *Repository {
	if fileStoragePath != configs.FileStorageDefault {
		log.Print("File storage is used")
		return &Repository{repositoryType: configs.RepositoryTypeFile, fileStoragePath: fileStoragePath,
			inMemoryMap: nil}
	}
	log.Print("Memory storage is used")
	return &Repository{repositoryType: configs.RepositoryTypeMemory, fileStoragePath: "",
		inMemoryMap: make(map[int]string)}
}

func (r *Repository) getNextID() int {
	//TODO: read from file to count all records or get len of inMemoryMap
	//TODO: remove counterURL
	return r.counterURL + 1
}

func (r *Repository) SaveURL(URL string) int {
	id := r.getNextID()
	if r.repositoryType == configs.RepositoryTypeFile {
		file := OpenFile(r.fileStoragePath, os.O_WRONLY|os.O_CREATE|os.O_APPEND)
		defer file.Close()

		encoder := json.NewEncoder(file)
		err := encoder.Encode(&FileRecord{ID: id, OriginalURL: URL})
		if err != nil {
			log.Fatalf("Cannot encode record: %s", err)
		}
	} else {
		r.inMemoryMap[id] = URL
	}
	r.counterURL += 1
	return id
}

func (r *Repository) GetURL(id int) (string, error) {
	if r.repositoryType == configs.RepositoryTypeFile {
		file := OpenFile(r.fileStoragePath, os.O_RDONLY|os.O_CREATE)
		defer file.Close()

		decoder := json.NewDecoder(file)
		for {
			record := &FileRecord{}
			if err := decoder.Decode(&record); err == io.EOF {
				break
			} else if err != nil {
				log.Fatal(err)
			}

			log.Printf("Record from file: %+v", record)
			if record.ID == id {
				return record.OriginalURL, nil
			}
		}
		return "", fmt.Errorf("not found")
	} else {
		URL, ok := r.inMemoryMap[id]
		if !ok {
			return "", fmt.Errorf("not found")
		}
		return URL, nil
	}
}

func OpenFile(fileName string, flag int) *os.File {
	file, err := os.OpenFile(fileName, flag, 0777)
	if err != nil {
		log.Fatal(err)
	}
	return file
}
