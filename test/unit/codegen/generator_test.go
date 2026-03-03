//go:build !js || !wasm

package codegen_test

import (
	"go/parser"
	"go/token"
	"strings"
	"testing"

	"github.com/Nu11ified/golem/internal/codegen"
)

func TestGenerateRoutes_SimpleRoute(t *testing.T) {
	// Single page.go at root → generates route for "/"
	tree := &codegen.ScannedRoute{
		Path:     "/",
		Segment:  "",
		DirPath:  "/app",
		PageFile: "/app/page.go",
	}

	output, err := codegen.GenerateRoutes(tree, "myapp")
	if err != nil {
		t.Fatalf("GenerateRoutes failed: %v", err)
	}

	// Verify build tag
	if !strings.Contains(output, "//go:build js && wasm") {
		t.Error("expected build tag '//go:build js && wasm'")
	}

	// Verify package main
	if !strings.Contains(output, "package main") {
		t.Error("expected 'package main'")
	}

	// Verify router import
	if !strings.Contains(output, `"github.com/Nu11ified/golem/router"`) {
		t.Error("expected router import")
	}

	// Verify registerRoutes function
	if !strings.Contains(output, "func registerRoutes(r *router.Router)") {
		t.Error("expected registerRoutes function")
	}

	// Verify route registration for "/"
	if !strings.Contains(output, `Path:`) && !strings.Contains(output, `"/"`) {
		t.Error("expected route with path /")
	}

	// Verify page function reference (root /app dir gets alias "app")
	if !strings.Contains(output, "app.Page(") {
		t.Error("expected app.Page function call")
	}
}

func TestGenerateRoutes_NestedRoutes(t *testing.T) {
	// about/page.go → generates route for "/about"
	tree := &codegen.ScannedRoute{
		Path:    "/",
		Segment: "",
		DirPath: "/app",
		Children: []*codegen.ScannedRoute{
			{
				Path:     "/about",
				Segment:  "about",
				DirPath:  "/app/about",
				PageFile: "/app/about/page.go",
			},
		},
	}

	output, err := codegen.GenerateRoutes(tree, "myapp")
	if err != nil {
		t.Fatalf("GenerateRoutes failed: %v", err)
	}

	// Verify import for the about page package
	if !strings.Contains(output, `about "myapp/app/about"`) {
		t.Errorf("expected about page import, got:\n%s", output)
	}

	// Verify route registration for "/about"
	if !strings.Contains(output, `"/about"`) {
		t.Error("expected route path /about")
	}

	// Verify about page function reference
	if !strings.Contains(output, "about.Page(") {
		t.Error("expected about.Page function call")
	}
}

func TestGenerateRoutes_DynamicRoute(t *testing.T) {
	// blog/[slug]/page.go → generates route for "/blog/:slug"
	tree := &codegen.ScannedRoute{
		Path:    "/",
		Segment: "",
		DirPath: "/app",
		Children: []*codegen.ScannedRoute{
			{
				Path:    "/blog",
				Segment: "blog",
				DirPath: "/app/blog",
				Children: []*codegen.ScannedRoute{
					{
						Path:      "/blog/:slug",
						Segment:   "[slug]",
						DirPath:   "/app/blog/[slug]",
						PageFile:  "/app/blog/[slug]/page.go",
						ParamName: "slug",
					},
				},
			},
		},
	}

	output, err := codegen.GenerateRoutes(tree, "myapp")
	if err != nil {
		t.Fatalf("GenerateRoutes failed: %v", err)
	}

	// Verify route with dynamic segment
	if !strings.Contains(output, `"/blog/:slug"`) {
		t.Error("expected route path /blog/:slug")
	}

	// Verify import for the slug page package
	if !strings.Contains(output, `"myapp/app/blog/[slug]"`) {
		t.Errorf("expected blog slug import, got:\n%s", output)
	}
}

