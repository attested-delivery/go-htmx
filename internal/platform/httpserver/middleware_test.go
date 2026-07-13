package httpserver

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSecurityHeaders(t *testing.T) {
	handler := SecurityHeaders()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	want := map[string]string{
		"X-Content-Type-Options":  "nosniff",
		"Referrer-Policy":         "no-referrer",
		"X-Frame-Options":         "DENY",
		"Content-Security-Policy": "default-src 'self'",
	}
	for header, wantValue := range want {
		if got := rec.Header().Get(header); got != wantValue {
			t.Errorf("%s = %q, want %q", header, got, wantValue)
		}
	}
}

// TestSecurityHeadersPreservesFlusher proves SecurityHeaders doesn't wrap
// the ResponseWriter in a type that drops http.Flusher support — it only
// needs to set headers before calling next, but a naive implementation
// wrapping w in a struct without forwarding Flush would silently break
// SSE streaming (Story #4's real-time layer) the same way an unwrapped
// statusRecorder would have, per middleware.go's own Flush doc comment.
func TestSecurityHeadersPreservesFlusher(t *testing.T) {
	var sawFlusher bool
	handler := SecurityHeaders()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, sawFlusher = w.(http.Flusher)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if !sawFlusher {
		t.Fatal("downstream handler should still see an http.Flusher-capable ResponseWriter")
	}
}
