// Command go-htmx is the template's single entrypoint. It wires config,
// the HTTP router, and the server together and contains no business logic
// itself (AD-6) — that lives under internal/.
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

	"github.com/attested-delivery/go-htmx/internal/platform/config"
	"github.com/attested-delivery/go-htmx/internal/platform/db"
	"github.com/attested-delivery/go-htmx/internal/platform/httpserver"
)

func main() {
	if err := run(); err != nil {
		slog.Error("fatal", "error", err)
		os.Exit(1)
	}
}

func run() error {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	store, err := db.Open(cfg.DBPath)
	if err != nil {
		return err
	}
	defer func() {
		if err := store.Close(); err != nil {
			logger.Error("db close failed", "error", err)
		}
	}()

	if err := db.Migrate(store.Write); err != nil {
		return err
	}

	// store threads through to feature handlers in Story #4, once there
	// are routes that actually read/write it.

	handler := httpserver.NewRouter(logger)
	srv := httpserver.New(cfg.Addr, handler)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	errCh := make(chan error, 1)
	go func() {
		logger.Info("listening", "addr", cfg.Addr, "env", cfg.Env)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	logger.Info("shutting down")
	return srv.Shutdown(shutdownCtx)
}
