//go:build !js || !wasm

package staticgen_test

import (
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Nu11ified/golem/dom"
	"github.com/Nu11ified/golem/internal/staticgen"
	"github.com/Nu11ified/golem/router"
)

func setupRouter() *router.Router {
	r := router.NewRouter()

	r.AddSimpleRoute("/", func(params map[string]string) *dom.Element {
		return dom.Div(dom.H1(dom.Text("Home")), dom.P(dom.Text("Welcome to the home page")))
	})

	r.AddSimpleRoute("/about", func(params map[string]string) *dom.Element {
		return dom.Div(dom.H1(dom.Text("About")), dom.P(dom.Text("About us")))
	})

	r.AddSimpleRoute("/blog/:slug", func(params map[string]string) *dom.Element {
		slug := params["slug"]
		return dom.Div(dom.H1(dom.Text("Blog: "+slug)), dom.P(dom.Text("Blog content for "+slug)))
	})

	r.AddSimpleRoute("/contact", func(params map[string]string) *dom.Element {
		return dom.Div(dom.H1(dom.Text("Contact")))
	})

	return r
}

func defaultDocOpts() dom.DocumentOptions {
	return dom.DocumentOptions{
		Title: "Test Site",
		Lang:  "en",
	}
}

func defaultConfig(outputDir string) staticgen.GeneratorConfig {
	return staticgen.GeneratorConfig{
		OutputDir:      outputDir,
		Concurrency:    2,
		DefaultDocOpts: defaultDocOpts(),
	}
}

func TestGenerateSinglePage(t *testing.T) {
	tmpDir := t.TempDir()
	r := setupRouter()
	gen := staticgen.NewGenerator(r, defaultDocOpts(), defaultConfig(tmpDir))

	results, err := gen.Generate([]staticgen.PageConfig{{Path: "/"}})
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Error != nil {
		t.Fatalf("page generation error: %v", results[0].Error)
	}
	if results[0].Path != "/" {
		t.Errorf("expected path '/', got '%s'", results[0].Path)
	}
	if results[0].OutFile != "index.html" {
		t.Errorf("expected outfile 'index.html', got '%s'", results[0].OutFile)
	}
	if results[0].HTML == "" {
		t.Error("expected non-empty HTML")
	}
}

func TestGenerateMultiplePagesConcurrently(t *testing.T) {
	tmpDir := t.TempDir()
	r := setupRouter()
	gen := staticgen.NewGenerator(r, defaultDocOpts(), defaultConfig(tmpDir))

	pages := []staticgen.PageConfig{
		{Path: "/"},
		{Path: "/about"},
		{Path: "/contact"},
	}

	results, err := gen.Generate(pages)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(results))
	}

	for _, result := range results {
		if result.Error != nil {
			t.Errorf("page %s error: %v", result.Path, result.Error)
		}
		if result.HTML == "" {
			t.Errorf("page %s has empty HTML", result.Path)
		}
	}

	// Verify ordering preserved
	if results[0].Path != "/" || results[1].Path != "/about" || results[2].Path != "/contact" {
		t.Error("result ordering not preserved")
	}
}

func TestCustomDocumentOptionsPerPage(t *testing.T) {
	tmpDir := t.TempDir()
	r := setupRouter()
	gen := staticgen.NewGenerator(r, defaultDocOpts(), defaultConfig(tmpDir))

	customOpts := &dom.DocumentOptions{
		Title: "About - Custom Title",
		Lang:  "fr",
		Meta:  map[string]string{"description": "A custom about page"},
	}

	pages := []staticgen.PageConfig{
		{Path: "/"},
		{Path: "/about", DocumentOptions: customOpts},
	}

	results, err := gen.Generate(pages)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	if !strings.Contains(results[0].HTML, "<title>Test Site</title>") {
		t.Error("default page should use default title")
	}
	if !strings.Contains(results[1].HTML, "<title>About - Custom Title</title>") {
		t.Error("custom page should use custom title")
	}
	if !strings.Contains(results[1].HTML, `lang="fr"`) {
		t.Error("custom page should use lang 'fr'")
	}
	if !strings.Contains(results[1].HTML, `name="description"`) {
		t.Error("custom page should have description meta tag")
	}
}

