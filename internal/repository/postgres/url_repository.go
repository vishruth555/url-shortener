package postgres

import (
	"context"
	"errors"
	"os"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = errors.New("not found")

func RunSchema(ctx context.Context, pool *pgxpool.Pool, schemaFilePath string) error {
	schemaSQL, err := os.ReadFile(schemaFilePath)
	if err != nil {
		return err
	}

	_, err = pool.Exec(ctx, string(schemaSQL))
	return err
}

type URLRepository struct {
	pool *pgxpool.Pool
}

func NewURLRepository(pool *pgxpool.Pool) *URLRepository {
	return &URLRepository{pool: pool}
}

func (r *URLRepository) Create(ctx context.Context, code string, originalURL string) error {
	_, err := r.pool.Exec(ctx, `INSERT INTO urls(code, original_url) VALUES($1, $2)`, code, originalURL)
	return err
}

func (r *URLRepository) GetOriginalByCode(ctx context.Context, code string) (string, error) {
	const q = `
		UPDATE urls
		SET hits = hits + 1
		WHERE code = $1
		RETURNING original_url
	`

	var originalURL string
	err := r.pool.QueryRow(ctx, q, code).Scan(&originalURL)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", ErrNotFound
	}
	if err != nil {
		return "", err
	}

	return originalURL, nil
}
