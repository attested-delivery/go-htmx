set shell := ["bash", "-euo", "pipefail", "-c"]

# List available recipes.
default:
    @just --list

# Regenerate templ Go code from .templ sources (AD-3) and sqlc's
# database/sql code from internal/platform/db/query/*.sql (Story #3).
generate:
    templ generate
    sqlc generate

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
# deliberate rendered-output change. Review the resulting testdata/*.golden
# diff before committing — an unreviewed -update run is how a real
# regression gets silently baked in as the new "expected" output.
test-golden-update: generate
    go test ./... -run TestGolden -update

# Lint (golangci-lint; config in .golangci.yml).
lint: generate
    golangci-lint run ./...

# Run the full CI sequence locally (generate + build + lint + test), so a
# green `just check` predicts a green ci.yml build-lint-test job before
# pushing.
check: build lint test

# Create a new zero-padded migration file (Task #16's padding contract —
# see internal/platform/db/migrations_test.go's TestMigrationFilenamesArePadded).
# -s selects goose's sequential (zero-padded numeric) naming over its
# default timestamp-based naming.
migrate-new name:
    go run github.com/pressly/goose/v3/cmd/goose@v3.27.2 -dir internal/platform/db/migrations -s create {{name}} sql

# Format Go and templ sources.
fmt:
    gofmt -l -w .
    templ fmt .

# Remove build output.
clean:
    rm -rf bin
