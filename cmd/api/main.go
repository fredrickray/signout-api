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

	"github.com/signout/signout-api/internal/config"
	"github.com/signout/signout-api/internal/server"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

	cfg, err := config.Load()
	if err != nil {
		logger.Error("config error", slog.String("err", err.Error()))
		os.Exit(1)
	}

	srv, err := server.New(cfg, logger)
	if err != nil {
		logger.Error("server init failed", slog.String("err", err.Error()))
		os.Exit(1)
	}

	errCh := make(chan error, 1)
	go func() {
		errCh <- srv.Start()
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-errCh:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("server stopped unexpectedly", slog.String("err", err.Error()))
			os.Exit(1)
		}
	case sig := <-stop:
		logger.Info("shutdown signal received", slog.String("signal", sig.String()))
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		if err := srv.Shutdown(ctx); err != nil {
			logger.Error("graceful shutdown failed", slog.String("err", err.Error()))
			os.Exit(1)
		}
	}
}
