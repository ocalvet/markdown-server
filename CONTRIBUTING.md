# Contributing to Markdown Server

Thank you for your interest in contributing! We welcome contributions from everyone.

## How to Contribute

### Reporting Bugs

Open an issue with:
- Clear, descriptive title
- Steps to reproduce
- Expected vs actual behavior
- Environment details (OS, Go version, browser)
- Screenshots if applicable

### Suggesting Features

Open an issue describing:
- The feature and its benefits
- Use cases
- Implementation ideas (optional)

### Pull Requests

1. Fork the repository and create a branch from `main`
2. Make your changes following the code style
3. Test thoroughly
4. Commit with clear messages
5. Push to your fork and submit a PR

## Development Setup

```bash
git clone https://github.com/yourusername/markdown-server.git
cd markdown-server/backend
go run main.go
```

Open `http://localhost:8703`

## Code Style

- **Go**: Follow standard conventions, use `gofmt`
- **JavaScript**: ES6+, clear variable names
- **CSS**: Use CSS variables, modular styles

## Testing Checklist

- [ ] Server starts without errors
- [ ] File browsing works
- [ ] Markdown renders correctly
- [ ] Mermaid diagrams render
- [ ] Code highlighting works
- [ ] Theme switching works
- [ ] Responsive on mobile/desktop

## License

By contributing, you agree your contributions will be licensed under the MIT License.
