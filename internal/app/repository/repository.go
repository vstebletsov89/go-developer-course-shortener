package repository

type Repository interface {
	SaveURL(URL string) (int, error)
	GetURL(id int) (string, error)
}
