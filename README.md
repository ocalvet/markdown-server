# Markdown Server

A lightweight web server for browsing and viewing markdown files with support for Mermaid diagrams, syntax highlighting, and dark/light themes.

## Features

- **Markdown Rendering**: Full GitHub Flavored Markdown support
- **Mermaid Diagrams**: Create flowcharts, sequence diagrams, and more
- **Syntax Highlighting**: Support for 180+ programming languages
- **Dark/Light Themes**: Toggle between themes with localStorage persistence
- **Recursive File Browsing**: Navigate through nested folder structures with sidebar navigation
- **Hot Reload**: Automatically updates when markdown files change
- **Responsive Design**: Mobile-friendly with collapsible sidebar
- **Configurable**: Set directory and port via environment variables
- **Docker Support**: Easy deployment with optimized container image (~10-15MB)

## Technology Stack

**Backend**:
- Go 1.25.2
- Standard library only (no external dependencies)

**Frontend**:
- Marked.js v12.0.0 (Markdown parsing)
- Mermaid.js v11.0.2 (Diagram rendering)
- Highlight.js v11.11.1 (Syntax highlighting)

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

# Use defaults
go run main.go

# With environment variables
MARKDOWN_DIR=/path/to/files PORT=8703 go run main.go
```

Open your browser to `http://localhost:8703`

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
│   └── markdown-files/      # Your markdown files (recursive)
│       ├── welcome.md
│       ├── tutorials/
│       │   ├── mermaid-diagrams.md
│       │   └── code-examples.md
│       └── examples/
│           └── markdown-features.md
├── frontend/
│   ├── index.html           # File browser
│   ├── viewer.html          # Markdown viewer
│   ├── styles.css           # Styles with theme support
│   └── app.js               # (optional)
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

## Configuration

The server can be configured using environment variables:

- **MARKDOWN_DIR**: Directory containing markdown files (default: `./markdown-files`)
- **PORT**: Server port (default: `8703`)
- **IGNORE_PATTERNS**: Comma-separated list of patterns to ignore (optional)

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
# Set custom directory
export MARKDOWN_DIR=/path/to/your/markdown/files
go run main.go

# Set custom port
export PORT=9000
go run main.go

# Custom ignore patterns (replaces defaults)
export IGNORE_PATTERNS="node_modules,.git,temp,cache"
go run main.go

# All together
MARKDOWN_DIR=/path/to/files PORT=9000 IGNORE_PATTERNS="build,dist" go run main.go
```

## Adding Your Own Markdown Files

1. Place your `.md` files in the `backend/markdown-files/` directory
2. Organize them in folders as needed
3. The server will automatically discover all files recursively
4. Refresh the browser to see new files

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
