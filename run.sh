#!/bin/bash

# Markdown Server - Quick Start Script
# Usage: ./run.sh [markdown_directory] [port] [ignore_patterns]

MARKDOWN_DIR=${1:-"./backend/markdown-files"}
PORT=${2:-8703}
IGNORE_PATTERNS=${3:-""}

echo "Starting Markdown Server..."
echo "Directory: $MARKDOWN_DIR"
echo "Port: $PORT"
if [ -n "$IGNORE_PATTERNS" ]; then
    echo "Custom Ignore Patterns: $IGNORE_PATTERNS"
fi
echo ""

cd backend
if [ -n "$IGNORE_PATTERNS" ]; then
    MARKDOWN_DIR="$MARKDOWN_DIR" PORT="$PORT" IGNORE_PATTERNS="$IGNORE_PATTERNS" go run main.go
else
    MARKDOWN_DIR="$MARKDOWN_DIR" PORT="$PORT" go run main.go
fi
