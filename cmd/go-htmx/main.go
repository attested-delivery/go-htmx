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

	// DB pool wiring (WAL + BEGIN IMMEDIATE + single-writer connection,
	// AD-1/AD-2) lands here in Story #3 (attested-delivery/go-htmx#15),
	// threaded through to feature handlers alongside the router below.

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
