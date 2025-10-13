# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.1.0] - 2025-10-13

### Added
- Full-text search functionality with Fuse.js
  - Fuzzy search across file names, paths, and content
  - Keyboard shortcut (Ctrl+K / Cmd+K) to open search
  - Live search results with 300ms debounce
  - Content snippets with match highlighting
  - Keyboard navigation (arrow keys, Enter, Esc)
  - Result count display (up to 20 results)
  - Responsive design for mobile and desktop
- Search button in header of viewer and files pages
- Search feature showcase on landing page
- Automatic search index building on page load

### Changed
- Added Fuse.js v7.0.0 as frontend dependency
- Updated header layout to include search button
- Enhanced documentation with search usage guide

## [1.0.0] - 2025-10-09

### Added
- Initial release
- Go backend server with file watching
- GitHub Flavored Markdown rendering
- Mermaid diagram support
- Syntax highlighting for 180+ languages
- Dark/light theme toggle with localStorage persistence
- Responsive sidebar navigation
- Hot reload functionality
- Configurable directory via `MARKDOWN_DIR` environment variable
- Configurable port via `PORT` environment variable
- Customizable ignore patterns via `IGNORE_PATTERNS` environment variable
- Default ignore patterns for common directories (node_modules, .git, etc.)
- Docker support with multi-stage builds
- SVG logo and favicon
- Quick start script (run.sh)
- Comprehensive documentation

### Features
- Recursive file tree browsing
- Breadcrumb navigation
- Collapsible folders in sidebar
- Active file highlighting
- Mobile-friendly with collapsible sidebar
- SSE-based hot reload
- Path traversal protection
- CORS support
- Optimized Docker image (~10-15MB)
