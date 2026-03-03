package codegen_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Nu11ified/golem/internal/codegen"
)

// Helper to create a test directory structure
func createTestApp(t *testing.T, structure map[string]string) string {
	t.Helper()
	dir := t.TempDir()
	for path, content := range structure {
		fullPath := filepath.Join(dir, path)
		os.MkdirAll(filepath.Dir(fullPath), 0755)
		os.WriteFile(fullPath, []byte(content), 0644)
	}
	return dir
}

func TestScanRoutes_RootPage(t *testing.T) {
	dir := createTestApp(t, map[string]string{
		"page.go": "package app",
	})
	root, err := codegen.ScanRoutes(dir)
	if err != nil {
		t.Fatal(err)
	}
	if root.PageFile == "" {
		t.Error("expected root page.go to be found")
	}
	if root.Path != "/" {
		t.Errorf("expected path '/', got '%s'", root.Path)
	}
}

func TestScanRoutes_NestedRoute(t *testing.T) {
	dir := createTestApp(t, map[string]string{
		"page.go":       "package app",
		"about/page.go": "package about",
	})
	root, err := codegen.ScanRoutes(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(root.Children) != 1 {
		t.Fatalf("expected 1 child, got %d", len(root.Children))
	}
	child := root.Children[0]
	if child.Path != "/about" {
		t.Errorf("expected path '/about', got '%s'", child.Path)
	}
	if child.PageFile == "" {
		t.Error("expected about/page.go to be found")
	}
}

func TestScanRoutes_DynamicSegment(t *testing.T) {
	dir := createTestApp(t, map[string]string{
		"blog/[slug]/page.go": "package slug",
	})
	root, err := codegen.ScanRoutes(dir)
	if err != nil {
		t.Fatal(err)
	}
	// Navigate to blog/[slug]
	if len(root.Children) != 1 {
		t.Fatalf("expected 1 child (blog), got %d", len(root.Children))
	}
	blog := root.Children[0]
	if len(blog.Children) != 1 {
		t.Fatalf("expected 1 child ([slug]), got %d", len(blog.Children))
	}
	slug := blog.Children[0]
	if slug.ParamName != "slug" {
		t.Errorf("expected param name 'slug', got '%s'", slug.ParamName)
	}
	if slug.Path != "/blog/:slug" {
		t.Errorf("expected path '/blog/:slug', got '%s'", slug.Path)
	}
}

func TestScanRoutes_CatchAll(t *testing.T) {
	dir := createTestApp(t, map[string]string{
		"docs/[...path]/page.go": "package docs",
	})
	root, err := codegen.ScanRoutes(dir)
	if err != nil {
		t.Fatal(err)
	}
	docs := root.Children[0]
	catchAll := docs.Children[0]
	if !catchAll.IsCatchAll {
		t.Error("expected catch-all flag to be true")
	}
	if catchAll.ParamName != "path" {
		t.Errorf("expected param name 'path', got '%s'", catchAll.ParamName)
	}
}

func TestScanRoutes_OptionalCatchAll(t *testing.T) {
	dir := createTestApp(t, map[string]string{
		"api/[[...path]]/page.go": "package api",
	})
	root, err := codegen.ScanRoutes(dir)
	if err != nil {
		t.Fatal(err)
	}
	api := root.Children[0]
	opt := api.Children[0]
	if !opt.IsOptionalCatchAll {
		t.Error("expected optional catch-all flag to be true")
	}
}

func TestScanRoutes_RouteGroup(t *testing.T) {
	dir := createTestApp(t, map[string]string{
		"(marketing)/pricing/page.go": "package pricing",
	})
	root, err := codegen.ScanRoutes(dir)
	if err != nil {
		t.Fatal(err)
	}
	// Group should be transparent -- pricing should be at /pricing not /(marketing)/pricing
	// Find the pricing route
	found := false
	var findPricing func(r *codegen.ScannedRoute)
	findPricing = func(r *codegen.ScannedRoute) {
		if r.Segment == "pricing" {
			found = true
			if r.Path != "/pricing" {
				t.Errorf("expected path '/pricing', got '%s'", r.Path)
			}
		}
		for _, child := range r.Children {
			findPricing(child)
		}
	}
	findPricing(root)
	if !found {
		t.Error("pricing route not found")
	}
}

func TestScanRoutes_Layout(t *testing.T) {
	dir := createTestApp(t, map[string]string{
		"layout.go":      "package app",
		"page.go":        "package app",
		"blog/layout.go": "package blog",
		"blog/page.go":   "package blog",
	})
	root, err := codegen.ScanRoutes(dir)
	if err != nil {
		t.Fatal(err)
	}
	if root.LayoutFile == "" {
		t.Error("expected root layout.go")
	}
	blog := root.Children[0]
	if blog.LayoutFile == "" {
		t.Error("expected blog layout.go")
	}
}

func TestScanRoutes_ErrorAndLoading(t *testing.T) {
	dir := createTestApp(t, map[string]string{
		"page.go":    "package app",
		"error.go":   "package app",
		"loading.go": "package app",
	})
	root, err := codegen.ScanRoutes(dir)
	if err != nil {
		t.Fatal(err)
	}
	if root.ErrorFile == "" {
		t.Error("expected error.go")
	}
	if root.LoadingFile == "" {
		t.Error("expected loading.go")
	}
}

func TestScanRoutes_ParallelSlot(t *testing.T) {
	dir := createTestApp(t, map[string]string{
		"layout.go":        "package app",
		"@sidebar/page.go": "package sidebar",
		"@content/page.go": "package content",
	})
	root, err := codegen.ScanRoutes(dir)
	if err != nil {
		t.Fatal(err)
	}
	if root.ParallelSlots == nil {
		t.Fatal("expected parallel slots map")
	}
	if _, ok := root.ParallelSlots["sidebar"]; !ok {
		t.Error("expected 'sidebar' slot")
	}
	if _, ok := root.ParallelSlots["content"]; !ok {
		t.Error("expected 'content' slot")
	}
}

func TestScanRoutes_InterceptingRoute(t *testing.T) {
	dir := createTestApp(t, map[string]string{
		"feed/page.go":                "package feed",
		"feed/(..)photo/[id]/page.go": "package photo",
	})
	root, err := codegen.ScanRoutes(dir)
	if err != nil {
		t.Fatal(err)
	}
	feed := root.Children[0]
	// Find the intercepting route
	found := false
	for _, child := range feed.Children {
		if child.InterceptPattern != "" {
			found = true
			if child.InterceptPattern != "(..)" {
				t.Errorf("expected '(..)' pattern, got '%s'", child.InterceptPattern)
			}
		}
	}
	if !found {
		t.Error("intercepting route not found")
	}
}

func TestScanRoutes_NoPageMeansNotRoute(t *testing.T) {
	dir := createTestApp(t, map[string]string{
		"layout.go":     "package app", // layout only, no page
		"about/page.go": "package about",
	})
	root, err := codegen.ScanRoutes(dir)
	if err != nil {
		t.Fatal(err)
	}
	// Root should still exist as a container but with no PageFile
	if root.PageFile != "" {
		t.Error("root should not have a page file")
	}
	if len(root.Children) != 1 {
		t.Fatalf("expected 1 child, got %d", len(root.Children))
	}
}
