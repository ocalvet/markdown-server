package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
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

type FileInfo struct {
	Path     string `json:"path"`
	Name     string `json:"name"`
	IsDir    bool   `json:"isDir"`
	Children []FileInfo `json:"children,omitempty"`
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

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8703"
	}

	go watchFiles()

	http.HandleFunc("/api/files", corsMiddleware(handleListFiles))
	http.HandleFunc("/api/file/", corsMiddleware(handleGetFile))
	http.HandleFunc("/api/events", corsMiddleware(handleSSE))
	http.Handle("/", http.FileServer(http.Dir("../frontend")))

	log.Printf("Server starting on port %s", port)
	log.Printf("Serving markdown files from: %s", getMarkdownDir())
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

	markdownDir := getMarkdownDir()

	ignorePatterns := getIgnorePatterns()

	addDirRecursive := func(dir string) error {
		return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				log.Printf("Skipping directory due to error: %s (%v)", path, err)
				return filepath.SkipDir
			}
			if info.IsDir() {
				if shouldIgnore(info.Name(), ignorePatterns) {
					return filepath.SkipDir
				}
				if err := watcher.Add(path); err != nil {
					log.Printf("Could not watch directory: %s (%v)", path, err)
				}
			}
			return nil
		})
	}

	if err := addDirRecursive(markdownDir); err != nil {
		log.Printf("Error adding directories to watcher: %v", err)
		return
	}

	log.Printf("Watching for file changes in: %s", markdownDir)

	debounce := time.NewTimer(100 * time.Millisecond)
	debounce.Stop()

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			if event.Op&(fsnotify.Write|fsnotify.Create|fsnotify.Remove) != 0 {
				if strings.HasSuffix(strings.ToLower(event.Name), ".md") {
					debounce.Reset(100 * time.Millisecond)
				}
				if event.Op&fsnotify.Create != 0 {
					if info, err := os.Stat(event.Name); err == nil && info.IsDir() {
						addDirRecursive(event.Name)
					}
				}
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			log.Printf("File watcher error: %v", err)
		case <-debounce.C:
			log.Println("File change detected, broadcasting reload")
			broadcaster.Broadcast("reload")
		}
	}
}

func handleSSE(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")

	client := make(chan string, 10)
	broadcaster.Register(client)
	defer broadcaster.Unregister(client)

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming not supported", http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "data: connected\n\n")
	flusher.Flush()

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case message := <-client:
			fmt.Fprintf(w, "data: %s\n\n", message)
			flusher.Flush()
		case <-ticker.C:
			fmt.Fprintf(w, ": keepalive\n\n")
			flusher.Flush()
		case <-r.Context().Done():
			return
		}
	}
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

func getMarkdownDir() string {
	dir := os.Getenv("MARKDOWN_DIR")
	if dir == "" {
		dir = "./markdown-files"
	}
	if absDir, err := filepath.Abs(dir); err == nil {
		return absDir
	}
	return dir
}

func getIgnorePatterns() []string {
	customPatterns := os.Getenv("IGNORE_PATTERNS")
	if customPatterns != "" {
		patterns := strings.Split(customPatterns, ",")
		for i, p := range patterns {
			patterns[i] = strings.TrimSpace(p)
		}
		return patterns
	}
	return defaultIgnorePatterns
}

func shouldIgnore(name string, ignorePatterns []string) bool {
	// Ignore all directories starting with a dot
	if strings.HasPrefix(name, ".") {
		return true
	}

	nameLower := strings.ToLower(name)
	for _, pattern := range ignorePatterns {
		if strings.Contains(nameLower, strings.ToLower(pattern)) {
			return true
		}
	}
	return false
}

func handleListFiles(w http.ResponseWriter, r *http.Request) {
	markdownDir := getMarkdownDir()

	files, err := buildFileTree(markdownDir, markdownDir)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error reading directory: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(files)
}

func buildFileTree(rootDir, currentDir string) ([]FileInfo, error) {
	var files []FileInfo
	ignorePatterns := getIgnorePatterns()

	entries, err := os.ReadDir(currentDir)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if shouldIgnore(entry.Name(), ignorePatterns) {
			continue
		}

		fullPath := filepath.Join(currentDir, entry.Name())
		relPath, _ := filepath.Rel(rootDir, fullPath)

		fileInfo := FileInfo{
			Path:  relPath,
			Name:  entry.Name(),
			IsDir: entry.IsDir(),
		}

		if entry.IsDir() {
			children, err := buildFileTree(rootDir, fullPath)
			if err == nil {
				fileInfo.Children = children
			}
		} else if strings.HasSuffix(strings.ToLower(entry.Name()), ".md") {
			files = append(files, fileInfo)
		}

		if entry.IsDir() && len(fileInfo.Children) > 0 {
			files = append(files, fileInfo)
		}
	}

	return files, nil
}

func handleGetFile(w http.ResponseWriter, r *http.Request) {
	filePath := strings.TrimPrefix(r.URL.Path, "/api/file/")

	if filePath == "" {
		http.Error(w, "File path required", http.StatusBadRequest)
		return
	}

	fullPath := filepath.Join(getMarkdownDir(), filePath)

	cleanPath := filepath.Clean(fullPath)
	if !strings.HasPrefix(cleanPath, filepath.Clean(getMarkdownDir())) {
		http.Error(w, "Invalid file path", http.StatusForbidden)
		return
	}

	if !strings.HasSuffix(strings.ToLower(cleanPath), ".md") {
		http.Error(w, "Only markdown files allowed", http.StatusForbidden)
		return
	}

	file, err := os.Open(cleanPath)
	if err != nil {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, "Error reading file", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Write(content)
}
