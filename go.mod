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

require (
	github.com/a-h/templ v0.3.1020
	github.com/pressly/goose/v3 v3.27.2
	github.com/sebdah/goldie/v2 v2.8.0
	modernc.org/sqlite v1.53.0
)

require (
	github.com/dustin/go-humanize v1.0.1 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/mattn/go-isatty v0.0.21 // indirect
	github.com/mfridman/interpolate v0.0.2 // indirect
	github.com/ncruces/go-strftime v1.0.0 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/remyoudompheng/bigfft v0.0.0-20230129092748-24d4a6f8daec // indirect
	github.com/sergi/go-diff v1.0.0 // indirect
	github.com/sethvargo/go-retry v0.3.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	golang.org/x/sync v0.21.0 // indirect
	golang.org/x/sys v0.45.0 // indirect
	modernc.org/libc v1.73.4 // indirect
	modernc.org/mathutil v1.7.1 // indirect
	modernc.org/memory v1.11.0 // indirect
)
