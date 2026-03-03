package dev

import (
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// FileOp represents a file operation type.
type FileOp int

const (
	// OpCreate indicates a file was created.
	OpCreate FileOp = iota
	// OpModify indicates a file was modified.
	OpModify
	// OpDelete indicates a file was deleted.
	OpDelete
)

// String returns the string representation of a FileOp.
func (op FileOp) String() string {
	switch op {
	case OpCreate:
		return "CREATE"
	case OpModify:
		return "MODIFY"
	case OpDelete:
		return "DELETE"
	default:
		return "UNKNOWN"
	}
}

// FileEvent represents a file system change event.
type FileEvent struct {
	Path string
	Op   FileOp
}

// FileWatcher watches directories for file changes and emits events.
type FileWatcher struct {
	dirs      []string
	Events    chan FileEvent
	done      chan struct{}
	mu        sync.Mutex
	running   bool
	interval  time.Duration
	lastMod   map[string]time.Time
	lastExist map[string]bool
}

// NewFileWatcher creates a new FileWatcher that watches the given directories.
func NewFileWatcher(dirs []string) *FileWatcher {
	return &FileWatcher{
		dirs:      dirs,
		Events:    make(chan FileEvent, 100),
		done:      make(chan struct{}),
		interval:  500 * time.Millisecond,
		lastMod:   make(map[string]time.Time),
		lastExist: make(map[string]bool),
	}
}

// Start begins watching for file changes. It runs in the background.
func (fw *FileWatcher) Start() error {
	fw.mu.Lock()
	defer fw.mu.Unlock()

	if fw.running {
		return nil
	}

	// Do initial scan
	if err := fw.scan(false); err != nil {
		return err
	}

	fw.running = true
	go fw.watchLoop()
	return nil
}

// Stop stops the file watcher.
func (fw *FileWatcher) Stop() {
	fw.mu.Lock()
	defer fw.mu.Unlock()

	if !fw.running {
		return
	}

	fw.running = false
	close(fw.done)
}

func (fw *FileWatcher) watchLoop() {
	ticker := time.NewTicker(fw.interval)
	defer ticker.Stop()

	for {
		select {
		case <-fw.done:
			return
		case <-ticker.C:
			if err := fw.scan(true); err != nil {
				log.Printf("File watcher scan error: %v", err)
			}
		}
	}
}

func (fw *FileWatcher) scan(emitEvents bool) error {
	currentFiles := make(map[string]bool)

	for _, dir := range fw.dirs {
		err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil // skip files we can't access
			}
			if info.IsDir() {
				return nil
			}

			currentFiles[path] = true
			modTime := info.ModTime()

			if emitEvents {
				if lastMod, exists := fw.lastMod[path]; exists {
					if modTime.After(lastMod) {
						fw.Events <- FileEvent{Path: path, Op: OpModify}
					}
				} else {
					fw.Events <- FileEvent{Path: path, Op: OpCreate}
				}
			}

			fw.lastMod[path] = modTime
			fw.lastExist[path] = true
			return nil
		})
		if err != nil {
			return err
		}
	}

	if emitEvents {
		// Check for deleted files
		for path := range fw.lastExist {
			if !currentFiles[path] {
				fw.Events <- FileEvent{Path: path, Op: OpDelete}
				delete(fw.lastMod, path)
				delete(fw.lastExist, path)
			}
		}
	}

	return nil
}
