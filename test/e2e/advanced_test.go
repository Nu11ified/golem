//go:build !js || !wasm

package e2e_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/Nu11ified/golem/dom"
	"github.com/Nu11ified/golem/internal/cache"
	"github.com/Nu11ified/golem/internal/isr"
	"github.com/Nu11ified/golem/internal/middleware"
	"github.com/Nu11ified/golem/internal/ssr"
	"github.com/Nu11ified/golem/internal/staticgen"
	"github.com/Nu11ified/golem/router"
)

// ---------------------------------------------------------------------------
// Helper: build a router with various route types for reuse across tests.
// ---------------------------------------------------------------------------

func buildMultiRouteRouter() *router.Router {
	r := router.NewRouter()

	r.AddSimpleRoute("/", func(params map[string]string) *dom.Element {
		return dom.H1(dom.Text("Home"))
	})

	r.AddSimpleRoute("/about", func(params map[string]string) *dom.Element {
		return dom.Div(dom.Text("About Us"))
	})

	r.AddSimpleRoute("/blog/:slug", func(params map[string]string) *dom.Element {
		return dom.NewElement("article",
			dom.H1(dom.Text("Blog: "+params["slug"])),
			dom.P(dom.Text("Content for "+params["slug"])),
		)
	})

	r.NotFound(func() *dom.Element {
		return dom.Div(
			dom.H1(dom.Text("404 - Page Not Found")),
			dom.P(dom.Text("The requested page does not exist.")),
		)
	})

	return r
}

// ===========================================================================
// 1. ISR + SSR Integration Tests (5 tests)
// ===========================================================================

// TestISRRequestLifecycleMissCacheHit verifies the full ISR lifecycle:
// first request is a cache miss (triggers SSR), second request is a cache hit.
func TestISRRequestLifecycleMissCacheHit(t *testing.T) {
	r := router.NewRouter()
	renderCount := 0
	r.AddSimpleRoute("/", func(params map[string]string) *dom.Element {
		renderCount++
		return dom.H1(dom.Text(fmt.Sprintf("Home render #%d", renderCount)))
	})

	c := cache.New()
	handler := isr.NewHandler(r, c, isr.Config{
		DefaultTTL:   10 * time.Second,
		DocumentOpts: dom.DocumentOptions{Title: "ISR App"},
	})

	ts := httptest.NewServer(handler)
	defer ts.Close()

	// First request: cache MISS - should trigger SSR render
	resp1, err := http.Get(ts.URL + "/")
	if err != nil {
		t.Fatal(err)
	}
	defer resp1.Body.Close()

	if resp1.Header.Get("X-Cache") != "MISS" {
		t.Errorf("expected X-Cache=MISS on first request, got %q", resp1.Header.Get("X-Cache"))
	}
	if renderCount != 1 {
		t.Errorf("expected 1 render call on MISS, got %d", renderCount)
	}

	buf := make([]byte, 32*1024)
	n, _ := resp1.Body.Read(buf)
	body1 := string(buf[:n])

	if !strings.Contains(body1, "Home render #1") {
		t.Errorf("expected first render content in response, got: %s", body1[:min(len(body1), 200)])
	}
	if !strings.Contains(body1, "<title>ISR App</title>") {
		t.Error("expected document title in first response")
	}

	// Second request: cache HIT - should NOT re-render
	resp2, err := http.Get(ts.URL + "/")
	if err != nil {
		t.Fatal(err)
	}
	defer resp2.Body.Close()

	if resp2.Header.Get("X-Cache") != "HIT" {
		t.Errorf("expected X-Cache=HIT on second request, got %q", resp2.Header.Get("X-Cache"))
	}
	if renderCount != 1 {
		t.Errorf("expected render count to remain 1 on cache hit, got %d", renderCount)
	}
}

// TestISRWithCustomDocumentOptions verifies that custom document options
// (title, meta tags, stylesheets) are reflected in the ISR-rendered output.
func TestISRWithCustomDocumentOptions(t *testing.T) {
	r := router.NewRouter()
	r.AddSimpleRoute("/styled", func(params map[string]string) *dom.Element {
		return dom.Div(dom.Text("Styled Page"))
	})

	c := cache.New()
	handler := isr.NewHandler(r, c, isr.Config{
		DefaultTTL: 30 * time.Second,
		DocumentOpts: dom.DocumentOptions{
			Title: "Styled ISR Page",
			Lang:  "fr",
			Meta: map[string]string{
				"description": "A page with custom styling",
				"author":      "Golem Team",
			},
			Styles: []string{"/css/main.css", "/css/theme.css"},
		},
	})

	ts := httptest.NewServer(handler)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/styled")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	buf := make([]byte, 32*1024)
	n, _ := resp.Body.Read(buf)
	body := string(buf[:n])

	checks := map[string]string{
		"<title>Styled ISR Page</title>":                "title tag",
		`lang="fr"`:                                     "lang attribute",
		`name="description" content="A page with custom styling"`: "description meta",
		`name="author" content="Golem Team"`:                      "author meta",
		`href="/css/main.css"`:                                    "main stylesheet",
		`href="/css/theme.css"`:                                   "theme stylesheet",
		"Styled Page":                                             "page content",
	}

	for needle, desc := range checks {
		if !strings.Contains(body, needle) {
			t.Errorf("missing %s: expected %q in output", desc, needle)
		}
	}
}

