# Markdown Server

A lightweight web server for browsing and viewing markdown files with support for Mermaid diagrams, syntax highlighting, and dark/light themes.

## Features

- **Knowledge Graph**: Interactive visualization of connections between markdown files
  - Click-to-navigate graph with physics simulation
  - Auto-highlights current file and shows link statistics
  - Discover orphaned files and connection patterns
  - Supports both `[text](link.md)` and `[[wiki-style]]` links
- **Collapsible Sidebar**: Toggle sidebar on all screen sizes with state persistence
- **Full-Text Search**: Fuzzy search across all files and content with keyboard shortcut (Ctrl+K / Cmd+K)
- **Markdown Rendering**: Full GitHub Flavored Markdown support
- **Mermaid Diagrams**: Create flowcharts, sequence diagrams, and more
- **Syntax Highlighting**: Support for 180+ programming languages
- **Dark/Light Themes**: Toggle between themes with localStorage persistence
- **Recursive File Browsing**: Navigate through nested folder structures with sidebar navigation
- **Hot Reload**: Automatically updates when markdown files change
- **Responsive Design**: Mobile-friendly interface
- **Configurable**: Set directory and port via CLI flags or environment variables
- **Offline-Ready**: All frontend dependencies bundled locally (no CDN required)
- **Docker Support**: Easy deployment with optimized container image (~10-15MB)

## Technology Stack

**Backend**:
- Go 1.25.2
- fsnotify for file watching (only external dependency)

**Frontend** (all dependencies bundled locally):
- Marked.js v12.0.0 (Markdown parsing)
- Mermaid.js v11.0.2 (Diagram rendering)
- Highlight.js v11.11.1 (Syntax highlighting)
- Fuse.js v7.0.0 (Fuzzy search)
- Vis.js v9.1.9 (Knowledge graph visualization)
- js-yaml v4.1.0 (YAML frontmatter parsing)

## Quick Start

### Running Locally

#### Quick Start with Script

```bash
cd markdown-server

# Use default directory (./backend/markdown-files) and port (8703)
./run.sh

# Specify custom directory
./run.sh /path/to/your/markdown/files

# Specify custom directory and port
./run.sh /path/to/your/markdown/files 9000

# Specify custom directory, port, and ignore patterns
./run.sh /path/to/your/markdown/files 9000 "node_modules,.git,temp"
```

#### Manual Start

```bash
cd markdown-server/backend

# Use defaults (current directory, port 8703)
go run main.go

# With command-line flags
go run main.go -p 9000 -d /path/to/files

# With positional argument
go run main.go /path/to/files

# With environment variables
MARKDOWN_DIR=/path/to/files PORT=8703 go run main.go
```

**Note:** Command-line flags take precedence over environment variables.

Open your browser to `http://localhost:8703`

### Installing with Fuseki Installer

The project includes a Fuseki installer for easy setup:

```bash
# Download and run the installer
curl -sSL https://raw.githubusercontent.com/ocalvet/markdown-server/main/fuseki-install.sh | bash

# Or clone and run locally
git clone https://github.com/ocalvet/markdown-server.git
cd markdown-server
./fuseki-install.sh
```

After installation, use the `fuki` command:

```bash
# Show help
fuki --help

# Start server with markdown directory
fuki -d /path/to/your/markdown

# Start server with custom port
fuki -p 9000 -d /path/to/your/markdown

# Use current directory (default)
fuki
```

See [FUSEKI_INSTALL.md](FUSEKI_INSTALL.md) for detailed instructions.

### Running with Docker

1. Build the Docker image:
```bash
cd markdown-server
docker build -t markdown-server .
```

2. Run the container:
```bash
# Using default directory
docker run -p 8703:8703 -v $(pwd)/backend/markdown-files:/app/markdown-files markdown-server

# Using custom directory, port, and ignore patterns
docker run -p 9000:9000 \
  -e MARKDOWN_DIR=/app/docs \
  -e PORT=9000 \
  -e IGNORE_PATTERNS="temp,cache" \
  -v /path/to/your/markdown:/app/docs \
  markdown-server
```

3. Access the server:
```
http://localhost:8703
```

## Project Structure

```
markdown-server/
├── backend/
│   ├── main.go              # Go server
│   ├── go.mod               # Go module definition
│   └── markdown-files/      # Default markdown files
├── frontend/
│   ├── index.html           # Landing page
│   ├── files.html           # File browser
│   ├── viewer.html          # Markdown viewer
│   ├── styles.css           # Styles with theme support
│   └── vendor/              # Bundled frontend dependencies
│       ├── css/             # highlight.js themes, vis-network
│       └── js/              # marked, mermaid, highlight, fuse, vis-network, js-yaml
├── fuseki-install.sh        # Installation script (installs as 'fuki')
├── fuseki-uninstall.sh      # Uninstallation script
├── Dockerfile               # Multi-stage Docker build
├── .dockerignore
└── README.md
```

## API Endpoints

### GET /api/files
Lists all markdown files in the directory tree.

**Response**:
```json
[
  {
    "path": "welcome.md",
    "name": "welcome.md",
    "isDir": false
  },
  {
    "path": "tutorials",
    "name": "tutorials",
    "isDir": true,
    "children": [...]
  }
]
```

### GET /api/file/:path
Retrieves the content of a specific markdown file.

