package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Anupal/stopwatch/internal/config"
	"github.com/Anupal/stopwatch/internal/handlers"
	"github.com/Anupal/stopwatch/internal/repositories"
	"github.com/Anupal/stopwatch/internal/services"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Load config
	cfg := config.Load()

	// setup logger
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: cfg.LogLevel,
	}))

	// Setup database connection pool
	pool, err := pgxpool.New(ctx, cfg.DatabaseURL())
	if err != nil {
		logger.Error("Unable to connect to database", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		logger.Error("Database ping failed", "error", err)
		os.Exit(1)
	}
	logger.Info("Successfully connected to database server")

	// Initialize repos, services and handlers
	stopRepo := repositories.NewStopRepository(pool, logger)
	stopService := services.NewStopService(stopRepo, logger)
	stopHandler := handlers.NewStopHandler(stopService, logger)

	// Setup HTTP server
	mux := http.NewServeMux()
	mux.HandleFunc("GET /stops/{stopID}", stopHandler.GetStopByID)

	server := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: mux,
	}

	// Handle Shutdown
	shutdownChan := make(chan os.Signal, 1)
	signal.Notify(shutdownChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		logger.Info("Server listening on port", "port", cfg.Port)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("Server failed", "error", err)
		}
	}()

	<-shutdownChan
	logger.Info("Shutting down server gracefully")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("Forced shutdown error", "error", err)
	}

	logger.Info("Server stopped cleanly")
}
