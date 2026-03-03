//go:build !js || !wasm

package testutil

import (
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// ---------------------------------------------------------------------------
// Fixture tests
// ---------------------------------------------------------------------------

func TestNewFixture_CreatesDirectoryStructure(t *testing.T) {
	fix := NewFixture(t)

	// Verify root directory exists
	if _, err := os.Stat(fix.Dir()); err != nil {
		t.Fatalf("fixture root directory does not exist: %v", err)
	}

	// Verify standard subdirectories
	expectedDirs := []string{
		"src/app",
		"src/server",
		"build",
	}
	for _, rel := range expectedDirs {
		abs := filepath.Join(fix.Dir(), rel)
		info, err := os.Stat(abs)
		if err != nil {
			t.Errorf("expected directory %s to exist: %v", rel, err)
			continue
		}
		if !info.IsDir() {
			t.Errorf("expected %s to be a directory", rel)
		}
	}
}

func TestFixture_AddPage(t *testing.T) {
	fix := NewFixture(t)

	code := `package main

import "fmt"

func Page() { fmt.Println("hello") }
`
	fix.AddPage("", code)

	// Check root page
	rootPage := filepath.Join(fix.Dir(), "src", "app", "page.go")
	data, err := os.ReadFile(rootPage)
	if err != nil {
		t.Fatalf("root page.go not found: %v", err)
	}
	if string(data) != code {
		t.Error("root page.go content does not match")
	}

	// Add a nested page
	fix.AddPage("about", `package about

func Page() {}
`)

	aboutPage := filepath.Join(fix.Dir(), "src", "app", "about", "page.go")
	if _, err := os.Stat(aboutPage); err != nil {
		t.Fatalf("about/page.go not found: %v", err)
	}
}

func TestFixture_AddLayout(t *testing.T) {
	fix := NewFixture(t)

	layoutCode := `package main

func Layout() {}
`
	fix.AddLayout("", layoutCode)

	layoutPath := filepath.Join(fix.Dir(), "src", "app", "layout.go")
	data, err := os.ReadFile(layoutPath)
	if err != nil {
		t.Fatalf("layout.go not found: %v", err)
	}
	if string(data) != layoutCode {
		t.Error("layout.go content does not match")
	}
}

func TestFixture_AddServerFunction(t *testing.T) {
	fix := NewFixture(t)

	code := `package server

func Hello(name string) string {
	return "Hello, " + name
}
`
	fix.AddServerFunction("hello", code)

	funcPath := filepath.Join(fix.Dir(), "src", "server", "hello.go")
	data, err := os.ReadFile(funcPath)
	if err != nil {
		t.Fatalf("server function file not found: %v", err)
	}
	if string(data) != code {
		t.Error("server function file content does not match")
	}
}

func TestFixture_WriteGoMod(t *testing.T) {
	fix := NewFixture(t)
	fix.WriteGoMod()

	modPath := filepath.Join(fix.Dir(), "go.mod")
	data, err := os.ReadFile(modPath)
	if err != nil {
		t.Fatalf("go.mod not found: %v", err)
	}

	content := string(data)

	// Verify the go.mod has required fields
	checks := []string{
		"module testapp",
		"go ",
		"require github.com/Nu11ified/golem",
		"replace github.com/Nu11ified/golem",
	}
	for _, check := range checks {
		if !strings.Contains(content, check) {
			t.Errorf("go.mod missing expected content: %q\nFull content:\n%s", check, content)
		}
	}
}

func TestFixture_WriteFile(t *testing.T) {
	fix := NewFixture(t)
	fix.WriteFile("custom/dir/file.txt", "hello world")

	data, err := os.ReadFile(filepath.Join(fix.Dir(), "custom", "dir", "file.txt"))
	if err != nil {
		t.Fatalf("custom file not found: %v", err)
	}
	if string(data) != "hello world" {
		t.Errorf("expected 'hello world', got %q", string(data))
	}
}

func TestFixture_WriteGolemConfig(t *testing.T) {
	fix := NewFixture(t)
	fix.WriteGolemConfig()

	configPath := filepath.Join(fix.Dir(), "golem.config.json")
	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("golem.config.json not found: %v", err)
	}

	content := string(data)
	if !strings.Contains(content, `"projectName"`) {
		t.Error("golem.config.json missing projectName")
	}
	if !strings.Contains(content, `"testapp"`) {
		t.Error("golem.config.json missing testapp project name")
	}
}

func TestFixture_Dir(t *testing.T) {
	fix := NewFixture(t)
	if fix.Dir() == "" {
		t.Fatal("Dir() returned empty string")
	}
	if !filepath.IsAbs(fix.Dir()) {
		t.Errorf("Dir() should return an absolute path, got %q", fix.Dir())
	}
}

// ---------------------------------------------------------------------------
// WASM utility tests
// ---------------------------------------------------------------------------

func TestFindWASMExecJS(t *testing.T) {
	path, err := FindWASMExecJS()
	if err != nil {
		t.Fatalf("FindWASMExecJS failed: %v", err)
	}

	if !strings.HasSuffix(path, "wasm_exec.js") {
		t.Errorf("expected path ending in wasm_exec.js, got %q", path)
	}

	if _, err := os.Stat(path); err != nil {
		t.Fatalf("wasm_exec.js not found at returned path: %v", err)
	}
}

func TestSetupWASMExecJS(t *testing.T) {
	dir := t.TempDir()

	if err := SetupWASMExecJS(dir); err != nil {
		t.Fatalf("SetupWASMExecJS failed: %v", err)
	}

	destPath := filepath.Join(dir, "wasm_exec.js")
	info, err := os.Stat(destPath)
	if err != nil {
		t.Fatalf("wasm_exec.js not found in destination: %v", err)
	}
	if info.Size() == 0 {
		t.Error("wasm_exec.js is empty")
	}
}

