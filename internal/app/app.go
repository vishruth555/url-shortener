package app

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"

	"urlshortener/internal/config"
	"urlshortener/internal/httpapi"
	"urlshortener/internal/middleware"
	"urlshortener/internal/repository/postgres"
	"urlshortener/internal/service"
)

func Run(ctx context.Context, cfg config.Config) error {
	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("connect database: %w", err)
	}
	defer pool.Close()

	dbCtx, cancel := context.WithTimeout(ctx, cfg.DBTimeout)
	defer cancel()
	if err := pool.Ping(dbCtx); err != nil {
		return fmt.Errorf("ping database: %w", err)
	}

	if err := postgres.RunSchema(dbCtx, pool, "db/schema.sql"); err != nil {
		return fmt.Errorf("run schema: %w", err)
	}

	repo := postgres.NewURLRepository(pool)
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
