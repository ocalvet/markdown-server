#!/bin/bash
# fuseki-uninstall.sh
# Uninstaller for Fuseki

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

# Function to find binary location
find_binary() {
    # Check common locations
    local locations=(
        "${HOME}/.local/bin/${BINARY_NAME}"
        "/usr/local/bin/${BINARY_NAME}"
        "${HOME}/bin/${BINARY_NAME}"
    )
    
    for location in "${locations[@]}"; do
        if [ -f "$location" ]; then
            echo "$location"
            return 0
        fi
    done
    
    # Check PATH
    if command -v "${BINARY_NAME}" >/dev/null 2>&1; then
        which "${BINARY_NAME}"
        return 0
    fi
    
    return 1
}

# Function to remove binary
remove_binary() {
    local binary_path="$1"
    
    if [ -f "$binary_path" ]; then
        rm -f "$binary_path"
        print_success "Removed binary: ${binary_path}"
        return 0
    else
        print_warning "Binary not found: ${binary_path}"
        return 1
    fi
}

# Function to remove from shell configuration
remove_from_shell_config() {
    local install_dir=$(dirname "$1")
    local shell_configs=(
        "${HOME}/.bashrc"
        "${HOME}/.zshrc"
        "${HOME}/.config/fish/config.fish"
    )
    
    for config in "${shell_configs[@]}"; do
        if [ -f "$config" ]; then
            # Create backup
            cp "$config" "${config}.bak" 2>/dev/null || true
            
            # Remove lines added by installer
            sed -i '/# Added by Fuseki installer/d' "$config" 2>/dev/null || true
            sed -i "\|export PATH=\"${install_dir}:\$PATH\"|d" "$config" 2>/dev/null || true
            
            print_info "Cleaned up ${config}"
        fi
    done
}

# Function to ask about markdown directory
ask_markdown_dir() {
    echo -n "Do you want to remove your markdown directory? [y/N]: "
    read response
    response=$(echo "$response" | tr '[:upper:]' '[:lower:]')
    
    if [ "$response" = "y" ] || [ "$response" = "yes" ]; then
        echo -n "Enter markdown directory path to remove: "
        read markdown_dir
        
        if [ -d "$markdown_dir" ]; then
            rm -rf "$markdown_dir"
            print_success "Removed markdown directory: ${markdown_dir}"
        else
            print_warning "Directory not found: ${markdown_dir}"
        fi
    fi
}

# Main function
main() {
    print_info "Starting ${APP_NAME} uninstaller..."
    echo ""
    
    # Find binary
    local binary_path=$(find_binary)
    
    if [ -z "$binary_path" ]; then
        print_error "${BINARY_NAME} binary not found in common locations or PATH"
        print_info "You may have installed it in a custom location."
        print_info "Please remove it manually."
        exit 1
    fi
    
    print_info "Found ${BINARY_NAME} at: ${binary_path}"
    
    # Confirm uninstallation
    echo ""
    echo -n "Are you sure you want to uninstall ${APP_NAME}? [y/N]: "
    read confirm
    confirm=$(echo "$confirm" | tr '[:upper:]' '[:lower:]')
    
    if [ "$confirm" != "y" ] && [ "$confirm" != "yes" ]; then
        print_info "Uninstallation cancelled."
        exit 0
    fi
    
    # Remove binary
    remove_binary "$binary_path"
    
    # Remove from shell configuration
    remove_from_shell_config "$binary_path"
    
    # Ask about markdown directory
    ask_markdown_dir
    
    # Final instructions
    echo ""
    print_success "=== ${APP_NAME} Uninstallation Complete ==="
    echo ""
    print_info "If you updated your shell configuration, you may want to:"
    print_info "  - Restart your shell"
    print_info "  - Or run: source ~/.bashrc (or ~/.zshrc)"
    echo ""
    print_info "Thank you for using ${APP_NAME}!"
}

# Run main function
main
