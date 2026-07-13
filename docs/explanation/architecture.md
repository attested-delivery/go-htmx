---
id: urn:mif:go-htmx:explanation:architecture
type: semantic
created: 2026-07-12T00:00:00Z
namespace: go-htmx/docs/explanation
tags: [explanation, architecture, decisions]
title: "Understanding this template's architecture"
relationships:
  - type: relates-to
    target: docs/tutorial/getting-started.md
  - type: relates-to
    target: docs/reference/concurrency-contract.md
provenance:
  '@type': Provenance
  agent: claude-code/claude-sonnet-5
  wasGeneratedBy:
    '@id': urn:mif:activity:claude-code-session:91f08ccf-1fbc-4d5c-843d-9ff6d4050ce8
    '@type': prov:Activity
  trustLevel: user_stated
  agentVersion: 2.1.207
modified: '2026-07-13T02:21:11.637Z'
---

# Understanding this template's architecture

Every default in this template was a choice, not an accident. This
document walks through why, so you can judge for yourself when a
default has stopped fitting your situation — and reach for the matching
escalation guide instead of fighting the default in place.

## Why `modernc.org/sqlite` instead of `mattn/go-sqlite3` (AD-1)

The two dominant Go SQLite drivers make an opposite trade at the same
point: `mattn/go-sqlite3` wraps the real C SQLite via CGO, which means
a CGO toolchain at build time and no trivial static cross-compilation.
`modernc.org/sqlite` is a pure-Go transpilation of SQLite, so `go
build` alone produces a single static binary that cross-compiles the
normal Go way — no C toolchain, no CGO cross-compilation matrix. For a
template whose stated goal is "one self-contained embedded binary"
(AD-9), that property mattered more than the marginal performance or
extension-surface differences between the two drivers. If you need a
SQLite extension or feature only the CGO driver exposes, that's a real
reason to escalate — see [Escalating beyond the
defaults](../how-to/escalate-beyond-the-defaults.md) — but it's a
conscious trade you're making back, not a default you were fighting.

## Why the fixed WAL + `BEGIN IMMEDIATE` single-writer contract (AD-2)

SQLite's classic footgun under concurrent writers is a "deferred"
transaction that starts read-only and only requests the write lock
when it hits its first write statement — by which point another
writer may have already taken the lock, producing an `SQLITE_BUSY`
that surfaces deep inside application code, often mid-transaction. The
[concurrency contract](../reference/concurrency-contract.md) closes
that failure mode structurally rather than by convention: a dedicated
write pool capped at one connection (`SetMaxOpenConns(1)`) means only
one write transaction can ever be in flight, and every transaction
from that pool opens with `_txlock=immediate`, so the write lock is
acquired at `BEGIN`, not discovered missing partway through. The
trade-off is throughput — writes serialize completely — which is the
right trade for a template whose target is a single-instance app with
a local SQLite file, not a write-heavy distributed system. Reads get
their own pool with `deferred` locking and no connection cap, so they
are never blocked by this write serialization beyond WAL's normal
reader/writer isolation.

This is deliberately not configurable. A flag to loosen it would
invite exactly the failure mode it exists to prevent, the first time
someone flips it under load-testing pressure without fully
understanding the consequence.

## Why `templ` for views

`templ` compiles `.templ` files into type-checked Go functions at build
time — a malformed template is a compile error, not a runtime panic or
a silently blank fragment. That compile-time check is worth the extra
build step (`templ generate`, wired into every recipe that touches Go
code via `justfile`'s dependency graph) for a template meant to be
copied and extended by teams who may not have deep familiarity with
whatever templating library they inherit. `html/template` remains a
documented escalation for environments that can't tolerate an extra
codegen step — see [Escalating beyond the
defaults](../how-to/escalate-beyond-the-defaults.md) — at the cost of
losing that compile-time guarantee.

## Why `net/http.ServeMux` instead of `chi` by default (AD-4)

Go 1.22 added method- and wildcard-pattern routing to the standard
library's `ServeMux`, closing most of the gap that used to justify
reaching for a router library by default. Starting from the stdlib
keeps this template's dependency surface smaller and its routing layer
legible without a library's own documentation. `chi`'s middleware
share the exact same `func(http.Handler) http.Handler` signature this
template's own `httpserver.Middleware` type uses, so adopting `chi`
later is a drop-in swap, not a rewrite — see [Escalating beyond the
defaults](../how-to/escalate-beyond-the-defaults.md) for exactly what
changes.

## Why Litestream-only by default, not LiteFS or libSQL (AD-5)

Litestream replicates the WAL to object storage out-of-process — it
adds no Go dependency and leaves the single-file, single-writer
contract (AD-1/AD-2) completely unchanged. That makes it the right
default for a template: durability against volume/instance loss with
zero architectural cost. LiteFS and libSQL/Turso solve a different
problem — multiple app instances reading (and in libSQL's case,
possibly writing through a remote primary) the same data — which is a
real architecture change, not a tuning knob, so this template documents
them as escalation paths rather than wiring either in by default. See
[Deploy with Litestream](../how-to/deploy-with-litestream.md) for the
default and [Escalating beyond the
defaults](../how-to/escalate-beyond-the-defaults.md) for both
alternatives.

## Why one internal/ layout boundary, enforced by lint (AD-6)

`internal/platform/*` (infrastructure) and `internal/web/*` (shared UI)
must never import `internal/<feature>/*` — otherwise infrastructure
code silently accumulates feature-specific assumptions and stops being
infrastructure. Go's compiler enforces that `internal/` isn't
importable from outside the module, but nothing in the language
enforces boundaries *within* `internal/`. This template makes the rule
real by encoding it in `.golangci.yml`'s `platform-no-feature-imports`
depguard rule: a violation fails `just lint`, and therefore CI, rather
than depending on a reviewer noticing it in a diff. `internal/notes` is
the worked example of a compliant feature package — see [Add a feature
package](../how-to/add-a-feature-package.md).

## Why htmx + Server-Sent Events for real-time UI

Server-rendered HTML plus htmx keeps all rendering logic on the server
in one language (Go, via `templ`), rather than splitting it between a
Go backend and a separate JavaScript frontend rendering the same
domain model twice. SSE was chosen over WebSockets for the
`internal/notes` example specifically because the real-time need here
is one-directional (server pushes updates; the client's own writes go
through a normal `POST`) — SSE is simpler to reason about and to test
(a plain HTTP response with `httptest.NewServer`) for that shape, and
htmx's SSE extension hands broadcast fragments to the same core OOB
swap logic normal htmx responses use, so there's no separate mental
model for "live" versus "on click." A feature with genuine
bidirectional real-time needs would be a real reason to reach for
WebSockets instead — that's not something this template's default
tries to preempt.

## Why a single self-contained embedded binary (AD-9)

Static assets and templates are embedded via `go:embed` (with a
build-tag split — `assets_dev.go` serves from disk under `-tags dev`
for edit-and-reload convenience; the default build embeds everything),
so the built binary in `bin/` is the entire deployable unit: no
separate asset bundle to ship, no risk of a binary and its assets
drifting apart across a deploy. This is the same instinct behind
picking a CGO-free SQLite driver (AD-1) — minimizing what has to travel
together for the app to run correctly, all the way down to "one file."

## Where these decisions came from

These are the architectural decisions (AD-1 through AD-9) from this
template's seed specification, adapted here for people using the
template rather than people who wrote the spec — the full decision
records live in the seed spec this template was built from, not
re-litigated in this repo. `AGENTS.md` treats these as already
settled: an agent working mechanically in this repo (scaffolding,
codegen, wiring a new feature) should not re-open them, only apply
them.
