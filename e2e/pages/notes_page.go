//go:build e2e

// Package pages holds Page Objects: one file per feature, each wrapping
// playwright-go's Page with named methods targeting that feature's stable
// DOM hooks instead of scattering raw selectors across test files. This
// file (notes_page.go) is the template: to add E2E coverage for your own
// feature, copy it to <feature>_page.go, swap in your feature's own
// stable id="..." hooks (the same convention internal/notes/views.templ
// establishes), and copy e2e/functional/notes_test.go the same way — see
// docs/how-to/add-e2e-coverage.md.
package pages

import "github.com/mxschmitt/playwright-go"

// NotesPage wraps the notes feature's page (internal/notes/views.templ):
// a create form (#note-form), a live count badge (#notes-count), and the
// notes list itself (#notes-list), kept in sync across clients over SSE.
type NotesPage struct {
	page    playwright.Page
	assertT playwright.PlaywrightAssertions
}

// New wraps an already-created playwright-go Page. Callers own the
// Page's lifecycle (creation/close); NotesPage only adds behavior.
func New(page playwright.Page) *NotesPage {
	return &NotesPage{page: page, assertT: playwright.NewPlaywrightAssertions()}
}

// Goto navigates to the notes page at the given base URL (e.g. an
// e2e/internal/testapp httptest.Server's URL).
func (p *NotesPage) Goto(url string) error {
	_, err := p.page.Goto(url)
	return err
}

// CreateNote fills the form and submits it. It does not wait for the
// resulting SSE broadcast to arrive — pair with WaitForNoteVisible or
// WaitForNoteCount to observe the effect.
func (p *NotesPage) CreateNote(body string) error {
	if err := p.page.Locator("#note-form input[name=body]").Fill(body); err != nil {
		return err
	}
	return p.page.Locator("#note-form button[type=submit]").Click()
}

// SubmitEmpty clicks the submit button with the input left empty. The
// input's `required` attribute (views.templ) makes the browser's native
// constraint validation block the request client-side — this exercises
// that behavior, not the server's own 422 (handler.go's belt-and-braces
// check on an empty/whitespace-only body, which a required-bypassing
// client would hit instead).
func (p *NotesPage) SubmitEmpty() error {
	return p.page.Locator("#note-form button[type=submit]").Click()
}

// NoteCount reads the #notes-count badge's current text (e.g. "3 notes")
// without waiting for it to reach any particular value.
func (p *NotesPage) NoteCount() (string, error) {
	return p.page.Locator("#notes-count").InnerText()
}

// WaitForNoteCount auto-retries (playwright-go's LocatorAssertions, not a
// hand-rolled poll loop) until #notes-count reads exactly text, or fails
// after the assertion's default timeout.
func (p *NotesPage) WaitForNoteCount(text string) error {
	return p.assertT.Locator(p.page.Locator("#notes-count")).ToHaveText(text)
}

// WaitForNoteVisible auto-retries until a note whose body contains text
// is visible in #notes-list — the SSE-broadcast-arrived signal, usable
// from any Page sharing the same server (e.g. a second simulated client
// in a multi-client test).
func (p *NotesPage) WaitForNoteVisible(text string) error {
	note := p.page.Locator("#notes-list .note p", playwright.PageLocatorOptions{
		HasText: text,
	}).First()
	return p.assertT.Locator(note).ToBeVisible()
}

// WaitForInputCleared auto-retries until #note-form's body input reads
// empty again. This is the regression test for Epic #67's CSP fix
// (internal/web/assets/static/js/app.js's htmx:afterRequest listener
// replacing hx-on::after:request="this.reset()", which CSP's script-src
// blocks as an implicit eval) — reverting that fix makes this assertion
// time out instead of passing, per Story #78's plan verification step.
func (p *NotesPage) WaitForInputCleared() error {
	return p.assertT.Locator(p.page.Locator("#note-form input[name=body]")).ToHaveValue("")
}

// NoteBodies reads every note's body text currently in #notes-list, in
// display order (newest first — see views.templ's hx-swap="afterbegin").
func (p *NotesPage) NoteBodies() ([]string, error) {
	return p.page.Locator("#notes-list .note p").AllInnerTexts()
}

// NoteBackgroundColor reads the first note's computed background-color
// (e.g. "rgb(255, 255, 255)") — used to confirm a color-scheme-dependent
// Tailwind variant (dark:bg-gray-800) actually took effect, rather than
// just checking the class string is present in the DOM.
func (p *NotesPage) NoteBackgroundColor() (string, error) {
	result, err := p.page.Locator("#notes-list .note").First().Evaluate(
		"el => getComputedStyle(el).backgroundColor", nil,
	)
	if err != nil {
		return "", err
	}
	color, _ := result.(string)
	return color, nil
}
