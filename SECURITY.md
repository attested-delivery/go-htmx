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

Container/release artifact signing and attestation verification instructions
will be added here once this repo ships a release or container pipeline —
see the org's `attested-delivery/rust-template` `SECURITY.md` for the shape
that section will take.
