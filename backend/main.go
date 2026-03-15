package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

var defaultIgnorePatterns = []string{
	"node_modules",
	".git",
	".svn",
	".hg",
	".idea",
	".vscode",
	"__pycache__",
	".pytest_cache",
	".mypy_cache",
	"vendor",
	"dist",
	"build",
	"target",
	".next",
	".nuxt",
	"coverage",
	".DS_Store",
	"Thumbs.db",
}

// allowedAssetExtensions is the allowlist for /api/asset/ serving.
var allowedAssetExtensions = map[string]bool{
	".png": true, ".jpg": true, ".jpeg": true, ".gif": true,
	".svg": true, ".webp": true, ".ico": true,
	".pdf": true,
	".mp4": true, ".webm": true, ".ogg": true,
	".mp3": true, ".wav": true,
	".txt": true, ".csv": true, ".json": true,
}

type FileInfo struct {
	Path     string     `json:"path"`
	Name     string     `json:"name"`
	IsDir    bool       `json:"isDir"`
	Children []FileInfo `json:"children,omitempty"`
}

type GraphNode struct {
	ID    string `json:"id"`
	Label string `json:"label"`
	Title string `json:"title"`
}

type GraphEdge struct {
	From string `json:"from"`
	To   string `json:"to"`
}

type GraphData struct {
	Nodes []GraphNode `json:"nodes"`
	Edges []GraphEdge `json:"edges"`
}

type BacklinkEntry struct {
	Path    string `json:"path"`
	Name    string `json:"name"`
	Context string `json:"context"`
}

type BacklinksResponse struct {
	Backlinks []BacklinkEntry `json:"backlinks"`
}

type ReloadBroadcaster struct {
	clients map[chan string]bool
	mu      sync.Mutex
}

var broadcaster = &ReloadBroadcaster{
	clients: make(map[chan string]bool),
}

func (b *ReloadBroadcaster) Register(client chan string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.clients[client] = true
}

func (b *ReloadBroadcaster) Unregister(client chan string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	delete(b.clients, client)
	close(client)
}

func (b *ReloadBroadcaster) Broadcast(message string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	for client := range b.clients {
		select {
		case client <- message:
		default:
		}
	}
}

// linkCache holds a precomputed map: targetFile -> []sourceFile for backlinks.
// It is rebuilt on startup and on file-change events.
var linkCache struct {
	mu   sync.RWMutex
	data map[string][]BacklinkEntry // key: target path, value: list of sources
}

func rebuildLinkCache() {
	newData := make(map[string][]BacklinkEntry)

	var mdFiles []string
	filepath.Walk(markdownDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() && strings.HasSuffix(strings.ToLower(info.Name()), ".md") {
			rel, _ := filepath.Rel(markdownDir, path)
			mdFiles = append(mdFiles, rel)
		}
		return nil
	})

	for _, srcFile := range mdFiles {
		content, err := os.ReadFile(filepath.Join(markdownDir, srcFile))
		if err != nil {
			continue
		}
		links := extractLinks(string(content), srcFile)

		// Deduplicate resolved targets per source file
		seenTargets := make(map[string]bool)
		for _, link := range links {
			resolved := resolveLink(srcFile, link)
			if seenTargets[resolved] {
				continue
			}
			seenTargets[resolved] = true

			ctx := extractLinkContext(string(content), link)
			entry := BacklinkEntry{
				Path:    srcFile,
				Name:    filepath.Base(srcFile),
				Context: ctx,
			}
			newData[resolved] = append(newData[resolved], entry)
		}
	}

	linkCache.mu.Lock()
	linkCache.data = newData
	linkCache.mu.Unlock()
	log.Printf("Link cache rebuilt: %d target files tracked", len(newData))
}

// extractLinkContext returns up to 120 chars of surrounding text for a link.
func extractLinkContext(content, link string) string {
	idx := strings.Index(content, link)
	if idx < 0 {
		return ""
	}
	start := idx - 60
	if start < 0 {
		start = 0
	}
	end := idx + len(link) + 60
	if end > len(content) {
		end = len(content)
	}
	snippet := strings.ReplaceAll(content[start:end], "\n", " ")
	snippet = strings.TrimSpace(snippet)
	if start > 0 {
		snippet = "…" + snippet
	}
	if end < len(content) {
		snippet = snippet + "…"
	}
	return snippet
}