// TestISRStaleWhileRevalidateFlow verifies that stale cached content is served
// immediately while a background revalidation is triggered. After the background
// revalidation completes, subsequent requests serve fresh, updated content.
func TestISRStaleWhileRevalidateFlow(t *testing.T) {
	r := router.NewRouter()
	renderCount := 0
	r.AddSimpleRoute("/data", func(params map[string]string) *dom.Element {
		renderCount++
		return dom.Div(dom.Text(fmt.Sprintf("Data v%d", renderCount)))
	})

	c := cache.New()
	// Use a short TTL so the first entry becomes stale quickly.
	// The ISR handler's Revalidate re-caches with this same TTL, so we need
	// to make the follow-up request before the new entry also goes stale.
	handler := isr.NewHandler(r, c, isr.Config{
		DefaultTTL:   50 * time.Millisecond,
		DocumentOpts: dom.DocumentOptions{Title: "Stale Test"},
	})

	ts := httptest.NewServer(handler)
	defer ts.Close()

	// First request: MISS
	resp1, err := http.Get(ts.URL + "/data")
	if err != nil {
		t.Fatal(err)
	}
	resp1.Body.Close()

	if resp1.Header.Get("X-Cache") != "MISS" {
		t.Errorf("expected MISS, got %q", resp1.Header.Get("X-Cache"))
	}
	if renderCount != 1 {
		t.Errorf("expected renderCount=1 after miss, got %d", renderCount)
	}

	// Wait for the cache entry to become stale
	time.Sleep(70 * time.Millisecond)

	// Second request: STALE - serves stale content, triggers background revalidation
	resp2, err := http.Get(ts.URL + "/data")
	if err != nil {
		t.Fatal(err)
	}
	defer resp2.Body.Close()

	if resp2.Header.Get("X-Cache") != "STALE" {
		t.Errorf("expected STALE, got %q", resp2.Header.Get("X-Cache"))
	}

	buf := make([]byte, 32*1024)
	n, _ := resp2.Body.Read(buf)
	body := string(buf[:n])

	// Should still contain the original v1 content (stale-while-revalidate)
	if !strings.Contains(body, "Data v1") {
		t.Errorf("expected stale content 'Data v1' in response, got: %s", body[:min(len(body), 200)])
	}

	// Wait briefly for the background revalidation goroutine to complete,
	// then poll quickly to catch the fresh entry before it expires again.
	time.Sleep(30 * time.Millisecond)

	// The background revalidation should have completed by now. The new entry
	// has a 50ms TTL from the point it was set. We need to request before it
	// becomes stale.
	resp3, err := http.Get(ts.URL + "/data")
	if err != nil {
		t.Fatal(err)
	}
	defer resp3.Body.Close()

	cacheHeader := resp3.Header.Get("X-Cache")

	buf2 := make([]byte, 32*1024)
	n2, _ := resp3.Body.Read(buf2)
	body3 := string(buf2[:n2])

	// The key verification is that content was updated from v1 to v2
	// via background revalidation. The cache header may be HIT or STALE
	// depending on timing, but the content must reflect the revalidation.
	if !strings.Contains(body3, "Data v2") {
		t.Errorf("expected revalidated content 'Data v2' after stale-while-revalidate, got: %s", body3[:min(len(body3), 200)])
	}

	// Verify that it was served from cache (HIT or STALE, not MISS)
	if cacheHeader == "MISS" {
		t.Error("expected cached response (HIT or STALE) after background revalidation, got MISS")
	}

	// Verify render count: first render on MISS, second render during background revalidation
	if renderCount != 2 {
		t.Errorf("expected 2 total render calls (MISS + revalidation), got %d", renderCount)
	}
}

// TestISROnDemandRevalidationAPI verifies that the on-demand revalidation
// endpoint re-renders and updates the cache for a specific path.
func TestISROnDemandRevalidationAPI(t *testing.T) {
	r := router.NewRouter()
	version := "v1"
	r.AddSimpleRoute("/page", func(params map[string]string) *dom.Element {
		return dom.Div(dom.Text("Page " + version))
	})

	c := cache.New()
	handler := isr.NewHandler(r, c, isr.Config{
		DefaultTTL:     5 * time.Minute,
		DocumentOpts:   dom.DocumentOptions{Title: "Revalidate Test"},
		OnDemandSecret: "test-secret",
	})

	mux := http.NewServeMux()
	mux.Handle("/", handler)
	mux.Handle("/api/revalidate", handler.RevalidationHandler())

	ts := httptest.NewServer(mux)
	defer ts.Close()

	// Initial request - populates cache with v1
	resp1, err := http.Get(ts.URL + "/page")
	if err != nil {
		t.Fatal(err)
	}
	resp1.Body.Close()

	if resp1.Header.Get("X-Cache") != "MISS" {
		t.Errorf("expected MISS, got %q", resp1.Header.Get("X-Cache"))
	}

	// Update the content
	version = "v2"

	// Trigger on-demand revalidation (with correct secret)
	revalReq, _ := http.NewRequest("POST", ts.URL+"/api/revalidate?path=/page&secret=test-secret", nil)
	revalResp, err := http.DefaultClient.Do(revalReq)
	if err != nil {
		t.Fatal(err)
	}
	revalResp.Body.Close()

	if revalResp.StatusCode != http.StatusOK {
		t.Errorf("expected 200 for revalidation, got %d", revalResp.StatusCode)
	}

	// Next request should get the updated content from cache
	resp2, err := http.Get(ts.URL + "/page")
	if err != nil {
		t.Fatal(err)
	}
	defer resp2.Body.Close()

	if resp2.Header.Get("X-Cache") != "HIT" {
		t.Errorf("expected HIT after revalidation, got %q", resp2.Header.Get("X-Cache"))
	}

	buf := make([]byte, 32*1024)
	n, _ := resp2.Body.Read(buf)
	body := string(buf[:n])

	if !strings.Contains(body, "Page v2") {
		t.Errorf("expected revalidated 'Page v2', got: %s", body[:min(len(body), 200)])
	}

	// Verify invalid secret is rejected
	badReq, _ := http.NewRequest("POST", ts.URL+"/api/revalidate?path=/page&secret=wrong", nil)
	badResp, err := http.DefaultClient.Do(badReq)
	if err != nil {
		t.Fatal(err)
	}
	badResp.Body.Close()

	if badResp.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected 401 for bad secret, got %d", badResp.StatusCode)
	}
}

