package repository

import (
	"context"
	"github.com/jackc/pgx/v4"
	"go-developer-course-shortener/internal/app/types"
	"go-developer-course-shortener/internal/app/utils"
	"log"
)

// DBRepository implements Repository interface
type DBRepository struct {
	conn *pgx.Conn
}

func (r *DBRepository) SaveURL(userID string, URL string) (int, error) {
	row := r.conn.QueryRow(context.Background(), "SELECT COUNT(*) FROM urls")
	var id int
	err := row.Scan(&id)
	if err != nil {
		return 0, err
	}

	sql := `INSERT INTO urls (user_id, original_url) VALUES ($1, $2)`
	_, err = r.conn.Exec(context.Background(), sql, userID, URL)
	if err != nil {
		return 0, err
	}
	return id + 1, nil
}

func (r *DBRepository) GetURL(id int) (string, error) {
	sql := `SELECT original_url FROM urls WHERE id = $1`
	row := r.conn.QueryRow(context.Background(), sql, id)
	var originalURL string
	err := row.Scan(&originalURL)
	if err != nil {
		return "", err
	}
	return originalURL, nil
}

func (r *DBRepository) GetUserStorage(userID string, baseURL string) ([]types.Link, error) {
	var links []types.Link
	sql := `SELECT id, original_url FROM urls WHERE user_id = $1`
	rows, err := r.conn.Query(context.Background(), sql, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var id int
		var url string
		err = rows.Scan(&id, &url)
		if err != nil {
			return nil, err
		}
		links = append(links, types.Link{ShortURL: utils.MakeShortURL(baseURL, id), OriginalURL: url})
	}
	return links, nil
}

func NewDBRepository(connection *pgx.Conn) (*DBRepository, error) {
	log.Print("DB storage is used")
	sql := `create table if not exists urls (
		id           serial not null primary key,
		user_id      text,
		original_url text
	);`
	_, err := connection.Exec(context.Background(), sql)
	if err != nil {
		return nil, err
	}
	return &DBRepository{conn: connection}, nil
}
