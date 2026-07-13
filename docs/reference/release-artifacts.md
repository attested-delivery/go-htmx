---
id: urn:mif:go-htmx:reference:release-artifacts
type: semantic
created: 2026-07-12T00:00:00Z
namespace: go-htmx/docs/reference
tags: [reference, release, artifacts, attestation]
title: "Reference: release artifact naming and verification"
relationships:
  - type: relates-to
    target: docs/explanation/architecture.md
provenance:
  '@type': Provenance
  agent: claude-code/claude-sonnet-5
  wasGeneratedBy:
    '@id': urn:mif:activity:claude-code-session:91f08ccf-1fbc-4d5c-843d-9ff6d4050ce8
    '@type': prov:Activity
  trustLevel: user_stated
  agentVersion: 2.1.207
modified: '2026-07-13T02:20:34.957Z'
---

# Reference: release artifact naming and verification

This repo follows the `attested-delivery` org's release and artifact
standards. The naming convention below applies to any release this
template's CI produces; the verification commands apply once the
attested release pipeline (Story #9) is in place.

## Naming

| Kind | Pattern | Example |
| --- | --- | --- |
| Release build | `{name}-{version}-{platform}` | `myapp-0.1.0-linux-amd64` |
| Nightly build | `{name}-nightly-YYYYMMDD-SHORT_SHA-{platform}` | `myapp-nightly-20260712-a1b2c3d-linux-amd64` |

Supported platforms: `linux-amd64`, `linux-arm64`, `macos-amd64`,
`macos-arm64`, `windows-amd64.exe`. Every artifact in one release shares
the same `{name}-{version}` prefix.

## SLSA build level and SBOM format

- **SLSA Build Level 3** — provenance signed by a workflow the calling
  repo cannot modify, isolating the build identity from the source
  repo.
- **SBOM format: CycloneDX.**

## Verification

Once this repo's attested release pipeline is wired (Story #9), every
release artifact carries SLSA provenance and a CycloneDX SBOM,
independently verifiable with no shared secret:

```sh
gh attestation verify <artifact-file> --owner attested-delivery
```

This is a static (non-container) artifact path: provenance comes from
`actions/attest-build-provenance` and the SBOM from `anchore/sbom-action`
(Syft) + `actions/attest-sbom`, both running in this repo's own release
workflow — not the image-signing (`sign-and-attest.yml`) path, which
applies to container images only.

## Status

The naming convention and verification command above are the org
standard this repo will conform to; the release workflow that actually
produces signed artifacts under this naming is Story #9's scope and is
not yet implemented as of this document's writing.