// TestISRNotFoundReturns404 verifies that requesting a route that does not
// exist returns a 404 status code from the ISR handler.
func TestISRNotFoundReturns404(t *testing.T) {
	r := router.NewRouter()
	r.AddSimpleRoute("/", func(params map[string]string) *dom.Element {
		return dom.Div(dom.Text("Home"))
	})
	// Intentionally no not-found handler set and no /missing route

	c := cache.New()
	handler := isr.NewHandler(r, c, isr.Config{
		DefaultTTL:   10 * time.Second,
		DocumentOpts: dom.DocumentOptions{Title: "404 Test"},
	})

	ts := httptest.NewServer(handler)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/missing")
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected 404 for missing route, got %d", resp.StatusCode)
	}

	// Cache should remain empty for the missing page
	if c.Len() != 0 {
		t.Errorf("expected empty cache after 404, got %d entries", c.Len())
	}
}

// ===========================================================================
// 2. Static Generation + ISR Integration Tests (3 tests)
// ===========================================================================

// TestStaticGenerationFollowedByISRServing verifies that statically generated
// pages can be loaded into the ISR cache and served with cache HIT headers.
func TestStaticGenerationFollowedByISRServing(t *testing.T) {
	r := router.NewRouter()
	r.AddSimpleRoute("/", func(params map[string]string) *dom.Element {
		return dom.Div(dom.Text("Static Home"))
	})
	r.AddSimpleRoute("/about", func(params map[string]string) *dom.Element {
		return dom.Div(dom.Text("Static About"))
	})

	docOpts := dom.DocumentOptions{Title: "Static Site"}

	// Step 1: Static generation
	tmpDir := t.TempDir()
	gen := staticgen.NewGenerator(r, docOpts, staticgen.GeneratorConfig{
		OutputDir:      tmpDir,
		Concurrency:    2,
		DefaultDocOpts: docOpts,
	})

	results, err := gen.Generate([]staticgen.PageConfig{
		{Path: "/"},
		{Path: "/about"},
	})
	if err != nil {
		t.Fatalf("static generation failed: %v", err)
	}

	for _, result := range results {
		if result.Error != nil {
			t.Fatalf("page %s failed: %v", result.Path, result.Error)
		}
		if result.HTML == "" {
			t.Fatalf("page %s produced empty HTML", result.Path)
		}
	}

	// Step 2: Load static content into ISR cache
	c := cache.New()
	for _, result := range results {
		c.Set(result.Path, []byte(result.HTML), 5*time.Minute)
	}

	// Step 3: Serve via ISR handler
	handler := isr.NewHandler(r, c, isr.Config{
		DefaultTTL:   5 * time.Minute,
		DocumentOpts: docOpts,
	})
	ts := httptest.NewServer(handler)
	defer ts.Close()

	// Request should be a cache HIT (pre-populated from static gen)
	resp, err := http.Get(ts.URL + "/")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.Header.Get("X-Cache") != "HIT" {
		t.Errorf("expected HIT for pre-cached static page, got %q", resp.Header.Get("X-Cache"))
	}

	buf := make([]byte, 32*1024)
	n, _ := resp.Body.Read(buf)
	body := string(buf[:n])

	if !strings.Contains(body, "Static Home") {
		t.Error("expected static home content in ISR response")
	}
	if !strings.Contains(body, "<title>Static Site</title>") {
		t.Error("expected title from static generation in ISR response")
	}
}

