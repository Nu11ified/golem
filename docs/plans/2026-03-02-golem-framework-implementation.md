# Golem Framework Overhaul Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Transform Golem from a broken proof-of-concept into a working Go/WASM framework with Next.js-equivalent core features, validated by a comprehensive test suite.

**Architecture:** Two-phase compiler (convention scanning + code generation → Go WASM compilation). Dual rendering via build tags (server HTML + client WASM hydration). Shell WASM + swappable page modules for HMR.

**Tech Stack:** Go 1.24+, WebAssembly, Playwright (e2e), standard `testing` package, `net/http` server, WebSocket (nhooyr.io/websocket)

**Design Doc:** `docs/plans/2026-03-02-golem-framework-overhaul-design.md`

---

## Phase 1: Fix Build Pipeline (Foundation)

The build pipeline is the first thing that must work. Nothing else matters until `go build` produces a valid WASM binary and the dev server can serve it.

### Task 1: Unit Tests for DOM RenderToHTML

The stub file `dom/element_stub.go` currently returns `"<div>"` from Render(). We need it to produce complete HTML. Start with tests.

**Files:**
- Create: `test/unit/dom/render_html_test.go`

**Step 1: Write failing tests**

```go
//go:build !js || !wasm

package dom_test

import (
	"strings"
	"testing"

	"github.com/Nu11ified/golem/dom"
)

func TestRenderToHTML_SimpleDiv(t *testing.T) {
	el := dom.Div()
	html := dom.RenderToHTML(el)
	if html != "<div></div>" {
		t.Errorf("expected <div></div>, got %s", html)
	}
}

func TestRenderToHTML_DivWithClass(t *testing.T) {
	el := dom.Div(dom.Class("container"))
	html := dom.RenderToHTML(el)
	if html != `<div class="container"></div>` {
		t.Errorf("expected <div class=\"container\"></div>, got %s", html)
	}
}

func TestRenderToHTML_DivWithId(t *testing.T) {
	el := dom.Div(dom.Id("app"))
	html := dom.RenderToHTML(el)
	if html != `<div id="app"></div>` {
		t.Errorf("expected <div id=\"app\"></div>, got %s", html)
	}
}

func TestRenderToHTML_NestedElements(t *testing.T) {
	el := dom.Div(
		dom.H1("Hello"),
		dom.P("World"),
	)
	html := dom.RenderToHTML(el)
	if !strings.Contains(html, "<h1>Hello</h1>") {
		t.Errorf("expected h1 in output, got %s", html)
	}
	if !strings.Contains(html, "<p>World</p>") {
		t.Errorf("expected p in output, got %s", html)
	}
}

func TestRenderToHTML_TextContent(t *testing.T) {
	el := dom.Span(dom.Text("hello"))
	html := dom.RenderToHTML(el)
	if html != "<span>hello</span>" {
		t.Errorf("expected <span>hello</span>, got %s", html)
	}
}

func TestRenderToHTML_InputSelfClosing(t *testing.T) {
	el := dom.Input(dom.Type("text"), dom.Placeholder("Enter name"))
	html := dom.RenderToHTML(el)
	if !strings.Contains(html, "type=\"text\"") {
		t.Errorf("expected type attr, got %s", html)
	}
	if !strings.Contains(html, "placeholder=\"Enter name\"") {
		t.Errorf("expected placeholder attr, got %s", html)
	}
}

func TestRenderToHTML_BooleanAttributes(t *testing.T) {
	el := dom.Input(dom.Checked(true), dom.Disabled(true))
	html := dom.RenderToHTML(el)
	if !strings.Contains(html, "checked") {
		t.Errorf("expected checked attr, got %s", html)
	}
	if !strings.Contains(html, "disabled") {
		t.Errorf("expected disabled attr, got %s", html)
	}
}

func TestRenderToHTML_ComplexTree(t *testing.T) {
	el := dom.Div(dom.Class("app"),
		dom.Div(dom.Class("sidebar"),
			dom.Ul(
				dom.Li("Item 1"),
				dom.Li("Item 2"),
			),
		),
		dom.Div(dom.Class("content"),
			dom.H1("Title"),
			dom.P("Paragraph text"),
		),
	)
	html := dom.RenderToHTML(el)
	if !strings.Contains(html, `class="app"`) {
		t.Errorf("expected app class, got %s", html)
	}
	if !strings.Contains(html, `class="sidebar"`) {
		t.Errorf("expected sidebar class, got %s", html)
	}
	if !strings.Contains(html, "<li>Item 1</li>") {
		t.Errorf("expected list item, got %s", html)
	}
}

func TestRenderToHTML_EventHandlersOmitted(t *testing.T) {
	el := dom.Button(dom.Text("Click"), dom.OnClick(func() {}))
	html := dom.RenderToHTML(el)
	if strings.Contains(html, "onclick") {
		t.Errorf("event handlers should not appear in HTML, got %s", html)
	}
	if !strings.Contains(html, "Click") {
		t.Errorf("expected button text, got %s", html)
	}
}
```