func TestErrorHandlingRouteNotFound(t *testing.T) {
	tmpDir := t.TempDir()
	r := setupRouter()
	gen := staticgen.NewGenerator(r, defaultDocOpts(), defaultConfig(tmpDir))

	results, err := gen.Generate([]staticgen.PageConfig{{Path: "/nonexistent"}})
	if err != nil {
		t.Fatalf("Generate should not return top-level error: %v", err)
	}

	if results[0].Error == nil {
		t.Error("expected error for nonexistent route")
	}
	if !strings.Contains(results[0].Error.Error(), "route not found") {
		t.Errorf("expected 'route not found' error, got: %v", results[0].Error)
	}
}

func TestOutputFileCreation(t *testing.T) {
	tmpDir := t.TempDir()
	r := setupRouter()
	gen := staticgen.NewGenerator(r, defaultDocOpts(), defaultConfig(tmpDir))

	pages := []staticgen.PageConfig{
		{Path: "/"},
		{Path: "/about"},
		{Path: "/blog/hello-world"},
	}

	results, err := gen.GenerateToFiles(pages)
	if err != nil {
		t.Fatalf("GenerateToFiles failed: %v", err)
	}

	expectedFiles := map[string]string{
		"/":                 "index.html",
		"/about":            filepath.Join("about", "index.html"),
		"/blog/hello-world": filepath.Join("blog", "hello-world", "index.html"),
	}

	for _, result := range results {
		if result.Error != nil {
			t.Errorf("page %s error: %v", result.Path, result.Error)
			continue
		}
		expected := expectedFiles[result.Path]
		if result.OutFile != expected {
			t.Errorf("path %s: expected outfile '%s', got '%s'", result.Path, expected, result.OutFile)
		}

		fullPath := filepath.Join(tmpDir, result.OutFile)
		data, err := os.ReadFile(fullPath)
		if err != nil {
			t.Errorf("failed to read file %s: %v", fullPath, err)
			continue
		}
		if string(data) != result.HTML {
			t.Errorf("file content mismatch for %s", result.Path)
		}
	}
}

func TestFallbackPageGeneration(t *testing.T) {
	tmpDir := t.TempDir()
	r := setupRouter()
	gen := staticgen.NewGenerator(r, defaultDocOpts(), defaultConfig(tmpDir))

	results, err := gen.Generate([]staticgen.PageConfig{
		{Path: "/dynamic/unknown", Fallback: staticgen.FallbackStatic},
	})
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	if results[0].Error != nil {
		t.Errorf("expected no error with FallbackStatic, got: %v", results[0].Error)
	}
	if !strings.Contains(results[0].HTML, "Loading...") {
		t.Error("fallback page should contain 'Loading...'")
	}
	if !strings.Contains(results[0].HTML, "<!DOCTYPE html>") {
		t.Error("fallback page should be a valid HTML document")
	}
}

func TestRevalidationConfig(t *testing.T) {
	tmpDir := t.TempDir()
	r := setupRouter()
	gen := staticgen.NewGenerator(r, defaultDocOpts(), defaultConfig(tmpDir))

	pages := []staticgen.PageConfig{
		{Path: "/", Revalidate: 0},
		{Path: "/about", Revalidate: 60 * time.Second},
		{Path: "/contact", Revalidate: 5 * time.Minute},
	}

	results, err := gen.Generate(pages)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	if results[0].Config.Revalidate != 0 {
		t.Errorf("expected revalidate 0, got %v", results[0].Config.Revalidate)
	}
	if results[1].Config.Revalidate != 60*time.Second {
		t.Errorf("expected revalidate 60s, got %v", results[1].Config.Revalidate)
	}
	if results[2].Config.Revalidate != 5*time.Minute {
		t.Errorf("expected revalidate 5m, got %v", results[2].Config.Revalidate)
	}
}

