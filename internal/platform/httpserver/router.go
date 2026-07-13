// Package httpserver provides the template's HTTP layer: a router built
// on Go 1.22+'s enhanced http.ServeMux (method + wildcard patterns) and a
// net/http-compatible middleware chain. chi is a documented escalation
// (drop-in, since chi middleware share this same func(http.Handler)
// http.Handler signature) for teams that outgrow ServeMux's routing
// features (AD-4).
package httpserver

import (
	"context"
	"database/sql"
	"log/slog"
	"net/http"
	"time"

	"github.com/attested-delivery/go-htmx/internal/web/assets"
)

// NewMux builds the application's route table with only the routes
// platform itself owns (static assets, health check). readDB is pinged
// by the health check below — pass store.ReadDB() from main.go. It
// deliberately does not own "/" or any application route — per
// internal/doc.go's import boundary, httpserver (platform) must never
// import a feature package, so feature packages register their own
// routes on the returned mux from the caller (main.go), which is free
// to import both. See Wrap for the middleware chain around whatever the
// caller ends up registering.
func NewMux(readDB *sql.DB) *http.ServeMux {
	mux := http.NewServeMux()
	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServerFS(assets.Static())))
	mux.HandleFunc("GET /healthz", handleHealthz(readDB))
	return mux
}

// handleHealthz pings the database on every call rather than just
// confirming the HTTP listener answers — a handler that always writes
// 200 regardless of application state isn't a health check, it can
// never catch the failure modes a check exists to catch (SQLite file
// corrupted, disk full, connection pool poisoned). PingContext is a
// real round-trip to the database, not a no-op on a healthy connection
// pool. Used by the Dockerfile's HEALTHCHECK (via `go-htmx healthcheck`,
// see cmd/go-htmx/main.go — distroless has no shell/curl to probe an
// HTTP endpoint with directly) and is equally usable as a Kubernetes
// liveness probe target if this image is ever deployed that way.
func handleHealthz(readDB *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()

		if err := readDB.PingContext(ctx); err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}

// Wrap applies the standard middleware chain around mux (or any
// http.Handler), after every route — platform's and every feature's —
// has been registered.
func Wrap(h http.Handler, logger *slog.Logger) http.Handler {
	return Chain(h,
		Recover(logger),
		Logging(logger),
		SecurityHeaders(),
	)
}
