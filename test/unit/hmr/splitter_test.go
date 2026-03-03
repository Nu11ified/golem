//go:build !js || !wasm

package hmr_test

import (
	"go/format"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Nu11ified/golem/internal/hmr"
)

// createTestApp creates a temporary directory structure for testing.
func createTestApp(t *testing.T, structure map[string]string) string {
	t.Helper()
	dir := t.TempDir()
	for path, content := range structure {
		fullPath := filepath.Join(dir, path)
		if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

func TestModuleSplitter_AnalyzeSimpleApp(t *testing.T) {
	appDir := createTestApp(t, map[string]string{
		"src/app/page.go":   "package app\n\nfunc Page(params map[string]string) interface{} { return nil }",
		"src/app/layout.go": "package app\n\nfunc Layout() interface{} { return nil }",
	})

	outputDir := filepath.Join(appDir, "build")
	splitter := hmr.NewModuleSplitter(filepath.Join(appDir, "src", "app"), outputDir, "github.com/example/myapp")

	modules, err := splitter.Analyze()
	if err != nil {
		t.Fatal(err)
	}

	// Should produce exactly 2 modules: 1 shell + 1 page
	if len(modules) != 2 {
		t.Fatalf("expected 2 modules (shell + 1 page), got %d", len(modules))
	}

	// Find shell and page modules
	var shell, page *hmr.ModuleInfo
	for i := range modules {
		if modules[i].IsShell {
			shell = &modules[i]
		} else {
			page = &modules[i]
		}
	}

	if shell == nil {
		t.Fatal("expected shell module to exist")
	}
	if page == nil {
		t.Fatal("expected page module to exist")
	}

	if shell.Name != "shell" {
		t.Errorf("expected shell module name 'shell', got '%s'", shell.Name)
	}
	if page.RoutePath != "/" {
		t.Errorf("expected page route path '/', got '%s'", page.RoutePath)
	}
}

func TestModuleSplitter_AnalyzeMultiplePages(t *testing.T) {
	appDir := createTestApp(t, map[string]string{
		"src/app/page.go":         "package app\n\nfunc Page(params map[string]string) interface{} { return nil }",
		"src/app/layout.go":       "package app\n\nfunc Layout() interface{} { return nil }",
		"src/app/about/page.go":   "package about\n\nfunc Page(params map[string]string) interface{} { return nil }",
		"src/app/contact/page.go": "package contact\n\nfunc Page(params map[string]string) interface{} { return nil }",
	})

	outputDir := filepath.Join(appDir, "build")
	splitter := hmr.NewModuleSplitter(filepath.Join(appDir, "src", "app"), outputDir, "github.com/example/myapp")

	modules, err := splitter.Analyze()
	if err != nil {
		t.Fatal(err)
	}

	// Should produce 4 modules: 1 shell + 3 pages
	if len(modules) != 4 {
		t.Fatalf("expected 4 modules (shell + 3 pages), got %d", len(modules))
	}

	shellCount := 0
	pageCount := 0
	for _, m := range modules {
		if m.IsShell {
			shellCount++
		} else {
			pageCount++
		}
	}
	if shellCount != 1 {
		t.Errorf("expected 1 shell module, got %d", shellCount)
	}
	if pageCount != 3 {
		t.Errorf("expected 3 page modules, got %d", pageCount)
	}
}

func TestModuleSplitter_PageModuleNaming(t *testing.T) {
	appDir := createTestApp(t, map[string]string{
		"src/app/page.go":                "package app\n\nfunc Page(params map[string]string) interface{} { return nil }",
		"src/app/about/page.go":          "package about\n\nfunc Page(params map[string]string) interface{} { return nil }",
		"src/app/blog/[slug]/page.go":    "package slug\n\nfunc Page(params map[string]string) interface{} { return nil }",
		"src/app/docs/[...path]/page.go": "package path\n\nfunc Page(params map[string]string) interface{} { return nil }",
	})

	outputDir := filepath.Join(appDir, "build")
	splitter := hmr.NewModuleSplitter(filepath.Join(appDir, "src", "app"), outputDir, "github.com/example/myapp")

	modules, err := splitter.Analyze()
	if err != nil {
		t.Fatal(err)
	}

	// Build a map of module names for easy lookup
	nameMap := make(map[string]bool)
	for _, m := range modules {
		if !m.IsShell {
			nameMap[m.Name] = true
		}
	}

	expectedNames := []string{
		"page_root",
		"page_about",
		"page_blog_slug",
		"page_docs_path",
	}

	for _, expected := range expectedNames {
		if !nameMap[expected] {
			t.Errorf("expected module name '%s' not found; got names: %v", expected, nameMap)
		}
	}
}

func TestModuleSplitter_DetermineAffected_PageChange(t *testing.T) {
	appDir := createTestApp(t, map[string]string{
		"src/app/page.go":       "package app\n\nfunc Page(params map[string]string) interface{} { return nil }",
		"src/app/layout.go":     "package app\n\nfunc Layout() interface{} { return nil }",
		"src/app/about/page.go": "package about\n\nfunc Page(params map[string]string) interface{} { return nil }",
	})

	outputDir := filepath.Join(appDir, "build")
	splitter := hmr.NewModuleSplitter(filepath.Join(appDir, "src", "app"), outputDir, "github.com/example/myapp")

	modules, err := splitter.Analyze()
	if err != nil {
		t.Fatal(err)
	}

	// Changing about/page.go should only affect the about page module
	changedFiles := []string{filepath.Join(appDir, "src", "app", "about", "page.go")}
	affected := splitter.DetermineAffectedModules(changedFiles, modules)

	if len(affected) != 1 {
		t.Fatalf("expected 1 affected module, got %d", len(affected))
	}
	if affected[0].Name != "page_about" {
		t.Errorf("expected affected module 'page_about', got '%s'", affected[0].Name)
	}
	if affected[0].IsShell {
		t.Error("affected module should not be the shell")
	}
}

func TestModuleSplitter_DetermineAffected_LayoutChange(t *testing.T) {
	appDir := createTestApp(t, map[string]string{
		"src/app/page.go":       "package app\n\nfunc Page(params map[string]string) interface{} { return nil }",
		"src/app/layout.go":     "package app\n\nfunc Layout() interface{} { return nil }",
		"src/app/about/page.go": "package about\n\nfunc Page(params map[string]string) interface{} { return nil }",
	})

	outputDir := filepath.Join(appDir, "build")
	splitter := hmr.NewModuleSplitter(filepath.Join(appDir, "src", "app"), outputDir, "github.com/example/myapp")

	modules, err := splitter.Analyze()
	if err != nil {
		t.Fatal(err)
	}

	// Changing layout.go should affect the shell module
	changedFiles := []string{filepath.Join(appDir, "src", "app", "layout.go")}
	affected := splitter.DetermineAffectedModules(changedFiles, modules)

	if len(affected) != 1 {
		t.Fatalf("expected 1 affected module, got %d", len(affected))
	}
	if !affected[0].IsShell {
		t.Error("expected affected module to be the shell")
	}
	if affected[0].Name != "shell" {
		t.Errorf("expected affected module name 'shell', got '%s'", affected[0].Name)
	}
}

func TestModuleSplitter_DetermineAffected_SharedCode(t *testing.T) {
	appDir := createTestApp(t, map[string]string{
		"src/app/page.go":              "package app\n\nfunc Page(params map[string]string) interface{} { return nil }",
		"src/app/layout.go":            "package app\n\nfunc Layout() interface{} { return nil }",
		"src/app/components/header.go": "package components\n\nfunc Header() interface{} { return nil }",
		"src/app/about/page.go":        "package about\n\nfunc Page(params map[string]string) interface{} { return nil }",
	})

	outputDir := filepath.Join(appDir, "build")
	splitter := hmr.NewModuleSplitter(filepath.Join(appDir, "src", "app"), outputDir, "github.com/example/myapp")

	modules, err := splitter.Analyze()
	if err != nil {
		t.Fatal(err)
	}

	// Changing a shared component should affect the shell module (full rebuild)
	changedFiles := []string{filepath.Join(appDir, "src", "app", "components", "header.go")}
	affected := splitter.DetermineAffectedModules(changedFiles, modules)

	if len(affected) != 1 {
		t.Fatalf("expected 1 affected module, got %d", len(affected))
	}
	if !affected[0].IsShell {
		t.Error("expected affected module to be the shell")
	}
}

func TestModuleSplitter_GeneratePageEntry_ValidGo(t *testing.T) {
	appDir := createTestApp(t, map[string]string{
		"src/app/about/page.go": "package about\n\nfunc Page(params map[string]string) interface{} { return nil }",
	})

	outputDir := filepath.Join(appDir, "build")
	splitter := hmr.NewModuleSplitter(filepath.Join(appDir, "src", "app"), outputDir, "github.com/example/myapp")

	module := hmr.ModuleInfo{
		Name:       "page_about",
		RoutePath:  "/about",
		IsShell:    false,
		EntryFile:  filepath.Join(appDir, "src", "app", "about", "page.go"),
		OutputPath: filepath.Join(outputDir, "page_about.wasm"),
	}

	code, err := splitter.GeneratePageEntry(module)
	if err != nil {
		t.Fatal(err)
	}

	// Verify the generated code is valid Go
	fset := token.NewFileSet()
	_, parseErr := parser.ParseFile(fset, "page_about_main.go", code, parser.AllErrors)
	if parseErr != nil {
		t.Fatalf("generated page entry is not valid Go:\n%s\nerror: %v", code, parseErr)
	}

	// Verify the generated code contains expected elements
	if !strings.Contains(code, "//go:wasmexport RenderPage") {
		t.Error("generated code should contain //go:wasmexport RenderPage directive")
	}
	if !strings.Contains(code, "func RenderPage()") {
		t.Error("generated code should contain RenderPage function")
	}
	if !strings.Contains(code, "func main()") {
		t.Error("generated code should contain main function")
	}
	if !strings.Contains(code, "package main") {
		t.Error("generated code should be in package main")
	}

	// Verify the code is properly formatted
	formatted, fmtErr := format.Source([]byte(code))
	if fmtErr != nil {
		t.Fatalf("generated code is not properly formatted: %v", fmtErr)
	}
	if string(formatted) != code {
		t.Error("generated code should already be formatted by go/format")
	}
}

func TestModuleSplitter_GenerateShellEntry_ValidGo(t *testing.T) {
	appDir := createTestApp(t, map[string]string{
		"src/app/page.go":       "package app\n\nfunc Page(params map[string]string) interface{} { return nil }",
		"src/app/layout.go":     "package app\n\nfunc Layout() interface{} { return nil }",
		"src/app/about/page.go": "package about\n\nfunc Page(params map[string]string) interface{} { return nil }",
	})

	outputDir := filepath.Join(appDir, "build")
	splitter := hmr.NewModuleSplitter(filepath.Join(appDir, "src", "app"), outputDir, "github.com/example/myapp")

	modules, err := splitter.Analyze()
	if err != nil {
		t.Fatal(err)
	}

	code, err := splitter.GenerateShellEntry(modules)
	if err != nil {
		t.Fatal(err)
	}

	// Verify the generated code is valid Go
	fset := token.NewFileSet()
	_, parseErr := parser.ParseFile(fset, "shell_main.go", code, parser.AllErrors)
	if parseErr != nil {
		t.Fatalf("generated shell entry is not valid Go:\n%s\nerror: %v", code, parseErr)
	}

	// Verify the generated code contains expected elements
	if !strings.Contains(code, "package main") {
		t.Error("generated code should be in package main")
	}
	if !strings.Contains(code, "func main()") {
		t.Error("generated code should contain main function")
	}

	// Verify the code is properly formatted
	formatted, fmtErr := format.Source([]byte(code))
	if fmtErr != nil {
		t.Fatalf("generated code is not properly formatted: %v", fmtErr)
	}
	if string(formatted) != code {
		t.Error("generated code should already be formatted by go/format")
	}
}
