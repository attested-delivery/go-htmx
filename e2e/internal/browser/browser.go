//go:build e2e

// Package browser starts a Playwright-driven Chromium instance for E2E
// tests, wiring cleanup via t.Cleanup so callers never have to unwind it
// themselves. Shared by every e2e/ package that needs a real browser
// (e2e/functional, e2e/accessibility, ...) instead of each duplicating
// the same launch/cleanup boilerplate.
package browser

import (
	"testing"

	"github.com/mxschmitt/playwright-go"
)

// NewChromium starts Playwright and launches Chromium.
func NewChromium(t *testing.T) playwright.Browser {
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

	b, err := pw.Chromium.Launch()
	if err != nil {
		t.Fatalf("Chromium.Launch: %v", err)
	}
	t.Cleanup(func() {
		if err := b.Close(); err != nil {
			t.Errorf("browser.Close: %v", err)
		}
	})
	return b
}
