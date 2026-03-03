package dev_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/Nu11ified/golem/internal/dev"
)

func TestFileWatcher_DetectsNewFile(t *testing.T) {
	dir := t.TempDir()
	changed := make(chan string, 10)

	watcher := dev.NewFileWatcher(dir, 50*time.Millisecond)
	watcher.OnChange(func(path string) { changed <- path })
	go watcher.Start()
	defer watcher.Stop()

	time.Sleep(150 * time.Millisecond) // wait for initial scan

	testFile := filepath.Join(dir, "test.go")
	os.WriteFile(testFile, []byte("package main"), 0644)

	select {
	case p := <-changed:
		if p != testFile {
			t.Errorf("expected %s, got %s", testFile, p)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("timeout waiting for new file detection")
	}
}

func TestFileWatcher_DetectsModifiedFile(t *testing.T) {
	dir := t.TempDir()
	testFile := filepath.Join(dir, "test.go")
	os.WriteFile(testFile, []byte("package main"), 0644)

	changed := make(chan string, 10)

	watcher := dev.NewFileWatcher(dir, 50*time.Millisecond)
	watcher.OnChange(func(path string) { changed <- path })
	go watcher.Start()
	defer watcher.Stop()

	time.Sleep(150 * time.Millisecond)

	// Need a delay so mtime is different
	time.Sleep(100 * time.Millisecond)
	os.WriteFile(testFile, []byte("package main\n// modified"), 0644)

	select {
	case <-changed:
		// success
	case <-time.After(3 * time.Second):
		t.Fatal("timeout waiting for modification detection")
	}
}

func TestFileWatcher_IgnoresNonGoFiles(t *testing.T) {
	dir := t.TempDir()
	changed := make(chan string, 10)

	watcher := dev.NewFileWatcher(dir, 50*time.Millisecond)
	watcher.OnChange(func(path string) { changed <- path })
	go watcher.Start()
	defer watcher.Stop()

	time.Sleep(150 * time.Millisecond)

	os.WriteFile(filepath.Join(dir, "notes.txt"), []byte("hello"), 0644)

	select {
	case <-changed:
		t.Fatal("should not detect non-Go file changes")
	case <-time.After(500 * time.Millisecond):
		// success
	}
}

func TestFileWatcher_DetectsDeletedFile(t *testing.T) {
	dir := t.TempDir()
	testFile := filepath.Join(dir, "test.go")
	os.WriteFile(testFile, []byte("package main"), 0644)

	changed := make(chan string, 10)

	watcher := dev.NewFileWatcher(dir, 50*time.Millisecond)
	watcher.OnChange(func(path string) { changed <- path })
	go watcher.Start()
	defer watcher.Stop()

	time.Sleep(150 * time.Millisecond)

	os.Remove(testFile)

	select {
	case <-changed:
		// success
	case <-time.After(3 * time.Second):
		t.Fatal("timeout waiting for deletion detection")
	}
}

func TestFileWatcher_IgnoresHiddenDirs(t *testing.T) {
	dir := t.TempDir()
	hiddenDir := filepath.Join(dir, ".hidden")
	os.MkdirAll(hiddenDir, 0755)

	changed := make(chan string, 10)

	watcher := dev.NewFileWatcher(dir, 50*time.Millisecond)
	watcher.OnChange(func(path string) { changed <- path })
	go watcher.Start()
	defer watcher.Stop()

	time.Sleep(150 * time.Millisecond)

	os.WriteFile(filepath.Join(hiddenDir, "secret.go"), []byte("package hidden"), 0644)

	select {
	case <-changed:
		t.Fatal("should not detect files in hidden directories")
	case <-time.After(500 * time.Millisecond):
		// success
	}
}
