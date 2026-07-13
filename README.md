# go-htmx

A production-ready Go + HTMX + SQLite project template: `modernc.org/sqlite`
(pure Go, CGO-free), a fixed WAL + `BEGIN IMMEDIATE` single-writer
concurrency contract, `templ` views, htmx v4 + Server-Sent Events for
real-time UI, `goose` migrations, `sqlc` codegen, and a single
self-contained embedded binary.

## Quickstart

```sh
just init myapp github.com/you/myapp   # rewrite this template's identity to yours — see below
just build                              # produces bin/myapp, a single self-contained binary
just run                                # dev server against on-disk static assets (no rebuild needed
                                         # to see a CSS/JS edit — Go source changes still need a restart)
just test                               # httptest + in-memory SQLite + golden-file snapshot tests
```

Run `just` with no arguments to list every recipe.

## Post-copy checklist

After clicking "Use this template" (or otherwise copying this tree):

1. **Run `just init <name> <module>`** — e.g. `just init myapp github.com/you/myapp`. This
   rewrites the module path, binary/`cmd/` name, env var prefix, and every
   other identity-bearing reference in one deterministic, idempotent pass
   (see `tools/init/main.go`'s doc comment for the full audit of what it
   touches). Run it exactly once, right after copying.
2. **Verify**: `just build && just test` should be green. `just smoke-init`
   re-runs this whole flow against a throwaway copy and is the template's
   own CI-enforced proof that step 1 actually works end to end.
3. **Optional: remove the example feature.** `internal/notes/` is a
   worked example (Story #4) demonstrating the htmx/SSE/OOB-swap pattern
   over the data layer — delete it and its route registration in
   `cmd/<name>/main.go` once you've built your own feature the same way,
   or keep it as a live reference.
4. **Owner-side settings this template cannot carry** (a copied repo
   starts with none of these — set them explicitly):
   - Branch protection / rulesets on `main`: require the CI status checks
     this repo actually runs (`pin-check`, `sca`, `trivy`, and this
     template's own `build-lint-test` job — check `.github/workflows/`
     for the current job names), required reviews, dismiss-stale,
     signed commits if your org requires them.
   - Repo secrets: `GITLEAKS_LICENSE` if you enable Gitleaks; anything
     Story #9's attested release pipeline needs (SLSA/SBOM signing is
     keyless/OIDC by default — no long-lived signing key secret needed).
   - If you're inside the `attested-delivery` org: this template's CI
     already calls the org's central reusable workflows
     (`attested-delivery/.github/...`, SHA-pinned) — see the org's
     `CLAUDE.md` §7 new-repo onboarding checklist for the full org-side
     setup (CODEOWNERS, secret scanning, dependabot). If you're
     *outside* the org, replace those `uses:` references with your own
     equivalent scanning/signing setup — they won't resolve otherwise.
   - Dependabot: `.github/dependabot.yml` is already wired for the
     `gomod`/`github-actions` ecosystems; nothing to add unless you want
     a different schedule.

## Architecture

See `internal/doc.go` for the `internal/platform/*` vs `internal/<feature>/*`
import boundary, and `AGENTS.md` (once Story #7 lands) for the full
conventions doc.
