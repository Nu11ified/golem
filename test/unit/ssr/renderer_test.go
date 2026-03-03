//go:build !js || !wasm

package ssr_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/Nu11ified/golem/dom"
	"github.com/Nu11ified/golem/internal/ssr"
	"github.com/Nu11ified/golem/router"
)

// Helper: creates a simple page component that renders a div with text
func simplePage(text string) func(params map[string]string) *dom.Element {
	return func(params map[string]string) *dom.Element {
		return dom.Div(dom.Text(text))
	}
}

// Helper: creates a page that uses a route param
func paramPage() func(params map[string]string) *dom.Element {
	return func(params map[string]string) *dom.Element {
		id := params["id"]
		return dom.Div(dom.Text("Item: "+id))
	}
}

// Helper: creates a layout that wraps content in a header/footer structure
func testLayout(content *dom.Element) *dom.Element {
	return dom.Div(
		dom.Class("layout"),
		dom.Div(dom.Class("header"), dom.H1(dom.Text("My App"))),
		dom.Div(dom.Class("content"), content),
		dom.Div(dom.Class("footer"), dom.P(dom.Text("Footer"))),
	)
}

func TestRenderPage_BasicPage(t *testing.T) {
	r := router.NewRouter()
	r.AddRoute(&router.Route{
		Path:      "/",
		Component: simplePage("Hello World"),
	})

	renderer := ssr.NewSSRRenderer(r, dom.DocumentOptions{
		Title: "Test App",
	})

	html, err := renderer.RenderPage("/")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should contain the page content
	if !strings.Contains(html, "Hello World") {
		t.Errorf("expected html to contain 'Hello World', got:\n%s", html)
	}

	// Should be a full HTML document
	if !strings.Contains(html, "<!DOCTYPE html>") {
		t.Error("expected html to start with DOCTYPE")
	}

	if !strings.Contains(html, "<html") {
		t.Error("expected html to contain <html> tag")
	}

	if !strings.Contains(html, "</html>") {
		t.Error("expected html to end with </html>")
	}

	// Should contain the title
	if !strings.Contains(html, "<title>Test App</title>") {
		t.Error("expected html to contain the title")
	}
}

func TestRenderPage_WithRouteParams(t *testing.T) {
	r := router.NewRouter()
	r.AddRoute(&router.Route{
		Path:      "/items/:id",
		Component: paramPage(),
	})

	renderer := ssr.NewSSRRenderer(r, dom.DocumentOptions{
		Title: "Items",
	})

	html, err := renderer.RenderPage("/items/42")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(html, "Item: 42") {
		t.Errorf("expected html to contain 'Item: 42', got:\n%s", html)
	}
}

func TestRenderPage_WithLayout(t *testing.T) {
	r := router.NewRouter()
	r.AddRoute(&router.Route{
		Path:      "/about",
		Component: simplePage("About Page"),
		Layout:    testLayout,
	})

	renderer := ssr.NewSSRRenderer(r, dom.DocumentOptions{
		Title: "About",
	})

	html, err := renderer.RenderPage("/about")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should contain layout structure
	if !strings.Contains(html, `class="layout"`) {
		t.Error("expected html to contain layout class")
	}
	if !strings.Contains(html, `class="header"`) {
		t.Error("expected html to contain header class")
	}
	if !strings.Contains(html, `class="footer"`) {
		t.Error("expected html to contain footer class")
	}
	if !strings.Contains(html, "My App") {
		t.Error("expected html to contain layout header text")
	}

	// Should also contain page content
	if !strings.Contains(html, "About Page") {
		t.Error("expected html to contain page content")
	}
}

func TestRenderPage_WithErrorBoundary_NoError(t *testing.T) {
	r := router.NewRouter()
	r.AddRoute(&router.Route{
		Path:      "/safe",
		Component: simplePage("Safe Page"),
		ErrorHandler: func(err error) *dom.Element {
			return dom.Div(dom.Text("Error: " + err.Error()))
		},
	})

	renderer := ssr.NewSSRRenderer(r, dom.DocumentOptions{})

	html, err := renderer.RenderPage("/safe")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(html, "Safe Page") {
		t.Error("expected html to contain normal page content")
	}
	if strings.Contains(html, "Error:") {
		t.Error("expected html NOT to contain error content")
	}
}

func TestRenderPage_WithErrorBoundary_WithPanic(t *testing.T) {
	r := router.NewRouter()
	r.AddRoute(&router.Route{
		Path: "/broken",
		Component: func(params map[string]string) *dom.Element {
			panic("something went wrong")
		},
		ErrorHandler: func(err error) *dom.Element {
			return dom.Div(dom.Class("error"), dom.Text("Error: "+err.Error()))
		},
	})

	renderer := ssr.NewSSRRenderer(r, dom.DocumentOptions{})

	html, err := renderer.RenderPage("/broken")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(html, "Error: something went wrong") {
		t.Errorf("expected html to contain error message, got:\n%s", html)
	}
	if !strings.Contains(html, `class="error"`) {
		t.Error("expected html to contain error class")
	}
}

