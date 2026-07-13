// Package notes is the template's example feature vertical slice
// (Story #4): a live, multi-client notes list demonstrating htmx +
// Server-Sent Events + OOB swaps over the data layer Story #3 built.
// Per internal/doc.go's import boundary, this package may import
// internal/platform/* and internal/web/*, but no other
// internal/<feature>/* package may import this one directly.
package notes

import (
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/attested-delivery/go-htmx/internal/platform/db/sqlc"
)

// Handler holds this feature's dependencies: separate read/write query
// wrappers (matching internal/platform/db's Read/Write pool split — see
// that package's doc comment for why writes must go through the write
// pool) and the Broadcaster used to fan new notes out to every connected
// SSE client, including the one that created it.
type Handler struct {
	read   *sqlc.Queries
	write  *sqlc.Queries
	bus    *Broadcaster
	logger *slog.Logger
}

// NewHandler builds a Handler. read/write should wrap db.DB's
// ReadDB()/WriteDB() pools respectively (e.g. sqlc.New(store.ReadDB())).
func NewHandler(read, write *sqlc.Queries, bus *Broadcaster, logger *slog.Logger) *Handler {
	return &Handler{read: read, write: write, bus: bus, logger: logger}
}

// Register mounts this feature's routes on mux. Called from main.go,
// which is free to import both this package and internal/platform/httpserver
// — httpserver itself never imports notes.
func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /{$}", h.handlePage)
	mux.HandleFunc("POST /notes", h.handleCreate)
	mux.HandleFunc("GET /notes/stream", h.handleStream)
}

func (h *Handler) handlePage(w http.ResponseWriter, r *http.Request) {
	list, err := h.read.ListNotes(r.Context())
	if err != nil {
		h.logger.Error("list notes failed", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := Page(list).Render(r.Context(), w); err != nil {
		h.logger.Error("render page failed", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
}

func (h *Handler) handleCreate(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	body := strings.TrimSpace(r.PostFormValue("body"))
	if body == "" {
		http.Error(w, "body is required", http.StatusUnprocessableEntity)
		return
	}

	note, err := h.write.CreateNote(r.Context(), body)
	if err != nil {
		h.logger.Error("create note failed", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	count, err := h.read.CountNotes(r.Context())
	if err != nil {
		h.logger.Error("count notes failed", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	var buf strings.Builder
	if err := Broadcast(note, count).Render(r.Context(), &buf); err != nil {
		h.logger.Error("render broadcast fragment failed", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	h.bus.Publish(buf.String())

	// The POST response itself swaps nothing (hx-swap="none" on the
	// form, see views.templ) — the SSE stream is the single source of
	// truth for updating the list, including for the client that just
	// submitted. Without this, that client would see its own note
	// twice: once from this response, once from the broadcast.
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) handleStream(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.WriteHeader(http.StatusOK)
	flusher.Flush()

	// Subscribe before reading the current state: any note created
	// between this Subscribe and the snapshot query below lands in both
	// the snapshot and (redundantly, but harmlessly — same id, same
	// content) a subsequent broadcast, whereas subscribing after the
	// snapshot could drop a note created in that gap entirely. Given
	// the choice, an occasional duplicate is fine; a silently missing
	// note is not.
	ch, unsubscribe := h.bus.Subscribe()
	defer unsubscribe()

	ctx := r.Context()

	// Full-state sync as the first message: closes the gap between
	// handlePage's render and this connection actually opening (a
	// separate round-trip — the client parses HTML, then opens
	// EventSource). Without this, a note created by another client in
	// that window would never reach this client until a manual reload
	// — see Sync's doc comment in views.templ.
	list, err := h.read.ListNotes(ctx)
	if err != nil {
		h.logger.Error("sync: list notes failed", "error", err)
		return
	}
	count, err := h.read.CountNotes(ctx)
	if err != nil {
		h.logger.Error("sync: count notes failed", "error", err)
		return
	}
	var syncBuf strings.Builder
	if err := Sync(list, count).Render(ctx, &syncBuf); err != nil {
		h.logger.Error("sync: render failed", "error", err)
		return
	}
	if err := writeSSEEvent(w, syncBuf.String()); err != nil {
		h.logger.Warn("sse sync write failed, dropping client", "error", err)
		return
	}
	flusher.Flush()

	for {
		select {
		case <-ctx.Done():
			return
		case msg, ok := <-ch:
			if !ok {
				return
			}
			if err := writeSSEEvent(w, msg); err != nil {
				h.logger.Warn("sse write failed, dropping client", "error", err)
				return
			}
			flusher.Flush()
		}
	}
}

// writeSSEEvent writes msg as one SSE event. Per the SSE spec a
// multi-line payload needs one "data:" line per source line — a naive
// single "data: <msg-with-embedded-newlines>" would truncate the event
// at the first newline.
func writeSSEEvent(w http.ResponseWriter, data string) error {
	for _, line := range strings.Split(data, "\n") {
		if _, err := fmt.Fprintf(w, "data: %s\n", line); err != nil {
			return err
		}
	}
	_, err := fmt.Fprint(w, "\n")
	return err
}
