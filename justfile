set shell := ["bash", "-euo", "pipefail", "-c"]

# List available recipes.
default:
    @just --list

# Regenerate templ Go code from .templ sources (AD-3).
generate:
    templ generate

# Build the single self-contained binary (AD-9). Depends on generate so a
# stale _templ.go never silently ships.
build: generate
    go build -o bin/go-htmx ./cmd/go-htmx

# Run the dev server against on-disk assets (-tags dev, see
# internal/web/assets/assets_dev.go) instead of the embedded build.
run: generate
    GO_HTMX_ENV=dev go run -tags dev ./cmd/go-htmx

# Run the test suite.
test: generate
    go test ./...

# Lint (golangci-lint; config in .golangci.yml).
lint: generate
    golangci-lint run ./...

# Format Go and templ sources.
fmt:
    gofmt -l -w .
    templ fmt .

# Remove build output.
clean:
    rm -rf bin
