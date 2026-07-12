module github.com/attested-delivery/go-htmx

go 1.26

// Pinned to the actual current stable release (verified at go.dev/dl,
// 2026-07-12), separate from the `go` line above: the `go` directive is
// deliberately left at the bare minor version because the org's SCA gate
// (OSV-Scanner's govulncheck, run in a container pinned to Go 1.26.2)
// treats it as a hard minimum and fails outright if it's newer than what
// that container ships — the `go` line is a language-version floor, not
// a "use this exact toolchain" pin. `go build`/`go test` invoked directly
// (GOTOOLCHAIN=auto, the default) still resolve and use the toolchain
// pinned below.
toolchain go1.26.5

require github.com/a-h/templ v0.3.1020
