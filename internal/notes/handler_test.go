package notes

import (
	"bufio"
	"context"
	"errors"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/attested-delivery/go-htmx/internal/platform/db"
	"github.com/attested-delivery/go-htmx/internal/platform/db/sqlc"
)

// newTestHandler opens a fresh in-memory SQLite database (AD-8's test
// triad), migrates it, and returns a ready-to-use Handler plus a
// t.Cleanup-registered close. Every test gets its own isolated database
// — see db.Open's doc comment for why ":memory:" is safe to share
// across a single Handler's read/write pools but not across tests.
func newTestHandler(t *testing.T) *Handler {
	t.Helper()

	store, err := db.Open(":memory:")
	if err != nil {
		t.Fatalf("db.Open: %v", err)
	}
	t.Cleanup(func() { _ = store.Close() })

	if err := db.Migrate(store.WriteDB()); err != nil {
		t.Fatalf("db.Migrate: %v", err)
	}

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	return NewHandler(sqlc.New(store.ReadDB()), sqlc.New(store.WriteDB()), NewBroadcaster(), logger)
}

func TestHandlePage(t *testing.T) {
	h := newTestHandler(t)
	mux := http.NewServeMux()
	h.Register(mux)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	body := rec.Body.String()
	for _, want := range []string{`id="notes-stream"`, `id="notes-list"`, `id="notes-count"`, `0 notes`} {
		if !strings.Contains(body, want) {
			t.Errorf("response body missing %q\nfull body: %s", want, body)
		}
	}
}

func TestHandleCreate(t *testing.T) {
	h := newTestHandler(t)
	mux := http.NewServeMux()
	h.Register(mux)

	t.Run("valid body", func(t *testing.T) {
		form := url.Values{"body": {"first note"}}
		req := httptest.NewRequest(http.MethodPost, "/notes", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)

		if rec.Code != http.StatusNoContent {
			t.Fatalf("status = %d, want %d (body: %s)", rec.Code, http.StatusNoContent, rec.Body.String())
		}
		if rec.Body.Len() != 0 {
			t.Errorf("expected empty body on 204, got %q", rec.Body.String())
		}

		notes, err := h.read.ListNotes(context.Background())
		if err != nil {
			t.Fatalf("ListNotes: %v", err)
		}
		if len(notes) != 1 || notes[0].Body != "first note" {
			t.Fatalf("expected exactly one note with body %q, got %+v", "first note", notes)
		}
	})

	t.Run("empty body is rejected", func(t *testing.T) {
		form := url.Values{"body": {"   "}}
		req := httptest.NewRequest(http.MethodPost, "/notes", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)

		if rec.Code != http.StatusUnprocessableEntity {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnprocessableEntity)
		}
	})
}

