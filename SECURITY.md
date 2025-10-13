# Security Policy

## Supported Versions

We release patches for security vulnerabilities for the following versions:

| Version | Supported          |
| ------- | ------------------ |
| 1.0.x   | :white_check_mark: |

## Reporting a Vulnerability

We take the security of Markdown Server seriously. If you discover a security vulnerability, please follow these steps:

### How to Report

1. **Do NOT** open a public GitHub issue for security vulnerabilities
2. Email details of the vulnerability to the maintainers (create an issue with label `security` and mark it as private when GitHub private vulnerability reporting is enabled)
3. Include detailed steps to reproduce the vulnerability
4. Include the potential impact and severity assessment

### What to Include

Please provide the following information:

- Type of vulnerability (e.g., path traversal, XSS, code injection)
- Full paths of affected source files
- Location of the affected code (tag/branch/commit or direct URL)
- Step-by-step instructions to reproduce the issue
- Proof-of-concept or exploit code (if possible)
- Impact of the issue and how an attacker might exploit it

### Response Timeline

- We will acknowledge receipt of your vulnerability report within 48 hours
- We will provide a detailed response within 7 days, including next steps
- We will keep you informed about the progress toward a fix
- We will notify you when the vulnerability is fixed

### Disclosure Policy

- We request that you do not publicly disclose the vulnerability until we have had a chance to address it
- We will credit you in the security advisory (unless you wish to remain anonymous)
- Once the vulnerability is patched, we will publish a security advisory

## Security Best Practices

When deploying Markdown Server:

1. **Access Control**: Use a reverse proxy (nginx, caddy) to add authentication if exposing publicly
2. **File Permissions**: Ensure the markdown directory has appropriate read-only permissions
3. **Network Security**: Use HTTPS in production (terminate SSL at reverse proxy)
4. **Docker**: Keep the Docker image updated to the latest version
5. **Environment Variables**: Do not commit `.env` files with sensitive configuration
6. **Input Validation**: Only serve markdown files from trusted sources

## Known Security Features

The application includes the following security features:

- Path traversal protection (prevents access to files outside the markdown directory)
- File type restriction (only `.md` files can be accessed)
- CORS configuration for cross-origin requests
- File path sanitization
- No execution of user-provided code on the server

## Dependencies

This project has minimal dependencies:

- **Backend**: Go standard library + fsnotify (file system notifications)
- **Frontend**: CDN-delivered libraries (Marked.js, Mermaid.js, Highlight.js)

We regularly monitor security advisories for our dependencies.
