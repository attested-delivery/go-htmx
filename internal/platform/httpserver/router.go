// Package httpserver provides the template's HTTP layer: a router built
// on Go 1.22+'s enhanced http.ServeMux (method + wildcard patterns) and a
// net/http-compatible middleware chain. chi is a documented escalation
// (drop-in, since chi middleware share this same func(http.Handler)
// http.Handler signature) for teams that outgrow ServeMux's routing
// features (AD-4).
package httpserver

import (
	"io/fs"
	"log/slog"
	"net/http"

	"github.com/attested-delivery/go-htmx/internal/web/assets"
	"github.com/attested-delivery/go-htmx/internal/web/templates"
)

// NewRouter builds the application's top-level handler: route
// registration plus the standard middleware chain. Feature packages add
// their own routes by taking a *http.ServeMux in their constructor and
// calling Handle/HandleFunc on it — see internal/doc.go for the import
// boundary this implies.
func NewRouter(logger *slog.Logger) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /{$}", handleHome(logger))
	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServerFS(staticFS())))

	return Chain(mux,
		Recover(logger),
		Logging(logger),
	)
}

func handleHome(logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if err := templates.Home().Render(r.Context(), w); err != nil {
			logger.Error("render failed", "path", r.URL.Path, "error", err)
		}
	}
}

func staticFS() fs.FS {
	return assets.Static()
}
