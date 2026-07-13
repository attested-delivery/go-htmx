---
id: urn:mif:go-htmx:tutorial:getting-started
type: procedural
created: 2026-07-12T00:00:00Z
namespace: go-htmx/docs/tutorial
tags: [tutorial, onboarding, quickstart]
title: "Use this template: from copy to a live change"
temporal:
  validFrom: 2026-07-12T00:00:00Z
relationships:
  - type: relates-to
    target: docs/how-to/add-a-feature-package.md
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
modified: '2026-07-13T16:47:22.784Z'
---

# Use this template: from copy to a live change

By the end of this tutorial you will have a running go-htmx app with its
own identity (not the template's), and you'll have made a real change to
its UI and watched it appear in the browser.

## Prerequisites

- Go (current Active LTS — check `go.mod`'s `toolchain` line), and
  [`just`](https://github.com/casey/just) on your `PATH`.
- `templ`, `sqlc`, and `golangci-lint` installed and pinned to the
  versions in `AGENTS.md`'s Toolchain section — run those `go install`
  commands now if you haven't already.
- The Tailwind CSS standalone CLI, also pinned in `AGENTS.md`'s
  Toolchain section (a separate `curl`-and-checksum-verify install, not
  `go install` — Tailwind's CLI isn't a Go module).
- A copy of this template on disk (via GitHub's "Use this template", or
  a plain clone), with a terminal open at its root.

## Step 1 — Give it your own identity

Run:

```sh
just init myapp github.com/you/myapp
```

Replace `myapp` and `github.com/you/myapp` with your project's actual
name and module path. This rewrites the module path, the `cmd/`
directory name, the environment-variable prefix, and every other
identity-bearing reference from `go-htmx` to `myapp` in one pass.

You should now see output ending with a line like:

```
init complete: N file(s) rewritten, module github.com/attested-delivery/go-htmx -> github.com/you/myapp, name go-htmx -> myapp
```

and a `cmd/myapp/` directory where `cmd/go-htmx/` used to be.

## Step 2 — Build and run it

```sh
just build
just run
```

`just run` starts a dev server on `:8080` using on-disk static assets
(so CSS/JS edits show up without a rebuild). Leave it running.

Open `http://localhost:8080/` in a browser. You should now see a page
titled "go-htmx — notes" (or "myapp — notes" if you already ran `just
init`) with a form, a note count badge, and an empty notes list.

## Step 3 — Try the live part

Type something into the text field and click **Add**. The note appears
in the list and the count badge updates — over Server-Sent Events, not
a page reload. Open a second browser tab to the same URL and add
another note: it appears in *both* tabs live.

## Step 4 — Make one small visible change

Stop the server (Ctrl-C) and open `internal/notes/views.templ` in an
editor. Find this line:

```templ
<h1 class="text-2xl font-semibold tracking-tight">Notes</h1>
```

Change the text (leave the `class` attribute — that's Tailwind's
styling for the heading, not part of what you're editing) to:

```templ
<h1 class="text-2xl font-semibold tracking-tight">My Notes</h1>
```

Save, then run `just run` again (it re-runs `templ generate` for you).
Reload `http://localhost:8080/` in the browser.

You should now see the page heading read "My Notes".

## What you've done

You copied this template, gave it a real identity, ran it, watched its
real-time UI update live across two tabs, and changed its source and
saw the change appear. From here:

- To add a whole new feature the same way `internal/notes` demonstrates,
  see [Add a feature package](../how-to/add-a-feature-package.md).
- To understand *why* the app is built this way — the SQLite concurrency
  contract, the choice of `templ`, the single-binary design — see
  [Architecture rationale](../explanation/architecture.md).
