#!/bin/bash
# fuseki-install.sh
# Installer for Fuseki (formerly Markdown Server)
# A Go-inspired markdown server with knowledge graph visualization

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
APP_NAME="Fuseki"
BINARY_NAME="fuki"
REPO="ocalvet/markdown-server"
GITHUB_API_URL="https://api.github.com/repos/${REPO}/releases/latest"

# Default installation directory (user-local, no sudo needed)
DEFAULT_INSTALL_DIR="${HOME}/.local/bin"
DEFAULT_MARKDOWN_DIR=""

# Function to print colored output
print_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Function to detect OS and architecture
detect_platform() {
    local os=$(uname -s)
    local arch=$(uname -m)
    
    case "$os" in
        Linux)
            OS="linux"
            ;;
        Darwin)
            OS="macos"
            ;;
        *)
            print_error "Unsupported OS: $os"
            exit 1
            ;;
    esac
    
    case "$arch" in
        x86_64|amd64)
            ARCH="amd64"
            ;;
        aarch64|arm64)
            ARCH="arm64"
            ;;
        *)
            print_error "Unsupported architecture: $arch"
            exit 1
            ;;
    esac
    
    print_info "Detected platform: ${OS}-${ARCH}"
}

# Function to check if a command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Function to get the binary source
get_binary_source() {
    # First, try to check GitHub releases
    if command_exists curl; then
        local response=$(curl -s "${GITHUB_API_URL}")
    elif command_exists wget; then
        local response=$(wget -q -O - "${GITHUB_API_URL}")
    else
        print_warning "Neither curl nor wget found. Will try local binary." >&2
        response=""
    fi
    
    # Check if the response is an error (no releases)
    if echo "$response" | grep -q '"message": "Not Found"'; then
        print_warning "No releases found on GitHub for ${REPO}" >&2
        print_info "Checking for local binary..." >&2
        
        # Try to find local binary in the current directory or backend directory
        local local_binary=""
        if [ -f "./backend/markdown-server" ]; then
            local_binary="./backend/markdown-server"
        elif [ -f "../backend/markdown-server" ]; then
            local_binary="../backend/markdown-server"
        elif [ -f "markdown-server" ]; then
            local_binary="./markdown-server"
        fi
        
        if [ -n "$local_binary" ] && [ -x "$local_binary" ]; then
            print_info "Found local binary: ${local_binary}" >&2
            echo "local:${local_binary}"
            return 0
        else
            print_error "No binary found locally and no releases on GitHub." >&2
            print_info "Please build the binary first or download a release." >&2
            exit 1
        fi
    fi
    
    # Extract download URL for the binary
    local download_url=$(echo "$response" | grep -o '"browser_download_url": "[^"]*"' | grep -E "markdown-server.*${OS}.*${ARCH}" | head -1 | sed 's/"browser_download_url": "//' | sed 's/"$//')
    
    if [ -z "$download_url" ]; then
        # Try alternative pattern - look for any binary asset
        download_url=$(echo "$response" | grep -o '"browser_download_url": "[^"]*"' | grep -E "\.(bin|exe|markdown-server)" | head -1 | sed 's/"browser_download_url": "//' | sed 's/"$//')
    fi
    
    if [ -z "$download_url" ]; then
        print_error "Could not find a download URL for ${OS}-${ARCH}" >&2
        print_info "Available assets:" >&2
        echo "$response" | grep -o '"name": "[^"]*"' | head -10 >&2
        exit 1
    fi
    
    echo "remote:${download_url}"
}

# Function to install the binary
install_binary() {
    local install_dir="$1"
    local binary_source="$2"
    
    # Create install directory if it doesn't exist
    mkdir -p "$install_dir"
    
    local target_file="${install_dir}/${BINARY_NAME}"
    local frontend_dir="${install_dir}/frontend"
    
    # Check if source is local or remote
    if [[ "$binary_source" == local:* ]]; then
        local local_path="${binary_source#local:}"
        local project_root=$(dirname "$(dirname "$local_path")")
        
        print_info "Copying local binary from ${local_path}..."
        cp "$local_path" "$target_file"
        
        # Copy frontend directory
        if [ -d "${project_root}/frontend" ]; then
            print_info "Copying frontend files..."
            mkdir -p "$frontend_dir"
            cp -r "${project_root}/frontend/." "$frontend_dir/"
        else
            print_warning "Frontend directory not found at ${project_root}/frontend"
        fi
    else
        local download_url="${binary_source#remote:}"
        print_info "Downloading ${APP_NAME} from ${download_url}..."
        
        local temp_file=$(mktemp)
        
        if command_exists curl; then
            curl -L -o "$temp_file" "$download_url"
        elif command_exists wget; then
            wget -O "$temp_file" "$download_url"
        fi
        
        mv "$temp_file" "$target_file"
        
        # For remote downloads, frontend files need to be bundled separately
        print_warning "Frontend files not included in remote download. Please ensure frontend directory is available."
    fi
    
    chmod +x "$target_file"
    
    print_success "Installed ${BINARY_NAME} to ${target_file}"
}

