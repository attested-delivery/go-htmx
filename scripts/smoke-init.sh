#!/usr/bin/env bash
# The template's real acceptance test (Story #6, Task #27): copy the
# tree to a temp dir, run `just init` with a throwaway identity, build
# + test the copy, and grep-gate for any leftover template identity
# string. Run via `just smoke-init`.
set -euo pipefail

repo_root=$(git rev-parse --show-toplevel)
smoke_dir=$(mktemp -d)
trap 'rm -rf "$smoke_dir"' EXIT

echo "==> copying tree to $smoke_dir"
cd "$repo_root"
# Tracked + untracked-but-not-gitignored (e.g. tools/init itself, or any
# other work in progress not yet committed) — a straight `git archive`
# would miss the latter and silently smoke-test a stale HEAD instead of
# the actual working tree.
git ls-files --cached --others --exclude-standard >"$smoke_dir/.filelist"
tar -cf "$smoke_dir/.tree.tar" -T "$smoke_dir/.filelist"
tar -xf "$smoke_dir/.tree.tar" -C "$smoke_dir"
rm -f "$smoke_dir/.filelist" "$smoke_dir/.tree.tar"

cd "$smoke_dir"

echo "==> just init (throwaway identity)"
just init smoketest github.com/smoke/test

echo "==> just build"
just build

echo "==> just test"
just test

echo "==> just docker-build"
# Proves the Dockerfile's cmd/<name> build path actually got rewritten
# (Task #52 found this the hard way: Dockerfile has no extension, so it
# needs its own tools/init allowName entry, not just allowExt -- a
# leftover ./cmd/go-htmx reference here fails the docker build outright
# once cmd/go-htmx has been renamed, which build/test above can't catch).
just docker-build

echo "==> grep-gate: no leftover template identity"
if grep -rlIE 'go-htmx|attested-delivery/go-htmx' \
    --include='*.go' --include='*.templ' --include='*.md' \
    --include='*.yml' --include='*.yaml' --include='*.golden' \
    --include='justfile' --include='go.mod' --include='Dockerfile' .
then
    echo "FAIL: template identity leaked into the initialized copy (see files above)" >&2
    exit 1
fi

echo "OK: smoke-init passed — init, build, test, docker-build, and identity grep-gate all green"
