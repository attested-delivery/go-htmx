---
id: urn:mif:go-htmx:changelog
type: episodic
created: 2026-07-12T00:00:00Z
namespace: go-htmx
tags: [changelog, release-history]
title: "go-htmx Changelog"
temporal:
  validFrom: 2026-07-12T00:00:00Z
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
modified: '2026-07-13T17:19:07.549Z'
---

# Changelog

All notable changes to this project are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.1.3] - 2026-07-13

### Added

- Tailwind CSS v4 build pipeline: a standalone, pinned, checksum-verified
  CLI (no Node/npm), a new `just tailwind` recipe wired into `generate`,
  and `internal/web/assets/tailwind/input.css` as the styling source of
  truth. CI and the container build both install and run it.
- `internal/notes` restyled with Tailwind: card-based notes list, a
  styled form/button, a pill-shaped count badge, and dark mode via
  `prefers-color-scheme` — the existing SSE/OOB real-time layer is
  unchanged.
- A proportionate security-headers pass: `X-Content-Type-Options`,
  `Referrer-Policy`, `X-Frame-Options`, and a same-origin
  `Content-Security-Policy`, plus a 4 KiB request-body size cap on the
  notes-create endpoint. Scoped deliberately to this example app's real
  threat model (no auth/sessions) — no CSRF tokens, no rate limiting.
- A new how-to guide, `docs/how-to/style-with-tailwind.md`, plus
  supporting reference and explanation updates documenting the styling
  pipeline and the security-hardening rationale.

### Fixed

- Two Content-Security-Policy compatibility issues surfaced by testing
  the running app: `hx-on::after:request="this.reset()"` (which
  evaluates via an implicit `eval`, CSP-blocked) is now a real event
  listener on htmx's own `htmx:afterRequest` event; an inline
  `style="display:none"` attribute (also CSP-blocked) now uses a
  Tailwind utility class instead.
- Every document under `docs/` was missing the MIF `temporal`
  frontmatter field required for Level 2 conformance — added across the
  whole set.

## [0.1.2] - 2026-07-13

### Fixed

- The container image's vulnerability-scan job (`gate-image`)
  no longer fails on tagged releases — the upstream `reusable-trivy.yml`
  reusable never authenticated to GHCR before scanning an image by
  digest, so every one of Trivy's image-source backends failed
  (attested-delivery/.github#33). `v0.1.1`'s container image is signed
  with SLSA provenance and an SBOM attestation but is missing this one;
  `v0.1.2` completes the full attestation set.

## [0.1.1] - 2026-07-13

### Added

- Application skeleton: Go module, `cmd/go-htmx` entrypoint,
  `internal/platform/*` vs `internal/<feature>/*` layout boundary
  enforced by `.golangci.yml`'s `platform-no-feature-imports` depguard
  rule, `net/http.ServeMux`-based router with a `chi`-compatible
  middleware chain, `templ` + `go:embed` asset bundling with a dev/prod
  build-tag split.
- Data layer: `modernc.org/sqlite` (pure Go, CGO-free) with a fixed
  WAL + `BEGIN IMMEDIATE` single-writer concurrency contract (separate
  read/write connection pools), `goose` migrations embedded via
  `go:embed`, `sqlc` codegen, Litestream durability config.
- HTMX + Server-Sent Events real-time layer: `internal/notes` worked
  example demonstrating live multi-client updates, `hx-swap-oob`
  multi-region OOB swap patterns, a sync-on-connect fix for the
  render-to-connect replay gap.
- Test triad (AD-8): `httptest` handler tests, in-memory SQLite
  (`db.Open(":memory:")`, shared-cache), `goldie` golden-file snapshot
  tests, wired into `justfile` and CI.
- Template-instantiation tooling: `just init <name> <module>` rewrites
  the template's own identity into a new project's in one deterministic,
  idempotent pass; `just smoke-init` proves it end to end against a
  throwaway copy, including a container build.
- AI-agent conventions: `AGENTS.md` (package boundaries, commands,
  compile-time interface-satisfaction idiom) and a checked-in
  `add-feature-package` scaffolding skill.
- Full Diátaxis documentation set under `docs/`: a tutorial, five
  how-to guides, five reference documents, and an explanation document
  covering this template's architecture rationale.
- Attested release pipeline: cross-platform CGO-free binary builds
  (`linux-amd64`, `linux-arm64`, `macos-amd64`, `macos-arm64`,
  `windows-amd64.exe`), SLSA Build L3 provenance and a CycloneDX SBOM
  per binary, merge-time SAST/SCA/IaC gate verdicts re-run and
  seam-signed against the tagged commit, a fail-closed
  `gh attestation verify` before publish.
- A distroless container image (`Dockerfile`, `gcr.io/distroless/static-debian12:nonroot`,
  digest-pinned, non-root, ~5 MB), with its own image-only SLSA Build
  L3 signing, independent fail-closed verify, and a seam-signed Trivy
  image-scan verdict.
- A real `/healthz` endpoint (pings the database, not a stub that
  always returns 200) and a `go-htmx healthcheck` binary subcommand,
  wired into the container's `HEALTHCHECK` directive since distroless
  has no shell to probe an HTTP endpoint with directly.
- Git hooks (Lefthook): pre-commit auto-formats staged Go and templ
  files; pre-push runs the same build+lint+test sequence CI's
  build-lint-test job runs.

### Changed

- `quality-gates.yml`'s `sast` job no longer gates on a
  `detect-go-source` check now that real Go source exists; `sast /
  analyze` is a required status check on `main`.

### Fixed

- The container image's SBOM attestation step (`docker-sign`) no
  longer fails with a permissions error on tagged releases — the
  upstream `sign-and-attest.yml` reusable was trying to attach the
  SBOM as a release asset with insufficient permissions
  (attested-delivery/.github#32).
