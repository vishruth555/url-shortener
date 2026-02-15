package redis

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

type Cache struct {
	rdb *redis.Client
}

func NewCache(rdb *redis.Client) *Cache {
	return &Cache{rdb: rdb}
}

func (r *Cache) Create(ctx context.Context, code string, originalURL string) error {
	key := "url_code:" + code
	//r.rdb.Expire(ctx, key, 10*time.Second).Err()
	return r.rdb.HSet(ctx, key,
		"url", originalURL,
		"hits", 0,
	).Err()
}

func (r *Cache) GetOriginalByCode(ctx context.Context, code string) (string, error) {
	key := "url_code:" + code
	r.IncrementHits(ctx, code, key)
	hits, _ := r.GetHits(ctx, code, key)
	fmt.Printf("Code: %s, Hits: %d\n", code, hits)
	return r.rdb.HGet(ctx, key, "url").Result()
}

func (r *Cache) IncrementHits(ctx context.Context, code string, key string) error {
	return r.rdb.HIncrBy(ctx, key, "hits", 1).Err()
}

func (r *Cache) GetHits(ctx context.Context, code string, key string) (int64, error) {
	return r.rdb.HGet(ctx, key, "hits").Int64()
}
