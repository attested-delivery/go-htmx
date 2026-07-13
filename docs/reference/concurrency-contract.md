---
id: urn:mif:go-htmx:reference:concurrency-contract
type: semantic
created: 2026-07-12T00:00:00Z
namespace: go-htmx/docs/reference
tags: [reference, sqlite, concurrency, wal]
title: "Reference: the SQLite concurrency contract"
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
modified: '2026-07-13T16:48:20.112Z'
---

# Reference: the SQLite concurrency contract

The normative statement of `internal/platform/db`'s concurrency
guarantees (AD-1/AD-2). This is deliberately non-configurable — there
are no flags to loosen it.

## Pool shape

`db.Open(path)` returns a `*db.DB` holding **two** separate
`*sql.DB` pools over the same SQLite file:

| Pool | Accessor | `_txlock` | `SetMaxOpenConns` |
| --- | --- | --- | --- |
| Read | `ReadDB()` | `deferred` | unset (driver default, multiple connections) |
| Write | `WriteDB()` | `immediate` | `1` |

## Guarantees

- **All writes serialize through a single connection.** The write
  pool's `SetMaxOpenConns(1)` means at most one write transaction is
  ever in flight; a second concurrent write blocks (up to
  `busy_timeout`) rather than racing.
- **Every write transaction is `BEGIN IMMEDIATE`.** The write pool's
  DSN sets `_txlock=immediate`, so every transaction opened against it
  acquires SQLite's write lock at `BEGIN`, not at the first actual
  write statement — eliminating the classic SQLite "deferred
  transaction upgrades mid-transaction and hits SQLITE_BUSY" failure
  mode.
- **Reads never block on writes, and vice versa, beyond WAL's normal
  rules.** The read pool's `_txlock=deferred` and separate connection
  pool let reads proceed concurrently with the single in-flight write,
  per WAL mode's reader/writer isolation.
- **`busy_timeout(5000)` and `foreign_keys(ON)` are set on every
  connection**, read and write pools alike (via `_pragma` in the DSN
  both `dsn()` calls share).

## API surface

| Method | Behavior |
| --- | --- |
| `db.Open(path string) (*DB, error)` | Opens both pools. `path == ":memory:"` generates a random per-call shared-cache name (`mode=memory&cache=shared`) instead of a literal `:memory:` DSN — a bare `:memory:` gives each connection its own private database, which would break the read/write pool split entirely. |
| `(*DB).ReadDB() *sql.DB` | The read pool. Use for all `SELECT`s. |
| `(*DB).WriteDB() *sql.DB` | The write pool. Use for all `INSERT`/`UPDATE`/`DELETE`s and for `sqlc.New(...)` wrappers that issue them. |
| `(*DB).BeginWrite(ctx) (*sql.Tx, error)` | Convenience: `WriteDB().BeginTx(ctx, nil)` — every transaction from this pool is `BEGIN IMMEDIATE` by construction of the pool's DSN, not by a parameter to this call. |
| `(*DB).Close() error` | Closes both pools. |

## What this contract does not cover

- It does not make SQLite a distributed database. All writes still go
  through one process's single write connection — see [Deploy with
  Litestream](../how-to/deploy-with-litestream.md) and [Escalating
  beyond the defaults](../how-to/escalate-beyond-the-defaults.md) for
  multi-instance options (LiteFS, libSQL/Turso), which are a different
  architecture, not a config change to this contract.
- It does not retry a busy write beyond `busy_timeout`'s 5-second
  window; a caller that needs application-level retry on
  `SQLITE_BUSY` must implement it itself.