var (
	port        string
	markdownDir string
)

func main() {
	// Define flags
	flag.StringVar(&port, "port", "", "Server port")
	flag.StringVar(&port, "p", "", "Server port (shorthand)")
	flag.StringVar(&markdownDir, "dir", "", "Markdown directory")
	flag.StringVar(&markdownDir, "d", "", "Markdown directory (shorthand)")

	// Custom usage message
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s [options] [markdown_directory]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  %s                      # Use current directory, port 8703\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s /path/to/docs        # Use specified directory, port 8703\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -p 9000              # Use current directory, port 9000\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -p 9000 /path/to/docs # Use specified directory, port 9000\n", os.Args[0])
	}

	// Parse flags
	flag.Parse()

	// Handle positional arguments (markdown directory)
	args := flag.Args()
	if len(args) > 0 {
		markdownDir = args[0]
	}

	// Determine port with precedence: flag > env var > default
	if port == "" {
		port = os.Getenv("PORT")
		if port == "" {
			port = "8703"
		}
	}

	// Validate port
	if _, err := strconv.Atoi(port); err != nil {
		log.Fatalf("Error: Invalid port '%s'. Must be a number.", port)
	}

	// Determine markdown directory with precedence: flag > env var > default (current dir)
	if markdownDir == "" {
		markdownDir = os.Getenv("MARKDOWN_DIR")
		if markdownDir == "" {
			markdownDir = "."
		}
	}

	// Resolve markdown directory path
	absMarkdownDir, err := filepath.Abs(markdownDir)
	if err != nil {
		log.Fatalf("Error resolving markdown directory path: %v", err)
	}
	markdownDir = absMarkdownDir

	// Validate markdown directory exists
	if info, err := os.Stat(markdownDir); err != nil {
		log.Fatalf("Error: Markdown directory does not exist: %s", markdownDir)
	} else if !info.IsDir() {
		log.Fatalf("Error: Markdown path is not a directory: %s", markdownDir)
	}

	// Build initial link cache for backlinks
	go rebuildLinkCache()

	go watchFiles()

	// Determine frontend directory path
	// Try to find it relative to the executable
	execPath, execErr := os.Executable()
	if execErr != nil {
		log.Printf("Warning: Could not determine executable path: %v", execErr)
	}
	execDir := filepath.Dir(execPath)

	// Try different paths for frontend directory
	// Priority: 1) Next to binary (frontend/), 2) Relative to project (../frontend)
	frontendDir := ""
	possiblePaths := []string{
		filepath.Join(execDir, "frontend"), // Installed: ~/.local/bin/frontend
		"../frontend",                      // Development: backend/../frontend
		"./frontend",                       // Current directory
	}

	for _, path := range possiblePaths {
		if _, statErr := os.Stat(path); statErr == nil {
			frontendDir = path
			break
		}
	}

	if frontendDir == "" {
		log.Printf("Warning: Could not find frontend directory. Tried: %v", possiblePaths)
		frontendDir = "../frontend" // Fallback to original path
	}

	http.HandleFunc("/api/files", corsMiddleware(handleListFiles))
	http.HandleFunc("/api/file/", corsMiddleware(handleGetFile))
	http.HandleFunc("/api/graph", corsMiddleware(handleGraph))
	http.HandleFunc("/api/events", corsMiddleware(handleSSE))
	http.Handle("/", http.FileServer(http.Dir(frontendDir)))

	log.Printf("Server starting on port %s", port)
	log.Printf("Serving markdown files from: %s", markdownDir)
	log.Printf("Hot reload enabled")
	log.Printf("Ignoring patterns: %v", getIgnorePatterns())
	log.Printf("Set MARKDOWN_DIR environment variable to change the directory")
	log.Printf("Set IGNORE_PATTERNS environment variable to customize ignore patterns (comma-separated)")
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}

