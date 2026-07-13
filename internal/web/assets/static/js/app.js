// First-party, not vendored (see js/VENDORED.md for htmx itself) — a
// same-origin external script so it runs under the Content-Security-Policy
// this template ships (default-src 'self', no unsafe-eval): htmx's
// hx-on::after:request="this.reset()" attribute form evaluates its
// argument via new AsyncFunction(...), which CSP blocks outright. Using
// htmx's own htmx:afterRequest custom event with a real addEventListener
// call here is functionally identical without needing eval.
document.getElementById("note-form")?.addEventListener("htmx:afterRequest", function () {
	this.reset();
});
