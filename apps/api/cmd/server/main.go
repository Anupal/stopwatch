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
)

func main() {
	_, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Load config
	cfg := config.Load()

	// setup logger
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: cfg.LogLevel,
	}))

	// Setup HTTP server
	mux := http.NewServeMux()

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
