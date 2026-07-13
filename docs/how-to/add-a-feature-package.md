---
id: urn:mif:go-htmx:how-to:add-a-feature-package
type: procedural
created: 2026-07-12T00:00:00Z
namespace: go-htmx/docs/how-to
tags: [how-to, feature, scaffolding]
title: "How to add a feature package"
relationships:
  - type: relates-to
    target: docs/reference/project-layout.md
provenance:
  '@type': Provenance
  agent: claude-code/claude-sonnet-5
  wasGeneratedBy:
    '@id': urn:mif:activity:claude-code-session:91f08ccf-1fbc-4d5c-843d-9ff6d4050ce8
    '@type': prov:Activity
  trustLevel: user_stated
  agentVersion: 2.1.207
modified: '2026-07-13T02:18:21.904Z'
---

# How to add a feature package

When you need a new self-contained feature (its own routes, views, and
optionally its own database table) alongside the `internal/notes`
example.

## Prerequisites

- A working copy of this template, already `just init`-ed.
- If the feature needs its own table: know its schema up front.

## Steps

1. Invoke the checked-in `add-feature-package` skill
   (`.claude/skills/add-feature-package/SKILL.md`) with your feature's
   name (a short, lowercase, valid Go identifier — no hyphens or
   underscores). It scaffolds `internal/<feature>/handler.go`,
   `views.templ`, and `handler_test.go` from the `internal/notes`
   pattern.

2. If the feature needs its own table, add a migration first:

   ```sh
   just migrate-new create_<feature>
   ```

   Edit the generated file under
   `internal/platform/db/migrations/`, then add its queries in
   `internal/platform/db/query/<feature>.sql` following
   `internal/platform/db/query/notes.sql` as the pattern.

3. Run `just generate` so `sqlc` regenerates
   `internal/platform/db/sqlc/` against the new schema/queries.

4. Wire the handler into `cmd/<app>/main.go`: import the new package
   and call its `Register(mux)` alongside `notes.NewHandler(...)`'s
   call, following the existing wiring for `notes`.

5. Run `just check`.

## Completion

`just check` passes: the new package builds, lints clean (its imports
satisfy `.golangci.yml`'s `platform-no-feature-imports` rule if it
touches `internal/platform/*`), and its scaffolded `handler_test.go`
passes.