func TestGenerateRoutes_WithLayout(t *testing.T) {
	// Route with layout → generates layout wrapping
	tree := &codegen.ScannedRoute{
		Path:       "/",
		Segment:    "",
		DirPath:    "/app",
		PageFile:   "/app/page.go",
		LayoutFile: "/app/layout.go",
	}

	output, err := codegen.GenerateRoutes(tree, "myapp")
	if err != nil {
		t.Fatalf("GenerateRoutes failed: %v", err)
	}

	// Verify layout is set on the route
	if !strings.Contains(output, "Layout:") {
		t.Error("expected Layout field in route")
	}

	// Verify layout function reference (root /app dir gets alias "app")
	if !strings.Contains(output, "app.Layout(") || !strings.Contains(output, "Layout: func(") {
		t.Error("expected layout function reference")
	}
}

func TestGenerateRoutes_WithErrorHandler(t *testing.T) {
	// Route with error.go → generates error boundary
	tree := &codegen.ScannedRoute{
		Path:      "/",
		Segment:   "",
		DirPath:   "/app",
		PageFile:  "/app/page.go",
		ErrorFile: "/app/error.go",
	}

	output, err := codegen.GenerateRoutes(tree, "myapp")
	if err != nil {
		t.Fatalf("GenerateRoutes failed: %v", err)
	}

	// Verify error handler is set
	if !strings.Contains(output, "ErrorHandler:") {
		t.Error("expected ErrorHandler field in route")
	}

	// Verify error handler function reference (root /app dir gets alias "app")
	if !strings.Contains(output, "app.Error(") {
		t.Error("expected app.Error function call")
	}
}

func TestGenerateRoutes_CatchAll(t *testing.T) {
	// docs/[...path]/page.go → generates catch-all route
	tree := &codegen.ScannedRoute{
		Path:    "/",
		Segment: "",
		DirPath: "/app",
		Children: []*codegen.ScannedRoute{
			{
				Path:    "/docs",
				Segment: "docs",
				DirPath: "/app/docs",
				Children: []*codegen.ScannedRoute{
					{
						Path:       "/docs/*path",
						Segment:    "[...path]",
						DirPath:    "/app/docs/[...path]",
						PageFile:   "/app/docs/[...path]/page.go",
						ParamName:  "path",
						IsCatchAll: true,
					},
				},
			},
		},
	}

	output, err := codegen.GenerateRoutes(tree, "myapp")
	if err != nil {
		t.Fatalf("GenerateRoutes failed: %v", err)
	}

	// Verify catch-all route path
	if !strings.Contains(output, `"/docs/*path"`) {
		t.Errorf("expected catch-all route path /docs/*path, got:\n%s", output)
	}
}

func TestGenerateRoutes_RouteGroup(t *testing.T) {
	// (marketing)/about/page.go → /about (no group in URL)
	tree := &codegen.ScannedRoute{
		Path:    "/",
		Segment: "",
		DirPath: "/app",
		Children: []*codegen.ScannedRoute{
			{
				Path:      "/",
				Segment:   "(marketing)",
				DirPath:   "/app/(marketing)",
				IsGroup:   true,
				GroupName: "marketing",
				Children: []*codegen.ScannedRoute{
					{
						Path:     "/about",
						Segment:  "about",
						DirPath:  "/app/(marketing)/about",
						PageFile: "/app/(marketing)/about/page.go",
					},
				},
			},
		},
	}

	output, err := codegen.GenerateRoutes(tree, "myapp")
	if err != nil {
		t.Fatalf("GenerateRoutes failed: %v", err)
	}

	// Verify group does NOT appear in the registered route path
	if strings.Contains(output, `"/(marketing)/about"`) {
		t.Error("route group name should not appear in URL path")
	}

	// Verify the route is registered at /about
	if !strings.Contains(output, `"/about"`) {
		t.Error("expected route path /about")
	}

	// Verify import uses the filesystem path (including group dir)
	if !strings.Contains(output, `"myapp/app/(marketing)/about"`) {
		t.Errorf("expected import with group directory, got:\n%s", output)
	}
}

