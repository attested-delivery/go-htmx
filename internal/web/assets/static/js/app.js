// First-party, not vendored (see VENDORED.md for htmx itself) — a
// same-origin external script so it runs under the Content-Security-Policy
// this template ships (default-src 'self', no unsafe-eval): htmx's
// hx-on::after:request="this.reset()" attribute form evaluates its
// argument via new AsyncFunction(...), which CSP blocks outright. Using
// htmx's own custom event with a real addEventListener call here is
// functionally identical without needing eval.
//
// The event name is htmx:after:request (colon-separated), not the v1/v2
// name htmx:afterRequest (camelCase) — htmx v4.0.0-beta5 (see
// VENDORED.md) renamed its whole event set to colon-separated form.
// The camelCase name silently never fires (addEventListener doesn't
// error on an event name nothing ever dispatches), so this listener
// never ran and the form never reset — caught by
// e2e/functional/notes_test.go's TestSmoke_ResetAfterSubmit.
document.getElementById("note-form")?.addEventListener("htmx:after:request", function () {
	this.reset();
});