func watchFiles() {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Printf("Error creating file watcher: %v", err)
		return
	}
	defer watcher.Close()

	// Watch markdown directory recursively
	err = filepath.Walk(markdownDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			ignorePatterns := getIgnorePatterns()
			shouldIgnore := false
			for _, pattern := range ignorePatterns {
				if matchesIgnorePattern(path, markdownDir, pattern) {
					shouldIgnore = true
					break
				}
			}
			if !shouldIgnore {
				err = watcher.Add(path)
				if err != nil {
					log.Printf("Error watching directory %s: %v", path, err)
				}
			}
		}
		return nil
	})

	if err != nil {
		log.Printf("Error walking directory: %v", err)
		return
	}

	// Debounce timer
	var timer *time.Timer
	debounceDuration := 100 * time.Millisecond

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			// Filter for create, write, remove, rename
			if event.Op&(fsnotify.Write|fsnotify.Create|fsnotify.Remove|fsnotify.Rename) != 0 {
				// Check if the file is a markdown file
				if strings.HasSuffix(event.Name, ".md") || strings.HasSuffix(event.Name, ".MD") {
					if timer != nil {
						timer.Stop()
					}
					timer = time.AfterFunc(debounceDuration, func() {
						log.Printf("File changed: %s", event.Name)
						go rebuildLinkCache()
						broadcaster.Broadcast("reload")
					})
				}
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			log.Printf("File watcher error: %v", err)
		}
	}
}

// matchesIgnorePattern checks if any segment of the path (relative to rootDir)
// exactly matches the given ignore pattern. This prevents false positives like
// "00-build-system" matching the "build" pattern.
func matchesIgnorePattern(path, rootDir, pattern string) bool {
	relPath, err := filepath.Rel(rootDir, path)
	if err != nil {
		relPath = path
	}
	segments := strings.Split(relPath, string(filepath.Separator))
	for _, seg := range segments {
		if seg == pattern {
			return true
		}
	}
	return false
}

func getIgnorePatterns() []string {
	envPatterns := os.Getenv("IGNORE_PATTERNS")
	if envPatterns != "" {
		patterns := strings.Split(envPatterns, ",")
		for i := range patterns {
			patterns[i] = strings.TrimSpace(patterns[i])
		}
		return patterns
	}
	return defaultIgnorePatterns
}

func corsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next(w, r)
	}
}

func handleListFiles(w http.ResponseWriter, r *http.Request) {
	fileTree, err := buildFileTree(markdownDir)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(fileTree)
}