func TestRenderPage_WithErrorBoundary_NoBoundary(t *testing.T) {
	r := router.NewRouter()
	r.AddRoute(&router.Route{
		Path: "/panic",
		Component: func(params map[string]string) *dom.Element {
			panic("unhandled panic")
		},
	})

	renderer := ssr.NewSSRRenderer(r, dom.DocumentOptions{})

	_, err := renderer.RenderPage("/panic")
	if err == nil {
		t.Fatal("expected an error for unhandled panic")
	}
	if !strings.Contains(err.Error(), "unhandled panic") {
		t.Errorf("expected error to contain panic message, got: %v", err)
	}
}

func TestRenderPage_WithParallelSlots(t *testing.T) {
	r := router.NewRouter()
	r.AddRoute(&router.Route{
		Path:      "/dashboard",
		Component: simplePage("Dashboard Main"),
		ParallelSlots: map[string]func(params map[string]string) *dom.Element{
			"sidebar": func(params map[string]string) *dom.Element {
				return dom.Div(dom.Class("sidebar"), dom.Text("Sidebar Content"))
			},
			"footer": func(params map[string]string) *dom.Element {
				return dom.Div(dom.Class("slot-footer"), dom.Text("Footer Slot"))
			},
		},
	})

	renderer := ssr.NewSSRRenderer(r, dom.DocumentOptions{})

	html, err := renderer.RenderPage("/dashboard")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Main content
	if !strings.Contains(html, "Dashboard Main") {
		t.Error("expected html to contain main content")
	}

	// Parallel slots
	if !strings.Contains(html, "Sidebar Content") {
		t.Error("expected html to contain sidebar slot content")
	}
	if !strings.Contains(html, "Footer Slot") {
		t.Error("expected html to contain footer slot content")
	}

	// Slot containers should have data-slot attribute
	if !strings.Contains(html, `data-golem-slot="sidebar"`) {
		t.Error("expected html to contain data-golem-slot for sidebar")
	}
	if !strings.Contains(html, `data-golem-slot="footer"`) {
		t.Error("expected html to contain data-golem-slot for footer")
	}
}

func TestRenderPage_404NotFound(t *testing.T) {
	r := router.NewRouter()
	r.AddRoute(&router.Route{
		Path:      "/",
		Component: simplePage("Home"),
	})

	renderer := ssr.NewSSRRenderer(r, dom.DocumentOptions{})

	_, err := renderer.RenderPage("/nonexistent")
	if err == nil {
		t.Fatal("expected error for non-existent route")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("expected 'not found' error, got: %v", err)
	}
}

func TestRenderPage_404WithCustomHandler(t *testing.T) {
	r := router.NewRouter()
	r.AddRoute(&router.Route{
		Path:      "/",
		Component: simplePage("Home"),
	})
	r.NotFound(func() *dom.Element {
		return dom.Div(dom.Class("not-found"), dom.H1(dom.Text("404 - Page Not Found")))
	})

	renderer := ssr.NewSSRRenderer(r, dom.DocumentOptions{})

	html, err := renderer.RenderPage("/nonexistent")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(html, "404 - Page Not Found") {
		t.Errorf("expected html to contain 404 message, got:\n%s", html)
	}
	if !strings.Contains(html, `class="not-found"`) {
		t.Error("expected html to contain not-found class")
	}
}

func TestRenderPage_ValidHTMLStructure(t *testing.T) {
	r := router.NewRouter()
	r.AddRoute(&router.Route{
		Path:      "/",
		Component: simplePage("Test"),
	})

	renderer := ssr.NewSSRRenderer(r, dom.DocumentOptions{
		Title: "Valid HTML Test",
		Lang:  "en",
		Meta: []dom.MetaTag{
			{Name: "description", Content: "A test page"},
		},
		Styles: []string{"/styles/main.css"},
	})

	html, err := renderer.RenderPage("/")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check overall structure
	checks := []struct {
		desc     string
		expected string
	}{
		{"DOCTYPE", "<!DOCTYPE html>"},
		{"html lang", `<html lang="en">`},
		{"head open", "<head>"},
		{"charset", `<meta charset="UTF-8">`},
		{"viewport", `<meta name="viewport"`},
		{"title", "<title>Valid HTML Test</title>"},
		{"meta description", `<meta name="description" content="A test page">`},
		{"stylesheet", `<link rel="stylesheet" href="/styles/main.css">`},
		{"head close", "</head>"},
		{"body open", "<body>"},
		{"body close", "</body>"},
		{"html close", "</html>"},
	}

	for _, check := range checks {
		if !strings.Contains(html, check.expected) {
			t.Errorf("missing %s: expected html to contain %q", check.desc, check.expected)
		}
	}
}

