# AGENTS.md

Conventions for any coding agent (or human) working in this repo. This
targets **mechanical work** — scaffolding, codegen, wiring a new feature
into the existing patterns — not autonomous architecture decisions; the
architectural decisions (AD-1 … AD-9) are already made and are described
in the seed spec this template was built from, not re-litigated here.

Per this repo's own package-boundary rules (below), an agent that reads
only this file should be able to run every command it needs without
discovering them from CI config — that's the actual acceptance bar this
file is held to (NFR-7), so keep it accurate as commands change.

## Toolchain

Go (current Active LTS — check `go.mod`'s `toolchain` line for the exact
pinned version), [`just`](https://github.com/casey/just), and network
access for `go install`/`go run` to fetch pinned tool versions on first
use (`templ`, `sqlc`, `golangci-lint`, `goose`). No other tools required
— see `just smoke-init`, which proves a fresh copy needs nothing beyond
this.

## Commands

Run `just` with no arguments to list every recipe with its description.
The ones that matter day to day:

| Command | What it does |
| --- | --- |
| `just generate` | Regenerate `templ` view code and `sqlc` query code. Every other recipe below that touches Go code depends on this — never run `go build`/`go test` directly without it first if you've touched a `.templ` or `internal/platform/db/query/*.sql` file. |
| `just build` | `generate`, then build the single self-contained binary to `bin/`. |
| `just run` | Dev server against on-disk static assets (`-tags dev`) instead of the embedded build — see `internal/web/assets/assets_dev.go`. |
| `just test` | `generate`, then `go test -race -cover ./...` — the full test triad (AD-8): `httptest` handler tests, in-memory SQLite (`db.Open(":memory:")`), and golden-file snapshots, all in the same run. |
| `just test-golden-update` | Regenerate golden fixtures under `internal/notes/testdata/*.golden` after a deliberate rendered-output change. Scoped to `internal/notes` specifically — `goldie` registers `-update` per test binary, so running this against `./...` fails on every package that doesn't import `goldie`. **Review the resulting diff before committing** — an unreviewed `-update` run silently bakes a real regression in as the new "expected" output. |
| `just lint` | `golangci-lint run ./...` (config: `.golangci.yml`). |
| `just check` | `build` + `lint` + `test` — the full CI sequence, locally. Run this before pushing; a green `just check` predicts a green `ci.yml` `build-lint-test` job. |
| `just migrate-new <name>` | Create a new zero-padded goose migration file under `internal/platform/db/migrations/`. Never hand-name a migration file — the numeric prefix must stay zero-padded (see `internal/platform/db/migrations_test.go`'s `TestMigrationFilenamesArePadded`) or migrations silently apply out of order. |
| `just init <name> <module>` | One-time, run right after copying this template — rewrites the template's own identity into the new project's. See `tools/init/main.go`'s doc comment for exactly what it touches. |
| `just smoke-init` | The template's real acceptance test: copies the tree, runs `init` with a throwaway identity, builds + tests the copy, grep-gates for leftover identity strings. |
| `just fmt` | `gofmt` + `templ fmt`. |

## Layout & package boundaries

Everything under `internal/` is one of three kinds of package (this is
also stated, more tersely, in `internal/doc.go` — that's the
compiler-adjacent copy, this is the prose one):

- **`internal/platform/*`** — infrastructure: config loading
  (`internal/platform/config`), the SQLite access layer and its
  WAL/`BEGIN IMMEDIATE`/single-writer concurrency contract
  (`internal/platform/db`, AD-1/AD-2 — read that package's doc comment
  before touching it, the contract is deliberately non-configurable),
  and the HTTP router/middleware chain (`internal/platform/httpserver`).
- **`internal/web/*`** — shared UI: the `templ` layout shell
  (`internal/web/templates`) and embedded static assets
  (`internal/web/assets`, with the AD-9 dev/prod build-tag split).
  Treated as a platform package for import-boundary purposes.
- **`internal/<feature>/*`** — one self-contained vertical slice per
  product feature (`internal/notes` is the worked example — Story #4).
  A feature package may import `internal/platform/*` and
  `internal/web/*`. It must not import another
  `internal/<feature>/*` package directly; cross-feature composition
  happens in `cmd/<name>/main.go` at wiring time, not between features.

**The rule that actually matters**: `internal/platform/*` and
`internal/web/*` must never import an `internal/<feature>/*` package —
infrastructure can't depend on a specific feature. This isn't just
documented, it's enforced: `.golangci.yml`'s `platform-no-feature-imports`
depguard rule denies anything outside an explicit allow-list from files
under those two paths, so a violation fails `just lint`, not code review.
When you add a new feature package, its dependencies (a database driver,
a third-party UI library, etc.) need adding to that allow-list too if
`internal/platform/*` code ends up importing them — see the rule's own
comment in `.golangci.yml` for why it's an allow-list, not a deny-list.

## Compile-time interface-satisfaction idiom (NFR-8)

Where a concrete type is meant to satisfy an interface, assert it at
compile time with `var _ Interface = (*Struct)(nil)` next to the type,
rather than relying on a caller somewhere to notice a mismatch at usage
time (or, worse, not noticing until a test fails). A worked example
already in this codebase: `sqlc generate` (with `emit_interface: true`
in `sqlc.yaml`) emits exactly this pattern in
`internal/platform/db/sqlc/querier.go`:

```go
type Querier interface {
    CountNotes(ctx context.Context) (int64, error)
    CreateNote(ctx context.Context, body string) (Note, error)
    // ...
}

var _ Querier = (*Queries)(nil)
```

Use the same idiom in hand-written code when a type is genuinely meant
to satisfy an interface someone else depends on — not preemptively on
every struct. An assertion with no real interface behind it is noise,
not documentation.

## Adding a new feature

Use the checked-in `add-feature-package` skill (`.claude/skills/
add-feature-package/SKILL.md`) rather than hand-rolling the same
boilerplate `internal/notes` already demonstrates — it scaffolds the
handler, templ view, sqlc query stub, and test-triad wiring for a new
`internal/<feature>/*` package in one pass, using `internal/notes` as
the canonical pattern. The skill's own instructions reference the
commands in this file rather than duplicating them; if a command here
changes, the skill doesn't need a matching edit.
