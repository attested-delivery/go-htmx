---
name: add-feature-package
description: Scaffold a new internal/<feature>/* vertical-slice package (handler, templ view, test) following this template's internal/notes example, and wire it into cmd/<app>/main.go. Use when the user asks to add a new feature, a new page, or a new internal package to this repo.
---

# Add a feature package

Scaffolds a new `internal/<feature>/*` package — the same shape as
`internal/notes` (this template's worked example, Story #4) — and wires
it into the app. Targets **mechanical scaffolding**: the boilerplate a
new feature always needs, not the feature's actual logic. See
`AGENTS.md` for the commands and package-boundary rules this skill
assumes; this skill doesn't repeat them.

## Before starting

Confirm with the user (if not already given):
1. The feature name — a short, lowercase, valid Go package name (e.g.
   `todos`, `comments`). No hyphens or underscores; `internal/<feature>/*`
   becomes a Go import path segment.
2. Whether the feature needs its own database table. If yes, tell the
   user they'll need to add a migration (`just migrate-new <name>`) and
   `sqlc` queries (`internal/platform/db/query/<feature>.sql`)
   themselves first, following `internal/platform/db/query/notes.sql`
   as the pattern — this skill scaffolds the HTTP/view layer, not a
   guessed-at schema for data it doesn't know the shape of.

## Steps

1. **Create `internal/<feature>/handler.go`**:

   ```go
   // Package <feature> is a <one-line description — ask the user, or
   // leave as TODO> feature package.
   package <feature>

   import (
       "log/slog"
       "net/http"
   )

   // Handler holds this feature's dependencies. Add read/write query
   // wrappers here once this feature has its own database queries — see
   // internal/notes/handler.go for the pattern (separate read/write
   // *sqlc.Queries fields, matching internal/platform/db's Read/Write
   // pool split).
   type Handler struct {
       logger *slog.Logger
   }

   // NewHandler builds a Handler.
   func NewHandler(logger *slog.Logger) *Handler {
       return &Handler{logger: logger}
   }

   // Register mounts this feature's routes on mux. Called from
   // cmd/<app>/main.go, which is free to import both this package and
   // internal/platform/httpserver — httpserver itself never imports a
   // feature package (see AGENTS.md's package-boundary rule).
   func (h *Handler) Register(mux *http.ServeMux) {
       mux.HandleFunc("GET /<feature>", h.handlePage)
   }

   func (h *Handler) handlePage(w http.ResponseWriter, r *http.Request) {
       w.Header().Set("Content-Type", "text/html; charset=utf-8")
       if err := Page().Render(r.Context(), w); err != nil {
           h.logger.Error("render page failed", "error", err)
           http.Error(w, "internal server error", http.StatusInternalServerError)
       }
   }
   ```

2. **Create `internal/<feature>/views.templ`**:

   ```templ
   package <feature>

   import "github.com/<module>/internal/web/templates"

   templ Page() {
       @templates.Layout("<Feature>") {
           <h1><Feature></h1>
           <p>TODO: replace this placeholder with your feature's UI.</p>
       }
   }
   ```

   Replace `github.com/<module>` with this repo's actual module path
   (from `go.mod`'s `module` line) and `<Feature>` with a
   human-readable capitalized form of the feature name.

3. **Create `internal/<feature>/handler_test.go`** (httptest tier, AD-8):

   ```go
   package <feature>

   import (
       "io"
       "log/slog"
       "net/http"
       "net/http/httptest"
       "strings"
       "testing"
   )

   func TestHandlePage(t *testing.T) {
       logger := slog.New(slog.NewTextHandler(io.Discard, nil))
       h := NewHandler(logger)
       mux := http.NewServeMux()
       h.Register(mux)

       req := httptest.NewRequest(http.MethodGet, "/<feature>", nil)
       rec := httptest.NewRecorder()
       mux.ServeHTTP(rec, req)

       if rec.Code != http.StatusOK {
           t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
       }
       if !strings.Contains(rec.Body.String(), "<Feature>") {
           t.Errorf("response body missing feature heading")
       }
   }
   ```

4. **Wire it into `cmd/<app>/main.go`**: add the import and, alongside
   the existing feature handler's `Register(mux)` call, add:

   ```go
   <feature>Handler := <feature>.NewHandler(logger)
   <feature>Handler.Register(mux)
   ```

   If the feature has its own database queries (step 0 above), follow
   `internal/notes`' pattern instead: pass `sqlc.New(store.ReadDB())`/
   `sqlc.New(store.WriteDB())` into `NewHandler` alongside `logger`.

5. **Verify**: run `just generate && just build && just test` (or
   `just check` for the full lint+test pass). The scaffolded package
   must compile and its test must pass *unmodified* — that's this
   skill's own acceptance bar (Task #30). If anything fails, the
   scaffold itself has a bug; fix the skill, not just this instance.

## Adding real functionality after scaffolding

Once the skeleton above passes `just test`, replace the `Page()`
placeholder body and add real routes/handlers the same way
`internal/notes/handler.go` does: `POST` handlers for
mutations, `GET` handlers for reads, and — if the feature needs live
updates — `internal/notes/broadcaster.go` and `internal/notes/handler.go`'s
`handleStream`/`Sync` pattern for Server-Sent Events with the
render-gap fix (sync-on-connect before entering the broadcast loop).
Don't copy that machinery preemptively; only add it if the feature
actually needs real-time updates.
