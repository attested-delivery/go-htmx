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
	"strings"
	"syscall"
	"time"

	"github.com/attested-delivery/go-htmx/internal/notes"
	"github.com/attested-delivery/go-htmx/internal/platform/config"
	"github.com/attested-delivery/go-htmx/internal/platform/db"
	"github.com/attested-delivery/go-htmx/internal/platform/db/sqlc"
	"github.com/attested-delivery/go-htmx/internal/platform/httpserver"
)

func main() {
	// `go-htmx healthcheck` is a separate mode, not a flag: the
	// Dockerfile's HEALTHCHECK needs an executable to run, but the
	// distroless base image has no shell, curl, or wget to reach for —
	// only this binary itself. It probes the running server's own
	// GET /healthz over loopback and exits 0/1 accordingly; it does not
	// go through run() at all, so it opens no database connection of its
	// own.
	if len(os.Args) > 1 && os.Args[1] == "healthcheck" {
		if err := healthcheck(); err != nil {
			slog.Error("healthcheck failed", "error", err)
			os.Exit(1)
		}
		return
	}

	if err := run(); err != nil {
		slog.Error("fatal", "error", err)
		os.Exit(1)
	}
}

// healthcheck probes this same process's own GET /healthz over loopback.
// cfg.Addr (e.g. ":8080") is a bind address, not a dial target — an empty
// host means "all interfaces" to a listener but isn't valid to dial, so
// a bare port is rewritten to 127.0.0.1:<port>.
func healthcheck() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	addr := cfg.Addr
	if strings.HasPrefix(addr, ":") {
		addr = "127.0.0.1" + addr
	}

	client := http.Client{Timeout: 3 * time.Second}
	resp, err := client.Get("http://" + addr + "/healthz")
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return errors.New("healthz returned " + resp.Status)
	}
	return nil
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

	if err := db.Migrate(store.WriteDB()); err != nil {
		return err
	}

	mux := httpserver.NewMux(store.ReadDB())

	notesHandler := notes.NewHandler(
		sqlc.New(store.ReadDB()),
		sqlc.New(store.WriteDB()),
		notes.NewBroadcaster(),
		logger,
	)
	notesHandler.Register(mux)

	handler := httpserver.Wrap(mux, logger)
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