func TestGenerateRoutes_ValidGoSyntax(t *testing.T) {
	// Build a comprehensive tree and verify the output parses as valid Go
	tree := &codegen.ScannedRoute{
		Path:       "/",
		Segment:    "",
		DirPath:    "/app",
		PageFile:   "/app/page.go",
		LayoutFile: "/app/layout.go",
		Children: []*codegen.ScannedRoute{
			{
				Path:     "/about",
				Segment:  "about",
				DirPath:  "/app/about",
				PageFile: "/app/about/page.go",
			},
			{
				Path:    "/blog",
				Segment: "blog",
				DirPath: "/app/blog",
				Children: []*codegen.ScannedRoute{
					{
						Path:      "/blog/:slug",
						Segment:   "[slug]",
						DirPath:   "/app/blog/[slug]",
						PageFile:  "/app/blog/[slug]/page.go",
						ParamName: "slug",
					},
				},
			},
			{
				Path:      "/",
				Segment:   "(marketing)",
				DirPath:   "/app/(marketing)",
				IsGroup:   true,
				GroupName: "marketing",
				Children: []*codegen.ScannedRoute{
					{
						Path:       "/pricing",
						Segment:    "pricing",
						DirPath:    "/app/(marketing)/pricing",
						PageFile:   "/app/(marketing)/pricing/page.go",
						LayoutFile: "/app/(marketing)/pricing/layout.go",
						ErrorFile:  "/app/(marketing)/pricing/error.go",
					},
				},
			},
		},
	}

	output, err := codegen.GenerateRoutes(tree, "myapp")
	if err != nil {
		t.Fatalf("GenerateRoutes failed: %v", err)
	}

	// Parse the output as Go source
	fset := token.NewFileSet()
	_, parseErr := parser.ParseFile(fset, "routes_gen.go", output, parser.AllErrors)
	if parseErr != nil {
		t.Errorf("generated code is not valid Go:\n%v\n\nGenerated code:\n%s", parseErr, output)
	}
}

func TestGenerateRoutes_ParallelSlots(t *testing.T) {
	// Route with parallel slots
	tree := &codegen.ScannedRoute{
		Path:       "/",
		Segment:    "",
		DirPath:    "/app",
		LayoutFile: "/app/layout.go",
		ParallelSlots: map[string]*codegen.ScannedRoute{
			"sidebar": {
				Path:     "/",
				Segment:  "@sidebar",
				DirPath:  "/app/@sidebar",
				PageFile: "/app/@sidebar/page.go",
			},
			"content": {
				Path:     "/",
				Segment:  "@content",
				DirPath:  "/app/@content",
				PageFile: "/app/@content/page.go",
			},
		},
	}

	output, err := codegen.GenerateRoutes(tree, "myapp")
	if err != nil {
		t.Fatalf("GenerateRoutes failed: %v", err)
	}

	// Verify parallel slot routes are registered
	if !strings.Contains(output, "sidebar") {
		t.Error("expected sidebar slot registration")
	}
	if !strings.Contains(output, "content") {
		t.Error("expected content slot registration")
	}

	// Verify valid Go syntax
	fset := token.NewFileSet()
	_, parseErr := parser.ParseFile(fset, "routes_gen.go", output, parser.AllErrors)
	if parseErr != nil {
		t.Errorf("generated code is not valid Go:\n%v\n\nGenerated code:\n%s", parseErr, output)
	}
}

