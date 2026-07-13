---
id: urn:mif:go-htmx:reference:justfile-commands
type: semantic
created: 2026-07-12T00:00:00Z
namespace: go-htmx/docs/reference
tags: [reference, justfile, commands]
title: "Reference: justfile commands"
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
modified: '2026-07-13T02:19:43.765Z'
---

# Reference: justfile commands

Every recipe defined in this repo's `justfile`. Run `just` with no
arguments to list them from the file itself; this table is the same
information as narrative reference.

| Command | Depends on | What it does |
| --- | --- | --- |
| `just` (no args) | — | Lists every recipe with its description. |
| `just generate` | — | Runs `templ generate` then `sqlc generate`. |
| `just build` | `generate` | Builds the single self-contained binary to `bin/<name>`. |
| `just run` | `generate` | Starts the dev server (`-tags dev`, `GO_HTMX_ENV=dev`) against on-disk static assets. |
| `just test` | `generate` | `go test -race -cover ./...`. |
| `just test-golden-update` | `generate` | `go test ./internal/notes/... -run TestGolden -update`. Scoped to `internal/notes` — `goldie` registers `-update` per test binary. |
| `just lint` | `generate` | `golangci-lint run ./...` (config: `.golangci.yml`). |
| `just check` | `build`, `lint`, `test` | The full CI sequence, locally. |
| `just migrate-new <name>` | — | Creates a new zero-padded goose migration file under `internal/platform/db/migrations/`. |
| `just init <name> <module>` | — | Rewrites the template's own identity into a new project's. One-time, run once right after copying. |
| `just smoke-init` | — | Copies the tree, runs `init` with a throwaway identity, builds + tests the copy, grep-gates for leftover identity strings. |
| `just fmt` | — | `gofmt -l -w .` then `templ fmt .`. |
| `just clean` | — | `rm -rf bin`. |

## Exit codes

Every recipe is a thin wrapper around the underlying tool it calls
(`go`, `templ`, `sqlc`, `golangci-lint`, `goose`) and exits with that
tool's own exit code — `just` itself adds no additional exit-code
semantics.

## Files read

- `justfile` — the recipe definitions themselves.
- `.golangci.yml` — read by `just lint`/`just check`.
- `sqlc.yaml` — read by `just generate`'s `sqlc generate` step.