// TestStaticGenerationWithLayoutsProducesValidHTML verifies that static
// generation correctly applies layout functions and produces valid HTML
// that can be cached and served by ISR.
func TestStaticGenerationWithLayoutsProducesValidHTML(t *testing.T) {
	r := router.NewRouter()
	r.AddSimpleRouteWithLayout("/products",
		func(params map[string]string) *dom.Element {
			return dom.NewElement("main",
				dom.H1(dom.Text("Product List")),
				dom.P(dom.Text("Browse our products")),
			)
		},
		func(child *dom.Element) *dom.Element {
			return dom.Div(
				dom.NewElement("header", dom.Text("Shop Header")),
				child,
				dom.NewElement("footer", dom.Text("Shop Footer")),
			)
		},
	)

	docOpts := dom.DocumentOptions{Title: "Product Page"}

	tmpDir := t.TempDir()
	gen := staticgen.NewGenerator(r, docOpts, staticgen.GeneratorConfig{
		OutputDir:      tmpDir,
		Concurrency:    1,
		DefaultDocOpts: docOpts,
	})

	results, err := gen.Generate([]staticgen.PageConfig{
		{Path: "/products"},
	})
	if err != nil {
		t.Fatalf("generation failed: %v", err)
	}

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Error != nil {
		t.Fatalf("generation error: %v", results[0].Error)
	}

	html := results[0].HTML

	// Verify layout wrapping is present
	if !strings.Contains(html, "Shop Header") {
		t.Error("missing layout header in generated HTML")
	}
	if !strings.Contains(html, "Product List") {
		t.Error("missing page content in generated HTML")
	}
	if !strings.Contains(html, "Shop Footer") {
		t.Error("missing layout footer in generated HTML")
	}

	// Verify the document structure is valid
	if !strings.HasPrefix(html, "<!DOCTYPE html>") {
		t.Error("missing DOCTYPE")
	}
	if !strings.Contains(html, "<title>Product Page</title>") {
		t.Error("missing title")
	}

	// Verify correct order: header before content before footer
	headerIdx := strings.Index(html, "Shop Header")
	contentIdx := strings.Index(html, "Product List")
	footerIdx := strings.Index(html, "Shop Footer")
	if headerIdx >= contentIdx || contentIdx >= footerIdx {
		t.Error("layout elements not in expected order (header, content, footer)")
	}

	// Load into ISR cache and verify it can be served
	c := cache.New()
	c.Set("/products", []byte(html), 5*time.Minute)

	isrHandler := isr.NewHandler(r, c, isr.Config{
		DefaultTTL:   5 * time.Minute,
		DocumentOpts: docOpts,
	})
	ts := httptest.NewServer(isrHandler)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/products")
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()

	if resp.Header.Get("X-Cache") != "HIT" {
		t.Errorf("expected HIT for pre-cached layout page, got %q", resp.Header.Get("X-Cache"))
	}
}

// TestPreGeneratedPagesWithDifferentRevalidationIntervals verifies that
// static pages can be generated with different revalidation TTLs and
// cached accordingly.
func TestPreGeneratedPagesWithDifferentRevalidationIntervals(t *testing.T) {
	r := router.NewRouter()
	r.AddSimpleRoute("/static-page", func(params map[string]string) *dom.Element {
		return dom.Div(dom.Text("Static Content"))
	})
	r.AddSimpleRoute("/dynamic-page", func(params map[string]string) *dom.Element {
		return dom.Div(dom.Text("Dynamic Content"))
	})

	docOpts := dom.DocumentOptions{Title: "Multi-TTL"}

	tmpDir := t.TempDir()
	gen := staticgen.NewGenerator(r, docOpts, staticgen.GeneratorConfig{
		OutputDir:      tmpDir,
		Concurrency:    2,
		DefaultDocOpts: docOpts,
	})

	pages := []staticgen.PageConfig{
		{Path: "/static-page", Revalidate: 0},                     // never revalidate
		{Path: "/dynamic-page", Revalidate: 50 * time.Millisecond}, // short revalidation
	}

	results, err := gen.Generate(pages)
	if err != nil {
		t.Fatal(err)
	}

	c := cache.New()
	for i, result := range results {
		if result.Error != nil {
			t.Fatalf("page %s failed: %v", result.Path, result.Error)
		}
		ttl := pages[i].Revalidate
		if ttl == 0 {
			ttl = 24 * time.Hour // "never expire" for static pages
		}
		c.Set(result.Path, []byte(result.HTML), ttl)
	}

	// Both should be fresh immediately
	_, found1, stale1 := c.Get("/static-page")
	_, found2, stale2 := c.Get("/dynamic-page")

	if !found1 || stale1 {
		t.Error("expected /static-page to be fresh")
	}
	if !found2 || stale2 {
		t.Error("expected /dynamic-page to be fresh initially")
	}

	// Wait for the short TTL to expire
	time.Sleep(70 * time.Millisecond)

	_, _, staleStatic := c.Get("/static-page")
	_, foundDynamic, staleDynamic := c.Get("/dynamic-page")

	if staleStatic {
		t.Error("expected /static-page to remain fresh (long TTL)")
	}
	if !foundDynamic || !staleDynamic {
		t.Error("expected /dynamic-page to be stale after short TTL")
	}
}

// ===========================================================================
// 3. Complex Feature Combinations (4 tests)
// ===========================================================================