**Step 2: Run tests to verify they fail**

Run: `cd /data/github/golem && go test ./test/unit/dom/ -v -count=1`
Expected: FAIL — `RenderToHTML` not defined

**Step 3: Implement RenderToHTML in element_stub.go**

Replace the current `element_stub.go` with a full server-side renderer. The file at `dom/element_stub.go` needs these additions:
- `RenderToHTML(*Element) string` function
- `renderElementToHTML(*Element) string` recursive helper
- Proper attribute rendering, self-closing tags, text nodes
- HTML escaping for attribute values and text content
- Event handlers omitted from HTML output (they're client-side only)

Self-closing tags: `input`, `img`, `br`, `hr`, `meta`, `link`

Key implementation details:
- `textContent` prop renders as inner text, not an attribute
- `class` prop renders as `class="value"`
- Boolean props (`checked`, `disabled`, `autofocus`) render as bare attributes when true
- Event handler props (`onclick`) are skipped entirely
- Children render recursively
- Text nodes (Type == "text") render their `textContent` prop directly

Also add the missing helper functions that exist in element.go but not in the stub:
- `Placeholder`, `Value`, `Type`, `Checked`, `Autofocus` (some missing from stub)
- `OnInput`, `OnChange`, `OnKeyDown` (as no-ops for server)
- `Checkbox`, `Label`, `H3`, `H4` (missing from stub)

**Step 4: Run tests to verify they pass**

Run: `cd /data/github/golem && go test ./test/unit/dom/ -v -count=1`
Expected: PASS

**Step 5: Commit**

```bash
git add test/unit/dom/render_html_test.go dom/element_stub.go
git commit -m "feat(dom): add RenderToHTML for server-side rendering"
```

---

### Task 2: Fix WASM Compilation Pipeline

The dev server's `buildDevWasm()` uses a broken `createWasmMainFile()` that merges two Go files. Replace with direct compilation of `./src/app/`.

**Files:**
- Modify: `internal/dev/server.go` (lines 400-601)
- Create: `test/integration/build_pipeline_test.go`
- Create: `src/app/main.go` (minimal test app for the framework)

**Step 1: Write failing integration test**

```go
package integration_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestWASMCompilation(t *testing.T) {
	// Verify we can compile the src/app/ to WASM
	outDir := t.TempDir()
	outFile := filepath.Join(outDir, "app.wasm")

	cmd := exec.Command("go", "build", "-o", outFile, "./src/app/")
	cmd.Env = append(os.Environ(), "GOOS=js", "GOARCH=wasm")
	cmd.Dir = "/data/github/golem"
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("WASM compilation failed: %v\nOutput: %s", err, output)
	}

	// Verify the file exists and is non-empty
	info, err := os.Stat(outFile)
	if err != nil {
		t.Fatalf("WASM output file not found: %v", err)
	}
	if info.Size() == 0 {
		t.Fatal("WASM output file is empty")
	}
}

func TestWasmExecJSExists(t *testing.T) {
	// Verify we can find wasm_exec.js from Go installation
	goRootCmd := exec.Command("go", "env", "GOROOT")
	goRootBytes, err := goRootCmd.Output()
	if err != nil {
		t.Fatalf("Could not get GOROOT: %v", err)
	}
	goRoot := string(goRootBytes)
	goRoot = goRoot[:len(goRoot)-1] // trim newline

	paths := []string{
		filepath.Join(goRoot, "lib", "wasm", "wasm_exec.js"),
		filepath.Join(goRoot, "misc", "wasm", "wasm_exec.js"),
	}

	found := false
	for _, p := range paths {
		if _, err := os.Stat(p); err == nil {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("wasm_exec.js not found in Go installation")
	}
}
```

**Step 2: Run test to verify compilation state**

Run: `cd /data/github/golem && go test ./test/integration/ -run TestWASM -v -count=1`
Expected: FAIL — `src/app/main.go` may not compile cleanly for WASM

**Step 3: Create a minimal working src/app/main.go**

Replace the current `src/app/main.go` (in the golem-test-app, or create one at the repo root) with a minimal app that compiles:

```go
//go:build js && wasm

package main

import (
	"github.com/Nu11ified/golem/dom"
)

func main() {
	app := dom.Div(
		dom.Class("app"),
		dom.H1("Golem Framework"),
		dom.P("App loaded successfully."),
	)
	dom.Render(app, "#app")

	// Keep the Go runtime alive
	select {}
}
```

Create `src/app/` directory at repo root if it doesn't exist.

**Step 4: Fix buildDevWasm in server.go**

Replace the `buildDevWasm()` method to:
1. Find and copy `wasm_exec.js` (keep existing logic, it works)
2. Remove `createWasmMainFile()` entirely
3. Compile directly: `go build -o .golem/dev/app.wasm ./src/app/`

The key change in `buildDevWasm()`:
```go
// Build command: compile src/app/ package directly
wasmOutput := filepath.Join(devDir, "app.wasm")
buildArgs := []string{"build", "-o", wasmOutput, "./src/app/"}
cmd := exec.Command("go", buildArgs...)
cmd.Dir = "."
cmd.Env = env
cmd.Stdout = os.Stdout
cmd.Stderr = os.Stderr
```

Remove the `createWasmMainFile()` method entirely.

**Step 5: Run tests to verify they pass**

Run: `cd /data/github/golem && go test ./test/integration/ -run TestWASM -v -count=1`
Expected: PASS

**Step 6: Commit**

```bash
git add src/app/main.go internal/dev/server.go test/integration/build_pipeline_test.go
git commit -m "fix(build): compile src/app/ directly instead of merging files"
```

---

### Task 3: Fix Dev Server Static File Serving

The dev server generates HTML and serves static files but has issues with the HTML template and CORS.

**Files:**
- Modify: `internal/dev/server.go` (generateDevHTML method)
- Create: `test/integration/dev_server_test.go`

**Step 1: Write failing test**

```go
package integration_test

import (
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/Nu11ified/golem/internal/config"
	"github.com/Nu11ified/golem/internal/dev"
)

func TestDevServerStartsAndServesHTML(t *testing.T) {
	cfg := &config.Config{
		ProjectName: "test-app",
		Output:      ".golem/build",
		Dev: config.DevConfig{
			Port:      0, // will need a free port
			HotReload: false,
		},
		Server: config.ServerConfig{
			GRPC: config.GRPCConfig{Port: 0},
		},
	}

	// We test the HTML generation directly instead of starting a full server
	s := dev.NewServer(cfg)
	html := s.GenerateDevHTML()

	if !strings.Contains(html, `<div id="app">`) {
		t.Error("HTML should contain app div")
	}
	if !strings.Contains(html, "wasm_exec.js") {
		t.Error("HTML should reference wasm_exec.js")
	}
	if !strings.Contains(html, "app.wasm") {
		t.Error("HTML should reference app.wasm")
	}
	if !strings.Contains(html, "test-app") {
		t.Error("HTML should contain project name")
	}
}
```

**Step 2: Run to verify failure**

Run: `cd /data/github/golem && go test ./test/integration/ -run TestDevServer -v -count=1`
Expected: FAIL — `GenerateDevHTML` is not exported

**Step 3: Export GenerateDevHTML and clean up HTML template**

In `internal/dev/server.go`:
- Rename `generateDevHTML()` to `GenerateDevHTML()` (export it)
- Update the caller `generateDevFiles()` to use `s.GenerateDevHTML()`
- Clean up the HTML template: remove hardcoded styles (they'll come from CSS-in-Go), improve error display

**Step 4: Run tests**

Run: `cd /data/github/golem && go test ./test/integration/ -run TestDevServer -v -count=1`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/dev/server.go test/integration/dev_server_test.go
git commit -m "fix(dev): export GenerateDevHTML and clean up HTML template"
```

---

### Task 4: Implement File Watching

The dev server's `watchFiles()` is stubbed. Implement polling-based file watching.

**Files:**
- Create: `internal/dev/watcher.go`
- Create: `test/unit/dev/watcher_test.go`
- Modify: `internal/dev/server.go` (integrate watcher)

**Step 1: Write failing tests for FileWatcher**

```go
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

	// Wait for initial scan
	time.Sleep(100 * time.Millisecond)

	// Create a new file
	testFile := filepath.Join(dir, "test.go")
	os.WriteFile(testFile, []byte("package main"), 0644)

	select {
	case <-changed:
		// success
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for file change detection")
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

	// Wait for initial scan
	time.Sleep(100 * time.Millisecond)

	// Modify the file
	os.WriteFile(testFile, []byte("package main\n// modified"), 0644)

	select {
	case <-changed:
		// success
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for file modification detection")
	}
}

func TestFileWatcher_IgnoresNonGoFiles(t *testing.T) {
	dir := t.TempDir()
	changed := make(chan string, 10)

	watcher := dev.NewFileWatcher(dir, 50*time.Millisecond)
	watcher.OnChange(func(path string) { changed <- path })
	go watcher.Start()
	defer watcher.Stop()

	time.Sleep(100 * time.Millisecond)

	// Create a non-Go file
	os.WriteFile(filepath.Join(dir, "notes.txt"), []byte("hello"), 0644)

	select {
	case <-changed:
		t.Fatal("should not detect non-Go file changes")
	case <-time.After(300 * time.Millisecond):
		// success — no change detected
	}
}
```

**Step 2: Verify failure**

Run: `cd /data/github/golem && go test ./test/unit/dev/ -v -count=1`
Expected: FAIL — `NewFileWatcher` not defined

**Step 3: Implement FileWatcher**

Create `internal/dev/watcher.go`:
- `FileWatcher` struct with: root dir, poll interval, file mod times map, callback, stop channel
- `NewFileWatcher(dir string, interval time.Duration) *FileWatcher`
- `OnChange(callback func(path string))`
- `Start()` — polling loop that walks dir, compares mtimes, calls callback on changes
- `Stop()` — sends on stop channel
- Only watches `.go` files (skip `_test.go`, skip `.golem/` directory)

**Step 4: Verify pass**

Run: `cd /data/github/golem && go test ./test/unit/dev/ -v -count=1`
Expected: PASS

**Step 5: Integrate watcher into dev server**

In `internal/dev/server.go`, replace the stubbed `watchFiles()`:
```go
func (s *Server) watchFiles() {
	watcher := NewFileWatcher("src", 500*time.Millisecond)
	watcher.OnChange(func(path string) {
		log.Printf("File changed: %s — rebuilding...", path)
		start := time.Now()
		if err := s.buildDevWasm(); err != nil {
			log.Printf("Build failed: %v", err)
			s.broadcastError(err.Error())
			return
		}
		log.Printf("Rebuild completed in %v", time.Since(start))
		s.broadcastReload()
	})
	watcher.Start()
}
```

**Step 6: Commit**

```bash
git add internal/dev/watcher.go test/unit/dev/watcher_test.go internal/dev/server.go
git commit -m "feat(dev): implement file watching with polling"
```

---

### Task 5: Implement WebSocket Hot Reload Broadcast

The WebSocket handler accepts connections but never sends messages. Implement broadcast.

**Files:**
- Create: `internal/dev/broadcast.go`
- Modify: `internal/dev/server.go` (WebSocket handler)
- Create: `test/unit/dev/broadcast_test.go`

**Step 1: Write failing tests**

```go
package dev_test

import (
	"testing"

	"github.com/Nu11ified/golem/internal/dev"
)

func TestBroadcaster_AddRemoveClient(t *testing.T) {
	b := dev.NewBroadcaster()
	ch := make(chan string, 1)
	id := b.AddClient(ch)

	if b.ClientCount() != 1 {
		t.Errorf("expected 1 client, got %d", b.ClientCount())
	}

	b.RemoveClient(id)
	if b.ClientCount() != 0 {
		t.Errorf("expected 0 clients, got %d", b.ClientCount())
	}
}

func TestBroadcaster_SendReload(t *testing.T) {
	b := dev.NewBroadcaster()
	ch := make(chan string, 1)
	b.AddClient(ch)

	b.SendReload()

	msg := <-ch
	if msg != "reload" {
		t.Errorf("expected 'reload', got %s", msg)
	}
}

func TestBroadcaster_SendError(t *testing.T) {
	b := dev.NewBroadcaster()
	ch := make(chan string, 1)
	b.AddClient(ch)

	b.SendError("compile error: syntax issue")

	msg := <-ch
	if msg != `{"type":"error","message":"compile error: syntax issue"}` {
		t.Errorf("unexpected error message: %s", msg)
	}
}
```

**Step 2: Verify failure**

Run: `cd /data/github/golem && go test ./test/unit/dev/ -v -count=1`
Expected: FAIL — `NewBroadcaster` not defined

**Step 3: Implement Broadcaster**

Create `internal/dev/broadcast.go`:
- `Broadcaster` struct with: clients map (id→chan string), mutex, counter
- `NewBroadcaster() *Broadcaster`
- `AddClient(ch chan string) int` — returns client ID
- `RemoveClient(id int)`
- `ClientCount() int`
- `SendReload()` — sends `"reload"` to all clients
- `SendError(msg string)` — sends JSON error message to all clients

**Step 4: Update WebSocket handler in server.go**

Replace `handleWebSocket` to:
1. Create a channel for the client
2. Register with broadcaster
3. Write loop: read from channel, write to WebSocket
4. On disconnect: remove from broadcaster

Add `broadcaster *Broadcaster` field to Server struct. Initialize in `NewServer()`.

Add `broadcastReload()` and `broadcastError(msg string)` methods to Server.

**Step 5: Verify pass**

Run: `cd /data/github/golem && go test ./test/unit/dev/ -v -count=1`
Expected: PASS

**Step 6: Commit**

```bash
git add internal/dev/broadcast.go test/unit/dev/broadcast_test.go internal/dev/server.go
git commit -m "feat(dev): implement WebSocket hot reload broadcasting"
```

---

## Phase 2: Core Routing & Code Generation

### Task 6: Route Convention Scanner

Build the core of the two-phase compiler: scan `src/app/` directory conventions and produce a route tree.

**Files:**
- Create: `internal/codegen/scanner.go`
- Create: `test/unit/codegen/scanner_test.go`

**Step 1: Write failing tests**

Test cases:
- `src/app/page.go` → route `/`
- `src/app/about/page.go` → route `/about`
- `src/app/blog/[slug]/page.go` → route `/blog/:slug`
- `src/app/docs/[...path]/page.go` → route `/docs/*path`
- `src/app/(marketing)/pricing/page.go` → route `/pricing` (group ignored)
- `src/app/layout.go` → root layout detected
- `src/app/blog/layout.go` → nested layout detected
- `src/app/error.go` → error boundary detected
- `src/app/loading.go` → loading state detected
- `src/app/@sidebar/page.go` → parallel slot "sidebar" detected
- Directory with no `page.go` → not a route (layout-only segment)

The scanner should produce a `RouteTree` data structure:
```go
type ScannedRoute struct {
    Path        string            // URL path pattern
    Segment     string            // directory name
    PageFile    string            // absolute path to page.go
    LayoutFile  string            // path to layout.go (if exists)
    ErrorFile   string            // path to error.go (if exists)
    LoadingFile string            // path to loading.go (if exists)
    NotFoundFile string           // path to notfound.go (if exists)
    TemplateFile string           // path to template.go (if exists)
    Children    []*ScannedRoute
    ParamName   string            // for dynamic segments like [slug]
    IsCatchAll  bool              // for [...path]
    IsOptionalCatchAll bool       // for [[...path]]
    IsGroup     bool              // for (marketing)
    ParallelSlots map[string]*ScannedRoute // for @sidebar
    InterceptPattern string       // for (.)route, (..)route
}
```

**Step 2: Implement scanner**

`internal/codegen/scanner.go`:
- `ScanRoutes(appDir string) (*ScannedRoute, error)`
- Walks directory tree recursively
- Detects special files by name
- Parses directory naming conventions: `[param]`, `[...catchAll]`, `[[...optional]]`, `(group)`, `@slot`, `(.)intercept`
- Returns tree of `ScannedRoute`

**Step 3: Verify pass**

Run: `cd /data/github/golem && go test ./test/unit/codegen/ -v -count=1`
Expected: PASS

**Step 4: Commit**

```bash
git add internal/codegen/scanner.go test/unit/codegen/scanner_test.go
git commit -m "feat(codegen): implement file-based route convention scanner"
```

---

### Task 7: Route Code Generator

Generate `routes_gen.go` from the scanned route tree.

**Files:**
- Create: `internal/codegen/generator.go`
- Create: `test/unit/codegen/generator_test.go`

**Step 1: Write failing tests**

Test that the generator produces valid Go source code that:
- Imports all page/layout/error packages
- Registers routes with the router
- Wires layouts as wrappers
- Wires error boundaries
- Handles dynamic params
- Handles parallel slots

The generated code should be parseable by `go/parser`.

**Step 2: Implement generator**

`internal/codegen/generator.go`:
- `GenerateRoutes(tree *ScannedRoute, moduleName string) (string, error)`
- Produces a valid Go file with `package main`
- Uses `text/template` for code generation
- Imports route packages based on directory paths
- Generates `func registerRoutes(r *router.Router)` function
- Each route calls `r.AddRoute(...)` with the correct path and component function
- Layouts wrap children: `layoutFn(childElement)`
- Error boundaries wrap in recover blocks

**Step 3: Verify the generated code parses**

Use `go/parser.ParseFile` in tests to verify output is valid Go.

Run: `cd /data/github/golem && go test ./test/unit/codegen/ -v -count=1`
Expected: PASS

**Step 4: Commit**

```bash
git add internal/codegen/generator.go test/unit/codegen/generator_test.go
git commit -m "feat(codegen): implement route code generator"
```

---

### Task 8: Integrate Code Generation into Build Pipeline

Wire the scanner + generator into `golem dev` and `golem build`.

**Files:**
- Modify: `internal/dev/server.go` (call codegen before WASM build)
- Modify: `internal/build/builder.go` (call codegen before WASM build)
- Create: `internal/codegen/codegen.go` (top-level API)

**Step 1: Create codegen entry point**

`internal/codegen/codegen.go`:
```go
package codegen

func Generate(appDir, outputFile, moduleName string) error {
    tree, err := ScanRoutes(appDir)
    if err != nil { return err }
    code, err := GenerateRoutes(tree, moduleName)
    if err != nil { return err }
    return os.WriteFile(outputFile, []byte(code), 0644)
}
```

**Step 2: Call from dev server's buildDevWasm()**

Before running `go build`, call:
```go
codegen.Generate("src/app", "src/app/routes_gen.go", moduleName)
```

**Step 3: Call from builder's Build()**

Same integration in the production build path.

**Step 4: Test end-to-end**

Create a test fixture directory with page.go/layout.go files, run codegen, compile to WASM, verify no errors.

**Step 5: Commit**

```bash
git add internal/codegen/codegen.go internal/dev/server.go internal/build/builder.go
git commit -m "feat(build): integrate route code generation into build pipeline"
```

---

### Task 9: Router Enhancements for Layouts and Error Boundaries

Enhance the router to support layout wrapping, error recovery, and loading states.

**Files:**
- Modify: `router/router.go`
- Create: `test/unit/router/layout_test.go`

**Step 1: Write failing tests**

Test layout wrapping:
```go
func TestRouter_LayoutWrapping(t *testing.T) {
    // Route with layout should wrap page in layout
}

func TestRouter_NestedLayouts(t *testing.T) {
    // Child route wrapped in parent layout + child layout
}

func TestRouter_ErrorBoundary(t *testing.T) {
    // Panicking component caught by error boundary
}
```

**Step 2: Add to Route struct**

```go
type Route struct {
    // ... existing fields ...
    Layout      func(children *dom.Element) *dom.Element
    ErrorHandler func(err error) *dom.Element
    LoadingHandler func() *dom.Element
    ParallelSlots map[string]*Route  // @slot routes
}
```

**Step 3: Update renderComponent**

When rendering a route:
1. If route has error handler, wrap rendering in `defer/recover`
2. Render page component
3. If route has layout, wrap: `layout(page)`
4. Walk up parent chain, wrapping in each parent's layout

**Step 4: Verify pass**

Run: `cd /data/github/golem && go test ./test/unit/router/ -v -count=1`
Expected: PASS

**Step 5: Commit**

```bash
git add router/router.go router/router_stub.go test/unit/router/
git commit -m "feat(router): add layout wrapping and error boundaries"
```

---

## Phase 3: Server-Side Rendering & Hydration

### Task 10: Complete Server-Side Element Rendering

Finish the full `RenderToHTML` implementation and ensure all element types work.

**Files:**
- Modify: `dom/element_stub.go`
- Extend: `test/unit/dom/render_html_test.go`

Add tests and implementation for:
- HTML escaping (XSS prevention)
- Style attribute rendering
- Data attributes (`data-golem-id` for hydration)
- Full document rendering (`<!DOCTYPE html>` wrapper)
- Proper whitespace handling

**Commit message:** `feat(dom): complete server-side HTML rendering with escaping`

---

### Task 11: SSR Request Handler

Create the server-side rendering HTTP handler that renders pages to HTML on request.

**Files:**
- Create: `internal/ssr/renderer.go`
- Create: `test/unit/ssr/renderer_test.go`

The renderer:
1. Takes a route path
2. Matches it against the route tree
3. Runs the page function (native Go, not WASM)
4. Wraps in layouts
5. Produces full HTML document with WASM bootstrap script
6. Returns HTML string

**Commit message:** `feat(ssr): implement server-side rendering request handler`

---

### Task 12: Client-Side Hydration

Create `dom.Hydrate()` that attaches event handlers to existing server-rendered DOM.

**Files:**
- Modify: `dom/element.go` (add Hydrate function, WASM build)
- Create: `test/e2e/ssr_hydration_test.go`

`Hydrate(element *Element, selector string)`:
1. Find existing DOM node via `querySelector`
2. Walk the element tree and the DOM tree in parallel
3. Attach event handlers to matching DOM nodes
4. Do NOT replace innerHTML — preserve server-rendered content
5. Use `data-golem-id` attributes for matching

**Commit message:** `feat(dom): implement client-side hydration`

---

## Phase 4: Static Generation & ISR

### Task 13: Static Page Generation at Build Time

**Files:**
- Create: `internal/staticgen/generator.go`
- Create: `test/unit/staticgen/generator_test.go`

At build time:
1. Scan all routes
2. For routes with no dynamic segments → render to HTML file
3. For routes with `GenerateStaticParams()` → render each param combination
4. Output to `.golem/build/` directory

**Commit message:** `feat(build): implement static page generation`

---

### Task 14: ISR Cache Layer

**Files:**
- Create: `internal/cache/cache.go`
- Create: `test/unit/cache/cache_test.go`

Implement:
- `Cache` struct with: entries map, mutex, TTL per entry
- `Get(key string) ([]byte, bool, bool)` — returns (data, found, stale)
- `Set(key string, data []byte, ttl time.Duration)`
- `Invalidate(key string)`
- `InvalidateByTag(tag string)`
- `SetTags(key string, tags []string)`
- Background goroutine for TTL expiry

**Commit message:** `feat(cache): implement ISR cache with TTL and tag invalidation`

---

### Task 15: ISR Server Integration

**Files:**
- Modify: `internal/server/server.go`
- Create: `test/integration/isr_test.go`

Wire ISR into the production server:
1. On request, check cache
2. If cached and fresh → serve immediately
3. If cached and stale → serve stale, trigger background regeneration
4. If not cached → render, cache, serve
5. Support `RevalidatePath` and `RevalidateTag` API endpoints

**Commit message:** `feat(server): integrate ISR caching into production server`

---

## Phase 5: Component Lifecycle & State

### Task 16: Component Lifecycle Hooks

**Files:**
- Modify: `dom/element.go` (add lifecycle hooks)
- Create: `test/unit/dom/lifecycle_test.go`

Add to Element:
- `OnMount(fn func() func())` — returns cleanup function
- `OnUnmount(fn func())`
- `OnUpdate(fn func())`
- Track mount state per element
- Call mount hooks after `appendChild`, unmount hooks before `removeChild`

**Commit message:** `feat(dom): implement component lifecycle hooks`

---

### Task 17: Enhanced Server Function Stubs

**Files:**
- Create: `internal/codegen/serverstubs.go`
- Create: `test/unit/codegen/serverstubs_test.go`

At build time, scan `src/server/` directory:
1. Parse exported function signatures
2. Generate type-safe WASM client stubs
3. Each stub marshals args to JSON, calls `/api/functions`, unmarshals response
4. Output to `src/server/stubs_gen.go` (WASM build tag)

**Commit message:** `feat(codegen): generate type-safe server function client stubs`

---

## Phase 6: Middleware

### Task 18: Server Middleware Pipeline

**Files:**
- Create: `internal/middleware/middleware.go`
- Create: `test/unit/middleware/middleware_test.go`

Implement:
- `Middleware` type: `func(req *Request, next func(*Request) *Response) *Response`
- `MiddlewareConfig` with `Matcher` field (glob patterns)
- Pipeline that chains middleware in order
- Discover `middleware.go` at project root
- Wire into HTTP server before route handling

**Commit message:** `feat(middleware): implement server middleware pipeline`

---

## Phase 7: HMR (Hot Module Replacement)

### Task 19: WASM Module Splitting

**Files:**
- Create: `internal/hmr/splitter.go`
- Create: `test/unit/hmr/splitter_test.go`

Split the app into:
1. **Shell module**: Router, state, layouts. Compiled as main WASM binary.
2. **Page modules**: Each page compiled as a separate WASM binary with `//go:wasmexport`.

The splitter:
- Analyzes import graph to determine shell vs page code
- Generates separate build commands for each module
- Manages module registry (which page → which WASM file)

**Commit message:** `feat(hmr): implement WASM module splitting`

---

### Task 20: HMR JS Bridge

**Files:**
- Create: `internal/dev/hmr_bridge.js` (embedded in Go via `embed`)
- Modify: `internal/dev/server.go` (embed and serve bridge)

JavaScript bridge that:
1. Listens for WebSocket messages
2. On `module_update` message: fetch new WASM module, instantiate, swap into shell
3. Preserve state by not touching shell WASM
4. Fallback to full reload for shell changes

**Commit message:** `feat(hmr): implement JS bridge for module hot-swapping`

---

### Task 21: HMR Integration

**Files:**
- Modify: `internal/dev/server.go`
- Create: `test/e2e/hmr_test.go`

Wire everything together:
1. File watcher detects change in `src/app/about/page.go`
2. Determine: is this a page or shell file?
3. If page: recompile only that page module
4. Send `module_update` via WebSocket with module URL
5. JS bridge swaps module
6. If shell: full rebuild + reload

**Commit message:** `feat(hmr): integrate hot module replacement into dev server`

---

## Phase 8: Parallel & Intercepting Routes

### Task 22: Parallel Routes

**Files:**
- Modify: `internal/codegen/scanner.go` (detect `@slot` dirs)
- Modify: `internal/codegen/generator.go` (generate slot-aware routing)
- Modify: `router/router.go` (render parallel slots)
- Create: `test/unit/router/parallel_test.go`

**Step 1:** Scanner detects `@` prefix, adds to `ParallelSlots` map
**Step 2:** Generator creates route registration with slot mappings
**Step 3:** Router renders each slot independently into layout's slots map
**Step 4:** Layout receives `map[string]*dom.Element` of slots

**Commit message:** `feat(router): implement parallel routes with @slot convention`

---

### Task 23: Intercepting Routes

**Files:**
- Modify: `internal/codegen/scanner.go` (detect `(.)` patterns)
- Modify: `router/router.go` (origin-aware routing)
- Create: `test/unit/router/intercepting_test.go`

**Step 1:** Scanner detects `(.)`, `(..)`, `(...)` prefixes
**Step 2:** Router tracks navigation origin (client-side vs direct URL)
**Step 3:** Client nav: render intercepted version in a slot/modal
**Step 4:** Direct URL: render full page version

**Commit message:** `feat(router): implement intercepting routes`

---

## Phase 9: E2E Test Suite

### Task 24: Test Infrastructure

**Files:**
- Create: `test/testutil/server.go` — start/stop dev server for tests
- Create: `test/testutil/wasm.go` — WASM compilation helpers
- Create: `test/testutil/fixture.go` — scaffolding test apps with known structure
- Create: `test/e2e/setup_test.go` — Playwright test setup

The test infrastructure needs to:
1. Create a temporary directory with a valid Golem app
2. Run `golem dev` (or start the server programmatically)
3. Wait for server to be ready
4. Provide URL to Playwright
5. Clean up after tests

**Commit message:** `feat(test): create test infrastructure for e2e tests`

---

### Task 25: Core E2E Tests

**Files:**
- Create: `test/e2e/basic_rendering_test.go`
- Create: `test/e2e/navigation_test.go`
- Create: `test/e2e/state_reactivity_test.go`
- Create: `test/e2e/server_functions_test.go`

Each test:
1. Scaffolds a minimal Golem app with specific features
2. Starts the dev server
3. Opens browser via Playwright
4. Verifies behavior (DOM content, navigation, state updates, server calls)
5. Tears down

**Commit message:** `feat(test): add core e2e browser tests`

---

### Task 26: Advanced E2E Tests

**Files:**
- Create: `test/e2e/ssr_hydration_test.go`
- Create: `test/e2e/layout_persistence_test.go`
- Create: `test/e2e/error_boundary_test.go`
- Create: `test/e2e/hmr_browser_test.go`
- Create: `test/e2e/isr_browser_test.go`
- Create: `test/e2e/parallel_rendering_test.go`
- Create: `test/e2e/intercepting_modal_test.go`

Each test verifies one advanced feature end-to-end in the browser.

**Commit message:** `feat(test): add advanced e2e tests for SSR, HMR, ISR, parallel routes`

---

## Summary: Implementation Order

| # | Task | Phase | Depends On |
|---|------|-------|------------|
| 1 | DOM RenderToHTML | 1 | — |
| 2 | Fix WASM compilation | 1 | — |
| 3 | Fix dev server HTML | 1 | 2 |
| 4 | File watching | 1 | — |
| 5 | WebSocket broadcast | 1 | 4 |
| 6 | Route convention scanner | 2 | — |
| 7 | Route code generator | 2 | 6 |
| 8 | Codegen build integration | 2 | 2, 7 |
| 9 | Router layouts/errors | 2 | — |
| 10 | Complete SSR rendering | 3 | 1 |
| 11 | SSR request handler | 3 | 10, 8 |
| 12 | Client hydration | 3 | 10 |
| 13 | Static generation | 4 | 11 |
| 14 | ISR cache layer | 4 | — |
| 15 | ISR server integration | 4 | 13, 14 |
| 16 | Component lifecycle | 5 | 9 |
| 17 | Server function stubs | 5 | 6 |
| 18 | Server middleware | 6 | 8 |
| 19 | WASM module splitting | 7 | 2 |
| 20 | HMR JS bridge | 7 | 19 |
| 21 | HMR integration | 7 | 5, 20 |
| 22 | Parallel routes | 8 | 6, 9 |
| 23 | Intercepting routes | 8 | 6, 9 |
| 24 | E2E test infra | 9 | 2, 3 |
| 25 | Core E2E tests | 9 | 24, 8 |
| 26 | Advanced E2E tests | 9 | 25, all features |

**Parallelizable groups:**
- Tasks 1, 2, 4, 6, 9, 14 can all start simultaneously (no dependencies)
- Tasks 3, 5, 7, 10 can start once their single dependency completes
- Tasks 19-21 (HMR) and 22-23 (parallel/intercepting) are independent tracks