# Function to update shell configuration
update_shell_config() {
    local install_dir="$1"
    local shell_config=""
    
    # Detect shell and config file
    case "$SHELL" in
        */bash)
            shell_config="${HOME}/.bashrc"
            ;;
        */zsh)
            shell_config="${HOME}/.zshrc"
            ;;
        */fish)
            shell_config="${HOME}/.config/fish/config.fish"
            ;;
        *)
            print_warning "Could not detect shell. Please add ${install_dir} to your PATH manually."
            return 1
            ;;
    esac
    
    # Check if already in PATH
    if grep -q "${install_dir}" "$shell_config" 2>/dev/null; then
        print_info "${install_dir} is already in ${shell_config}"
        return 0
    fi
    
    # Add to shell config
    if [ -f "$shell_config" ]; then
        echo "" >> "$shell_config"
        echo "# Added by ${APP_NAME} installer" >> "$shell_config"
        echo "export PATH=\"${install_dir}:\$PATH\"" >> "$shell_config"
        print_success "Added ${install_dir} to ${shell_config}"
        print_warning "Please restart your shell or run: source ${shell_config}"
    else
        print_warning "Shell config file not found: ${shell_config}"
        print_info "Please add ${install_dir} to your PATH manually"
    fi
}



# Function to verify installation
verify_installation() {
    local install_dir="$1"
    
    local binary_path="${install_dir}/${BINARY_NAME}"
    
    if [ ! -f "$binary_path" ]; then
        print_error "Binary not found: ${binary_path}"
        return 1
    fi
    
    if [ ! -x "$binary_path" ]; then
        print_error "Binary is not executable: ${binary_path}"
        return 1
    fi
    
    print_info "Verifying installation..."
    
    # Check file size to ensure binary is valid
    local file_size=$(stat -c%s "$binary_path" 2>/dev/null || stat -f%z "$binary_path" 2>/dev/null)
    if [ "$file_size" -lt 1000 ]; then
        print_error "Binary file seems too small (${file_size} bytes). It may be corrupted."
        return 1
    fi
    
    print_success "${APP_NAME} installed successfully!"
    print_info "Binary location: ${binary_path}"
    print_info "To start the server: MARKDOWN_DIR=/path/to/your/markdown ${BINARY_NAME}"
    return 0
}

# Main function
main() {
    print_info "Starting ${APP_NAME} installer..."
    print_info "This will install ${BINARY_NAME} (formerly markdown-server)"
    echo ""
    
    # Detect platform
    detect_platform
    
    # Get installation directory from user
    local install_dir=""
    while [ -z "$install_dir" ]; do
        echo -n "Enter installation directory [${DEFAULT_INSTALL_DIR}]: "
        read user_install_dir
        install_dir="${user_install_dir:-${DEFAULT_INSTALL_DIR}}"
        
        if [ ! -d "$(dirname "$install_dir")" ]; then
            print_error "Parent directory does not exist: $(dirname "$install_dir")"
            install_dir=""
        fi
    done
    
    # Ask about PATH update
    local update_path=""
    while [ -z "$update_path" ]; do
        echo -n "Add installation directory to PATH? [Y/n]: "
        read user_input
        user_input=$(echo "$user_input" | tr '[:upper:]' '[:lower:]')
        
        case "$user_input" in
            y|yes|"")
                update_path="yes"
                ;;
            n|no)
                update_path="no"
                ;;
            *)
                print_error "Please answer yes or no"
                ;;
        esac
    done
    
    # Get binary source (local or remote)
    local binary_source=$(get_binary_source)
    
    # Confirm installation
    echo ""
    print_info "Installation Summary:"
    echo "  Binary: ${BINARY_NAME}"
    echo "  Install directory: ${install_dir}"
    echo "  Update PATH: ${update_path}"
    
    if [[ "$binary_source" == local:* ]]; then
        local local_path="${binary_source#local:}"
        echo "  Binary source: Local (${local_path})"
    else
        local remote_url="${binary_source#remote:}"
        echo "  Binary source: Remote (${remote_url})"
    fi
    echo ""
    
    echo -n "Proceed with installation? [Y/n]: "
    read confirm
    confirm=$(echo "$confirm" | tr '[:upper:]' '[:lower:]')
    
    if [ "$confirm" = "n" ] || [ "$confirm" = "no" ]; then
        print_info "Installation cancelled."
        exit 0
    fi
    
    # Install binary
    install_binary "$install_dir" "$binary_source"
    
    # Update shell configuration if requested
    if [ "$update_path" = "yes" ]; then
        update_shell_config "$install_dir"
    fi
    
    # Verify installation
    verify_installation "$install_dir"
    
    # Final instructions
    echo ""
    print_success "=== ${APP_NAME} Installation Complete ==="
    echo ""
    print_info "Next steps:"
    echo "  1. If you updated PATH, restart your shell or run: source ~/.bashrc (or ~/.zshrc)"
    echo "  2. Start the server:"
    echo "     ${BINARY_NAME} -d /path/to/your/markdown"
    echo "     ${BINARY_NAME} -p 9000 -d /path/to/your/markdown"
    echo "  3. Open your browser to: http://localhost:8703"
    echo ""
    print_info "For more information, visit: https://github.com/${REPO}"
    print_info "Run ${BINARY_NAME} --help for usage information"
}

# Run main function
main
