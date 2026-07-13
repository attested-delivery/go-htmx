// Package app wires the application's platform and feature packages
// together into a single http.Handler. It exists specifically because
// internal/platform/httpserver may never import internal/notes (or any
// other internal/<feature>/* package — see internal/doc.go and
// .golangci.yml's platform-no-feature-imports depguard rule, AD-6); this
// package is neither platform nor a feature, so it's free to depend on
// both, the same way cmd/go-htmx/main.go already does. Extracting the
// wiring here (rather than leaving it inlined in main.go) gives the E2E
// test harness (e2e/internal/testapp) a single source of truth for how
// the app is assembled, instead of a hand-copied wiring sequence that
// could silently drift from the real thing.
package app

import (
	"database/sql"
	"log/slog"
	"net/http"

	"github.com/attested-delivery/go-htmx/internal/notes"
	"github.com/attested-delivery/go-htmx/internal/platform/db/sqlc"
	"github.com/attested-delivery/go-htmx/internal/platform/httpserver"
)

// New assembles the full application handler: platform routes
// (static assets, /healthz), every feature's routes, and the standard
// middleware chain — everything main.go's run() needs except the
// *http.Server itself (httpserver.New) and the listener lifecycle,
// which differ between a real process and a test harness.
func New(readDB, writeDB *sql.DB, logger *slog.Logger) http.Handler {
	mux := httpserver.NewMux(readDB)

	notesHandler := notes.NewHandler(
		sqlc.New(readDB),
		sqlc.New(writeDB),
		notes.NewBroadcaster(),
		logger,
	)
	notesHandler.Register(mux)

	return httpserver.Wrap(mux, logger)
}