func buildFileTree(rootDir string) ([]FileInfo, error) {
	var files []FileInfo
	ignorePatterns := getIgnorePatterns()

	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip ignored patterns (match against each path segment's name, not substrings)
		for _, pattern := range ignorePatterns {
			if matchesIgnorePattern(path, rootDir, pattern) {
				if info.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}
		}

		// Skip root directory itself
		if path == rootDir {
			return nil
		}

		relPath, err := filepath.Rel(rootDir, path)
		if err != nil {
			return err
		}

		// Only include .md files and directories
		if info.IsDir() || strings.HasSuffix(strings.ToLower(info.Name()), ".md") {
			fileInfo := FileInfo{
				Path:  relPath,
				Name:  info.Name(),
				IsDir: info.IsDir(),
			}
			files = append(files, fileInfo)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Build hierarchical structure
	return buildHierarchy(files), nil
}

func buildHierarchy(files []FileInfo) []FileInfo {
	// Use pointers so all mutations (appending children) are visible everywhere.
	dirMap := make(map[string]*FileInfo)

	// First pass: create stable pointer entries for every directory.
	for _, f := range files {
		if f.IsDir {
			entry := f // copy value
			dirMap[f.Path] = &entry
		}
	}

	// Second pass: attach files to their parent directories.
	var rootFiles []FileInfo
	for _, f := range files {
		if f.IsDir {
			continue
		}
		parentPath := filepath.Dir(f.Path)
		if parentPath != "." && parentPath != "" {
			if parent, ok := dirMap[parentPath]; ok {
				parent.Children = append(parent.Children, f)
			} else {
				rootFiles = append(rootFiles, f)
			}
		} else {
			rootFiles = append(rootFiles, f)
		}
	}

	// Third pass: attach child directories to parent directories.
	// Sort by depth (deepest first) so that when we snapshot a directory into
	// its parent, it already contains all of its own children.
	// Collect paths and sort by number of separators descending.
	type dirEntry struct {
		path  string
		depth int
	}
	var entries []dirEntry
	for path := range dirMap {
		entries = append(entries, dirEntry{path, strings.Count(path, string(filepath.Separator))})
	}
	// Simple sort: deepest first
	for i := 0; i < len(entries); i++ {
		for j := i + 1; j < len(entries); j++ {
			if entries[j].depth > entries[i].depth {
				entries[i], entries[j] = entries[j], entries[i]
			}
		}
	}

	placedDirs := make(map[string]bool)
	for _, e := range entries {
		parentPath := filepath.Dir(e.path)
		if parentPath != "." && parentPath != "" {
			if parent, ok := dirMap[parentPath]; ok {
				parent.Children = append(parent.Children, *dirMap[e.path])
				placedDirs[e.path] = true
			}
		}
	}

	// Build root: top-level directories + root-level files.
	var root []FileInfo
	for _, e := range entries {
		if !placedDirs[e.path] {
			root = append(root, *dirMap[e.path])
		}
	}
	root = append(root, rootFiles...)

	// Sort the entire tree alphabetically (directories first, then files).
	sortFileInfos(root)

	return root
}

// sortFileInfos sorts a slice of FileInfo entries: directories first (alphabetically),
// then files (alphabetically). It recurses into children.
func sortFileInfos(items []FileInfo) {
	sort.Slice(items, func(i, j int) bool {
		// Directories before files
		if items[i].IsDir != items[j].IsDir {
			return items[i].IsDir
		}
		// Alphabetical by name (case-insensitive)
		return strings.ToLower(items[i].Name) < strings.ToLower(items[j].Name)
	})
	// Recursively sort children
	for idx := range items {
		if len(items[idx].Children) > 0 {
			sortFileInfos(items[idx].Children)
		}
	}
}

func handleGetFile(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/file/")

	// Security: prevent directory traversal
	cleanPath := filepath.Clean(path)
	if strings.Contains(cleanPath, "..") {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}

	// Ensure path is relative to markdownDir
	fullPath := filepath.Join(markdownDir, cleanPath)

	// Check if file exists and is within markdownDir
	absPath, err := filepath.Abs(fullPath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	absMarkdownDir, _ := filepath.Abs(markdownDir)
	if !strings.HasPrefix(absPath, absMarkdownDir) {
		http.Error(w, "Access denied", http.StatusForbidden)
		return
	}

	content, err := os.ReadFile(fullPath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "text/markdown")
	w.Write(content)
}

// handleGetAsset serves non-markdown assets (images, PDFs, etc.) from the markdown directory.
func handleGetAsset(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/asset/")

	// Security: prevent directory traversal
	cleanPath := filepath.Clean(path)
	if strings.Contains(cleanPath, "..") {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}

	// Check extension against allowlist
	ext := strings.ToLower(filepath.Ext(cleanPath))
	if !allowedAssetExtensions[ext] {
		http.Error(w, "File type not allowed", http.StatusForbidden)
		return
	}

	fullPath := filepath.Join(markdownDir, cleanPath)

	// Verify the resolved path is still inside markdownDir
	absPath, err := filepath.Abs(fullPath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	absMarkdownDir, _ := filepath.Abs(markdownDir)
	if !strings.HasPrefix(absPath, absMarkdownDir) {
		http.Error(w, "Access denied", http.StatusForbidden)
		return
	}

	// Set Content-Type based on extension
	contentType := mime.TypeByExtension(ext)
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	w.Header().Set("Content-Type", contentType)

	// Use http.ServeFile for proper range request / cache header support
	http.ServeFile(w, r, fullPath)
}

func handleGraph(w http.ResponseWriter, r *http.Request) {
	graphData := buildGraphData()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(graphData)
}

func buildGraphData() GraphData {
	var nodes []GraphNode
	var edges []GraphEdge
	nodeMap := make(map[string]bool)

	// Get all markdown files
	var markdownFiles []string
	filepath.Walk(markdownDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() && strings.HasSuffix(strings.ToLower(info.Name()), ".md") {
			relPath, _ := filepath.Rel(markdownDir, path)
			markdownFiles = append(markdownFiles, relPath)
		}
		return nil
	})

	// Create nodes
	for _, file := range markdownFiles {
		nodeMap[file] = true
		nodes = append(nodes, GraphNode{
			ID:    file,
			Label: file,
			Title: file,
		})
	}

	// Extract links and create edges
	for _, file := range markdownFiles {
		content, err := os.ReadFile(filepath.Join(markdownDir, file))
		if err != nil {
			continue
		}

		links := extractLinks(string(content), file)
		for _, link := range links {
			resolvedLink := resolveLink(file, link)
			if nodeMap[resolvedLink] {
				edges = append(edges, GraphEdge{
					From: file,
					To:   resolvedLink,
				})
			}
		}
	}

	return GraphData{Nodes: nodes, Edges: edges}
}

// handleBacklinks returns all files that link to the requested file.
func handleBacklinks(w http.ResponseWriter, r *http.Request) {
	targetPath := strings.TrimPrefix(r.URL.Path, "/api/backlinks/")

	// Security: prevent directory traversal
	cleanPath := filepath.Clean(targetPath)
	if strings.Contains(cleanPath, "..") {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}

	linkCache.mu.RLock()
	entries := linkCache.data[cleanPath]
	linkCache.mu.RUnlock()

	if entries == nil {
		entries = []BacklinkEntry{}
	}

	resp := BacklinksResponse{Backlinks: entries}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func extractLinks(content, sourcePath string) []string {
	var links []string

	// Match [text](link.md) or [text](link)
	linkRegex := regexp.MustCompile(`\[([^\]]+)\]\(([^)#\s]+)`)
	matches := linkRegex.FindAllStringSubmatch(content, -1)
	for _, match := range matches {
		if len(match) > 2 {
			link := match[2]
			// Skip external URLs and pure anchor links
			if strings.HasPrefix(link, "http://") || strings.HasPrefix(link, "https://") || strings.HasPrefix(link, "#") {
				continue
			}
			// Only include .md links or links without extension (treat as .md)
			ext := strings.ToLower(filepath.Ext(link))
			if ext == ".md" || ext == "" {
				links = append(links, link)
			}
		}
	}

	// Match [[link]] or [[path/to/link]]
	wikiRegex := regexp.MustCompile(`\[\[([^\]]+)\]\]`)
	wikiMatches := wikiRegex.FindAllStringSubmatch(content, -1)
	for _, match := range wikiMatches {
		if len(match) > 1 {
			link := match[1]
			// Add .md extension if not present
			if !strings.HasSuffix(strings.ToLower(link), ".md") {
				link += ".md"
			}
			links = append(links, link)
		}
	}

	return links
}

func resolveLink(sourcePath, link string) string {
	// If link is absolute (starts with /), treat as relative to markdownDir
	if strings.HasPrefix(link, "/") {
		link = strings.TrimPrefix(link, "/")
	}

	// Ensure .md extension
	if !strings.HasSuffix(strings.ToLower(link), ".md") {
		link += ".md"
	}

	// Resolve relative path
	sourceDir := filepath.Dir(sourcePath)
	resolved := filepath.Join(sourceDir, link)

	// Clean the path
	resolved = filepath.Clean(resolved)

	return resolved
}

func handleSSE(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	client := make(chan string)
	broadcaster.Register(client)
	defer broadcaster.Unregister(client)

	// Listen for connection close
	notify := w.(http.CloseNotifier).CloseNotify()
	go func() {
		<-notify
		broadcaster.Unregister(client)
	}()

	for {
		select {
		case message := <-client:
			fmt.Fprintf(w, "data: %s\n\n", message)
			w.(http.Flusher).Flush()
		case <-notify:
			return
		}
	}
}