// TestErrorBoundaryWithSSRRendersErrorFallbackInDocument verifies that when
// a component panics, the SSR renderer uses the error boundary handler and
// wraps the result in a full HTML document.
func TestErrorBoundaryWithSSRRendersErrorFallbackInDocument(t *testing.T) {
	r := router.NewRouter()
	r.AddRouteWithErrorBoundary("/crashing",
		func(params map[string]string) *dom.Element {
			panic("component explosion")
		},
		func(err error) *dom.Element {
			return dom.Div(
				dom.Class("error-boundary"),
				dom.H1(dom.Text("Oops!")),
				dom.P(dom.Text("Error: "+err.Error())),
			)
		},
	)

	renderer := ssr.NewSSRRenderer(r, dom.DocumentOptions{
		Title: "Error Page",
		Meta:  map[string]string{"robots": "noindex"},
	})

	html, err := renderer.RenderPage("/crashing")
	if err != nil {
		t.Fatalf("expected SSR to handle error boundary, got error: %v", err)
	}

	// Should be a complete HTML document
	if !strings.HasPrefix(html, "<!DOCTYPE html>") {
		t.Error("expected full HTML document with DOCTYPE")
	}
	if !strings.Contains(html, "<title>Error Page</title>") {
		t.Error("expected document title even for error boundary pages")
	}
	if !strings.Contains(html, `name="robots"`) {
		t.Error("expected meta tags in error boundary document")
	}

	// Error boundary content should be rendered
	if !strings.Contains(html, "Oops!") {
		t.Error("expected error boundary heading in output")
	}
	if !strings.Contains(html, "component explosion") {
		t.Error("expected error message in boundary output")
	}

	// Should still have the app container and WASM bootstrap
	if !strings.Contains(html, `<div id="app">`) {
		t.Error("expected app container in error boundary document")
	}
	if !strings.Contains(html, "wasm_exec.js") {
		t.Error("expected WASM bootstrap even for error page")
	}
}

// TestParallelSlotsWithSSRProducesSlotWrappers verifies that routes with
// parallel slots produce slot wrapper divs with data-golem-slot attributes
// in the SSR-rendered HTML.
func TestParallelSlotsWithSSRProducesSlotWrappers(t *testing.T) {
	r := router.NewRouter()
	r.AddRouteWithParallelSlots("/dashboard",
		func(params map[string]string) *dom.Element {
			return dom.Div(dom.Class("main-content"), dom.Text("Dashboard Home"))
		},
		map[string]func(params map[string]string) *dom.Element{
			"sidebar": func(params map[string]string) *dom.Element {
				return dom.NewElement("nav", dom.Text("Navigation Menu"))
			},
			"metrics": func(params map[string]string) *dom.Element {
				return dom.Div(dom.Text("Key Metrics"))
			},
		},
	)

	renderer := ssr.NewSSRRenderer(r, dom.DocumentOptions{Title: "Dashboard"})

	html, err := renderer.RenderPage("/dashboard")
	if err != nil {
		t.Fatal(err)
	}

	// Main content should be present
	if !strings.Contains(html, "Dashboard Home") {
		t.Error("missing main dashboard content")
	}

	// Each parallel slot should be wrapped in a data-golem-slot div
	if !strings.Contains(html, `data-golem-slot="metrics"`) {
		t.Error("missing data-golem-slot wrapper for metrics slot")
	}
	if !strings.Contains(html, `data-golem-slot="sidebar"`) {
		t.Error("missing data-golem-slot wrapper for sidebar slot")
	}

	// Slot content should appear
	if !strings.Contains(html, "Navigation Menu") {
		t.Error("missing sidebar slot content")
	}
	if !strings.Contains(html, "Key Metrics") {
		t.Error("missing metrics slot content")
	}

	// Full document structure should be present
	if !strings.Contains(html, "<title>Dashboard</title>") {
		t.Error("missing document title")
	}
	if !strings.Contains(html, `<div id="app">`) {
		t.Error("missing app container")
	}
}

// TestLayoutChainPlusParallelSlotsPlusSSR verifies that a route combining
// a layout, parallel slots, and SSR all work together to produce correct output.
func TestLayoutChainPlusParallelSlotsPlusSSR(t *testing.T) {
	r := router.NewRouter()

	// A route with both a layout and parallel slots
	r.AddRoute(&router.Route{
		Path: "/app",
		Component: func(params map[string]string) *dom.Element {
			return dom.NewElement("main", dom.Text("App Content"))
		},
		Layout: func(child *dom.Element) *dom.Element {
			return dom.Div(
				dom.Class("layout"),
				dom.NewElement("header", dom.Text("App Header")),
				child,
				dom.NewElement("footer", dom.Text("App Footer")),
			)
		},
		ParallelSlots: map[string]*router.Route{
			"sidebar": {
				Component: func(params map[string]string) *dom.Element {
					return dom.NewElement("aside", dom.Text("Sidebar Content"))
				},
			},
		},
	})

	renderer := ssr.NewSSRRenderer(r, dom.DocumentOptions{Title: "Full App"})

	html, err := renderer.RenderPage("/app")
	if err != nil {
		t.Fatal(err)
	}

	// Layout elements should be present
	if !strings.Contains(html, "App Header") {
		t.Error("missing layout header")
	}
	if !strings.Contains(html, "App Content") {
		t.Error("missing main content")
	}
	if !strings.Contains(html, "App Footer") {
		t.Error("missing layout footer")
	}

	// Parallel slot wrapper and content
	if !strings.Contains(html, `data-golem-slot="sidebar"`) {
		t.Error("missing sidebar slot wrapper")
	}
	if !strings.Contains(html, "Sidebar Content") {
		t.Error("missing sidebar content")
	}

	// Document structure
	if !strings.Contains(html, "<!DOCTYPE html>") {
		t.Error("missing DOCTYPE")
	}
	if !strings.Contains(html, "<title>Full App</title>") {
		t.Error("missing title")
	}
}

