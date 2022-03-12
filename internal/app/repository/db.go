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

func (r *DBRepository) SaveURL(userID string, shortURL string, originalURL string) error {
	sql := `INSERT INTO urls (user_id, short_url, original_url) VALUES ($1, $2, $3)`
	_, err := r.conn.Exec(context.Background(), sql, userID, shortURL, originalURL)
	if err != nil {
		return err
	}
	return nil
}

func (r *DBRepository) SaveBatchURLS(userID string, links types.BatchLinks) (types.ResponseBatch, error) {
	ctx := context.Background()
	tx, err := r.conn.Begin(ctx)
	defer func() {
		if err != nil {
			tx.Rollback(ctx)
		}
	}()
	if err != nil {
		return nil, err
	}

	sql := `INSERT INTO urls (user_id, short_url, original_url) VALUES ($1, $2, $3)`

	var response types.ResponseBatch

	for _, v := range links {
		_, err := tx.Exec(ctx, sql, userID, v.ShortURL, v.OriginalURL)
		if err != nil {
			return nil, err
		}
		response = append(response, types.ResponseBatchJSON{CorrelationID: v.CorrelationID, ShortURL: v.ShortURL})
	}

	err = tx.Commit(ctx)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func (r *DBRepository) GetURL(shortURL string) (string, error) {
	sql := `SELECT original_url FROM urls WHERE short_url = $1`
	row := r.conn.QueryRow(context.Background(), sql, shortURL)
	var originalURL string
	err := row.Scan(&originalURL)
	if err != nil {
		return "", err
	}
	return originalURL, nil
}

func (r *DBRepository) GetShortURLByOriginalURL(originalURL string) (string, error) {
	sql := `SELECT short_url FROM urls WHERE original_url = $1`
	row := r.conn.QueryRow(context.Background(), sql, originalURL)
	var shortURL string
	err := row.Scan(&shortURL)
	if err != nil {
		return "", err
	}
	return shortURL, nil
}

func (r *DBRepository) GetUserStorage(userID string) ([]types.Link, error) {
	var links []types.Link
	sql := `SELECT short_url, original_url FROM urls WHERE user_id = $1`
	rows, err := r.conn.Query(context.Background(), sql, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var shortURL string
		var originalURL string
		err = rows.Scan(&shortURL, &originalURL)
		if err != nil {
			return nil, err
		}
		links = append(links, types.Link{ShortURL: shortURL, OriginalURL: originalURL})
	}
	return links, nil
}

func (r *DBRepository) Ping() bool {
	err := r.conn.Ping(context.Background())

	return err == nil
}

func NewDBRepository(connection *pgx.Conn) (*DBRepository, error) {
	log.Print("DB storage is used")
	_, err := connection.Exec(context.Background(), PostgreSQLTable)
	if err != nil {
		return nil, err
	}
	return &DBRepository{conn: connection}, nil
}
