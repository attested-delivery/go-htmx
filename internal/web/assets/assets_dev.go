//go:build dev

package assets

import (
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
)

// Static returns the template's static asset tree read directly off disk,
// so edits under internal/web/assets/static are visible without a rebuild.
// Build with `-tags dev` (see the justfile's `run` recipe) to select this
// file over assets_embed.go.
func Static() fs.FS {
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		panic("assets: could not determine source file location for dev asset root")
	}
	return os.DirFS(filepath.Join(filepath.Dir(thisFile), "static"))
}
