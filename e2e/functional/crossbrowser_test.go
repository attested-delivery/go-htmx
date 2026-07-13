//go:build e2e

package functional

import (
	"fmt"
	"testing"
	"time"

	"github.com/attested-delivery/go-htmx/e2e/internal/browser"
)

// TestCrossBrowser_CreateNoteAppears runs the same core create-note flow
// TestSmoke_CreateNoteAppears covers, against all three engines
// playwright-go can launch (Chromium, Firefox, WebKit). Full-suite tier
// only (just e2e-full, merge to main) — not Smoke-tagged, since running
// three browser engines on every PR would be slow for marginal signal
// beyond Chromium alone.
func TestCrossBrowser_CreateNoteAppears(t *testing.T) {
	for _, name := range []string{"chromium", "firefox", "webkit"} {
		t.Run(name, func(t *testing.T) {
			br := browser.New(t, name)
			notes, _ := newNotesPage(t, br)

			body := fmt.Sprintf("crossbrowser note %d", time.Now().UnixNano())
			if err := notes.CreateNote(body); err != nil {
				t.Fatalf("CreateNote: %v", err)
			}

			if err := notes.WaitForNoteVisible(body); err != nil {
				t.Errorf("note never appeared in %s: %v", name, err)
			}
			if err := notes.WaitForNoteCount("1 note"); err != nil {
				t.Errorf("count badge never updated in %s: %v", name, err)
			}
		})
	}
}
