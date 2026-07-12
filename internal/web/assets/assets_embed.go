//go:build !dev

// Package assets exposes the template's static files (CSS, and — once
// Story #4 lands — vendored htmx) as an fs.FS. This file is the default
// (prod) build: assets are embedded into the binary at compile time using
// the directive below, so `just build` produces a single self-contained
// executable (AD-9, NFR-4). Build with `-tags dev` to swap in
// assets_dev.go instead, which serves the same tree straight off disk for
// live-reload iteration.
package assets

import (
	"embed"
	"io/fs"
)

//go:embed static
var embedded embed.FS

// Static returns the template's static asset tree, rooted at "static".
func Static() fs.FS {
	sub, err := fs.Sub(embedded, "static")
	if err != nil {
		// static/ is embedded above; fs.Sub can only fail on a bad path,
		// which would be a compile-time-caught programmer error.
		panic(err)
	}
	return sub
}