// TestHandleStreamSyncAndBroadcast exercises the SSE endpoint end to
// end: a note created before the client connects must arrive via the
// initial sync message (the fix for the replay gap found in review —
// see internal/notes/views.templ's Sync doc comment), and a note
// created after connecting must arrive via a live broadcast. Needs a
// real listening server (httptest.NewServer, not ResponseRecorder)
// since streaming + explicit Flush doesn't work against a recorder.
func TestHandleStreamSyncAndBroadcast(t *testing.T) {
	h := newTestHandler(t)
	mux := http.NewServeMux()
	h.Register(mux)
	srv := httptest.NewServer(mux)
	defer srv.Close()

	// Created before any client connects — must reach this test's
	// client via the sync message, not be silently dropped.
	if _, err := h.write.CreateNote(context.Background(), "pre-existing"); err != nil {
		t.Fatalf("CreateNote (pre-existing): %v", err)
	}

	req, err := http.NewRequest(http.MethodGet, srv.URL+"/notes/stream", nil)
	if err != nil {
		t.Fatalf("NewRequest: %v", err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("connect to /notes/stream: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}
	if ct := resp.Header.Get("Content-Type"); ct != "text/event-stream" {
		t.Fatalf("Content-Type = %q, want %q", ct, "text/event-stream")
	}

	events := make(chan string, 4)
	go readSSEEvents(t, resp.Body, events)

	sync := readEventWithTimeout(t, events, 2*time.Second)
	if !strings.Contains(sync, "pre-existing") {
		t.Fatalf("sync message missing the pre-existing note: %q", sync)
	}
	if !strings.Contains(sync, `hx-swap-oob="innerHTML"`) {
		t.Errorf("sync message should OOB-replace #notes-list wholesale: %q", sync)
	}

	// Created after connecting, via the real POST /notes endpoint (not
	// h.write.CreateNote directly) — only handleCreate publishes to the
	// broadcaster, so this must go through the actual HTTP path to
	// reach this client via a live broadcast, distinct from the sync
	// message.
	createResp, err := http.PostForm(srv.URL+"/notes", url.Values{"body": {"live note"}})
	if err != nil {
		t.Fatalf("POST /notes: %v", err)
	}
	_ = createResp.Body.Close()
	if createResp.StatusCode != http.StatusNoContent {
		t.Fatalf("POST /notes status = %d, want %d", createResp.StatusCode, http.StatusNoContent)
	}

	broadcast := readEventWithTimeout(t, events, 2*time.Second)
	if !strings.Contains(broadcast, "live note") {
		t.Fatalf("broadcast missing the live note: %q", broadcast)
	}
	// Checked as independent conditions, not one exact adjacent substring
	// — asserting a fixed attribute order/adjacency makes this brittle
	// against unrelated markup changes (e.g. adding a class attribute)
	// that don't affect the actual behavior being tested: the count
	// badge OOB-updating to the right value.
	if !strings.Contains(broadcast, `id="notes-count"`) {
		t.Errorf("broadcast should include the #notes-count element: %q", broadcast)
	}
	if !strings.Contains(broadcast, `hx-swap-oob="true"`) {
		t.Errorf("broadcast should OOB-swap the count element: %q", broadcast)
	}
	if !strings.Contains(broadcast, "2 notes") {
		t.Errorf("broadcast should OOB-update the count to 2: %q", broadcast)
	}
}

// readSSEEvents accumulates "data: ..." lines into whole events
// (blank-line-terminated, per the SSE spec) and sends each completed
// event on ch. Runs in its own goroutine, so it reports scan failures
// via t.Errorf — safe to call from any goroutine, unlike t.Fatalf,
// which must run on the test's own goroutine. Without the enlarged
// buffer, a fragment exceeding bufio.Scanner's default 64K token limit
// would make Scan() return false with no event ever sent — the caller
// would see a bare "no SSE event received" timeout with no indication
// that a buffer overflow, not a missing broadcast, was the actual cause.
func readSSEEvents(t *testing.T, r io.Reader, ch chan<- string) {
	t.Helper()
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	var event strings.Builder
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			if event.Len() > 0 {
				ch <- event.String()
				event.Reset()
			}
			continue
		}
		event.WriteString(strings.TrimPrefix(line, "data: "))
		event.WriteString("\n")
	}
	// Ignore net.ErrClosed: the test closes resp.Body (deferred) once
	// it's done reading events, which unblocks this goroutine's
	// in-flight Read with exactly this error — expected teardown, not
	// a real failure. Anything else (notably bufio.ErrTooLong, if a
	// fragment ever exceeded the buffer above) is a genuine test bug.
	if err := scanner.Err(); err != nil && !errors.Is(err, net.ErrClosed) {
		t.Errorf("readSSEEvents: scanner error: %v", err)
	}
}

func readEventWithTimeout(t *testing.T, ch <-chan string, timeout time.Duration) string {
	t.Helper()
	select {
	case ev := <-ch:
		return ev
	case <-time.After(timeout):
		t.Fatalf("no SSE event received within %s", timeout)
		return ""
	}
}
