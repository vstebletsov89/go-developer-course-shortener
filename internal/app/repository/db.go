package repository

import (
	"context"
	"github.com/jackc/pgx/v4"
	"go-developer-course-shortener/internal/app/types"
	"log"
)

// DBRepository implements Repository interface
type DBRepository struct {
	conn *pgx.Conn
}

func (r *DBRepository) SaveURL(userID string, URL string) (int, error) {
	//TODO: implement
	return 0, nil
}

func (r *DBRepository) GetURL(userID string, id int) (string, error) {
	//TODO: implement
	return "", nil
}

func (r *DBRepository) GetUserStorage(userID string, baseURL string) ([]types.Link, error) {
	//TODO: implement
	return nil, nil
}

func NewDBRepository(connection *pgx.Conn) (*DBRepository, error) {
	log.Print("DB storage is used")
	const sql = `create table if not exists urls (
		id           serial not null primary key,
		user_id      text,
		short_url    text,
		original_url text
	);`
	_, err := connection.Exec(context.Background(), sql)
	if err != nil {
		return nil, err
	}
	return &DBRepository{conn: connection}, nil
}