// TestDynamicRoutesWithParamsRenderedInSSR verifies that dynamic route
// parameters are correctly extracted and passed to the component during SSR.
func TestDynamicRoutesWithParamsRenderedInSSR(t *testing.T) {
	r := router.NewRouter()
	r.AddSimpleRoute("/users/:id", func(params map[string]string) *dom.Element {
		return dom.Div(
			dom.H1(dom.Text("User Profile")),
			dom.P(dom.Text("User ID: "+params["id"])),
		)
	})
	r.AddSimpleRoute("/posts/:year/:month/:slug", func(params map[string]string) *dom.Element {
		return dom.NewElement("article",
			dom.H1(dom.Text(params["slug"])),
			dom.P(dom.Text(fmt.Sprintf("Published: %s/%s", params["month"], params["year"]))),
		)
	})

	renderer := ssr.NewSSRRenderer(r, dom.DocumentOptions{Title: "Dynamic Routes"})

	// Test single parameter
	html1, err := renderer.RenderPage("/users/42")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(html1, "User ID: 42") {
		t.Error("expected user ID 42 in rendered output")
	}

	// Test multiple parameters
	html2, err := renderer.RenderPage("/posts/2026/03/hello-world")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(html2, "hello-world") {
		t.Error("expected slug in rendered output")
	}
	if !strings.Contains(html2, "Published: 03/2026") {
		t.Error("expected formatted date in rendered output")
	}

	// Both should be full HTML documents
	if !strings.HasPrefix(html1, "<!DOCTYPE html>") || !strings.HasPrefix(html2, "<!DOCTYPE html>") {
		t.Error("expected full HTML documents for dynamic routes")
	}
}

// ===========================================================================
// 4. Full Pipeline Tests (3 tests)
// ===========================================================================

// TestCompletePipelineDefineRoutesStaticGenISRServeRevalidate tests the
// complete lifecycle: define routes -> static generate -> ISR serve -> revalidate.
func TestCompletePipelineDefineRoutesStaticGenISRServeRevalidate(t *testing.T) {
	// Step 1: Define routes
	version := "v1"
	r := router.NewRouter()
	r.AddSimpleRoute("/", func(params map[string]string) *dom.Element {
		return dom.H1(dom.Text("Home " + version))
	})
	r.AddSimpleRoute("/contact", func(params map[string]string) *dom.Element {
		return dom.Div(dom.Text("Contact " + version))
	})

	docOpts := dom.DocumentOptions{Title: "Pipeline App"}

	// Step 2: Static generation
	tmpDir := t.TempDir()
	gen := staticgen.NewGenerator(r, docOpts, staticgen.GeneratorConfig{
		OutputDir:      tmpDir,
		Concurrency:    2,
		DefaultDocOpts: docOpts,
	})

	results, err := gen.GenerateToFiles([]staticgen.PageConfig{
		{Path: "/"},
		{Path: "/contact"},
	})
	if err != nil {
		t.Fatal(err)
	}

	// Verify files were written
	for _, result := range results {
		if result.Error != nil {
			t.Fatalf("generation error for %s: %v", result.Path, result.Error)
		}
		outPath := filepath.Join(tmpDir, result.OutFile)
		if _, err := os.Stat(outPath); os.IsNotExist(err) {
			t.Fatalf("expected output file %s to exist", outPath)
		}
	}

	// Step 3: Load into ISR cache and serve
	c := cache.New()
	for _, result := range results {
		c.Set(result.Path, []byte(result.HTML), 5*time.Minute)
	}

	isrHandler := isr.NewHandler(r, c, isr.Config{
		DefaultTTL:     5 * time.Minute,
		DocumentOpts:   docOpts,
		OnDemandSecret: "pipeline-secret",
	})

	mux := http.NewServeMux()
	mux.Handle("/api/revalidate", isrHandler.RevalidationHandler())
	mux.Handle("/", isrHandler)

	ts := httptest.NewServer(mux)
	defer ts.Close()

	// Verify cached content is served
	resp, err := http.Get(ts.URL + "/")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.Header.Get("X-Cache") != "HIT" {
		t.Errorf("expected HIT for pre-generated page, got %q", resp.Header.Get("X-Cache"))
	}

	buf := make([]byte, 32*1024)
	n, _ := resp.Body.Read(buf)
	body := string(buf[:n])

	if !strings.Contains(body, "Home v1") {
		t.Error("expected v1 content from static generation")
	}

	// Step 4: Update content and revalidate
	version = "v2"

	revalReq, _ := http.NewRequest("POST", ts.URL+"/api/revalidate?path=/&secret=pipeline-secret", nil)
	revalResp, err := http.DefaultClient.Do(revalReq)
	if err != nil {
		t.Fatal(err)
	}
	revalResp.Body.Close()

	if revalResp.StatusCode != http.StatusOK {
		t.Errorf("expected 200 from revalidation, got %d", revalResp.StatusCode)
	}

	// Verify updated content
	resp2, err := http.Get(ts.URL + "/")
	if err != nil {
		t.Fatal(err)
	}
	defer resp2.Body.Close()

	buf2 := make([]byte, 32*1024)
	n2, _ := resp2.Body.Read(buf2)
	body2 := string(buf2[:n2])

	if !strings.Contains(body2, "Home v2") {
		t.Errorf("expected revalidated v2 content, got: %s", body2[:min(len(body2), 200)])
	}
}

