package postgres

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"
)

var ErrNotFound = errors.New("not found")

type URLRepository struct {
	db *gorm.DB
}

func NewURLRepository(db *gorm.DB) *URLRepository {
	return &URLRepository{db: db}
}

func (r *URLRepository) Create(ctx context.Context, code string, originalURL string) error {
	url := URL{
		Code:        code,
		OriginalURL: originalURL,
		Hits:        0,
	}

	return r.db.WithContext(ctx).Create(&url).Error
}

func (r *URLRepository) GetOriginalByCode(ctx context.Context, code string) (string, error) {
	var url URL

	result := r.db.WithContext(ctx).
		Where("code = ?", code).
		First(&url)

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return "", ErrNotFound
	}

	if result.Error != nil {
		return "", result.Error
	}

	// increment hits
	err := r.db.WithContext(ctx).
		Model(&URL{}).
		Where("code = ?", code).
		Update("hits", gorm.Expr("hits + 1")).Error

	if err != nil {
		return "", err
	}

	return url.OriginalURL, nil
}

func (r *URLRepository) GetAll(ctx context.Context) ([]URL, error) {
	var urls []URL

	err := r.db.WithContext(ctx).
		Find(&urls).Error

	return urls, err
}

func (r *URLRepository) PrintAll(ctx context.Context) error {
	urls, err := r.GetAll(ctx)
	if err != nil {
		return err
	}

	for _, u := range urls {
		fmt.Printf("Code: %s\n", u.Code)
		fmt.Printf("Original URL: %s\n", u.OriginalURL)
		fmt.Printf("Hits: %d\n", u.Hits)
		fmt.Println("------------")
	}

	return nil
}
