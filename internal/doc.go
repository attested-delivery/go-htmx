// Package internal is the root of this template's private application code.
//
// Layout (AD-6): code lives in one of two kinds of package tree under
// internal/:
//
//   - internal/platform/*  — infrastructure: config, HTTP server/router,
//     database access, embedded assets. Platform packages MUST NOT import
//     from internal/<feature>/*.
//   - internal/<feature>/* — a self-contained vertical slice (handlers,
//     data access, view models for one product feature). Feature packages
//     MAY import internal/platform/* and internal/web/* but MUST NOT import
//     another internal/<feature>/* package directly; cross-feature
//     composition happens in cmd/go-htmx/main.go at wiring time.
//   - internal/web/*        — shared UI: templ templates, embedded static
//     assets. Treated as a platform package for import-boundary purposes.
//
// This boundary is enforced by the depguard rules in .golangci.yml, not by
// Go's compiler alone (Go only enforces that internal/* isn't importable
// from outside the module, not the cross-package rules above).
package internal