func TestRenderPage_WASMBootstrapScript(t *testing.T) {
	r := router.NewRouter()
	r.AddRoute(&router.Route{
		Path:      "/",
		Component: simplePage("Test"),
	})

	renderer := ssr.NewSSRRenderer(r, dom.DocumentOptions{})

	html, err := renderer.RenderPage("/")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should contain WASM bootstrap script
	if !strings.Contains(html, "wasm_exec.js") {
		t.Error("expected html to contain wasm_exec.js script reference")
	}
	if !strings.Contains(html, "WebAssembly.instantiateStreaming") {
		t.Error("expected html to contain WASM instantiation code")
	}
}

func TestRenderPageWithOptions(t *testing.T) {
	r := router.NewRouter()
	r.AddRoute(&router.Route{
		Path:      "/custom",
		Component: simplePage("Custom Page"),
	})

	renderer := ssr.NewSSRRenderer(r, dom.DocumentOptions{
		Title: "Default Title",
	})

	customOpts := dom.DocumentOptions{
		Title: "Custom Title",
		Lang:  "fr",
	}

	html, err := renderer.RenderPageWithOptions("/custom", customOpts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should use custom options, not default
	if !strings.Contains(html, "<title>Custom Title</title>") {
		t.Error("expected html to contain custom title")
	}
	if !strings.Contains(html, `lang="fr"`) {
		t.Error("expected html to contain custom lang")
	}
	if strings.Contains(html, "Default Title") {
		t.Error("expected html NOT to contain default title")
	}
}

func TestRenderPage_LayoutAndErrorBoundaryCombined(t *testing.T) {
	r := router.NewRouter()
	r.AddRoute(&router.Route{
		Path: "/fail-in-layout",
		Component: func(params map[string]string) *dom.Element {
			panic(fmt.Errorf("render failed"))
		},
		Layout: testLayout,
		ErrorHandler: func(err error) *dom.Element {
			return dom.Div(dom.Class("error-fallback"), dom.Text("Oops: "+err.Error()))
		},
	})

	renderer := ssr.NewSSRRenderer(r, dom.DocumentOptions{})

	html, err := renderer.RenderPage("/fail-in-layout")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Error handler should catch the panic
	if !strings.Contains(html, "Oops: render failed") {
		t.Errorf("expected html to contain error fallback, got:\n%s", html)
	}
}

func TestRenderPage_AppContainerDiv(t *testing.T) {
	r := router.NewRouter()
	r.AddRoute(&router.Route{
		Path:      "/",
		Component: simplePage("App Content"),
	})

	renderer := ssr.NewSSRRenderer(r, dom.DocumentOptions{})

	html, err := renderer.RenderPage("/")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Content should be wrapped in an app container div for hydration
	if !strings.Contains(html, `id="app"`) {
		t.Error("expected html to contain app container div with id='app'")
	}
}

func TestRenderPage_NilComponent(t *testing.T) {
	r := router.NewRouter()
	r.AddRoute(&router.Route{
		Path:      "/empty",
		Component: nil,
	})

	renderer := ssr.NewSSRRenderer(r, dom.DocumentOptions{})

	_, err := renderer.RenderPage("/empty")
	if err == nil {
		t.Fatal("expected error for route with nil component")
	}
}

func TestRenderPage_MultipleRoutes(t *testing.T) {
	r := router.NewRouter()
	r.AddRoute(&router.Route{
		Path:      "/",
		Component: simplePage("Home Page"),
	})
	r.AddRoute(&router.Route{
		Path:      "/about",
		Component: simplePage("About Page"),
	})
	r.AddRoute(&router.Route{
		Path:      "/contact",
		Component: simplePage("Contact Page"),
	})

	renderer := ssr.NewSSRRenderer(r, dom.DocumentOptions{})

	tests := []struct {
		path     string
		expected string
	}{
		{"/", "Home Page"},
		{"/about", "About Page"},
		{"/contact", "Contact Page"},
	}

	for _, tc := range tests {
		html, err := renderer.RenderPage(tc.path)
		if err != nil {
			t.Fatalf("unexpected error for %s: %v", tc.path, err)
		}
		if !strings.Contains(html, tc.expected) {
			t.Errorf("for path %s, expected html to contain %q", tc.path, tc.expected)
		}
	}
}

func TestRenderPage_HydrationIDs(t *testing.T) {
	r := router.NewRouter()
	r.AddRoute(&router.Route{
		Path:      "/",
		Component: simplePage("Hydrate Me"),
	})

	renderer := ssr.NewSSRRenderer(r, dom.DocumentOptions{})

	html, err := renderer.RenderPage("/")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should contain hydration ID attributes
	if !strings.Contains(html, "data-golem-id") {
		t.Error("expected html to contain data-golem-id attributes for hydration")
	}
}
