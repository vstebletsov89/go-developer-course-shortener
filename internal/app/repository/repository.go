package repository

import "go-developer-course-shortener/internal/app/types"

type Repository interface {
	SaveURL(userID string, URL string) (int, error)
	SaveBatchURLS(userID string, request types.RequestBatch, baseURL string) (types.ResponseBatch, error)
	GetURL(id int) (string, error)
	GetUserStorage(userID string, baseURL string) ([]types.Link, error)
}