func TestGenerateRoutes_InterceptingRoute(t *testing.T) {
	// Route with intercepting pattern
	tree := &codegen.ScannedRoute{
		Path:    "/",
		Segment: "",
		DirPath: "/app",
		Children: []*codegen.ScannedRoute{
			{
				Path:    "/feed",
				Segment: "feed",
				DirPath: "/app/feed",
				Children: []*codegen.ScannedRoute{
					{
						Path:             "/feed/photo",
						Segment:          "photo",
						DirPath:          "/app/feed/(..)photo",
						PageFile:         "/app/feed/(..)photo/page.go",
						InterceptPattern: "(..)",
						InterceptDepth:   1,
					},
				},
			},
		},
	}

	output, err := codegen.GenerateRoutes(tree, "myapp")
	if err != nil {
		t.Fatalf("GenerateRoutes failed: %v", err)
	}

	// Verify intercepting route is registered
	if !strings.Contains(output, "IsIntercepting") {
		t.Error("expected IsIntercepting field in route")
	}

	// Verify valid Go syntax
	fset := token.NewFileSet()
	_, parseErr := parser.ParseFile(fset, "routes_gen.go", output, parser.AllErrors)
	if parseErr != nil {
		t.Errorf("generated code is not valid Go:\n%v\n\nGenerated code:\n%s", parseErr, output)
	}
}

func TestGenerateRoutes_NoPages(t *testing.T) {
	// Tree with no pages should still generate valid code
	tree := &codegen.ScannedRoute{
		Path:    "/",
		Segment: "",
		DirPath: "/app",
	}

	output, err := codegen.GenerateRoutes(tree, "myapp")
	if err != nil {
		t.Fatalf("GenerateRoutes failed: %v", err)
	}

	// Should still be valid Go
	fset := token.NewFileSet()
	_, parseErr := parser.ParseFile(fset, "routes_gen.go", output, parser.AllErrors)
	if parseErr != nil {
		t.Errorf("generated code is not valid Go:\n%v\n\nGenerated code:\n%s", parseErr, output)
	}
}

func TestGenerateRoutes_NestedLayout(t *testing.T) {
	// Nested layouts: root layout wraps child layout
	tree := &codegen.ScannedRoute{
		Path:       "/",
		Segment:    "",
		DirPath:    "/app",
		LayoutFile: "/app/layout.go",
		Children: []*codegen.ScannedRoute{
			{
				Path:       "/dashboard",
				Segment:    "dashboard",
				DirPath:    "/app/dashboard",
				PageFile:   "/app/dashboard/page.go",
				LayoutFile: "/app/dashboard/layout.go",
			},
		},
	}

	output, err := codegen.GenerateRoutes(tree, "myapp")
	if err != nil {
		t.Fatalf("GenerateRoutes failed: %v", err)
	}

	// Both layouts should be referenced
	if !strings.Contains(output, "Layout:") {
		t.Error("expected Layout in generated code")
	}

	// Verify valid Go
	fset := token.NewFileSet()
	_, parseErr := parser.ParseFile(fset, "routes_gen.go", output, parser.AllErrors)
	if parseErr != nil {
		t.Errorf("generated code is not valid Go:\n%v\n\nGenerated code:\n%s", parseErr, output)
	}
}

func TestGenerateRoutes_GroupWithLayout(t *testing.T) {
	// Route group with its own layout
	tree := &codegen.ScannedRoute{
		Path:    "/",
		Segment: "",
		DirPath: "/app",
		Children: []*codegen.ScannedRoute{
			{
				Path:       "/",
				Segment:    "(auth)",
				DirPath:    "/app/(auth)",
				IsGroup:    true,
				GroupName:  "auth",
				LayoutFile: "/app/(auth)/layout.go",
				Children: []*codegen.ScannedRoute{
					{
						Path:     "/login",
						Segment:  "login",
						DirPath:  "/app/(auth)/login",
						PageFile: "/app/(auth)/login/page.go",
					},
				},
			},
		},
	}

	output, err := codegen.GenerateRoutes(tree, "myapp")
	if err != nil {
		t.Fatalf("GenerateRoutes failed: %v", err)
	}

	// The login route should be at /login, not /(auth)/login
	if !strings.Contains(output, `"/login"`) {
		t.Error("expected route path /login")
	}

	// Verify valid Go
	fset := token.NewFileSet()
	_, parseErr := parser.ParseFile(fset, "routes_gen.go", output, parser.AllErrors)
	if parseErr != nil {
		t.Errorf("generated code is not valid Go:\n%v\n\nGenerated code:\n%s", parseErr, output)
	}
}
