package app

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"urlshortener/internal/config"
	"urlshortener/internal/httpapi"
	"urlshortener/internal/middleware"
	"urlshortener/internal/repository/postgres"
	"urlshortener/internal/repository/redis"
	"urlshortener/internal/service"

	"github.com/jackc/pgx/v5/pgxpool"
	r "github.com/redis/go-redis/v9"
)

func Run(ctx context.Context, cfg config.Config) error {
	//repo, err := getPostgresRepo(ctx, cfg)
	repo, err := getRedisRepo(ctx, cfg)
	if err != nil {
		return fmt.Errorf("Error connecting to Redis: %w", err)
	}

	shortener := service.NewShortener(repo, cfg.BaseURL, cfg.CodeLength, cfg.MaxGenerateRetries)
	api := httpapi.NewAPI(shortener)

	mux := http.NewServeMux()
	api.RegisterRoutes(mux)

	server := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      middleware.Logging(mux),
		ReadTimeout:  cfg.ServerReadTimeout,
		WriteTimeout: cfg.ServerWriteTimeout,
		IdleTimeout:  cfg.ServerIdleTimeout,
	}

	errCh := make(chan error, 1)
	go func() {
		log.Printf("server listening on %s", server.Addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
			return
		}
		errCh <- nil
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
		defer shutdownCancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("shutdown server: %w", err)
		}
		<-errCh
		return nil
	case err := <-errCh:
		return err
	}
}

func getRedisRepo(ctx context.Context, cfg config.Config) (service.URLRepository, error) {
	rdb := r.NewClient(&r.Options{Addr: cfg.RedisURL})
	err := rdb.Ping(ctx).Err()
	if err != nil {
		panic(err)
	}
	defer rdb.Close()
	fmt.Println("Connected to Redis!")

	return redis.NewCache(rdb), nil

}

func getPostgresRepo(ctx context.Context, cfg config.Config) (service.URLRepository, error) {
	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("connect database: %w", err)
	}
	defer pool.Close()

	dbCtx, cancel := context.WithTimeout(ctx, cfg.DBTimeout)
	defer cancel()
	if err := pool.Ping(dbCtx); err != nil {
		return nil, fmt.Errorf("ping database: %w", err)
	}

	if err := postgres.RunSchema(dbCtx, pool, "db/schema.sql"); err != nil {
		return nil, fmt.Errorf("run schema: %w", err)
	}
	fmt.Println("Connected to Postgres")
	return postgres.NewURLRepository(pool), nil
}
