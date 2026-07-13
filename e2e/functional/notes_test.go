//go:build e2e

// Package functional exercises the notes feature's real user-facing
// flows through a real browser via the NotesPage Page Object
// (e2e/pages/notes_page.go). Tests named "Smoke" are the lean,
// PR-blocking subset (just e2e-smoke); the rest run only on merge to
// main (just e2e-full) — see justfile and .github/workflows/ci.yml.
package functional

import (
	"fmt"
	"testing"
	"time"

	"github.com/attested-delivery/go-htmx/e2e/internal/browser"
	"github.com/attested-delivery/go-htmx/e2e/internal/testapp"
	"github.com/attested-delivery/go-htmx/e2e/pages"
	"github.com/mxschmitt/playwright-go"
)

// newNotesPage starts a real server (e2e/internal/testapp), opens a new
// page against it, and returns a ready-to-use NotesPage.
func newNotesPage(t *testing.T, br playwright.Browser, opts ...playwright.BrowserNewPageOptions) (*pages.NotesPage, string) {
	t.Helper()

	srv := testapp.New(t)

	page, err := br.NewPage(opts...)
	if err != nil {
		t.Fatalf("NewPage: %v", err)
	}
	t.Cleanup(func() {
		if err := page.Close(); err != nil {
			t.Errorf("page.Close: %v", err)
		}
	})

	notes := pages.NewNotesPage(page)
	if err := notes.Goto(srv.URL + "/"); err != nil {
		t.Fatalf("Goto: %v", err)
	}
	return notes, srv.URL
}

// TestSmoke_CreateNoteAppears is the core flow: create a note, see it
// appear in the list, see the count badge update. PR-blocking.
func TestSmoke_CreateNoteAppears(t *testing.T) {
	br := browser.NewChromium(t)
	notes, _ := newNotesPage(t, br)

	body := fmt.Sprintf("smoke note %d", time.Now().UnixNano())
	if err := notes.CreateNote(body); err != nil {
		t.Fatalf("CreateNote: %v", err)
	}

	if err := notes.WaitForNoteVisible(body); err != nil {
		t.Errorf("note never appeared: %v", err)
	}
	if err := notes.WaitForNoteCount("1 note"); err != nil {
		t.Errorf("count badge never updated: %v", err)
	}
}

// TestSmoke_ResetAfterSubmit is the regression test for Epic #67's CSP
// fix: hx-on::after:request="this.reset()" evaluates via an implicit
// eval CSP blocks, so it silently never ran under CSP. See
// e2e/pages/notes_page.go's WaitForInputCleared doc comment for how to
// falsify this test against the original bug. PR-blocking because a
// regression here means every note submission after the first leaves
// stale text in the input.
func TestSmoke_ResetAfterSubmit(t *testing.T) {
	br := browser.NewChromium(t)
	notes, _ := newNotesPage(t, br)

	if err := notes.CreateNote("this should clear"); err != nil {
		t.Fatalf("CreateNote: %v", err)
	}

	if err := notes.WaitForInputCleared(); err != nil {
		t.Errorf("input was not cleared after submit: %v", err)
	}
}

// TestSmoke_MultiClientBroadcast proves the real-time layer works: a
// note created by one simulated client is pushed, over SSE, to a second
// client that never submitted anything. PR-blocking — this is the
// feature's core value proposition (AD-4).
func TestSmoke_MultiClientBroadcast(t *testing.T) {
	br := browser.NewChromium(t)

	srv := testapp.New(t)

	pageA, err := br.NewPage()
	if err != nil {
		t.Fatalf("NewPage (A): %v", err)
	}
	t.Cleanup(func() {
		if err := pageA.Close(); err != nil {
			t.Errorf("pageA.Close: %v", err)
		}
	})
	notesA := pages.NewNotesPage(pageA)
	if err := notesA.Goto(srv.URL + "/"); err != nil {
		t.Fatalf("Goto (A): %v", err)
	}

	pageB, err := br.NewPage()
	if err != nil {
		t.Fatalf("NewPage (B): %v", err)
	}
	t.Cleanup(func() {
		if err := pageB.Close(); err != nil {
			t.Errorf("pageB.Close: %v", err)
		}
	})
	notesB := pages.NewNotesPage(pageB)
	if err := notesB.Goto(srv.URL + "/"); err != nil {
		t.Fatalf("Goto (B): %v", err)
	}

	body := fmt.Sprintf("broadcast note %d", time.Now().UnixNano())
	if err := notesA.CreateNote(body); err != nil {
		t.Fatalf("CreateNote: %v", err)
	}

	if err := notesB.WaitForNoteVisible(body); err != nil {
		t.Errorf("client B never received the SSE broadcast: %v", err)
	}
}

// TestEmptyBodyRejected exercises the input's `required` attribute
// (views.templ): submitting with an empty body must be blocked
// client-side by the browser's native constraint validation, so no note
// is ever created. Full-suite tier only — not part of the PR-blocking
// smoke subset.
func TestEmptyBodyRejected(t *testing.T) {
	br := browser.NewChromium(t)
	notes, _ := newNotesPage(t, br)

	if err := notes.SubmitEmpty(); err != nil {
		t.Fatalf("SubmitEmpty: %v", err)
	}

	count, err := notes.NoteCount()
	if err != nil {
		t.Fatalf("NoteCount: %v", err)
	}
	if count != "0 notes" {
		t.Errorf("note count = %q, want %q (empty submission should be rejected)", count, "0 notes")
	}
}

// TestDarkModeRendering emulates prefers-color-scheme: dark and asserts
// the dark: Tailwind variant actually applies — checking computed style,
// not just class presence, since a CSS load failure would leave the
// dark: classes present in markup but visually inert. Full-suite tier
// only.
func TestDarkModeRendering(t *testing.T) {
	br := browser.NewChromium(t)
	notes, _ := newNotesPage(t, br, playwright.BrowserNewPageOptions{
		ColorScheme: playwright.ColorSchemeDark,
	})

	if err := notes.CreateNote("dark mode check"); err != nil {
		t.Fatalf("CreateNote: %v", err)
	}
	if err := notes.WaitForNoteVisible("dark mode check"); err != nil {
		t.Fatalf("note never appeared: %v", err)
	}

	// internal/notes/views.templ: light-mode body background is
	// bg-white (rgb(255, 255, 255)); dark:bg-gray-800 overrides it
	// under prefers-color-scheme: dark. Reading the computed style
	// confirms Tailwind actually compiled and shipped the dark:
	// variant, not just that the class string is present in the DOM.
	bg, err := notes.NoteBackgroundColor()
	if err != nil {
		t.Fatalf("NoteBackgroundColor: %v", err)
	}
	if bg == "rgb(255, 255, 255)" {
		t.Errorf("note background = %q under dark color scheme, want a dark background (dark:bg-gray-800 did not apply)", bg)
	}
}
