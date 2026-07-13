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

## Gate Coverage

Disposition against the org's 12-gate map (see the docs site's
`github-native-attested-quality-gates` spec), so every gap is stated
explicitly rather than silently absent:

| Gate | Status | Where |
| --- | --- | --- |
| SAST | Covered | CodeQL, `quality-gates.yml`'s `sast` job (required check) + re-run and seam-signed at release |
| SCA / dependency | Covered | OSV-Scanner + dependency review, `quality-gates.yml`'s `sca` job (required check) + re-run and seam-signed at release |
| Secret detection | Covered | GitHub secret scanning + push protection, enabled (repo setting, confirmed via API) |
| Container / image scan | Covered | Trivy against the built image, `release.yml`'s `gate-image` job + seam-signed |
| IaC / misconfiguration | Covered | Trivy filesystem scan, `quality-gates.yml`'s `trivy` job (required check) + re-run and seam-signed at release |
| License compliance | Covered | Same Trivy job as IaC (`scan-iac: true` covers both) |
| SBOM | Covered | Binaries: `anchore/sbom-action` + `actions/attest-sbom` at release. Image: `sign-and-attest.yml`'s own SBOM attestation |
| Vuln disposition (VEX) | **Gap, documented** | Not wired. The org's `reusable-vex.yml` is opt-in; revisit if/when a real finding needs a disposition record, not before |
| Build provenance (SLSA) | Covered | Binaries: `actions/attest-build-provenance` (L3). Image: `sign-and-attest.yml` (L3, separate signer identity) |
| Supply-chain posture | Covered | OpenSSF Scorecard, `quality-gates.yml`'s `posture` job (push/schedule only, per Scorecard's own default-branch requirement) |
| Peer review | **Gap, documented, org-wide** | `requiredApprovingReviewCount: 0` on `main` — confirmed via `get_branch_protection`. This matches the org's current auto-merge pattern across its public repos, not a go-htmx-specific gap; revisit at the org level if that pattern changes |
| Load / performance | **Not applicable, documented** | Needs a running app to test against; no deployed instance of this template exists to point k6 at. A real deployment's own repo should wire `reusable-k6.yml` against itself |
| DAST | **Not applicable, documented** | Same reasoning as load/performance — `reusable-zap.yml` needs a running app. Documented opt-in for a real deployment, not this template repo |

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
shasum -a 256 -c go-htmx-<version>-checksums.txt
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

## Verifying the Container Image

Every tagged release also publishes a distroless container image to
`ghcr.io/attested-delivery/go-htmx`, built from the `Dockerfile` at the
repo root (`FROM` pinned by digest, not tag — see that file). Unlike the
binaries above, this is the **image-only** attested path: signed by the
org's central `sign-and-attest.yml` (a different signer identity from
the static-artifact path), keyed to the pushed image's digest, not its
tag:

```sh
IMAGE="ghcr.io/attested-delivery/go-htmx@<digest>"   # resolve the digest
                                                       # from the release,
                                                       # never trust a
                                                       # mutable tag alone
SIGNER="attested-delivery/.github/.github/workflows/sign-and-attest.yml"

cosign verify "$IMAGE" \
  --certificate-identity-regexp "^https://github.com/${SIGNER}@.*$" \
  --certificate-oidc-issuer https://token.actions.githubusercontent.com

gh attestation verify "oci://$IMAGE" \
  --repo attested-delivery/go-htmx \
  --signer-workflow "$SIGNER" \
  --predicate-type https://slsa.dev/provenance/v1

cosign verify-attestation "$IMAGE" \
  --type cyclonedx \
  --certificate-identity-regexp "^https://github.com/${SIGNER}@.*$" \
  --certificate-oidc-issuer https://token.actions.githubusercontent.com
```

The image's own Trivy scan verdict is seam-signed and bound to the same
digest, same "signed != passed" caveat as above:

```sh
gh attestation verify "oci://$IMAGE" --owner attested-delivery \
  --signer-workflow attested-delivery/.github/.github/workflows/reusable-attest-scan.yml \
  --predicate-type https://attested-delivery.github.io/attestations/container-scan/v1
```