func TestFindGolemRoot(t *testing.T) {
	root := findGolemRoot()
	if root == "" {
		t.Fatal("findGolemRoot returned empty string")
	}

	// Verify the root has a go.mod
	modPath := filepath.Join(root, "go.mod")
	data, err := os.ReadFile(modPath)
	if err != nil {
		t.Fatalf("go.mod not found at detected root %s: %v", root, err)
	}
	if !strings.Contains(string(data), "module github.com/Nu11ified/golem") {
		t.Errorf("go.mod at %s does not contain expected module path", root)
	}
}

// ---------------------------------------------------------------------------
// TestServer tests
// ---------------------------------------------------------------------------

func TestStartTestServer_HealthCheck(t *testing.T) {
	dir := t.TempDir()
	// Create a minimal build directory
	if err := os.MkdirAll(filepath.Join(dir, "build"), 0755); err != nil {
		t.Fatal(err)
	}

	srv, err := StartTestServer(dir)
	if err != nil {
		t.Fatalf("StartTestServer failed: %v", err)
	}
	defer srv.Stop()

	// Wait for server to be ready
	if err := srv.WaitReady(5 * time.Second); err != nil {
		t.Fatalf("server did not become ready: %v", err)
	}

	// Verify health endpoint
	resp, err := http.Get(srv.URL() + "/health")
	if err != nil {
		t.Fatalf("health check request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200 OK, got %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	if string(body) != "OK" {
		t.Errorf("expected body 'OK', got %q", string(body))
	}
}

func TestStartTestServer_ServesStaticFiles(t *testing.T) {
	dir := t.TempDir()
	buildDir := filepath.Join(dir, "build")
	if err := os.MkdirAll(buildDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Write a test file to the build directory
	testContent := "<html><body>hello</body></html>"
	if err := os.WriteFile(filepath.Join(buildDir, "index.html"), []byte(testContent), 0644); err != nil {
		t.Fatal(err)
	}

	srv, err := StartTestServer(dir)
	if err != nil {
		t.Fatalf("StartTestServer failed: %v", err)
	}
	defer srv.Stop()

	if err := srv.WaitReady(5 * time.Second); err != nil {
		t.Fatalf("server did not become ready: %v", err)
	}

	resp, err := http.Get(srv.URL() + "/index.html")
	if err != nil {
		t.Fatalf("static file request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200 OK, got %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	if string(body) != testContent {
		t.Errorf("expected %q, got %q", testContent, string(body))
	}
}

func TestTestServer_URL(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "build"), 0755); err != nil {
		t.Fatal(err)
	}

	srv, err := StartTestServer(dir)
	if err != nil {
		t.Fatalf("StartTestServer failed: %v", err)
	}
	defer srv.Stop()

	url := srv.URL()
	if !strings.HasPrefix(url, "http://127.0.0.1:") {
		t.Errorf("expected URL starting with http://127.0.0.1:, got %q", url)
	}
}

func TestTestServer_Port(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "build"), 0755); err != nil {
		t.Fatal(err)
	}

	srv, err := StartTestServer(dir)
	if err != nil {
		t.Fatalf("StartTestServer failed: %v", err)
	}
	defer srv.Stop()

	port := srv.Port()
	if port <= 0 || port > 65535 {
		t.Errorf("expected valid port number, got %d", port)
	}
}

func TestStartTestServerWithHandler(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("custom handler"))
	})

	srv, err := StartTestServerWithHandler(handler)
	if err != nil {
		t.Fatalf("StartTestServerWithHandler failed: %v", err)
	}
	defer srv.Stop()

	// Small delay to ensure server is listening
	time.Sleep(10 * time.Millisecond)

	resp, err := http.Get(srv.URL() + "/anything")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if string(body) != "custom handler" {
		t.Errorf("expected 'custom handler', got %q", string(body))
	}
}

func TestTestServer_WaitReady_Timeout(t *testing.T) {
	// Create a server that never responds with 200 on /health
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	})

	srv, err := StartTestServerWithHandler(handler)
	if err != nil {
		t.Fatalf("StartTestServerWithHandler failed: %v", err)
	}
	defer srv.Stop()

	// WaitReady should time out
	err = srv.WaitReady(100 * time.Millisecond)
	if err == nil {
		t.Error("expected WaitReady to return an error for non-ready server")
	}
	if !strings.Contains(err.Error(), "did not become ready") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestTestServer_Stop(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "build"), 0755); err != nil {
		t.Fatal(err)
	}

	srv, err := StartTestServer(dir)
	if err != nil {
		t.Fatalf("StartTestServer failed: %v", err)
	}

	if err := srv.WaitReady(5 * time.Second); err != nil {
		t.Fatalf("server did not become ready: %v", err)
	}

	url := srv.URL()
	srv.Stop()

	// Give the server a moment to fully shut down
	time.Sleep(50 * time.Millisecond)

	// After stop, the server should refuse connections
	client := &http.Client{Timeout: 1 * time.Second}
	_, err = client.Get(url + "/health")
	if err == nil {
		t.Error("expected connection error after Stop, but request succeeded")
	}
}

func TestTestServer_BuildDir(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "build"), 0755); err != nil {
		t.Fatal(err)
	}

	srv, err := StartTestServer(dir)
	if err != nil {
		t.Fatalf("StartTestServer failed: %v", err)
	}
	defer srv.Stop()

	expected := filepath.Join(dir, "build")
	if srv.BuildDir() != expected {
		t.Errorf("expected BuildDir() = %q, got %q", expected, srv.BuildDir())
	}
}
