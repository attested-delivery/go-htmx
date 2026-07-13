# go-htmx

A production-ready Go + HTMX + SQLite project template: `modernc.org/sqlite`
(pure Go, CGO-free), a fixed WAL + `BEGIN IMMEDIATE` single-writer
concurrency contract, `templ` views, htmx v4 + Server-Sent Events for
real-time UI, `goose` migrations, `sqlc` codegen, a single self-contained
embedded binary, a distroless container image, and an attested release
pipeline (SLSA provenance, CycloneDX SBOM, seam-signed gate verdicts,
fail-closed verify).

## Quickstart

```sh
just init myapp github.com/you/myapp   # rewrite this template's identity to yours — see below
just build                              # produces bin/myapp, a single self-contained binary
just run                                # dev server against on-disk static assets (no rebuild needed
                                         # to see a CSS/JS edit — Go source changes still need a restart)
just test                               # httptest + in-memory SQLite + golden-file snapshot tests
```

Run `just` with no arguments to list every recipe.

## Documentation

Full Diátaxis documentation set under [`docs/`](docs/):

- **Tutorial**: [Use this template: from copy to a live
  change](docs/tutorial/getting-started.md) — start here if you're new.
- **How-to guides**: [add a feature
  package](docs/how-to/add-a-feature-package.md), [write and run a
  migration](docs/how-to/write-and-run-a-migration.md), [add an
  SSE-pushed live region](docs/how-to/add-an-sse-live-region.md),
  [deploy with Litestream](docs/how-to/deploy-with-litestream.md),
  [escalate beyond the
  defaults](docs/how-to/escalate-beyond-the-defaults.md) (chi,
  mattn/go-sqlite3, LiteFS/libSQL, html/template).
- **Reference**: [justfile
  commands](docs/reference/justfile-commands.md), [project
  layout](docs/reference/project-layout.md),
  [configuration](docs/reference/configuration.md), [the SQLite
  concurrency contract](docs/reference/concurrency-contract.md),
  [release artifact naming and
  verification](docs/reference/release-artifacts.md).
- **Explanation**: [understanding this template's
  architecture](docs/explanation/architecture.md) — the *why* behind
  every default above.

`AGENTS.md` has conventions for any coding agent (or human) working
mechanically in this repo — scaffolding, codegen, wiring a new feature.

## Verifying a release

Every tagged release publishes five platform binaries, a checksums
manifest, a distroless container image, SLSA Build L3 provenance, a
CycloneDX SBOM, and seam-signed SAST/SCA/IaC/container-scan gate
verdicts — all independently, keylessly verifiable. See
[`SECURITY.md`](SECURITY.md) for the exact, copy-pasteable
`gh attestation verify`/`cosign verify` commands.

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
     this repo actually runs (`pin-check`, `sast`, `sca`, `trivy` —
     check `.github/workflows/quality-gates.yml` for the current job
     names), required reviews, dismiss-stale, signed commits if your
     org requires them.
   - Repo secrets: `GITLEAKS_LICENSE` if you enable Gitleaks. Nothing
     else — the release pipeline's SLSA/SBOM/image signing is
     keyless/OIDC by default, no long-lived signing key secret needed.
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
import boundary, `AGENTS.md` for the full conventions doc, and
[`docs/explanation/architecture.md`](docs/explanation/architecture.md)
for the rationale behind every default this template picks.
