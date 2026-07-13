# Vendored JavaScript (test-only)

Unlike `internal/web/assets/static/js/VENDORED.md`'s `htmx.min.js`/
`hx-sse.min.js` — which ship in the app binary and are downloaded from
GitHub release assets — `axe.min.js` here is **test-only** (injected into
a page by `e2e/accessibility/axe_test.go` via Playwright's `AddScriptTag`,
never served by the app itself), so it lives under `e2e/testdata/` instead.

**Source differs from the htmx pattern deliberately**: `dequelabs/axe-core`
does not publish GitHub release assets at all (verified empirically —
every one of its 5 most recent releases has zero attached assets, and
`axe.min.js` is not committed anywhere in the tagged source tree). Its
canonical, authoritative distribution point is the npm registry, so
`axe.min.js` here is extracted from the `axe-core` npm package's tarball,
fetched directly via `curl` (never `npm install`/`npm view` — no Node/npm
toolchain dependency is introduced, matching AD-9). The tarball itself is
checksum-verified against **two independently published values from the
npm registry's own metadata** before extraction — not just a single
self-consistency check — then the extracted `axe.min.js` file's own
SHA-256 is what's pinned below.

When bumping: repeat the same two-hash tarball verification, then update
this file.

```sh
curl -s https://registry.npmjs.org/axe-core/<version> \
  | python3 -c "import json,sys; d=json.load(sys.stdin)['dist']; print(d['shasum']); print(d['integrity'])"

curl -sLo axe-core.tgz https://registry.npmjs.org/axe-core/-/axe-core-<version>.tgz
shasum -a 1 axe-core.tgz                                    # must match dist.shasum above
openssl dgst -sha512 -binary axe-core.tgz | openssl base64 -A  # must match dist.integrity above

tar -xzf axe-core.tgz package/axe.min.js
shasum -a 256 package/axe.min.js                             # update the table below
```

| File | Source | Package version | SHA-256 |
| --- | --- | --- | --- |
| `axe.min.js` | `https://registry.npmjs.org/axe-core/-/axe-core-4.12.1.tgz` (`package/axe.min.js`) | `axe-core@4.12.1` | `66a8aaa95a8b044a7fd74a5435873bf04ff65a1ca75567c921b7509742085a14` |

axe-core 4.12.1 supports the `wcag22aa` rule tag (confirmed by grepping
the vendored file itself) — `e2e/accessibility/axe_test.go` runs `axe.run()`
scoped to that tag set, per Story #80's WCAG 2.2 AA acceptance criterion.
