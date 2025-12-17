# Security Policy

## Supported Versions

| Version | Supported          |
| ------- | ------------------ |
| 0.x.x   | :white_check_mark: |

## Reporting a Vulnerability

We take security seriously. If you discover a security vulnerability in GoFHIR, please report it responsibly.

### How to Report

1. **Do not** open a public GitHub issue for security vulnerabilities
2. Send an email to the project maintainers with details about the vulnerability
3. Include steps to reproduce the issue if possible
4. Allow reasonable time for us to address the issue before public disclosure

### What to Expect

- We will acknowledge receipt of your report within 48 hours
- We will provide an estimated timeline for a fix
- We will notify you when the issue is resolved
- We will credit you in the release notes (unless you prefer to remain anonymous)

### Scope

This security policy applies to:

- The GoFHIR library code
- Code generation templates
- Official documentation

### Out of Scope

- Third-party dependencies (please report to respective maintainers)
- Issues in user applications built with GoFHIR

## Security Best Practices

When using GoFHIR in your applications:

- Keep GoFHIR updated to the latest version
- Validate and sanitize all input data
- Follow FHIR security guidelines for handling healthcare data
- Implement proper access controls in your application

## Dependencies

We regularly update dependencies to address known vulnerabilities. You can check for vulnerabilities using:

```bash
go list -m -json all | nancy sleuth
```

Or using Go's built-in vulnerability checker:

```bash
govulncheck ./...
```
