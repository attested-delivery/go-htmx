---
id: urn:mif:go-htmx:how-to:add-an-sse-live-region
type: procedural
created: 2026-07-12T00:00:00Z
namespace: go-htmx/docs/how-to
tags: [how-to, sse, htmx, real-time]
title: "How to add an SSE-pushed live region"
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
modified: '2026-07-13T02:18:54.076Z'
---

# How to add an SSE-pushed live region

When you want part of a page to update live for every connected client,
the way `internal/notes`'s notes list and count badge do, following the
same pattern for a new feature.

## Prerequisites

- A feature package already scaffolded (see
  [Add a feature package](add-a-feature-package.md)).

## Steps

1. Add a `Broadcaster` to your feature, following
   `internal/notes/broadcaster.go`: an in-process pub/sub with a
   `map[chan string]struct{}` guarded by a mutex, exposing `Subscribe()
   (chan string, unsubscribe func())` and `Publish(msg string)`.

2. In your feature's view, give the connecting element its own stable,
   empty container — do not put `hx-sse:connect` on the same element
   your broadcasts OOB-replace, or a full-region refresh will detach
   the connection delivering it:

   ```templ
   <div
       id="<feature>-stream"
       hx-sse:connect="/<feature>/stream"
       hx-target="#<feature>-content"
       hx-swap="afterbegin"
       style="display:none"
   ></div>
   <div id="<feature>-content">
       // initial server-rendered content
   </div>
   ```

3. Add a `GET /<feature>/stream` handler that: sets
   `Content-Type: text/event-stream`, `Cache-Control: no-cache`, and
   `Connection: keep-alive`; flushes the header; subscribes to the
   `Broadcaster` *before* querying current state (subscribing after
   risks missing an update that lands in the gap); sends a full-state
   "sync" fragment as the first event (see below); then loops writing
   each subsequent published message until the request context is
   done. Follow `internal/notes/handler.go`'s `handleStream` for the
   exact sequencing and its SSE line-per-line write helper
   (`writeSSEEvent`) — a naive single `data: <multi-line>` write
   truncates at the first newline.

4. Add a "sync" view fragment (see `internal/notes/views.templ`'s
   `Sync`) that OOB-replaces your content container's children
   (`hx-swap-oob="innerHTML"`) with the full current state. This closes
   the gap between the page render and the SSE connection actually
   opening — a client that connects after another client's update
   would otherwise never see it until a manual reload.

5. Wherever your handler mutates state, render a broadcast fragment
   (the new/changed item, plus any OOB-updated summary elements) and
   call `bus.Publish(...)` on it, following
   `internal/notes/handler.go`'s `handleCreate`.

## Completion

Opening the page in two browser tabs and triggering an update in one
tab makes it appear live in both, without a page reload.
