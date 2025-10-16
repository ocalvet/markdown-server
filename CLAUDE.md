# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

A lightweight web server for browsing and viewing markdown files with real-time hot reload. The application serves markdown files from a configurable directory through a Go backend and renders them with a responsive frontend featuring search, syntax highlighting, Mermaid diagrams, and dark/light themes.

## Technology Stack

**Backend**:
- Go 1.25 with fsnotify for file watching
- Single dependency: github.com/fsnotify/fsnotify for hot reload
- Standard library HTTP server

**Frontend**:
- Pure HTML/CSS/JavaScript (no build process)
- Marked.js v12.0.0 for markdown parsing
- Mermaid.js v11.0.2 for diagrams
- Highlight.js v11.11.1 for syntax highlighting
- Fuse.js v7.0.0 for fuzzy search

## Development Commands

### Running Locally

```bash
# Quick start with defaults (port 8703, ./backend/markdown-files)
./run.sh

# Custom directory and port
./run.sh /path/to/markdown/files 8703

# Manual start from backend directory
cd backend
go run main.go

# With environment variables
MARKDOWN_DIR=/path/to/files PORT=8703 IGNORE_PATTERNS="temp,cache" go run main.go
```

### Building

```bash
# Build Go binary
cd backend
go build -o markdown-server

# Build Docker image
docker build -t markdown-server .

# Run Docker container
docker run -p 8703:8703 -v $(pwd)/backend/markdown-files:/app/markdown-files markdown-server
```

### Dependencies

```bash
# Download Go dependencies
cd backend
go mod download

# Tidy dependencies
go mod tidy
```

## Architecture

### Backend Structure (backend/main.go)

The Go server is implemented as a single file with these key components:

1. **File Tree Building**: Recursively scans markdown directory, respecting ignore patterns, and builds a hierarchical JSON structure
2. **Hot Reload System**: Uses fsnotify to watch for file changes and broadcasts reload events to connected clients via Server-Sent Events (SSE)
3. **ReloadBroadcaster**: Manages SSE client connections with sync.Mutex for thread safety
4. **API Endpoints**:
   - `GET /api/files` - Returns hierarchical file tree as JSON
   - `GET /api/file/:path` - Returns raw markdown content with path traversal protection
   - `GET /api/events` - SSE endpoint for hot reload notifications
   - `/` - Static file server for frontend

### Frontend Structure

- **index.html**: Landing page with logo
- **files.html**: File browser view with tree navigation
- **viewer.html**: Main markdown viewer with sidebar, search modal, and content area
- **styles.css**: Theme-aware styling with CSS variables for light/dark modes

### Key Frontend Features

1. **File Navigation**: Sidebar with collapsible folders, active file highlighting, and smooth scrolling to current file
2. **Search System**:
   - Builds client-side index of all files on page load
   - Searches across filenames, paths, and content (first 5000 chars per file)
   - Keyboard shortcut: Ctrl+K / Cmd+K
   - Arrow key navigation through results
3. **Hot Reload**: EventSource connection to /api/events triggers content and file tree refresh on file changes
4. **Theme System**: localStorage persistence, toggles between light/dark modes including syntax highlighting themes
5. **History API**: Uses pushState/popState for seamless navigation without page reloads

### Configuration

Environment variables:
- **MARKDOWN_DIR**: Directory to serve (default: ./markdown-files)
- **PORT**: Server port (default: 8703)
- **IGNORE_PATTERNS**: Comma-separated patterns to ignore (overrides defaults if set)

Default ignore patterns include: node_modules, .git, .vscode, .idea, __pycache__, vendor, dist, build, target, .next, coverage, .DS_Store, and all dot-prefixed directories.

## Important Implementation Details

### Security Features

- Path traversal protection in handleGetFile (filepath.Clean validation)
- Only .md files can be accessed via API
- All user-provided paths are sanitized
- CORS headers configured for API endpoints

### Hot Reload Mechanism

1. fsnotify watches all directories recursively (excluding ignored patterns)
2. On file changes, 100ms debounce timer prevents rapid-fire reloads
3. ReloadBroadcaster sends "reload" event to all connected SSE clients
4. Frontend receives event and refreshes both file tree and current content

### Search Implementation

- Index built on page load by fetching all files
- Content limited to first 5000 characters per file
- Fuse.js configuration: threshold 0.4, weighted keys (name: 2, path: 1.5, content: 1)
- 300ms debounce on search input
- Maximum 20 results displayed

### File Tree Rendering

- Cached in fileTreeData to avoid refetching on navigation
- Folders containing current file auto-expand
- Active file receives .active class and scrolls into view
- Mobile: sidebar collapses on file selection

## Development Notes

- No test suite exists currently
- No linting configuration
- Frontend has no build process - all libraries loaded via CDN
- Docker image uses multi-stage build resulting in ~10-15MB final size
- The backend must be run from backend/ directory as it serves ../frontend
