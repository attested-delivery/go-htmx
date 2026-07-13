//go:build e2e

// Package testapp starts a real instance of the application for E2E
// tests to drive with a browser. It calls the same internal/app.New
// wiring cmd/go-htmx/main.go's run() calls, over httptest.NewServer
// (real TCP socket, real headers — indistinguishable from a browser's
// perspective from a subprocess-exec'd binary, without the CI complexity
// of a build/port-allocation/health-check-polling/teardown cycle a
// subprocess needs) — so this harness can never silently drift from how
// the real app is actually assembled. The whole e2e/ tree is behind the
// "e2e" build tag (not just _test.go files) so `go build ./...`/`go vet
// ./...`/`just check` never touches it — this package's only purpose is
// browser-driven testing, which needs playwright-go's installed browser
// binaries (`just e2e-install`) to actually run.
package testapp

import (
	"log/slog"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/attested-delivery/go-htmx/internal/app"
	"github.com/attested-delivery/go-htmx/internal/platform/db"
)

// New starts a fresh application instance backed by an isolated,
// temp-file SQLite database (not :memory: — a real file + WAL journal
// is closer to how the app is actually deployed, which matters more for
// E2E tests than for the existing httptest-recorder-based unit tests).
// The returned server and its database are closed automatically via
// t.Cleanup; callers don't need to clean up themselves.
func New(t *testing.T) *httptest.Server {
	t.Helper()

	dbPath := filepath.Join(t.TempDir(), "e2e.db")
	store, err := db.Open(dbPath)
	if err != nil {
		t.Fatalf("db.Open: %v", err)
	}
	t.Cleanup(func() {
		if err := store.Close(); err != nil {
			t.Errorf("store.Close: %v", err)
		}
	})

	if err := db.Migrate(store.WriteDB()); err != nil {
		t.Fatalf("db.Migrate: %v", err)
	}

	logger := slog.New(slog.NewTextHandler(testWriter{t}, nil))
	handler := app.New(store.ReadDB(), store.WriteDB(), logger)

	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	return srv
}

// testWriter routes the app's own log output through t.Log, so it shows
// up attributed to the test that produced it instead of polluting stdout
// unconditionally (go test only prints t.Log output for failed tests, or
// under -v — matching normal Go test hygiene).
type testWriter struct{ t *testing.T }

func (w testWriter) Write(p []byte) (int, error) {
	w.t.Log(string(p))
	return len(p), nil
}
