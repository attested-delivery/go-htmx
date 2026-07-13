package httpserver

import (
	"log/slog"
	"net/http"
	"time"
)

// Middleware wraps an http.Handler with cross-cutting behavior. It uses
// the standard net/http-compatible signature so any stdlib- or
// chi-compatible middleware can be dropped in without adapting types
// (AD-4).
type Middleware func(http.Handler) http.Handler

// Chain composes middlewares around h, applying them in the order given:
// Chain(a, b)(h) handles a request as a(b(h)).
func Chain(h http.Handler, mws ...Middleware) http.Handler {
	for i := len(mws) - 1; i >= 0; i-- {
		h = mws[i](h)
	}
	return h
}

// Logging logs each request's method, path, status, and duration.
func Logging(logger *slog.Logger) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			rw := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
			next.ServeHTTP(rw, r)
			logger.Info("request",
				"method", r.Method,
				"path", r.URL.Path,
				"status", rw.status,
				"duration", time.Since(start),
			)
		})
	}
}

// Recover converts a panic in a downstream handler into a 500 response
// instead of crashing the server.
func Recover(logger *slog.Logger) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					logger.Error("panic recovered", "error", err, "path", r.URL.Path)
					http.Error(w, "internal server error", http.StatusInternalServerError)
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}

// statusRecorder captures the status code written by a downstream handler
// so Logging can report it after the fact.
type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

// Flush implements http.Flusher by delegating to the wrapped
// ResponseWriter, if it supports flushing. Without this, wrapping a
// ResponseWriter here would silently break streaming responses (SSE,
// Story #4's real-time layer): embedding http.ResponseWriter only
// promotes its own three methods, not Flush, so a handler's
// `w.(http.Flusher)` type assertion would fail on a *statusRecorder even
// though the underlying writer supports it.
func (r *statusRecorder) Flush() {
	if f, ok := r.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}
