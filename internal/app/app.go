package app

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"urlshortener/internal/config"
	"urlshortener/internal/httpapi"
	"urlshortener/internal/middleware"
	_ "urlshortener/internal/repository/postgres"
	"urlshortener/internal/repository/redis"
	"urlshortener/internal/service"

	_ "github.com/jackc/pgx/v5/pgxpool"
	r "github.com/redis/go-redis/v9"
)

func Run(ctx context.Context, cfg config.Config) error {
	// pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	// if err != nil {
	// 	return fmt.Errorf("connect database: %w", err)
	// }
	// defer pool.Close()

	// dbCtx, cancel := context.WithTimeout(ctx, cfg.DBTimeout)
	// defer cancel()
	// if err := pool.Ping(dbCtx); err != nil {
	// 	return fmt.Errorf("ping database: %w", err)
	// }

	// if err := postgres.RunSchema(dbCtx, pool, "db/schema.sql"); err != nil {
	// 	return fmt.Errorf("run schema: %w", err)
	// }
	// fmt.Println("Connected to Postgres")
	// repo := postgres.NewURLRepository(pool)

	rdb := r.NewClient(&r.Options{Addr: cfg.RedisURL})
	err := rdb.Ping(ctx).Err()
	if err != nil {
		panic(err)
	}
	defer rdb.Close()
	repo := redis.NewCache(rdb)

	fmt.Println("Connected to Redis!")

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
