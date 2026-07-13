---
id: urn:mif:go-htmx:reference:project-layout
type: semantic
created: 2026-07-12T00:00:00Z
namespace: go-htmx/docs/reference
tags: [reference, layout, packages]
title: "Reference: project layout"
temporal:
  validFrom: 2026-07-12T00:00:00Z
relationships:
  - type: relates-to
    target: docs/explanation/architecture.md
  - type: relates-to
    target: docs/how-to/style-with-tailwind.md
provenance:
  '@type': Provenance
  agent: claude-code/claude-sonnet-5
  wasGeneratedBy:
    '@id': urn:mif:activity:claude-code-session:91f08ccf-1fbc-4d5c-843d-9ff6d4050ce8
    '@type': prov:Activity
  trustLevel: user_stated
  agentVersion: 2.1.207
modified: '2026-07-13T16:46:44.916Z'
---

# Reference: project layout

Every top-level directory in this template and its contract.

| Path | Contract |
| --- | --- |
| `cmd/<name>/main.go` | The single entrypoint. Wires config, the HTTP router, and the server; contains no business logic itself. |
| `internal/platform/*` | Infrastructure: `config` (env-var loading), `db` (SQLite access layer, WAL/`BEGIN IMMEDIATE`/single-writer contract), `httpserver` (router/middleware chain). Must not import `internal/<feature>/*`. |
| `internal/web/*` | Shared UI: `templates` (the `templ` layout shell), `assets` (embedded static assets, dev/prod build-tag split). Treated as a platform package for import-boundary purposes. |
| `internal/web/assets/tailwind/input.css` | Tailwind CSS source (`@import`, `@source` globs, `@theme` tokens) — see [Style with Tailwind](../how-to/style-with-tailwind.md). |
| `internal/web/assets/static/css/app.css` | Tailwind's compiled output. Gitignored — a build artifact, not a source file; regenerate with `just tailwind` (or any recipe that depends on `generate`). |
| `internal/web/assets/static/js/app.js` | First-party (not vendored — see `static/js/VENDORED.md` for what is) client-side glue script, e.g. the notes form's reset-after-submit behavior. Loaded with `defer` in `layout.templ`. |
| `internal/<feature>/*` | One self-contained vertical slice per product feature (`internal/notes` is the worked example). May import `internal/platform/*` and `internal/web/*`. Must not import another `internal/<feature>/*` package directly. |
| `internal/doc.go` | The compiler-adjacent statement of the layout rule above. |
| `internal/platform/db/migrations/` | Zero-padded goose SQL migration files, embedded via `go:embed`. |
| `internal/platform/db/query/*.sql` | `sqlc` query source, one file per feature's queries. |
| `internal/platform/db/sqlc/` | `sqlc`-generated Go code. Gitignored — regenerate with `just generate`. |
| `tools/init/` | The `just init` identity-rewrite program. Never invoked by the built application; not part of `just build`'s output. |
| `scripts/smoke-init.sh` | The template's real acceptance test for `just init`, run via `just smoke-init`. |
| `.claude/skills/add-feature-package/` | The checked-in scaffolding skill for new feature packages. |
| `docs/` | This Diátaxis documentation set. |
| `AGENTS.md` | Conventions for any coding agent (or human) working in this repo. |
| `litestream.yml` | Litestream sidecar config for SQLite durability (AD-5). |

## Enforcement

`internal/platform/*` and `internal/web/*` must never import an
`internal/<feature>/*` package. This is enforced by
`.golangci.yml`'s `platform-no-feature-imports` depguard rule, which
denies anything outside an explicit allow-list from files under those
two path globs — a violation fails `just lint`, not just code review.
