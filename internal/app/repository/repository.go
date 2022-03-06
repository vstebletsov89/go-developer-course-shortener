package repository

type Repository interface {
	SaveURL(userID string, URL string) (int, error)
	GetURL(userID string, id int) (string, error)
}
