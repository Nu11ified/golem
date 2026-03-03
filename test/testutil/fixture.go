//go:build !js || !wasm

// Package testutil provides reusable test utilities for end-to-end testing
// of Golem applications. It includes helpers for scaffolding test app
// fixtures, compiling WASM binaries, and running dev servers.
package testutil

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

// Fixture represents a scaffolded test application with a known directory
// structure. It creates a temporary directory and provides methods to add
// pages, layouts, server functions, and a go.mod file that references the
// local Golem module.
type Fixture struct {
	t   *testing.T
	dir string
}

// NewFixture creates a new test fixture in a temporary directory.
// The directory is automatically removed when the test completes via
// t.Cleanup, but callers can also call Cleanup explicitly.
func NewFixture(t *testing.T) *Fixture {
	t.Helper()

	dir := t.TempDir() // automatically cleaned up by testing framework

	f := &Fixture{
		t:   t,
		dir: dir,
	}

	// Create standard directory structure
	dirs := []string{
		filepath.Join(dir, "src", "app"),
		filepath.Join(dir, "src", "server"),
		filepath.Join(dir, "build"),
	}
	for _, d := range dirs {
		if err := os.MkdirAll(d, 0755); err != nil {
			t.Fatalf("testutil.NewFixture: failed to create directory %s: %v", d, err)
		}
	}

	return f
}

// Dir returns the root directory of the fixture.
func (f *Fixture) Dir() string {
	return f.dir
}

// AddPage creates a page component file at the specified route path.
// The routePath is relative to src/app (e.g., "" for the root page,
// "about" for src/app/about/page.go). The componentCode should be
// valid Go source that builds under GOOS=js GOARCH=wasm.
func (f *Fixture) AddPage(routePath string, componentCode string) {
	f.t.Helper()

	pageDir := filepath.Join(f.dir, "src", "app")
	if routePath != "" {
		pageDir = filepath.Join(pageDir, routePath)
	}

	if err := os.MkdirAll(pageDir, 0755); err != nil {
		f.t.Fatalf("testutil.AddPage: failed to create directory %s: %v", pageDir, err)
	}

	pagePath := filepath.Join(pageDir, "page.go")
	if err := os.WriteFile(pagePath, []byte(componentCode), 0644); err != nil {
		f.t.Fatalf("testutil.AddPage: failed to write %s: %v", pagePath, err)
	}
}

// AddLayout creates a layout file at the specified route path.
// The routePath follows the same convention as AddPage (e.g., "" for root).
func (f *Fixture) AddLayout(routePath string, layoutCode string) {
	f.t.Helper()

	layoutDir := filepath.Join(f.dir, "src", "app")
	if routePath != "" {
		layoutDir = filepath.Join(layoutDir, routePath)
	}

	if err := os.MkdirAll(layoutDir, 0755); err != nil {
		f.t.Fatalf("testutil.AddLayout: failed to create directory %s: %v", layoutDir, err)
	}

	layoutPath := filepath.Join(layoutDir, "layout.go")
	if err := os.WriteFile(layoutPath, []byte(layoutCode), 0644); err != nil {
		f.t.Fatalf("testutil.AddLayout: failed to write %s: %v", layoutPath, err)
	}
}

// AddServerFunction creates a server-side function file in the src/server
// directory. The name is used as the filename (without .go extension).
func (f *Fixture) AddServerFunction(name string, code string) {
	f.t.Helper()

	serverDir := filepath.Join(f.dir, "src", "server")
	if err := os.MkdirAll(serverDir, 0755); err != nil {
		f.t.Fatalf("testutil.AddServerFunction: failed to create directory %s: %v", serverDir, err)
	}

	funcPath := filepath.Join(serverDir, name+".go")
	if err := os.WriteFile(funcPath, []byte(code), 0644); err != nil {
		f.t.Fatalf("testutil.AddServerFunction: failed to write %s: %v", funcPath, err)
	}
}

// WriteGoMod writes a go.mod file that uses a replace directive to point
// to the local Golem module. The moduleName defaults to "testapp" if empty.
func (f *Fixture) WriteGoMod() {
	f.t.Helper()

	golemRoot := findGolemRoot()
	if golemRoot == "" {
		f.t.Fatal("testutil.WriteGoMod: could not locate Golem module root (no go.mod found)")
	}

	goVersion := runtime.Version()
	// Extract numeric version (e.g., "go1.23" from "go1.23.0")
	goVer := "1.23.0"
	if len(goVersion) > 2 {
		goVer = goVersion[2:] // strip "go" prefix
	}

	content := fmt.Sprintf(`module testapp

go %s

require github.com/Nu11ified/golem v0.0.0

replace github.com/Nu11ified/golem => %s
`, goVer, golemRoot)

	modPath := filepath.Join(f.dir, "go.mod")
	if err := os.WriteFile(modPath, []byte(content), 0644); err != nil {
		f.t.Fatalf("testutil.WriteGoMod: failed to write %s: %v", modPath, err)
	}
}

// WriteFile writes arbitrary content to a path relative to the fixture root.
func (f *Fixture) WriteFile(relPath string, content string) {
	f.t.Helper()

	absPath := filepath.Join(f.dir, relPath)
	dir := filepath.Dir(absPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		f.t.Fatalf("testutil.WriteFile: failed to create directory %s: %v", dir, err)
	}

	if err := os.WriteFile(absPath, []byte(content), 0644); err != nil {
		f.t.Fatalf("testutil.WriteFile: failed to write %s: %v", absPath, err)
	}
}

// WriteGolemConfig writes a golem.config.json file with sensible defaults
// for testing.
func (f *Fixture) WriteGolemConfig() {
	f.t.Helper()

	config := `{
  "projectName": "testapp",
  "version": "0.0.1",
  "entry": "src/app/",
  "output": "build",
  "dev": {
    "port": 0,
    "hotReload": false,
    "watch": ["src"]
  },
  "build": {
    "minify": false,
    "target": "wasm",
    "sourcemap": false
  },
  "server": {
    "grpc": {
      "port": 0,
      "reflection": false
    },
    "functions": "src/server"
  },
  "wasm": {
    "optimizeSize": false,
    "enableFeatures": []
  }
}`

	configPath := filepath.Join(f.dir, "golem.config.json")
	if err := os.WriteFile(configPath, []byte(config), 0644); err != nil {
		f.t.Fatalf("testutil.WriteGolemConfig: failed to write %s: %v", configPath, err)
	}
}

// Cleanup removes the fixture directory. This is called automatically by
// testing.T.TempDir cleanup, but is provided for explicit use if needed.
func (f *Fixture) Cleanup() {
	os.RemoveAll(f.dir)
}

// findGolemRoot walks up from the current file's directory to find the
// Golem module root (the directory containing go.mod with module
// github.com/Nu11ified/golem).
func findGolemRoot() string {
	// Start from the directory of this source file
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		return ""
	}

	dir := filepath.Dir(filename)
	for {
		modPath := filepath.Join(dir, "go.mod")
		if data, err := os.ReadFile(modPath); err == nil {
			content := string(data)
			if len(content) > 0 && contains(content, "module github.com/Nu11ified/golem") {
				return dir
			}
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break // reached filesystem root
		}
		dir = parent
	}

	return ""
}

// contains checks if s contains substr. Avoids importing strings package
// just for this one check.
func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
