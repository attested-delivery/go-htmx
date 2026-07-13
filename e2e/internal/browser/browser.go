//go:build e2e

// Package browser starts a Playwright-driven browser instance for E2E
// tests, wiring cleanup via t.Cleanup so callers never have to unwind it
// themselves. Shared by every e2e/ package that needs a real browser
// (e2e/functional, e2e/accessibility, e2e/visual, ...) instead of each
// duplicating the same launch/cleanup boilerplate.
package browser

import (
	"testing"

	"github.com/mxschmitt/playwright-go"
)

// New starts Playwright and launches the named browser engine
// ("chromium", "firefox", or "webkit") — the three playwright-go
// supports launching. Used by the cross-browser tests
// (e2e/functional/crossbrowser_test.go); most tests only need
// NewChromium.
func New(t *testing.T, name string) playwright.Browser {
	t.Helper()

	pw, err := playwright.Run()
	if err != nil {
		t.Fatalf("playwright.Run: %v", err)
	}
	t.Cleanup(func() {
		if err := pw.Stop(); err != nil {
			t.Errorf("pw.Stop: %v", err)
		}
	})

	var bt playwright.BrowserType
	switch name {
	case "chromium":
		bt = pw.Chromium
	case "firefox":
		bt = pw.Firefox
	case "webkit":
		bt = pw.WebKit
	default:
		t.Fatalf("unknown browser %q, want chromium, firefox, or webkit", name)
	}

	b, err := bt.Launch()
	if err != nil {
		t.Fatalf("%s.Launch: %v", name, err)
	}
	t.Cleanup(func() {
		if err := b.Close(); err != nil {
			t.Errorf("browser.Close: %v", err)
		}
	})
	return b
}

// NewChromium is a convenience wrapper for New(t, "chromium") — the
// common case used by every test that only needs one, deterministic
// browser (e2e/functional, e2e/accessibility, e2e/visual).
func NewChromium(t *testing.T) playwright.Browser {
	t.Helper()
	return New(t, "chromium")
}
