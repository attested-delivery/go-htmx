set shell := ["bash", "-euo", "pipefail", "-c"]

# List available recipes.
default:
    @just --list

# Regenerate templ Go code from .templ sources (AD-3), sqlc's database/sql
# code from internal/platform/db/query/*.sql (Story #3), and the compiled
# Tailwind stylesheet from internal/web/assets/tailwind/input.css.
generate: tailwind
    templ generate
    sqlc generate

# Compile internal/web/assets/tailwind/input.css into the embedded
# stylesheet (internal/web/assets/static/css/app.css is gitignored — a
# build artifact, not a source file; see .gitignore for why).
tailwind:
    tailwindcss -i internal/web/assets/tailwind/input.css \
        -o internal/web/assets/static/css/app.css --minify

# Build the single self-contained binary (AD-9). Depends on generate so a
# stale _templ.go never silently ships.
build: generate
    go build -o bin/go-htmx ./cmd/go-htmx

# Run the dev server against on-disk assets (-tags dev, see
# internal/web/assets/assets_dev.go) instead of the embedded build.
run: generate
    GO_HTMX_ENV=dev go run -tags dev ./cmd/go-htmx

# Run the test suite. Matches ci.yml's Test step exactly, so a green
# `just test` means a green CI test step (local/CI parity).
test: generate
    go test -race -cover ./...

# Regenerate golden-file fixtures (AD-8's snapshot tier) after a
# deliberate rendered-output change. Scoped to internal/notes (the only
# package importing goldie right now) rather than ./... — goldie
# registers -update as a per-test-binary flag, so any other package's
# test binary rejects it as unrecognized and fails outright. Review the
# resulting testdata/*.golden diff before committing — an unreviewed
# -update run is how a real regression gets silently baked in as the
# new "expected" output.
test-golden-update: generate
    go test ./internal/notes/... -run TestGolden -update

# Lint (golangci-lint; config in .golangci.yml).
lint: generate
    golangci-lint run ./...

# Run the full CI sequence locally (generate + build + lint + test), so a
# green `just check` predicts a green ci.yml build-lint-test job before
# pushing.
check: build lint test

# Rewrite this template's own identity (module path, binary/cmd name,
# env-var prefix) into a new project's, in one deterministic, idempotent
# pass (Story #6, Task #25) — see tools/init/main.go's doc comment for
# the full identity audit this covers. Run once, right after copying the
# template via "Use this template".
init name module:
    go run ./tools/init {{name}} {{module}}

# The template's real acceptance test (Task #27): copy the tree to a
# temp dir, run `just init` with a throwaway identity, then build + test
# the copy, and grep-gate for any leftover template identity string.
# Proves `just init` actually works end to end, not just that it runs.
smoke-init:
    scripts/smoke-init.sh

# Create a new zero-padded migration file (Task #16's padding contract —
# see internal/platform/db/migrations_test.go's TestMigrationFilenamesArePadded).
# -s selects goose's sequential (zero-padded numeric) naming over its
# default timestamp-based naming.
migrate-new name:
    go run github.com/pressly/goose/v3/cmd/goose@v3.27.2 -dir internal/platform/db/migrations -s create {{name}} sql

# Build the distroless container image locally (Task #52). Matches what
# release.yml's `docker` job builds, minus the push — useful to verify a
# Dockerfile change builds and the app actually serves before pushing.
docker-build:
    docker build -t go-htmx:local .

# Format Go and templ sources.
fmt:
    gofmt -l -w .
    templ fmt .

# Remove build output.
clean:
    rm -rf bin
