//go:build e2e

// Package e2e_test proves the harness itself works: a real Playwright
// browser can load a real page served by a real instance of the app.
// This is deliberately the only test in the root e2e package — everything
// feature-specific lives under e2e/functional, e2e/accessibility, etc.,
// each with their own Page Object.
package e2e_test

import (
	"strings"
	"testing"

	"github.com/attested-delivery/go-htmx/e2e/internal/testapp"
	"github.com/mxschmitt/playwright-go"
)

// TestSmoke_PageLoads is the Smoke-tagged baseline that runs on every
// PR: does a real browser, against a real instance of the app, load the
// notes page at all. Story #79's functional tests build on this same
// pattern for the actual feature flows.
func TestSmoke_PageLoads(t *testing.T) {
	srv := testapp.New(t)

	pw, err := playwright.Run()
	if err != nil {
		t.Fatalf("playwright.Run: %v", err)
	}
	t.Cleanup(func() {
		if err := pw.Stop(); err != nil {
			t.Errorf("pw.Stop: %v", err)
		}
	})

	browser, err := pw.Chromium.Launch()
	if err != nil {
		t.Fatalf("Chromium.Launch: %v", err)
	}
	t.Cleanup(func() {
		if err := browser.Close(); err != nil {
			t.Errorf("browser.Close: %v", err)
		}
	})

	page, err := browser.NewPage()
	if err != nil {
		t.Fatalf("NewPage: %v", err)
	}

	if _, err := page.Goto(srv.URL + "/"); err != nil {
		t.Fatalf("Goto: %v", err)
	}

	title, err := page.Title()
	if err != nil {
		t.Fatalf("Title: %v", err)
	}
	if !strings.Contains(title, "notes") {
		t.Errorf("page title = %q, want it to contain %q", title, "notes")
	}
}
