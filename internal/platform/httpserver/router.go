// Package httpserver provides the template's HTTP layer: a router built
// on Go 1.22+'s enhanced http.ServeMux (method + wildcard patterns) and a
// net/http-compatible middleware chain. chi is a documented escalation
// (drop-in, since chi middleware share this same func(http.Handler)
// http.Handler signature) for teams that outgrow ServeMux's routing
// features (AD-4).
package httpserver

import (
	"log/slog"
	"net/http"

	"github.com/attested-delivery/go-htmx/internal/web/assets"
)

// NewMux builds the application's route table with only the routes
// platform itself owns (static assets). It deliberately does not own
// "/" or any application route — per internal/doc.go's import boundary,
// httpserver (platform) must never import a feature package, so feature
// packages register their own routes on the returned mux from the
// caller (main.go), which is free to import both. See Wrap for the
// middleware chain around whatever the caller ends up registering.
func NewMux() *http.ServeMux {
	mux := http.NewServeMux()
	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServerFS(assets.Static())))
	return mux
}

// Wrap applies the standard middleware chain around mux (or any
// http.Handler), after every route — platform's and every feature's —
// has been registered.
func Wrap(h http.Handler, logger *slog.Logger) http.Handler {
	return Chain(h,
		Recover(logger),
		Logging(logger),
	)
}
