---
id: urn:mif:go-htmx:how-to:deploy-with-litestream
type: procedural
created: 2026-07-12T00:00:00Z
namespace: go-htmx/docs/how-to
tags: [how-to, deploy, litestream, durability]
title: "How to deploy with Litestream, and drill a restore"
relationships:
  - type: relates-to
    target: docs/reference/release-artifacts.md
provenance:
  '@type': Provenance
  agent: claude-code/claude-sonnet-5
  wasGeneratedBy:
    '@id': urn:mif:activity:claude-code-session:91f08ccf-1fbc-4d5c-843d-9ff6d4050ce8
    '@type': prov:Activity
  trustLevel: user_stated
  agentVersion: 2.1.207
modified: '2026-07-13T02:19:06.534Z'
---

# How to deploy with Litestream, and drill a restore

When you're ready to run this app somewhere persistent and want the
SQLite file protected against volume/instance loss, using this
template's default durability story.

## Prerequisites

- An S3-compatible object storage bucket (or any Litestream-supported
  replica target) and credentials for it.
- The built binary (`just build` produces `bin/<name>`) and the
  `litestream` binary on the host.

## Steps

1. Set the environment variables `litestream.yml` reads:
   `GO_HTMX_DB_PATH` (or your renamed app's equivalent env prefix —
   `just init` rewrites this), `LITESTREAM_S3_BUCKET`,
   `LITESTREAM_S3_PATH`, `LITESTREAM_S3_REGION`,
   `LITESTREAM_S3_ENDPOINT`.

2. Start replication as a sidecar process, alongside the app (not
   inside it — nothing in `internal/` imports Litestream):

   ```sh
   litestream replicate -config litestream.yml
   ```

   With any `LITESTREAM_*` variable unset, `litestream replicate`
   fails outright on the unresolved environment substitution rather
   than silently running unreplicated — treat that failure as your
   deploy blocking on real config, not a bug to route around.

3. Start the app binary normally (`./bin/<name>`), pointed at the same
   `GO_HTMX_DB_PATH` file `litestream.yml` replicates.

## Drill a restore

Before trusting this in production, prove the restore path works:

1. Stop the app and delete (or move aside) the local database file.
2. Restore from the replica:

   ```sh
   litestream restore -config litestream.yml "$GO_HTMX_DB_PATH"
   ```

3. Start the app again and confirm your data is back — e.g. reload the
   notes page and see the notes that existed before the deletion.

## Completion

The restored database file matches what you had before the drill, and
the app serves correctly from it.

## Outgrowing Litestream

Litestream keeps the "single SQLite file, single writer" contract
unchanged because it replicates the WAL out-of-process. Two paths exist
for outgrowing it — see [Escalating beyond the
defaults](escalate-beyond-the-defaults.md) for LiteFS (multi-instance
read replicas) and libSQL/Turso (managed, edge-hosted primary).