func TestGenerationWithLayoutsAndParallelSlots(t *testing.T) {
	tmpDir := t.TempDir()
	r := router.NewRouter()

	r.AddRoute(&router.Route{
		Path: "/with-layout",
		Component: func(params map[string]string) *dom.Element {
			return dom.Div(dom.P(dom.Text("Page content")))
		},
		Layout: func(child *dom.Element) *dom.Element {
			return dom.Div(
				dom.Class("layout"),
				dom.NewElement("header", dom.Text("Site Header")),
				child,
				dom.NewElement("footer", dom.Text("Site Footer")),
			)
		},
	})

	r.AddRoute(&router.Route{
		Path: "/with-slots",
		Component: func(params map[string]string) *dom.Element {
			return dom.Div(dom.P(dom.Text("Main content")))
		},
		ParallelSlots: map[string]*router.Route{
			"sidebar": {
				Component: func(params map[string]string) *dom.Element {
					return dom.Div(dom.Text("Sidebar content"))
				},
			},
		},
	})

	gen := staticgen.NewGenerator(r, defaultDocOpts(), defaultConfig(tmpDir))

	results, err := gen.Generate([]staticgen.PageConfig{
		{Path: "/with-layout"},
		{Path: "/with-slots"},
	})
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	layoutHTML := results[0].HTML
	if !strings.Contains(layoutHTML, "Site Header") {
		t.Error("layout page should contain 'Site Header'")
	}
	if !strings.Contains(layoutHTML, "Page content") {
		t.Error("layout page should contain 'Page content'")
	}
	if !strings.Contains(layoutHTML, "Site Footer") {
		t.Error("layout page should contain 'Site Footer'")
	}

	slotsHTML := results[1].HTML
	if !strings.Contains(slotsHTML, "Main content") {
		t.Error("slots page should contain 'Main content'")
	}
	if !strings.Contains(slotsHTML, "Sidebar content") {
		t.Error("slots page should contain 'Sidebar content'")
	}
	if !strings.Contains(slotsHTML, `data-golem-slot="sidebar"`) {
		t.Error("slots page should contain slot wrapper with data-golem-slot attribute")
	}
}

func TestHTMLOutputContent(t *testing.T) {
	tmpDir := t.TempDir()
	r := setupRouter()

	opts := dom.DocumentOptions{
		Title:        "My Site",
		Lang:         "en",
		Meta:         map[string]string{"viewport": "width=device-width, initial-scale=1.0"},
		Scripts:      []string{"/js/app.js"},
		Styles:       []string{"/css/style.css"},
		WasmPath:     "/app.wasm",
		WasmExecPath: "/wasm_exec.js",
	}

	config := staticgen.GeneratorConfig{
		OutputDir:      tmpDir,
		Concurrency:    1,
		DefaultDocOpts: opts,
	}

	gen := staticgen.NewGenerator(r, opts, config)
	results, err := gen.Generate([]staticgen.PageConfig{{Path: "/"}})
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	html := results[0].HTML

	checks := map[string]string{
		"DOCTYPE":       "<!DOCTYPE html>",
		"html lang":     `<html lang="en">`,
		"title":         "<title>My Site</title>",
		"viewport meta": `name="viewport"`,
		"stylesheet":    `href="/css/style.css"`,
		"script":        `src="/js/app.js"`,
		"wasm_exec":     `src="/wasm_exec.js"`,
		"app container": `<div id="app">`,
		"page content":  "Home",
	}

	for name, substr := range checks {
		if !strings.Contains(html, substr) {
			t.Errorf("should contain %s (%s)", name, substr)
		}
	}
}

func TestConcurrencyLimiting(t *testing.T) {
	tmpDir := t.TempDir()
	r := router.NewRouter()

	var active int64
	var maxActive int64

	for i := 0; i < 20; i++ {
		path := "/" + string(rune('a'+i))
		r.AddSimpleRoute(path, func(params map[string]string) *dom.Element {
			current := atomic.AddInt64(&active, 1)
			for {
				old := atomic.LoadInt64(&maxActive)
				if current <= old || atomic.CompareAndSwapInt64(&maxActive, old, current) {
					break
				}
			}
			time.Sleep(10 * time.Millisecond)
			atomic.AddInt64(&active, -1)
			return dom.Div(dom.Text("Page"))
		})
	}

	config := staticgen.GeneratorConfig{
		OutputDir:      tmpDir,
		Concurrency:    3,
		DefaultDocOpts: defaultDocOpts(),
	}

	gen := staticgen.NewGenerator(r, defaultDocOpts(), config)

	pages := make([]staticgen.PageConfig, 20)
	for i := 0; i < 20; i++ {
		pages[i] = staticgen.PageConfig{Path: "/" + string(rune('a'+i))}
	}

	results, err := gen.Generate(pages)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	for _, result := range results {
		if result.Error != nil {
			t.Errorf("page %s error: %v", result.Path, result.Error)
		}
	}

	observed := atomic.LoadInt64(&maxActive)
	if observed > 3 {
		t.Errorf("concurrency exceeded limit: max active was %d, limit was 3", observed)
	}
	if len(results) != 20 {
		t.Errorf("expected 20 results, got %d", len(results))
	}
}

