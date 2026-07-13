# Changelog

All notable changes to this project are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

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
  throwaway copy.
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

### Changed

- `quality-gates.yml`'s `sast` job no longer gates on a
  `detect-go-source` check now that real Go source exists; `sast /
  analyze` is a required status check on `main`.

## [0.1.0] - Unreleased

First tagged release. See the [Unreleased] section above for its full
contents; this section is renamed and dated at tag time (Task #43).
