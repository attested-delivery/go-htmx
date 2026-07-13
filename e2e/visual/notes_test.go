//go:build e2e

package visual

import (
	"flag"
	"os"
	"path/filepath"
	"testing"

	"github.com/attested-delivery/go-htmx/e2e/internal/browser"
	"github.com/attested-delivery/go-htmx/e2e/internal/testapp"
	"github.com/attested-delivery/go-htmx/e2e/pages"
	"github.com/mxschmitt/playwright-go"
)

// update mirrors the existing goldie/`just test-golden-update` convention
// (see internal/notes' golden tests and the justfile's test-golden-update
// recipe): re-running with -update regenerates the baseline instead of
// comparing against it. This flag is scoped to this test binary only
// (each Go package's test binary gets its own flag set), so it doesn't
// collide with goldie's own -update flag registered in internal/notes.
var update = flag.Bool("update", false, "update visual regression baselines instead of comparing against them")

// diffThreshold is the maximum fraction of differing pixels a screenshot
// may have before failing — loose enough to absorb minor font-hinting/
// anti-aliasing differences across CI runs, tight enough to catch a real
// layout or styling regression.
const diffThreshold = 0.01

var colorSchemeCases = []struct {
	name        string
	colorScheme *playwright.ColorScheme
}{
	{name: "light", colorScheme: playwright.ColorSchemeLight},
	{name: "dark", colorScheme: playwright.ColorSchemeDark},
}

// TestVisual_NotesPage screenshots the notes page (with one note present)
// in both light and dark mode and compares each against its checked-in
// baseline at testdata/<name>.golden.png. Full-suite tier only — not
// Smoke-tagged, since screenshot comparison is inherently more prone to
// environment-specific noise than a functional assertion, and doesn't
// need to gate every PR.
func TestVisual_NotesPage(t *testing.T) {
	for _, tc := range colorSchemeCases {
		t.Run(tc.name, func(t *testing.T) {
			// testapp.New must run before b.NewPage so cleanup unwinds
			// in the right order (page/SSE connection closes before the
			// server does) — see e2e/accessibility/axe_test.go's
			// newNotesPageWithData doc comment for the failure mode this
			// avoids.
			srv := testapp.New(t)

			b := browser.NewChromium(t)
			page, err := b.NewPage(playwright.BrowserNewPageOptions{
				ColorScheme: tc.colorScheme,
				Viewport:    &playwright.Size{Width: 1024, Height: 768},
			})
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
			if err := notes.CreateNote("visual regression check note"); err != nil {
				t.Fatalf("CreateNote: %v", err)
			}
			if err := notes.WaitForNoteVisible("visual regression check note"); err != nil {
				t.Fatalf("note never appeared: %v", err)
			}

			// Each note's <time> element renders the real creation
			// timestamp, which is different on every run by design
			// (internal/notes/views.templ) — mask it out of the
			// screenshot rather than let it produce a spurious diff
			// every single run.
			screenshot, err := page.Screenshot(playwright.PageScreenshotOptions{
				Animations: playwright.ScreenshotAnimationsDisabled,
				Mask:       []playwright.Locator{page.Locator(".note time")},
			})
			if err != nil {
				t.Fatalf("Screenshot: %v", err)
			}

			actualPath := filepath.Join(t.TempDir(), "actual.png")
			if err := os.WriteFile(actualPath, screenshot, 0o644); err != nil {
				t.Fatalf("write actual screenshot: %v", err)
			}

			goldenPath := filepath.Join("testdata", tc.name+".golden.png")

			if *update {
				if err := copyFile(actualPath, goldenPath); err != nil {
					t.Fatalf("update baseline: %v", err)
				}
				t.Logf("updated baseline %s", goldenPath)
				return
			}

			if _, err := os.Stat(goldenPath); os.IsNotExist(err) {
				t.Fatalf("no baseline at %s — run `just test-visual-update` to create one, review the diff, then commit it", goldenPath)
			}

			diffPath := filepath.Join("testdata", tc.name+".diff.png")
			if _, err := Compare(goldenPath, actualPath, diffPath, diffThreshold); err != nil {
				t.Errorf("%v — diff overlay written to %s (uploaded as a CI artifact on failure)", err, diffPath)
				return
			}
			// No diff: remove any stale diff overlay from a previous
			// failing run so a passing run doesn't leave one behind.
			_ = os.Remove(diffPath)
		})
	}
}

func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0o644)
}
