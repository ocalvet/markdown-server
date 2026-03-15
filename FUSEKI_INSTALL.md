# Fuseki Installer

This directory contains installation scripts for **Fuseki**, a Go-inspired markdown server with knowledge graph visualization.

## What is Fuseki?

**Fuseki** (不記) is named after the Go term for the opening phase of the game, representing the beginning of your markdown knowledge base. It's a lightweight web server for browsing and viewing markdown files with:

- **Knowledge Graph**: Visualize connections between your files
- **Hot Reload**: Automatically updates when files change
- **Full-Text Search**: Find anything quickly with fuzzy matching
- **Dark/Light Themes**: Easy on the eyes
- **Mermaid Diagrams**: Render flowcharts and diagrams
- **Syntax Highlighting**: Support for 180+ programming languages

## Installation

### Quick Install

```bash
# Download and run the installer
curl -sSL https://raw.githubusercontent.com/ocalvet/markdown-server/main/fuseki-install.sh | bash

# Or download first and run
wget https://raw.githubusercontent.com/ocalvet/markdown-server/main/fuseki-install.sh
chmod +x fuseki-install.sh
./fuseki-install.sh
```

### Manual Install

```bash
# Clone the repository
git clone https://github.com/ocalvet/markdown-server.git
cd markdown-server

# Run the installer
chmod +x fuseki-install.sh
./fuseki-install.sh
```

### Installation Options

The installer will ask you for:

1. **Installation Directory**: Where to install the `fuki` binary
   - Default: `~/.local/bin` (user-local, no sudo needed)
   - Custom: Any directory you prefer

2. **Add to PATH**: Whether to add the installation directory to your shell PATH
   - Recommended: Yes (for easy access)

### Example Installation

```bash
$ ./fuseki-install.sh
[INFO] Starting Fuseki installer...
[INFO] This will install fuki (formerly markdown-server)
[INFO] Detected platform: linux-amd64
Enter installation directory [/home/user/.local/bin]:
Add installation directory to PATH? [Y/n]: Y
[INFO] Found local binary: ./backend/markdown-server
[INFO] Installation Summary:
  Binary: fuki
  Install directory: /home/user/.local/bin
  Update PATH: yes
  Binary source: Local (./backend/markdown-server)
Proceed with installation? [Y/n]: Y
[INFO] Copying local binary from ./backend/markdown-server...
[SUCCESS] Installed fuki to /home/user/.local/bin/fuki
[SUCCESS] Added /home/user/.local/bin to /home/user/.bashrc
[INFO] Verifying installation...
[SUCCESS] Fuseki installed successfully!
```

## Usage

After installation:

```bash
# Start the server with your markdown directory (using flags)
fuki -d ~/my-notes

# Or if you added to PATH
fuki -d /path/to/your/markdown

# The server will start on port 8703 by default
# Open your browser to http://localhost:8703

# Show help
fuki --help
```

### Command-Line Options

Fuseki supports the following command-line options:

```bash
# Short options
fuki -p 9000 -d /path/to/docs

# Long options
fuki --port 9000 --dir /path/to/docs

# Positional argument (markdown directory)
fuki /path/to/docs  # Uses default port 8703

# Combine options and positional argument
fuki -p 9000 /path/to/docs
```

### Environment Variables

You can also customize the server with environment variables (flags take precedence):

```bash
# Custom port (flag takes precedence over env var)
PORT=7777 fuki -p 9000 -d ~/my-notes  # Uses port 9000

# Custom directory (flag takes precedence over env var)
MARKDOWN_DIR=/old/path fuki -d ~/new/path  # Uses ~/new/path

# Custom ignore patterns (no flag equivalent)
IGNORE_PATTERNS="temp,cache,node_modules" fuki -d ~/my-notes

# All together
PORT=7777 MARKDOWN_DIR=/default/path IGNORE_PATTERNS="temp,cache" fuki -p 9000 -d ~/actual/path
```

### Precedence Order

Command-line arguments have the highest priority:

1. **Command-line flags** (highest priority)
   - `-p` / `--port` for port
   - `-d` / `--dir` for directory
2. **Environment variables**
   - `PORT` for port
   - `MARKDOWN_DIR` for directory
3. **Default values**
   - Port: 8703
   - Directory: Current directory (`.`)

**Example of precedence:**
```bash
# This will use port 9999 (from flag), not 7777 (from env var)
PORT=7777 fuki -p 9999 -d ~/my-notes
```

## Uninstallation

To uninstall Fuseki:

```bash
# Run the uninstaller
./fuseki-uninstall.sh

# Or manually remove the binary
rm ~/.local/bin/fuki

# And remove from shell configuration
# Edit ~/.bashrc or ~/.zshrc and remove the line:
# export PATH="$HOME/.local/bin:$PATH"
```

## Manual Build (Advanced)

If you prefer to build from source:

```bash
# Clone the repository
git clone https://github.com/ocalvet/markdown-server.git
cd markdown-server/backend

# Build the binary
go build -o markdown-server main.go

# Install manually
cp markdown-server ~/.local/bin/fuki
chmod +x ~/.local/bin/fuki
```

## Troubleshooting

### Command not found: fuki

If you get "command not found" after installation:

1. **Restart your shell**: Close and reopen your terminal
2. **Or source your config**: `source ~/.bashrc` (or `~/.zshrc`)
3. **Check PATH**: `echo $PATH` should include `~/.local/bin`
4. **Manual PATH addition**: Add this to your shell config:
   ```bash
   export PATH="$HOME/.local/bin:$PATH"
   ```

### Permission denied

If you get permission errors:

1. Ensure the installation directory exists and is writable:
   ```bash
   mkdir -p ~/.local/bin
   ```
2. Check file permissions:
   ```bash
   ls -la ~/.local/bin/fuki
   ```
3. Fix permissions if needed:
   ```bash
   chmod +x ~/.local/bin/fuki
   ```

### Binary not found locally

If the installer can't find a local binary:

1. Ensure you're running the installer from the repository directory
2. Build the binary first:
   ```bash
   cd backend
   go build -o markdown-server main.go
   ```
3. Run the installer again from the repository root

## Why "Fuseki"?

In the game of Go, **Fuseki** refers to the opening phase where players establish their positions on the board. This mirrors how Fuseki helps you establish and organize your markdown knowledge base from the very beginning.

The name also pays homage to the game that inspired many concepts in computer science and artificial intelligence.

## License

MIT License - see the main repository for details.

## Support

- **Repository**: https://github.com/ocalvet/markdown-server
- **Issues**: https://github.com/ocalvet/markdown-server/issues
