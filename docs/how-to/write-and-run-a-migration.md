---
id: urn:mif:go-htmx:how-to:write-and-run-a-migration
type: procedural
created: 2026-07-12T00:00:00Z
namespace: go-htmx/docs/how-to
tags: [how-to, migration, goose, sqlc]
title: "How to write and run a migration"
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
modified: '2026-07-13T02:18:34.481Z'
---

# How to write and run a migration

When your schema needs a new table, column, or index.

## Prerequisites

- A working copy of this template.
- Know the schema change you want to make.

## Steps

1. Create the migration file:

   ```sh
   just migrate-new <name>
   ```

   This uses `goose`'s sequential (zero-padded numeric) naming, not
   timestamp-based naming — never hand-name a migration file, the
   numeric prefix must stay zero-padded or migrations apply out of
   order (see `internal/platform/db/migrations_test.go`'s
   `TestMigrationFilenamesArePadded`).

2. Open the generated file under
   `internal/platform/db/migrations/` and write the schema change as a
   goose SQL migration (a `-- +goose Up` section and, unless the change
   is genuinely irreversible, a matching `-- +goose Down`).

3. If the change adds or alters a table your queries touch, update or
   add queries in `internal/platform/db/query/*.sql`.

4. Regenerate `sqlc`'s output:

   ```sh
   just generate
   ```

5. Run the app once, or `just test`, so `db.Migrate` applies the new
   migration against a real (or in-memory) database and you see it
   succeed. Migrations run automatically at startup
   (`cmd/<app>/main.go` calls `db.Migrate(store.WriteDB())` before
   serving any request) — there is no separate manual "apply" step in
   production.

## Completion

`just check` passes, and `internal/platform/db/sqlc/`'s generated
`Queries` methods reflect the new schema.
