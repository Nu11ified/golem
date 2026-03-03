package dev

import (
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// FileWatcher watches a directory tree for file changes using polling.
// It is goroutine-safe and does not depend on any external libraries.
type FileWatcher struct {
	root     string
	interval time.Duration
	modTimes map[string]time.Time
	callback func(path string)
	stopCh   chan struct{}
	mu       sync.Mutex
	started  chan struct{} // closed after the first scan completes
}

// NewFileWatcher creates a new polling-based file watcher that monitors
// .go files (excluding _test.go) under root. The interval controls how
// often the directory tree is re-scanned.
func NewFileWatcher(root string, interval time.Duration) *FileWatcher {
	return &FileWatcher{
		root:     root,
		interval: interval,
		modTimes: make(map[string]time.Time),
		stopCh:   make(chan struct{}),
		started:  make(chan struct{}),
	}
}

// OnChange registers a callback that is invoked with the path of every
// file that was created, modified, or deleted since the previous scan.
// The callback is called from the watcher goroutine, so it must be safe
// for concurrent use if it accesses shared state.
func (w *FileWatcher) OnChange(callback func(path string)) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.callback = callback
}

// Start begins watching for file changes. It blocks until Stop is called.
// The first scan is used to build a baseline of modification times;
// the callback is not invoked during the initial scan.
func (w *FileWatcher) Start() {
	// Initial scan to populate modTimes (no notifications).
	w.scan(false)
	close(w.started)

	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			w.scan(true)
		case <-w.stopCh:
			return
		}
	}
}

// Stop signals the watcher to stop. It is safe to call from any goroutine.
// Calling Stop more than once will panic (closing a closed channel).
func (w *FileWatcher) Stop() {
	close(w.stopCh)
}

// scan walks the root directory and compares modification times.
// When notify is true it invokes the callback for every new, modified,
// or deleted .go file.
func (w *FileWatcher) scan(notify bool) {
	w.mu.Lock()
	defer w.mu.Unlock()

	currentFiles := make(map[string]time.Time)

	err := filepath.Walk(w.root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			// Log and skip entries we cannot read.
			log.Printf("watcher: error accessing %s: %v", path, err)
			return nil
		}

		if info.IsDir() {
			name := info.Name()
			// Skip hidden directories (e.g. .git, .golem) and node_modules.
			if strings.HasPrefix(name, ".") || name == "node_modules" {
				return filepath.SkipDir
			}
			return nil
		}

		// Only watch .go source files, excluding tests.
		if !strings.HasSuffix(path, ".go") {
			return nil
		}
		if strings.HasSuffix(path, "_test.go") {
			return nil
		}

		modTime := info.ModTime()
		currentFiles[path] = modTime

		if notify && w.callback != nil {
			prevTime, exists := w.modTimes[path]
			if !exists || modTime.After(prevTime) {
				w.callback(path)
			}
		}

		return nil
	})

	if err != nil {
		log.Printf("watcher: walk error: %v", err)
	}

	// Detect deleted files.
	if notify && w.callback != nil {
		for path := range w.modTimes {
			if _, exists := currentFiles[path]; !exists {
				w.callback(path)
			}
		}
	}

	w.modTimes = currentFiles
}