// TestMultiPageAppWithDifferentConfigurations tests that a multi-page app
// with different route configurations (simple, layout, error boundary,
// parallel slots) all serve correctly through the ISR handler.
func TestMultiPageAppWithDifferentConfigurations(t *testing.T) {
	r := router.NewRouter()

	// Simple route
	r.AddSimpleRoute("/", func(params map[string]string) *dom.Element {
		return dom.H1(dom.Text("Welcome Home"))
	})

	// Route with layout
	r.AddSimpleRouteWithLayout("/docs",
		func(params map[string]string) *dom.Element {
			return dom.NewElement("article", dom.Text("Documentation"))
		},
		func(child *dom.Element) *dom.Element {
			return dom.Div(
				dom.NewElement("nav", dom.Text("Docs Nav")),
				child,
			)
		},
	)

	// Route with error boundary (component works fine)
	r.AddRouteWithErrorBoundary("/safe",
		func(params map[string]string) *dom.Element {
			return dom.Div(dom.Text("Safe Page"))
		},
		func(err error) *dom.Element {
			return dom.Div(dom.Text("Error: " + err.Error()))
		},
	)

	// Route with parallel slots
	r.AddRouteWithParallelSlots("/workspace",
		func(params map[string]string) *dom.Element {
			return dom.Div(dom.Text("Workspace Main"))
		},
		map[string]func(params map[string]string) *dom.Element{
			"tools": func(params map[string]string) *dom.Element {
				return dom.Div(dom.Text("Tool Panel"))
			},
		},
	)

	// Dynamic route
	r.AddSimpleRoute("/items/:id", func(params map[string]string) *dom.Element {
		return dom.Div(dom.Text("Item " + params["id"]))
	})

	c := cache.New()
	isrHandler := isr.NewHandler(r, c, isr.Config{
		DefaultTTL:   5 * time.Minute,
		DocumentOpts: dom.DocumentOptions{Title: "Multi-Page App"},
	})

	ts := httptest.NewServer(isrHandler)
	defer ts.Close()

	tests := []struct {
		path    string
		expect  []string
		desc    string
	}{
		{
			path:   "/",
			expect: []string{"Welcome Home"},
			desc:   "simple home page",
		},
		{
			path:   "/docs",
			expect: []string{"Docs Nav", "Documentation"},
			desc:   "page with layout",
		},
		{
			path:   "/safe",
			expect: []string{"Safe Page"},
			desc:   "page with error boundary (no error)",
		},
		{
			path:   "/workspace",
			expect: []string{"Workspace Main", "Tool Panel", `data-golem-slot="tools"`},
			desc:   "page with parallel slots",
		},
		{
			path:   "/items/123",
			expect: []string{"Item 123"},
			desc:   "dynamic route with params",
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			resp, err := http.Get(ts.URL + tc.path)
			if err != nil {
				t.Fatalf("request to %s failed: %v", tc.path, err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				t.Fatalf("expected 200 for %s, got %d", tc.path, resp.StatusCode)
			}

			buf := make([]byte, 32*1024)
			n, _ := resp.Body.Read(buf)
			body := string(buf[:n])

			for _, expected := range tc.expect {
				if !strings.Contains(body, expected) {
					t.Errorf("expected %q in response for %s", expected, tc.path)
				}
			}

			// All pages should be valid HTML documents
			if !strings.HasPrefix(body, "<!DOCTYPE html>") {
				t.Errorf("expected full HTML document for %s", tc.path)
			}
		})
	}

	// Verify that the second request to each path is a cache HIT
	for _, tc := range tests {
		t.Run(tc.desc+"_cache_hit", func(t *testing.T) {
			resp, err := http.Get(ts.URL + tc.path)
			if err != nil {
				t.Fatal(err)
			}
			resp.Body.Close()

			if resp.Header.Get("X-Cache") != "HIT" {
				t.Errorf("expected HIT on second request to %s, got %q", tc.path, resp.Header.Get("X-Cache"))
			}
		})
	}
}

