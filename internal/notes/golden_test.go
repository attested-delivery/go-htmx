package notes

import (
	"bytes"
	"context"
	"testing"
	"time"

	"github.com/sebdah/goldie/v2"

	"github.com/attested-delivery/go-htmx/internal/platform/db/sqlc"
)

// fixedNotes returns deterministic test fixtures — golden-file
// comparison needs byte-stable output, so CreatedAt is a fixed instant
// rather than time.Now() (AD-8's golden-file tier).
func fixedNotes() []sqlc.Note {
	t1 := time.Date(2026, 1, 15, 10, 30, 0, 0, time.UTC)
	t2 := t1.Add(5 * time.Minute)
	return []sqlc.Note{
		{ID: 2, Body: "second note", CreatedAt: t2},
		{ID: 1, Body: "first note", CreatedAt: t1},
	}
}

func TestGoldenPage(t *testing.T) {
	g := goldie.New(t)

	var buf bytes.Buffer
	if err := Page(fixedNotes()).Render(context.Background(), &buf); err != nil {
		t.Fatalf("render: %v", err)
	}
	g.Assert(t, "page", buf.Bytes())
}

func TestGoldenPageEmpty(t *testing.T) {
	g := goldie.New(t)

	var buf bytes.Buffer
	if err := Page(nil).Render(context.Background(), &buf); err != nil {
		t.Fatalf("render: %v", err)
	}
	g.Assert(t, "page-empty", buf.Bytes())
}

func TestGoldenNoteRow(t *testing.T) {
	g := goldie.New(t)

	var buf bytes.Buffer
	note := fixedNotes()[1]
	if err := NoteRow(note).Render(context.Background(), &buf); err != nil {
		t.Fatalf("render: %v", err)
	}
	g.Assert(t, "note-row", buf.Bytes())
}

func TestGoldenBroadcast(t *testing.T) {
	g := goldie.New(t)

	var buf bytes.Buffer
	note := fixedNotes()[1]
	if err := Broadcast(note, 3).Render(context.Background(), &buf); err != nil {
		t.Fatalf("render: %v", err)
	}
	g.Assert(t, "broadcast", buf.Bytes())
}

func TestGoldenSync(t *testing.T) {
	g := goldie.New(t)

	var buf bytes.Buffer
	if err := Sync(fixedNotes(), 2).Render(context.Background(), &buf); err != nil {
		t.Fatalf("render: %v", err)
	}
	g.Assert(t, "sync", buf.Bytes())
}
