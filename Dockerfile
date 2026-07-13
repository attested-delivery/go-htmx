# Multi-stage build producing a minimal, non-root distroless image. The
# binary is CGO_ENABLED=0 pure Go (modernc.org/sqlite, AD-1), so the final
# stage needs no libc at all -- gcr.io/distroless/static rather than
# distroless/base. Both stages are pinned by digest (not just tag), the
# same integrity bar this repo's GitHub Actions pins hold third-party
# actions to -- a tag can move, a digest can't.
#
# Builder: golang:1.26.5-bookworm, matching go.mod's toolchain line
# exactly (resolved live via `docker buildx imagetools inspect` at
# authoring time -- re-resolve if go.mod's toolchain version changes).
FROM golang:1.26.5-bookworm@sha256:18aedc16aa19b3fd7ded7245fc14b109e054d65d22ed53c355c899582bbb2113 AS builder

WORKDIR /src

# Module download is its own layer so a source-only change doesn't
# invalidate it.
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# templ/sqlc pinned to the exact versions AGENTS.md's Toolchain section
# documents -- keep both in sync by hand if either changes.
RUN go install github.com/a-h/templ/cmd/templ@v0.3.1020 && \
    go install github.com/sqlc-dev/sqlc/cmd/sqlc@v1.31.1 && \
    templ generate && \
    sqlc generate

RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/go-htmx ./cmd/go-htmx

# The distroless nonroot image runs as uid/gid 65532 with no shell, so
# there's no `chown` available in the final stage -- create and own the
# writable data directory here instead, then copy it across with
# --chown. GO_HTMX_DB_PATH defaults to this path below; a real deployment
# should mount a volume at /data for the SQLite file to persist across
# container restarts (see docs/how-to/deploy-with-litestream.md for the
# durability story on top of that).
RUN mkdir -p /data && chown 65532:65532 /data

# gcr.io/distroless/static-debian12:nonroot -- multi-arch index digest
# (covers both linux/amd64 and linux/arm64), resolved live at authoring
# time. No shell, no package manager, no libc: matches this binary's
# CGO-free build and minimizes the image's attack surface.
FROM gcr.io/distroless/static-debian12:nonroot@sha256:b7bb25d9f7c31d2bdd1982feb4dafcaf137703c7075dbe2febb41c24212b946f

COPY --from=builder /out/go-htmx /go-htmx
COPY --from=builder --chown=65532:65532 /data /data

ENV GO_HTMX_ADDR=:8080
ENV GO_HTMX_ENV=prod
ENV GO_HTMX_DB_PATH=/data/go-htmx.db

WORKDIR /data
EXPOSE 8080
USER 65532:65532

# Distroless has no shell/curl/wget to probe an HTTP endpoint with, so
# the check runs this same binary in its `healthcheck` mode (see
# cmd/go-htmx/main.go), not a shell command.
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD ["/go-htmx", "healthcheck"]

ENTRYPOINT ["/go-htmx"]
