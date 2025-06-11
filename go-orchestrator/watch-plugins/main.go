// To use this watcher, run: go get github.com/fsnotify/fsnotify
package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
)

func buildPlugins() {
	cmd := exec.Command("go", "run", filepath.Join(".", "build-plugins", "main.go"))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "[watch-plugins] Build failed: %v\n", err)
	}
}

func main() {
	goDir := filepath.Join("..", "user-app", "server", "go")
	w, err := fsnotify.NewWatcher()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create watcher: %v\n", err)
		os.Exit(1)
	}
	defer w.Close()

	err = w.Add(goDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to watch %s: %v\n", goDir, err)
		os.Exit(1)
	}

	fmt.Println("[watch-plugins] Initial build...")
	buildPlugins()
	lastBuild := time.Now()

	for {
		select {
		case event, ok := <-w.Events:
			if !ok {
				return
			}
			if event.Op&(fsnotify.Write|fsnotify.Create|fsnotify.Remove|fsnotify.Rename) != 0 && strings.HasSuffix(event.Name, ".go") {
				if time.Since(lastBuild) > 500*time.Millisecond {
					fmt.Printf("[watch-plugins] Detected change: %s\n", event)
					buildPlugins()
					lastBuild = time.Now()
				}
			}
		case err, ok := <-w.Errors:
			if !ok {
				return
			}
			fmt.Fprintf(os.Stderr, "[watch-plugins] Watch error: %v\n", err)
		}
	}
}
