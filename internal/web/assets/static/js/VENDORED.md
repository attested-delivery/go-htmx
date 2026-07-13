# Vendored JavaScript

Downloaded directly from `bigskysoftware/htmx`'s GitHub release assets
(not npm) and committed rather than fetched at build time, matching
AD-9's single-binary/no-external-fetch-at-startup goal. When bumping:
download the new release's assets from the same URLs, recompute the
checksums below, and update both this file and the version noted here.

htmx is pinned to **v4.0.0-beta5** (2026-06-26) — the latest htmx *is*
still beta at time of writing (v2.0.x is the latest stable line); this is
a deliberate choice to build against v4's current API, not an oversight.

| File | Source | SHA-256 |
| --- | --- | --- |
| `htmx.min.js` | https://github.com/bigskysoftware/htmx/releases/download/v4.0.0-beta5/htmx.min.js | `192d2d425dda6834bd15973a10f55940cea217a3a840f3f819ffd16063be9a68` |
| `hx-sse.min.js` | https://github.com/bigskysoftware/htmx/releases/download/v4.0.0-beta5/hx-sse.min.js | `fcc844a52779d8450c1c4796feea8d038943f908b9ee974322c276230e6c86cc` |

`hx-sse.min.js` is htmx v4's Server-Sent Events extension (Story #4,
AD-4's real-time layer). Unlike htmx v1/v2's extension model, v4
extensions self-register globally on load — no `hx-ext="sse"` opt-in
attribute is needed, both scripts just need to be present.
