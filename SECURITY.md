# Security Policy

## Supported Versions

| Version | Supported |
|---------|-----------|
| latest  | Yes       |
| < latest | No       |

## Reporting a Vulnerability

**Please do not report security vulnerabilities through public GitHub issues.**

Instead, please report them via [GitHub Security Advisories](https://github.com/attested-delivery/go-htmx/security/advisories/new).

### What to Include

- A description of the vulnerability
- Steps to reproduce the issue
- Potential impact
- Suggested fix (if any)

### Response Timeline

- **Acknowledgment**: Within 48 hours of the report
- **Initial assessment**: Within 1 week
- **Fix and disclosure**: Coordinated with the reporter, typically within 90 days

### Disclosure Policy

We follow responsible disclosure practices:

1. The reporter privately notifies us of the vulnerability.
2. We work together to understand and fix the issue.
3. We release a patched version.
4. The vulnerability is publicly disclosed after users have had time to update.

### Scope

This policy applies to the go-htmx module and its published artifacts.
Third-party dependencies are managed via Go modules and audited through the
CI pipeline.

## Security Measures

This project employs several security practices:

- **govulncheck**: Audits dependencies and the module's own code paths for
  known Go vulnerabilities
- **Dependabot**: Automated dependency updates for `go.sum` and GitHub Actions,
  with a 7-day cooldown on newly published versions
- **SHA-pinned actions**: Every GitHub Actions `uses:` reference is pinned to
  a full commit SHA, enforced by a `pin-check` CI gate
- **SARIF-normalized scanning**: SAST (CodeQL), SCA (OSV-Scanner), and
  IaC/license (Trivy) findings all land in the repository's Security tab

## Verifying Release Artifacts

Every tagged release (`.github/workflows/release.yml`) ships five
platform binaries (`linux-amd64`, `linux-arm64`, `macos-amd64`,
`macos-arm64`, `windows-amd64.exe`) plus a `go-htmx-<version>-checksums.txt`
manifest, named `go-htmx-<version>-<platform>` per the org's release
naming standard (see `docs/reference/release-artifacts.md`). This is a
static-artifact release — a plain Go binary, no container image — so
provenance and the SBOM are produced by GitHub's own attestation actions
running in this repo's `release.yml`, not the org's central image signer.

Verification is independent and keyless: no shared secret, just the `gh`
CLI with read access to this repo.

```sh
gh release download <tag> --repo attested-delivery/go-htmx
sha256sum -c go-htmx-<version>-checksums.txt
```

**SLSA provenance** (per binary):

```sh
gh attestation verify go-htmx-<version>-<platform> \
  --repo attested-delivery/go-htmx \
  --signer-workflow attested-delivery/go-htmx/.github/workflows/release.yml \
  --predicate-type https://slsa.dev/provenance/v1
```

**CycloneDX SBOM binding** (per binary):

```sh
gh attestation verify go-htmx-<version>-<platform> \
  --repo attested-delivery/go-htmx \
  --signer-workflow attested-delivery/go-htmx/.github/workflows/release.yml \
  --predicate-type https://cyclonedx.org/bom
```

## Verifying Quality-Gate Attestations

SAST (CodeQL), SCA (OSV-Scanner), and IaC/license (Trivy) are re-run at
release time against the exact tagged commit and seam-signed by the org's
central `reusable-attest-scan.yml`, bound to the sha256 digest of the
release's own `go-htmx-<version>-checksums.txt` — the one artifact whose
content (a hash of every shipped binary) identifies this exact release.
Under SLSA Build L3 the signer identity is the *central signer workflow*,
not this repo, so verification pins `--owner` + `--signer-workflow`
together (one signer per command) rather than `--repo`:

```sh
SEAM=attested-delivery/.github/.github/workflows/reusable-attest-scan.yml

gh attestation verify go-htmx-<version>-checksums.txt --owner attested-delivery \
  --signer-workflow "$SEAM" \
  --predicate-type https://attested-delivery.github.io/attestations/sast/v1

gh attestation verify go-htmx-<version>-checksums.txt --owner attested-delivery \
  --signer-workflow "$SEAM" \
  --predicate-type https://attested-delivery.github.io/attestations/sca/v1

gh attestation verify go-htmx-<version>-checksums.txt --owner attested-delivery \
  --signer-workflow "$SEAM" \
  --predicate-type https://attested-delivery.github.io/attestations/iac-license/v1
```

Signed is not the same as passed: each attestation proves the named gate
*ran and recorded a verdict* against this exact release, not that the
verdict was clean — inspect the predicate body (e.g. via
`gh attestation verify --format json`) to read the actual result.
