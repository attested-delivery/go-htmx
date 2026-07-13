---
id: urn:mif:go-htmx:how-to:add-e2e-coverage
type: procedural
created: 2026-07-13T00:00:00Z
namespace: go-htmx/docs/how-to
tags: [how-to, e2e, testing, playwright, accessibility, visual-regression]
title: "How to add E2E coverage for a new feature"
temporal:
  validFrom: 2026-07-13T00:00:00Z
relationships:
  - type: relates-to
    target: docs/how-to/add-a-feature-package.md
  - type: relates-to
    target: docs/reference/project-layout.md
  - type: relates-to
    target: docs/reference/justfile-commands.md
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
modified: '2026-07-13T22:11:02.677Z'
---

# How to add E2E coverage for a new feature

When your new `internal/<feature>` package needs the same browser-driven
test coverage `internal/notes` has: functional, accessibility,
cross-browser, and visual regression.

## Prerequisites

- A feature package already added — see [Add a feature
  package](add-a-feature-package.md) — with stable `id="..."` hooks on
  the DOM elements a test needs to find.
- `just e2e-install` run at least once.

## Steps

1. Copy `e2e/pages/notes_page.go` to `e2e/pages/<feature>_page.go`.
   Rename `NotesPage` to `<Feature>Page` and `NewNotesPage` to
   `New<Feature>Page`. Replace every CSS selector with your feature's
   own `id="..."` hooks. Keep each method's shape: an action method
   (`CreateNote` → your feature's equivalent) pairs with a wait method
   built on `playwright.NewPlaywrightAssertions()`'s `LocatorAssertions`
   (`ToHaveText`, `ToBeVisible`, `ToHaveValue`) — see [Understanding
   this template's architecture](../explanation/architecture.md) for why
   this replaces a hand-rolled poll loop.

2. Copy `e2e/functional/notes_test.go` to
   `e2e/functional/<feature>_test.go`. In its `newNotesPage` (or your
   renamed equivalent) helper, keep `testapp.New(t)` as the **first**
   call, before opening the browser page — see [Understanding this
   template's architecture](../explanation/architecture.md) for why this
   ordering matters. Swap in your Step 1 page object everywhere
   `pages.NewNotesPage` is called.

3. Name any test that must block every PR with `Smoke` in its function
   name (e.g. `TestSmoke_CreateOrderAppears`) — see [Understanding this
   template's architecture](../explanation/architecture.md) for the
   `Smoke`/full-suite split's rationale. Leave every other test
   unprefixed; it runs on merge to `main` only.

4. If your feature needs its own accessibility check, add a test
   alongside `e2e/accessibility/axe_test.go`'s pattern. When you inject
   `axe.min.js` via `Page.AddScriptTag`, set
   `BrowserNewPageOptions.BypassCSP: playwright.Bool(true)` on that
   page — see [Understanding this template's
   architecture](../explanation/architecture.md) for why.

5. If your feature needs its own visual regression baseline, add a case
   alongside `e2e/visual/notes_test.go`'s pattern using
   `e2e/visual.Compare`. Mask any element that legitimately differs
   between runs (a timestamp, a generated ID) via
   `PageScreenshotOptions.Mask` before taking the screenshot. Run `just
   test-visual-update` to generate the baseline, then review the
   resulting `e2e/visual/testdata/<name>.golden.png` before committing
   it.

6. Falsify each new test: temporarily revert the fix or behavior it's
   meant to catch, run it, and confirm it fails with a clear diagnostic.
   Restore the fix and confirm the test passes again.

7. Run `just e2e-install && just e2e-full`. If you added a
   `Smoke`-tagged test, also run `just e2e-smoke` on its own.

## Completion

`just e2e-full` passes, including your new test(s); `just check` is
unaffected, since the `e2e` build tag keeps this whole tree out of the
default `go build`/`go vet`/`go test ./...` path.