**Example**: `/api/file/tutorials/mermaid-diagrams.md`

**Response**: Raw markdown content

### GET /api/graph
Returns the knowledge graph data showing connections between files.

**Response**:
```json
{
  "nodes": [
    {
      "id": "file1.md",
      "label": "file1.md",
      "title": "file1.md"
    }
  ],
  "edges": [
    {
      "from": "file1.md",
      "to": "file2.md"
    }
  ]
}
```

### GET /api/events
Server-Sent Events endpoint for hot reload notifications.

## Configuration

The server can be configured using **command-line flags** or **environment variables**. Flags take precedence over environment variables, which take precedence over defaults.

### Command-Line Flags

| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| `--port` | `-p` | Server port | `8703` |
| `--dir` | `-d` | Markdown directory | Current directory (`.`) |
| `--help` | `-h` | Show usage information | |

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `MARKDOWN_DIR` | Directory containing markdown files | Current directory (`.`) |
| `PORT` | Server port | `8703` |
| `IGNORE_PATTERNS` | Comma-separated list of patterns to ignore | See below |

### Default Ignore Patterns

The server automatically ignores these common directories:
- `node_modules`
- `.git`, `.svn`, `.hg`
- `.idea`, `.vscode`
- `__pycache__`, `.pytest_cache`, `.mypy_cache`
- `vendor`, `dist`, `build`, `target`
- `.next`, `.nuxt`, `coverage`
- `.DS_Store`, `Thumbs.db`

### Examples

```bash
# Using CLI flags (recommended)
fuki -d /path/to/files -p 9000

# Using environment variables
MARKDOWN_DIR=/path/to/files PORT=9000 fuki

# Flags override environment variables
PORT=7777 fuki -p 9000  # Uses port 9000

# Custom ignore patterns (env var only, replaces defaults)
IGNORE_PATTERNS="node_modules,.git,temp,cache" fuki -d /path/to/files
```

## Adding Your Own Markdown Files

1. Place your `.md` files in the `backend/markdown-files/` directory
2. Organize them in folders as needed
3. The server will automatically discover all files recursively
4. Refresh the browser to see new files

## Search

The server includes a powerful full-text search feature powered by Fuse.js for fuzzy matching.

### How to Use

- **Keyboard Shortcut**: Press `Ctrl+K` (Windows/Linux) or `Cmd+K` (Mac) from any page
- **Search Button**: Click the 🔍 icon in the header
- **Live Results**: Search results update as you type (with 300ms debounce)
- **Navigation**: Use arrow keys (↑↓) to navigate results, Enter to open, Esc to close

### What's Searched

The search indexes:
- **File names** (highest priority)
- **File paths** (medium priority)
- **File content** (up to 5000 characters per file)

### Features

- **Fuzzy Matching**: Finds results even with typos or partial matches
- **Content Snippets**: Shows relevant excerpts with highlighted matches
- **Result Count**: Displays the number of matching files (up to 20 results)
- **Keyboard Navigation**: Fully accessible via keyboard
- **Responsive**: Works on desktop and mobile

### Search Index

The search index is built automatically when you load the viewer or files page. The index includes all markdown files and updates when the page reloads.

## Knowledge Graph

Visualize connections between your markdown files as an interactive network graph.

### How to Use

- **Open Graph**: Click the 🕸️ icon in the header
- **Single-click node**: Navigate to that file (graph stays open, highlights update)
- **Double-click**: Close the graph modal
- **Drag nodes**: Rearrange the graph layout
- **Zoom/Pan**: Use mouse wheel to zoom, drag background to pan
- **Navigation controls**: Use built-in controls in bottom-right corner

### Features

- **Current File Highlighting**: Your current file appears larger and in red/pink
- **Auto-centering**: Graph automatically centers on your current file
- **Link Statistics**: Shows outgoing and incoming link counts for current file
- **Orphaned Files**: Identifies files with no connections to other files
- **Physics Simulation**: Organic graph layout that settles over time
- **Theme Integration**: Adapts to light/dark mode

### Link Support

The knowledge graph detects two types of links:
- **Standard Markdown**: `[Link Text](path/to/file.md)`
- **Wiki-style**: `[[filename]]` or `[[path/to/file]]`

Relative paths are resolved correctly based on the source file location.

## Sidebar

The sidebar can be collapsed on all screen sizes for a distraction-free reading experience.

### How to Use

- **Toggle**: Click the ☰ (hamburger) button in the header
- **State Persistence**: Your preference is saved in localStorage
- **Default**: Sidebar is open by default on desktop, closed on mobile

## Theme Colors

### Light Theme
- Background: `#fafafa`
- Text: `#1a1a1a`
- Accent: `#0066cc`

### Dark Theme
- Background: `#1a1a1a`
- Text: `#e0e0e0`
- Accent: `#4da6ff`

## Docker Image Size

The production Docker image is approximately **10-15MB** thanks to:
- Multi-stage build
- Alpine Linux base
- Single Go binary with no runtime dependencies

## Security Features

- Path traversal protection
- Only `.md` files can be accessed
- CORS headers configured
- File path sanitization

## Browser Support

Modern browsers with ES6+ support:
- Chrome 60+
- Firefox 60+
- Safari 12+
- Edge 79+

## License

MIT
