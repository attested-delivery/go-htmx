---
id: urn:mif:go-htmx:how-to:style-with-tailwind
type: procedural
created: 2026-07-13T00:00:00Z
namespace: go-htmx/docs/how-to
tags: [how-to, tailwind, css, styling]
title: "How to style a feature with Tailwind"
temporal:
  validFrom: 2026-07-13T00:00:00Z
relationships:
  - type: relates-to
    target: docs/how-to/add-a-feature-package.md
  - type: relates-to
    target: docs/reference/project-layout.md
  - type: relates-to
    target: docs/reference/justfile-commands.md
provenance:
  '@type': Provenance
  agent: claude-code/claude-sonnet-5
  wasGeneratedBy:
    '@id': urn:mif:activity:claude-code-session:91f08ccf-1fbc-4d5c-843d-9ff6d4050ce8
    '@type': prov:Activity
  trustLevel: user_stated
  agentVersion: 2.1.207
modified: '2026-07-13T16:52:48.061Z'
---

# How to style a feature with Tailwind

This template ships with [Tailwind CSS](https://tailwindcss.com) v4,
compiled by its standalone CLI at build time — no Node.js/npm anywhere
in the toolchain (see [Understanding this template's
architecture](../explanation/architecture.md) for why). `internal/notes`
is the worked example: its form, note cards, and count badge are all
styled with Tailwind utility classes.

## Prerequisites

- A working copy of this template.
- The Tailwind CLI installed and pinned per `AGENTS.md`'s Toolchain
  section.

## How the build step works

- `internal/web/assets/tailwind/input.css` is the source: an
  `@import "tailwindcss";` line, explicit `@source` globs pointing at
  every directory containing `.templ` files (Tailwind v4's automatic
  content-detection doesn't reliably scan `.templ`), and an `@theme`
  block for design tokens.
- `just tailwind` compiles it into
  `internal/web/assets/static/css/app.css` — a **build artifact**,
  gitignored the same way `*_templ.go` and `sqlc`'s generated package
  are (`input.css` is the source of truth, not the compiled output).
- `just generate` runs `just tailwind` first, so `just build`/`run`/
  `test`/`lint`/`check` all pick up style changes automatically —
  there's no separate "remember to rebuild CSS" step.

## Adding Tailwind classes to a new feature

If you're adding a new `internal/<feature>` package (see [Add a
feature package](add-a-feature-package.md)), its `.templ` files are
covered automatically once you add a matching `@source` glob to
`input.css` — e.g. for `internal/orders`:

```css
@source "../../../orders/**/*.templ";
```

Paths are relative to `input.css`'s own location
(`internal/web/assets/tailwind/`), not the repo root — get this wrong
and Tailwind silently won't see classes used only in that feature's
markup (they'll be purged from the compiled output). Run `just
tailwind` and grep the compiled `app.css` for a class you just added
to confirm it actually made it in.

## Extending the design tokens

Add to the `@theme` block in `input.css` — e.g. a custom font:

```css
@theme {
  --font-sans: ui-sans-serif, system-ui, sans-serif;
}
```

Tailwind v4 uses CSS-based configuration (no `tailwind.config.js`);
everything lives in `input.css`.

## Dark mode

This template's styling uses Tailwind's `dark:` variant, which follows
the browser/OS `prefers-color-scheme` setting — there's no manual
light/dark toggle. Add `dark:`-prefixed classes alongside the light-mode
ones (see `internal/notes/views.templ`'s note cards for the pattern:
`bg-white dark:bg-gray-800`, etc.).

## A note on `<script>` execution order

If you add a `<script>` tag whose code needs to find an already-parsed
DOM element (the way `internal/web/assets/static/js/app.js` looks up
`#note-form`), it needs the `defer` attribute if it's loaded from
`<head>` — otherwise it executes before `<body>` is parsed and
`getElementById` silently returns `null`. htmx.min.js/hx-sse.min.js
don't need this since they self-initialize via their own internal
`DOMContentLoaded` handling before scanning for `hx-*` attributes; a
plain script with a top-level DOM lookup does.

## Completion

`just check` passes, and the compiled `app.css` (`just tailwind`, or
inspect it after `just build`) contains the classes you added.
