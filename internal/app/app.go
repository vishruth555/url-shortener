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

	r "github.com/redis/go-redis/v9"

	"github.com/gin-gonic/gin"
)

func Run(ctx context.Context, cfg config.Config) error {
	// Initialize repositories
	// postgresRepo, err := getPostgresRepo(ctx, cfg)
	// if err != nil {
	// 	return err
	// }

	redisRepo, err := getRedisRepo(ctx, cfg)
	if err != nil {
		return err
	}

	// Initialize service
	shortener := service.NewShortener(
		redisRepo,
		cfg.BaseURL,
		cfg.CodeLength,
		cfg.MaxGenerateRetries,
	)

	// Initialize API
	api := httpapi.NewAPI(shortener)

	// Setup Gin
	router := gin.New()

	// Attach middleware
	router.Use(
		gin.Recovery(),
		middleware.GinLogging(), // we'll define this below
	)

	// Register routes
	api.RegisterRoutes(router)

	// Create HTTP server (still needed for graceful shutdown + timeouts)
	server := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      router,
		ReadTimeout:  cfg.ServerReadTimeout,
		WriteTimeout: cfg.ServerWriteTimeout,
		IdleTimeout:  cfg.ServerIdleTimeout,
	}

	errCh := make(chan error, 1)

	go func() {
		log.Printf("server listening on %s", server.Addr)

		if err := server.ListenAndServe(); err != nil &&
			err != http.ErrServerClosed {

			errCh <- err
			return
		}

		errCh <- nil
	}()

	select {
	case <-ctx.Done():
		log.Println("shutdown signal received")

		shutdownCtx, cancel := context.WithTimeout(
			context.Background(),
			cfg.ShutdownTimeout,
		)
		defer cancel()

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
	rdb := r.NewClient(&r.Options{
		Addr: cfg.RedisURL,
	})

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("connect redis: %w", err)
	}

	fmt.Println("Connected to Redis!")

	return redis.NewCache(rdb), nil
}

func getPostgresRepo(ctx context.Context, cfg config.Config) (service.URLRepository, error) {
	db, err := postgres.NewGormDB(cfg.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("connect database: %w", err)
	}

	fmt.Println("Connected to Postgres")

	repo := postgres.NewURLRepository(db)

	if err := repo.PrintAll(ctx); err != nil {
		return nil, err
	}

	return repo, nil
}
