---
id: urn:mif:go-htmx:how-to:escalate-beyond-the-defaults
type: procedural
created: 2026-07-12T00:00:00Z
namespace: go-htmx/docs/how-to
tags: [how-to, escalation, chi, sqlite, litefs, libsql, html-template]
title: "How to escalate beyond this template's defaults"
relationships:
  - type: relates-to
    target: docs/explanation/architecture.md
provenance:
  '@type': Provenance
  agent: claude-code/claude-sonnet-5
  wasGeneratedBy:
    '@id': urn:mif:activity:claude-code-session:91f08ccf-1fbc-4d5c-843d-9ff6d4050ce8
    '@type': prov:Activity
  trustLevel: user_stated
  agentVersion: 2.1.207
modified: '2026-07-13T02:19:30.041Z'
---

# How to escalate beyond this template's defaults

This template picks one default per concern deliberately (see
[Architecture rationale](../explanation/architecture.md) for why). When
you genuinely outgrow a default, here is exactly what to change for
each escalation path. Each is a real code change, not a config flag —
that's intentional, so outgrowing a default is a conscious decision,
not an accidental one.

## Swap `net/http.ServeMux` for `chi`

When you need routing features ServeMux's method + wildcard patterns
don't cover (regex constraints, richer sub-router composition).

1. Add `github.com/go-chi/chi/v5` to `go.mod`.
2. In `internal/platform/httpserver/router.go`, replace
   `http.NewServeMux()` with `chi.NewRouter()`. `chi`'s middleware
   share this package's `func(http.Handler) http.Handler` signature
   (`httpserver.Middleware`), so `Chain`/`Wrap` and every existing
   `Middleware` (`Logging`, `Recover`) keep working unchanged.
3. Update `.golangci.yml`'s `platform-no-feature-imports` allow-list to
   include `github.com/go-chi/chi/v5`.
4. Update every feature package's `Register(mux *http.ServeMux)`
   signature to `Register(r chi.Router)` (or keep `http.ServeMux` if
   you only need chi in the platform layer — chi routers satisfy
   `http.Handler`, so this can be adopted incrementally).

## Swap `modernc.org/sqlite` for `mattn/go-sqlite3`

When you need a SQLite feature or extension only the CGO-based driver
exposes, and CGO in your build/deploy pipeline is acceptable.

1. Replace the import and driver name in
   `internal/platform/db/db.go`: `modernc.org/sqlite`'s driver name is
   `"sqlite"`; `mattn/go-sqlite3`'s is `"sqlite3"`.
2. Re-verify the DSN parameters `dsn()` builds (`_txlock`, `_pragma`)
   against `mattn/go-sqlite3`'s actual query-parameter names — they are
   not identical to `modernc.org/sqlite`'s.
3. `CGO_ENABLED=1` becomes required for `just build`; update CI's build
   step and any cross-compilation matrix accordingly — this is the
   direct cost of this swap, since AD-1 picked `modernc.org/sqlite`
   specifically to stay CGO-free and cross-compile as a single static
   binary.
4. Re-run the full `internal/platform/db` test suite
   (`internal/platform/db/db_test.go`) unmodified against the new
   driver to confirm the WAL/`BEGIN IMMEDIATE`/single-writer contract
   still holds — the contract is a behavior of the DSN parameters and
   pool configuration, not the driver package itself, but it must be
   re-verified, not assumed.

## Swap Litestream for LiteFS or libSQL/Turso

See [Deploy with Litestream](deploy-with-litestream.md)'s "Outgrowing
Litestream" section for when each applies.

- **LiteFS**: no application code change — it's a FUSE filesystem the
  app's normal file-path access sits on top of. Provision a
  FUSE-capable host (or Fly.io) and a consul-compatible lease backend,
  then point `GO_HTMX_DB_PATH` at the LiteFS-managed path.
- **libSQL/Turso**: a real code change. Replace the
  `modernc.org/sqlite` import and driver registration in
  `internal/platform/db/db.go` with libSQL's driver, and re-verify the
  DSN parameters and concurrency contract exactly as in the
  `mattn/go-sqlite3` swap above — libSQL's driver has its own DSN
  vocabulary.

## Swap `templ` for `html/template`

When you need to drop the `templ` build step entirely (e.g. a
constrained build environment with no `go generate`-style codegen
step).

1. Rewrite each `.templ` file's templates as `html/template` templates
   (`.tmpl` files or embedded strings), preserving the same rendering
   call sites (`Page(...).Render(ctx, w)` becomes
   `tmpl.ExecuteTemplate(w, "page", data)` or equivalent).
2. Remove the `templ generate` step from `justfile`'s `generate` recipe
   and from `ci.yml`.
3. `html/template` auto-escapes by context the same way `templ` does,
   so the XSS-safety property is preserved — but you lose `templ`'s
   compile-time template-syntax checking; budget for runtime template
   parse errors that `templ generate` would have caught at build time.