func TestFallbackBlockingMode(t *testing.T) {
	tmpDir := t.TempDir()
	r := setupRouter()
	gen := staticgen.NewGenerator(r, defaultDocOpts(), defaultConfig(tmpDir))

	results, err := gen.Generate([]staticgen.PageConfig{
		{Path: "/does-not-exist", Fallback: staticgen.FallbackBlocking},
	})
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	if results[0].Error != nil {
		t.Errorf("FallbackBlocking should not propagate error, got: %v", results[0].Error)
	}
	if !strings.Contains(results[0].HTML, "Loading...") {
		t.Error("FallbackBlocking should produce loading page")
	}
}

func TestFallbackNonePropagatesErrors(t *testing.T) {
	tmpDir := t.TempDir()
	r := setupRouter()
	gen := staticgen.NewGenerator(r, defaultDocOpts(), defaultConfig(tmpDir))

	results, err := gen.Generate([]staticgen.PageConfig{
		{Path: "/does-not-exist", Fallback: staticgen.FallbackNone},
	})
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	if results[0].Error == nil {
		t.Error("FallbackNone should propagate error for missing route")
	}
}

func TestGenerateToFilesSkipsErrors(t *testing.T) {
	tmpDir := t.TempDir()
	r := setupRouter()
	gen := staticgen.NewGenerator(r, defaultDocOpts(), defaultConfig(tmpDir))

	pages := []staticgen.PageConfig{
		{Path: "/"},
		{Path: "/nonexistent", Fallback: staticgen.FallbackNone},
		{Path: "/about"},
	}

	results, err := gen.GenerateToFiles(pages)
	if err != nil {
		t.Fatalf("GenerateToFiles failed: %v", err)
	}

	for _, idx := range []int{0, 2} {
		if results[idx].Error != nil {
			t.Errorf("page %s should succeed: %v", results[idx].Path, results[idx].Error)
			continue
		}
		fullPath := filepath.Join(tmpDir, results[idx].OutFile)
		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			t.Errorf("file should exist: %s", fullPath)
		}
	}

	if results[1].Error == nil {
		t.Error("nonexistent page should have error")
	}
}

func TestDefaultConcurrency(t *testing.T) {
	tmpDir := t.TempDir()
	r := setupRouter()

	config := staticgen.GeneratorConfig{
		OutputDir:      tmpDir,
		Concurrency:    0,
		DefaultDocOpts: defaultDocOpts(),
	}

	gen := staticgen.NewGenerator(r, defaultDocOpts(), config)
	results, err := gen.Generate([]staticgen.PageConfig{{Path: "/"}, {Path: "/about"}})
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	for _, result := range results {
		if result.Error != nil {
			t.Errorf("page %s error: %v", result.Path, result.Error)
		}
	}
}

func TestPathToFilename(t *testing.T) {
	tmpDir := t.TempDir()
	r := setupRouter()
	gen := staticgen.NewGenerator(r, defaultDocOpts(), defaultConfig(tmpDir))

	tests := []struct {
		path     string
		expected string
	}{
		{"/", "index.html"},
		{"/about", filepath.Join("about", "index.html")},
		{"/blog/my-post", filepath.Join("blog", "my-post", "index.html")},
	}

	for _, tc := range tests {
		results, _ := gen.Generate([]staticgen.PageConfig{{Path: tc.path}})
		if results[0].OutFile != tc.expected {
			t.Errorf("path %s: expected '%s', got '%s'", tc.path, tc.expected, results[0].OutFile)
		}
	}
}

func TestFallbackWithCustomTitle(t *testing.T) {
	tmpDir := t.TempDir()
	r := setupRouter()
	gen := staticgen.NewGenerator(r, defaultDocOpts(), defaultConfig(tmpDir))

	customOpts := &dom.DocumentOptions{Title: "Loading Page"}
	results, err := gen.Generate([]staticgen.PageConfig{
		{Path: "/dynamic/unknown", Fallback: staticgen.FallbackStatic, DocumentOptions: customOpts},
	})
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	if !strings.Contains(results[0].HTML, "<title>Loading Page</title>") {
		t.Error("fallback page should use custom title 'Loading Page'")
	}
}
