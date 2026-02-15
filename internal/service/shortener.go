package service

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"net/url"

	"github.com/jackc/pgx/v5/pgconn"

	"urlshortener/internal/repository/postgres"
)

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

var ErrInvalidURL = errors.New("invalid URL")

type URLRepository interface {
	Create(ctx context.Context, code string, originalURL string) error
	GetOriginalByCode(ctx context.Context, code string) (string, error)
}

type Shortener struct {
	repo               URLRepository
	baseURL            string
	codeLength         int
	maxGenerateRetries int
}

func NewShortener(repo URLRepository, baseURL string, codeLength int, maxGenerateRetries int) *Shortener {
	return &Shortener{
		repo:               repo,
		baseURL:            baseURL,
		codeLength:         codeLength,
		maxGenerateRetries: maxGenerateRetries,
	}
}

func (s *Shortener) CreateShortURL(ctx context.Context, rawURL string) (code string, shortURL string, err error) {
	if !isValidURL(rawURL) {
		return "", "", ErrInvalidURL
	}

	for i := 0; i < s.maxGenerateRetries; i++ {
		code, err = randomCode(s.codeLength)
		if err != nil {
			return "", "", err
		}

		err = s.repo.Create(ctx, code, rawURL)
		if err == nil {
			return code, fmt.Sprintf("%s/%s", s.baseURL, code), nil
		}

		if isUniqueViolation(err) {
			continue
		}
		return "", "", err
	}

	return "", "", fmt.Errorf("could not generate unique code after %d retries", s.maxGenerateRetries)
}

func (s *Shortener) ResolveCode(ctx context.Context, code string) (string, error) {
	if code == "" {
		return "", postgres.ErrNotFound
	}
	return s.repo.GetOriginalByCode(ctx, code)
}

func randomCode(n int) (string, error) {
	buf := make([]byte, n)
	for i := range buf {
		idx, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", err
		}
		buf[i] = charset[idx.Int64()]
	}
	return string(buf), nil
}

func isValidURL(raw string) bool {
	u, err := url.ParseRequestURI(raw)
	if err != nil {
		return false
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return false
	}
	return u.Host != ""
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505"
	}
	return false
}
