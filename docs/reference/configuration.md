---
id: urn:mif:go-htmx:reference:configuration
type: semantic
created: 2026-07-12T00:00:00Z
namespace: go-htmx/docs/reference
tags: [reference, configuration, environment-variables]
title: "Reference: configuration surface"
temporal:
  validFrom: 2026-07-12T00:00:00Z
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
modified: '2026-07-13T16:48:28.215Z'
---

# Reference: configuration surface

All runtime configuration is read from environment variables by
`internal/platform/config.Load()` — no config file, so the single
embedded binary needs nothing else to start. `just init` rewrites the
`GO_HTMX_` prefix below to match your app's name.

## Environment variables

| Variable | Type | Default | Constraints | Description |
| --- | --- | --- | --- | --- |
| `GO_HTMX_ADDR` | string | `:8080` | none | Address the HTTP server listens on. |
| `GO_HTMX_ENV` | string | `prod` | must be `dev` or `prod` | Informational only (logged at startup as `env`). Does not select between dev/prod asset serving — that split is a compile-time `-tags dev` decision (see `internal/web/assets`), independent of this value. |
| `GO_HTMX_DB_PATH` | string | `go-htmx.db` | none | SQLite database file path. Also read by `litestream.yml`. |

## Litestream variables (`litestream.yml`)

| Variable | Type | Default | Constraints | Description |
| --- | --- | --- | --- | --- |
| `LITESTREAM_S3_BUCKET` | string | none | required for replication | S3-compatible bucket for the replica. |
| `LITESTREAM_S3_PATH` | string | none | required for replication | Path within the bucket. |
| `LITESTREAM_S3_REGION` | string | none | required for replication | Bucket region. |
| `LITESTREAM_S3_ENDPOINT` | string | none | required for replication | S3-compatible endpoint URL. |

With any `LITESTREAM_*` variable unset, `litestream replicate -config
litestream.yml` fails on the unresolved environment substitution and
exits with a config error — it does not silently run unreplicated.

## Command-line flags

None. All configuration is environment-variable-only.

## Error behavior

`config.Load()` returns an error (causing the process to exit non-zero
via `main`'s `run()`) only if `GO_HTMX_ENV` is set to a value other than
`dev` or `prod`. Every other variable falls back to its default if
unset or empty.
