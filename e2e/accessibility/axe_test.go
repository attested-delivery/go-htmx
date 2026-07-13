//go:build e2e

// Package accessibility runs axe-core (e2e/testdata/axe.min.js, vendored
// from the axe-core npm package — see e2e/testdata/VENDORED.md) against
// the notes page and fails on any WCAG 2.2 AA violation. Not
// Smoke-tagged: this is a full-suite-tier check (merge to main), not a
// PR-blocking one — see justfile and .github/workflows/ci.yml.
package accessibility

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/attested-delivery/go-htmx/e2e/internal/browser"
	"github.com/attested-delivery/go-htmx/e2e/internal/testapp"
	"github.com/attested-delivery/go-htmx/e2e/pages"
	"github.com/mxschmitt/playwright-go"
)

// axeViolation mirrors the fields of axe-core's Result objects
// (https://github.com/dequelabs/axe-core/blob/master/doc/API.md#results-object)
// that are useful for a failing test's diagnostic, not the full shape.
type axeViolation struct {
	ID          string `json:"id"`
	Impact      string `json:"impact"`
	Description string `json:"description"`
	HelpURL     string `json:"helpUrl"`
	Nodes       []struct {
		Target []string `json:"target"`
		HTML   string   `json:"html"`
	} `json:"nodes"`
}

func (v axeViolation) String() string {
	var b strings.Builder
	b.WriteString(v.ID + " (" + v.Impact + "): " + v.Description + "\n  " + v.HelpURL)
	for _, n := range v.Nodes {
		b.WriteString("\n  - " + strings.Join(n.Target, " ") + ": " + n.HTML)
	}
	return b.String()
}

// runAxe injects the vendored axe-core build and runs it scoped to the
// WCAG 2.2 AA rule tag (Story #80's acceptance criterion), returning any
// violations found.
func runAxe(t *testing.T, page playwright.Page) []axeViolation {
	t.Helper()

	script, err := os.ReadFile("../testdata/axe.min.js")
	if err != nil {
		t.Fatalf("read axe.min.js: %v", err)
	}
	content := string(script)
	if _, err := page.AddScriptTag(playwright.PageAddScriptTagOptions{Content: &content}); err != nil {
		t.Fatalf("AddScriptTag: %v", err)
	}

	raw, err := page.Evaluate(`async () => {
		const results = await axe.run(document, { runOnly: { type: 'tag', values: ['wcag22aa'] } });
		return JSON.stringify(results.violations);
	}`, nil)
	if err != nil {
		t.Fatalf("axe.run: %v", err)
	}

	violationsJSON, ok := raw.(string)
	if !ok {
		t.Fatalf("axe.run result was %T, want string", raw)
	}

	var violations []axeViolation
	if err := json.Unmarshal([]byte(violationsJSON), &violations); err != nil {
		t.Fatalf("unmarshal axe violations: %v", err)
	}
	return violations
}

// newNotesPageWithData starts a real server, opens a new page against
// it, and creates one note so the check exercises real rendered content
// (form, count badge, note list item), not just the empty state.
//
// The server must be created (testapp.New) before the page, so its
// t.Cleanup (httptest.Server.Close) is registered before the page's
// (page.Close) — t.Cleanup runs LIFO, so the page (and the SSE
// connection its notes-stream element holds open) closes first, and the
// server isn't left blocking in Close() waiting for a still-active
// connection that will never end on its own. Reversing this ordering
// hung the whole test suite until go test's 10-minute default timeout
// killed it (confirmed empirically before this fix).
func newNotesPageWithData(t *testing.T, b playwright.Browser, opts ...playwright.BrowserNewPageOptions) playwright.Page {
	t.Helper()

	srv := testapp.New(t)

	page, err := b.NewPage(opts...)
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
	if err := notes.CreateNote("accessibility check note"); err != nil {
		t.Fatalf("CreateNote: %v", err)
	}
	if err := notes.WaitForNoteVisible("accessibility check note"); err != nil {
		t.Fatalf("note never appeared: %v", err)
	}
	return page
}

func TestNotesPage_NoWCAG22AAViolations(t *testing.T) {
	cases := []struct {
		name        string
		colorScheme *playwright.ColorScheme
	}{
		{name: "light", colorScheme: playwright.ColorSchemeLight},
		{name: "dark", colorScheme: playwright.ColorSchemeDark},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			b := browser.NewChromium(t)

			// BypassCSP is required here: this repo ships a strict CSP
			// (default-src 'self', no unsafe-inline — see
			// internal/platform/httpserver/middleware.go's
			// SecurityHeaders), and AddScriptTag injects axe.min.js as
			// an inline <script>. Without bypassing CSP, the browser
			// silently blocks that inline script from executing, its
			// load event never fires, and Playwright's AddScriptTag
			// call hangs waiting for it (confirmed empirically: it hit
			// go test's 10-minute default timeout before this fix).
			page := newNotesPageWithData(t, b, playwright.BrowserNewPageOptions{
				ColorScheme: tc.colorScheme,
				BypassCSP:   playwright.Bool(true),
			})

			violations := runAxe(t, page)
			if len(violations) == 0 {
				return
			}

			var b2 strings.Builder
			for _, v := range violations {
				b2.WriteString(v.String() + "\n")
			}
			t.Errorf("%d WCAG 2.2 AA violation(s) in %s mode:\n%s", len(violations), tc.name, b2.String())
		})
	}
}
