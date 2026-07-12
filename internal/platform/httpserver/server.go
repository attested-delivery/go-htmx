package httpserver

import (
	"net/http"
	"time"
)

// New builds an *http.Server for addr serving handler, with production
// timeouts set (Go's zero-value http.Server has none, which is a known
// slowloris-class footgun).
func New(addr string, handler http.Handler) *http.Server {
	return &http.Server{
		Addr:              addr,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
}