// TestCacheInvalidationUpdateContentServeFresh verifies that after updating
// content and invalidating cache entries by tag, fresh content is served.
func TestCacheInvalidationUpdateContentServeFresh(t *testing.T) {
	productVersion := "v1"
	r := router.NewRouter()
	r.AddSimpleRoute("/products", func(params map[string]string) *dom.Element {
		return dom.Div(dom.Text("Products " + productVersion))
	})
	r.AddSimpleRoute("/products/featured", func(params map[string]string) *dom.Element {
		return dom.Div(dom.Text("Featured " + productVersion))
	})
	r.AddSimpleRoute("/about", func(params map[string]string) *dom.Element {
		return dom.Div(dom.Text("About page"))
	})

	c := cache.New()
	docOpts := dom.DocumentOptions{Title: "Cache Test"}

	// Pre-render and cache pages with tags
	renderer := ssr.NewSSRRenderer(r, docOpts)

	html1, _ := renderer.RenderPage("/products")
	c.SetWithTags("/products", []byte(html1), 5*time.Minute, []string{"products", "catalog"})

	html2, _ := renderer.RenderPage("/products/featured")
	c.SetWithTags("/products/featured", []byte(html2), 5*time.Minute, []string{"products", "featured"})

	html3, _ := renderer.RenderPage("/about")
	c.SetWithTags("/about", []byte(html3), 5*time.Minute, []string{"static"})

	if c.Len() != 3 {
		t.Fatalf("expected 3 cached entries, got %d", c.Len())
	}

	// Update content
	productVersion = "v2"

	// Invalidate all product-tagged pages
	c.InvalidateByTag("products")

	// Product pages should be gone
	_, found1, _ := c.Get("/products")
	_, found2, _ := c.Get("/products/featured")
	_, found3, _ := c.Get("/about")

	if found1 {
		t.Error("expected /products to be invalidated")
	}
	if found2 {
		t.Error("expected /products/featured to be invalidated")
	}
	if !found3 {
		t.Error("expected /about to still be cached")
	}

	// Now serve via ISR - invalidated pages get re-rendered (MISS), about is HIT
	isrHandler := isr.NewHandler(r, c, isr.Config{
		DefaultTTL:   5 * time.Minute,
		DocumentOpts: docOpts,
	})

	ts := httptest.NewServer(isrHandler)
	defer ts.Close()

	// Products page should be a MISS and re-rendered with v2
	resp1, err := http.Get(ts.URL + "/products")
	if err != nil {
		t.Fatal(err)
	}
	defer resp1.Body.Close()

	if resp1.Header.Get("X-Cache") != "MISS" {
		t.Errorf("expected MISS for invalidated /products, got %q", resp1.Header.Get("X-Cache"))
	}

	buf := make([]byte, 32*1024)
	n, _ := resp1.Body.Read(buf)
	body := string(buf[:n])

	if !strings.Contains(body, "Products v2") {
		t.Errorf("expected fresh 'Products v2' after invalidation, got: %s", body[:min(len(body), 200)])
	}

	// About page should be a HIT (not invalidated)
	respAbout, err := http.Get(ts.URL + "/about")
	if err != nil {
		t.Fatal(err)
	}
	respAbout.Body.Close()

	if respAbout.Header.Get("X-Cache") != "HIT" {
		t.Errorf("expected HIT for /about, got %q", respAbout.Header.Get("X-Cache"))
	}
}

// ===========================================================================
// Bonus: Middleware + ISR Integration
// ===========================================================================

// TestMiddlewareWithISRHeaderInjection verifies that middleware can inject
// headers into responses served by the ISR handler.
func TestMiddlewareWithISRHeaderInjection(t *testing.T) {
	r := router.NewRouter()
	r.AddSimpleRoute("/", func(params map[string]string) *dom.Element {
		return dom.H1(dom.Text("Secure Home"))
	})

	c := cache.New()
	isrHandler := isr.NewHandler(r, c, isr.Config{
		DefaultTTL:   10 * time.Second,
		DocumentOpts: dom.DocumentOptions{Title: "Middleware ISR"},
	})

	// Build a middleware pipeline that adds security headers
	pipeline := middleware.NewPipeline()
	pipeline.Use(middleware.WithHeaders(map[string]string{
		"X-Frame-Options":  "DENY",
		"X-Custom-Header":  "golem",
	}))

	// Wrap the ISR handler with the middleware pipeline
	wrappedHandler := http.HandlerFunc(func(w http.ResponseWriter, httpReq *http.Request) {
		mwReq := &middleware.Request{Request: httpReq}
		mwResp := pipeline.Execute(mwReq, func(req *middleware.Request) *middleware.Response {
			// Use a recorder to capture the ISR handler output
			rec := httptest.NewRecorder()
			isrHandler.ServeHTTP(rec, req.Request)
			return &middleware.Response{
				StatusCode: rec.Code,
				Headers:    extractHeaders(rec),
				Body:       rec.Body.Bytes(),
			}
		})

		// Write middleware-processed response
		for k, v := range mwResp.Headers {
			w.Header().Set(k, v)
		}
		if mwResp.StatusCode > 0 {
			w.WriteHeader(mwResp.StatusCode)
		}
		w.Write(mwResp.Body)
	})

	ts := httptest.NewServer(wrappedHandler)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.Header.Get("X-Frame-Options") != "DENY" {
		t.Errorf("expected X-Frame-Options=DENY, got %q", resp.Header.Get("X-Frame-Options"))
	}
	if resp.Header.Get("X-Custom-Header") != "golem" {
		t.Errorf("expected X-Custom-Header=golem, got %q", resp.Header.Get("X-Custom-Header"))
	}

	buf := make([]byte, 32*1024)
	n, _ := resp.Body.Read(buf)
	body := string(buf[:n])

	if !strings.Contains(body, "Secure Home") {
		t.Error("expected page content through middleware pipeline")
	}
}

// ---------------------------------------------------------------------------
// Helper functions
// ---------------------------------------------------------------------------

// extractHeaders copies response recorder headers into a flat map.
func extractHeaders(rec *httptest.ResponseRecorder) map[string]string {
	headers := make(map[string]string)
	for k, vals := range rec.Header() {
		if len(vals) > 0 {
			headers[k] = vals[0]
		}
	}
	return headers
}

// min returns the smaller of two integers.
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
