//go:build !js || !wasm

package isr_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Nu11ified/golem/dom"
	"github.com/Nu11ified/golem/internal/cache"
	"github.com/Nu11ified/golem/internal/isr"
	"github.com/Nu11ified/golem/router"
)

// setupRouter creates a router with some test routes.
func setupRouter() *router.Router {
	r := router.NewRouter()
	r.AddRoute(&router.Route{
		Path: "/",
		Component: func(params map[string]string) *dom.Element {
			return dom.Div(dom.Text("Home Page"))
		},
	})
	r.AddRoute(&router.Route{
		Path: "/about",
		Component: func(params map[string]string) *dom.Element {
			return dom.Div(dom.Text("About Page"))
		},
	})
	r.AddRoute(&router.Route{
		Path: "/contact",
		Component: func(params map[string]string) *dom.Element {
			return dom.Div(dom.Text("Contact Page"))
		},
	})
	return r
}

// setupHandler creates an ISR handler with default config.
func setupHandler() (*isr.Handler, *cache.Cache) {
	r := setupRouter()
	c := cache.New()
	config := isr.Config{
		DefaultTTL: 60 * time.Second,
		DocumentOpts: dom.DocumentOptions{
			Title: "Test App",
		},
	}
	h := isr.NewHandler(r, c, config)
	return h, c
}

// Test 1: Cache miss renders and caches
func TestServeHTTP_CacheMissRendersAndCaches(t *testing.T) {
	h, c := setupHandler()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	h.ServeHTTP(w, req)

	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}

	if !strings.Contains(string(body), "Home Page") {
		t.Errorf("expected body to contain 'Home Page', got:\n%s", string(body))
	}

	// Verify the page was cached
	if c.Len() != 1 {
		t.Errorf("expected cache to have 1 entry, got %d", c.Len())
	}

	data, found, stale := c.Get("/")
	if !found {
		t.Fatal("expected '/' to be cached")
	}
	if stale {
		t.Error("expected cached entry to be fresh")
	}
	if !strings.Contains(string(data), "Home Page") {
		t.Error("expected cached data to contain 'Home Page'")
	}
}

// Test 2: Fresh cache hit serves without re-rendering
func TestServeHTTP_FreshCacheHitServesWithoutReRendering(t *testing.T) {
	h, c := setupHandler()

	// Pre-populate cache with custom content
	c.Set("/", []byte("<html>Cached Content</html>"), 60*time.Second)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	h.ServeHTTP(w, req)

	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}

	// Should serve the cached content, not re-rendered
	if string(body) != "<html>Cached Content</html>" {
		t.Errorf("expected cached content, got:\n%s", string(body))
	}
}

// Test 3: Stale cache hit serves stale content
func TestServeHTTP_StaleCacheHitServesStaleContent(t *testing.T) {
	h, c := setupHandler()

	// Pre-populate cache with a very short TTL so it becomes stale immediately
	c.Set("/", []byte("<html>Stale Content</html>"), 1*time.Nanosecond)

	// Wait for it to become stale
	time.Sleep(2 * time.Millisecond)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	h.ServeHTTP(w, req)

	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}

	// Should serve the stale content
	if string(body) != "<html>Stale Content</html>" {
		t.Errorf("expected stale content to be served, got:\n%s", string(body))
	}
}

// Test 4: Stale cache triggers background revalidation
func TestServeHTTP_StaleCacheTriggersBackgroundRevalidation(t *testing.T) {
	h, c := setupHandler()

	// Pre-populate cache with stale content
	c.Set("/", []byte("<html>Stale Content</html>"), 1*time.Nanosecond)
	time.Sleep(2 * time.Millisecond)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	h.ServeHTTP(w, req)

	// Wait for background revalidation to complete
	time.Sleep(100 * time.Millisecond)

	// The cache should now have fresh content from re-rendering
	data, found, stale := c.Get("/")
	if !found {
		t.Fatal("expected '/' to still be cached after revalidation")
	}
	if stale {
		t.Error("expected cached entry to be fresh after revalidation")
	}

	// The content should now be the re-rendered version containing "Home Page"
	if !strings.Contains(string(data), "Home Page") {
		t.Errorf("expected revalidated cache to contain 'Home Page', got:\n%s", string(data))
	}

	// It should no longer be the stale content
	if strings.Contains(string(data), "Stale Content") {
		t.Error("expected stale content to be replaced after revalidation")
	}
}

// Test 5: Revalidate updates cache
func TestRevalidate_UpdatesCache(t *testing.T) {
	h, c := setupHandler()

	// Cache should start empty
	if c.Len() != 0 {
		t.Fatalf("expected empty cache, got %d entries", c.Len())
	}

	err := h.Revalidate("/about")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, found, _ := c.Get("/about")
	if !found {
		t.Fatal("expected '/about' to be cached after revalidation")
	}
	if !strings.Contains(string(data), "About Page") {
		t.Errorf("expected cached data to contain 'About Page', got:\n%s", string(data))
	}
}

// Test 6: RevalidateByTag invalidates tagged entries
func TestRevalidateByTag_InvalidatesTaggedEntries(t *testing.T) {
	h, c := setupHandler()

	// Set up tagged entries
	c.SetWithTags("/", []byte("<html>Home</html>"), 60*time.Second, []string{"pages"})
	c.SetWithTags("/about", []byte("<html>About</html>"), 60*time.Second, []string{"pages"})
	c.SetWithTags("/contact", []byte("<html>Contact</html>"), 60*time.Second, []string{"other"})

	if c.Len() != 3 {
		t.Fatalf("expected 3 cache entries, got %d", c.Len())
	}

	h.RevalidateByTag("pages")

	// Tagged entries should be invalidated
	if c.Len() != 1 {
		t.Errorf("expected 1 cache entry after tag invalidation, got %d", c.Len())
	}

	_, found, _ := c.Get("/")
	if found {
		t.Error("expected '/' to be invalidated")
	}

	_, found, _ = c.Get("/about")
	if found {
		t.Error("expected '/about' to be invalidated")
	}

	// Non-tagged entry should remain
	_, found, _ = c.Get("/contact")
	if !found {
		t.Error("expected '/contact' to remain cached")
	}
}

// Test 7: X-Cache header values (HIT, MISS, STALE)
func TestServeHTTP_XCacheHeaders(t *testing.T) {
	h, c := setupHandler()

	// Test MISS
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	if got := w.Result().Header.Get("X-Cache"); got != "MISS" {
		t.Errorf("expected X-Cache MISS, got %q", got)
	}

	// Test HIT (entry was cached from the MISS request)
	req = httptest.NewRequest(http.MethodGet, "/", nil)
	w = httptest.NewRecorder()
	h.ServeHTTP(w, req)
	if got := w.Result().Header.Get("X-Cache"); got != "HIT" {
		t.Errorf("expected X-Cache HIT, got %q", got)
	}

	// Test STALE
	c.Set("/", []byte("<html>Stale</html>"), 1*time.Nanosecond)
	time.Sleep(2 * time.Millisecond)

	req = httptest.NewRequest(http.MethodGet, "/", nil)
	w = httptest.NewRecorder()
	h.ServeHTTP(w, req)
	if got := w.Result().Header.Get("X-Cache"); got != "STALE" {
		t.Errorf("expected X-Cache STALE, got %q", got)
	}
}

// Test 8: 404 for non-existent routes
func TestServeHTTP_404ForNonExistentRoutes(t *testing.T) {
	h, _ := setupHandler()

	req := httptest.NewRequest(http.MethodGet, "/nonexistent", nil)
	w := httptest.NewRecorder()

	h.ServeHTTP(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", resp.StatusCode)
	}
}

// Test 9: On-demand revalidation API with valid secret
func TestRevalidationHandler_ValidSecret(t *testing.T) {
	r := setupRouter()
	c := cache.New()
	config := isr.Config{
		DefaultTTL: 60 * time.Second,
		DocumentOpts: dom.DocumentOptions{
			Title: "Test App",
		},
		OnDemandSecret: "my-secret",
	}
	h := isr.NewHandler(r, c, config)

	handler := h.RevalidationHandler()

	req := httptest.NewRequest(http.MethodPost, "/api/revalidate?path=/about&secret=my-secret", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}

	if !strings.Contains(string(body), `"revalidated":true`) {
		t.Errorf("expected revalidated response, got:\n%s", string(body))
	}

	// Verify the page was cached
	data, found, _ := c.Get("/about")
	if !found {
		t.Fatal("expected '/about' to be cached after on-demand revalidation")
	}
	if !strings.Contains(string(data), "About Page") {
		t.Error("expected cached data to contain 'About Page'")
	}
}

// Test 10: On-demand revalidation API with invalid secret
func TestRevalidationHandler_InvalidSecret(t *testing.T) {
	r := setupRouter()
	c := cache.New()
	config := isr.Config{
		DefaultTTL:     60 * time.Second,
		OnDemandSecret: "my-secret",
	}
	h := isr.NewHandler(r, c, config)

	handler := h.RevalidationHandler()

	req := httptest.NewRequest(http.MethodPost, "/api/revalidate?path=/about&secret=wrong-secret", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", resp.StatusCode)
	}
}

// Test 11: On-demand revalidation API with tag parameter
func TestRevalidationHandler_TagParameter(t *testing.T) {
	r := setupRouter()
	c := cache.New()
	config := isr.Config{
		DefaultTTL:     60 * time.Second,
		OnDemandSecret: "my-secret",
	}
	h := isr.NewHandler(r, c, config)

	// Pre-populate cache with tagged entries
	c.SetWithTags("/", []byte("<html>Home</html>"), 60*time.Second, []string{"pages"})
	c.SetWithTags("/about", []byte("<html>About</html>"), 60*time.Second, []string{"pages"})

	handler := h.RevalidationHandler()

	req := httptest.NewRequest(http.MethodPost, "/api/revalidate?tag=pages&secret=my-secret", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}

	if !strings.Contains(string(body), `"revalidated":true`) {
		t.Errorf("expected revalidated response, got:\n%s", string(body))
	}

	// Tagged entries should be invalidated
	if c.Len() != 0 {
		t.Errorf("expected all tagged entries to be invalidated, got %d remaining", c.Len())
	}
}

// Test 12: On-demand revalidation API rejects GET requests
func TestRevalidationHandler_RejectsGET(t *testing.T) {
	h, _ := setupHandler()

	handler := h.RevalidationHandler()

	req := httptest.NewRequest(http.MethodGet, "/api/revalidate?path=/about", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Errorf("expected status 405, got %d", resp.StatusCode)
	}
}

// Test 13: Content-Type header is set correctly
func TestServeHTTP_ContentTypeHeader(t *testing.T) {
	h, _ := setupHandler()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	ct := w.Result().Header.Get("Content-Type")
	if ct != "text/html; charset=utf-8" {
		t.Errorf("expected Content-Type 'text/html; charset=utf-8', got %q", ct)
	}
}

// Test 14: Default TTL is used when config TTL is zero
func TestServeHTTP_DefaultTTLWhenConfigZero(t *testing.T) {
	r := setupRouter()
	c := cache.New()
	config := isr.Config{
		DefaultTTL: 0, // zero TTL should default to 60s
		DocumentOpts: dom.DocumentOptions{
			Title: "Test App",
		},
	}
	h := isr.NewHandler(r, c, config)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Result().StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Result().StatusCode)
	}

	// The page should be cached
	_, found, stale := c.Get("/")
	if !found {
		t.Fatal("expected '/' to be cached")
	}
	if stale {
		t.Error("expected cached entry to be fresh (default 60s TTL)")
	}
}

// Test 15: Revalidation API requires path or tag parameter
func TestRevalidationHandler_RequiresPathOrTag(t *testing.T) {
	r := setupRouter()
	c := cache.New()
	config := isr.Config{
		DefaultTTL: 60 * time.Second,
	}
	h := isr.NewHandler(r, c, config)

	handler := h.RevalidationHandler()

	req := httptest.NewRequest(http.MethodPost, "/api/revalidate", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", resp.StatusCode)
	}
}

// Test 16: Revalidation API for non-existent route returns error
func TestRevalidationHandler_NonExistentRouteReturnsError(t *testing.T) {
	r := setupRouter()
	c := cache.New()
	config := isr.Config{
		DefaultTTL: 60 * time.Second,
	}
	h := isr.NewHandler(r, c, config)

	handler := h.RevalidationHandler()

	req := httptest.NewRequest(http.MethodPost, "/api/revalidate?path=/nonexistent", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", resp.StatusCode)
	}
}

// Test 17: No secret configured allows any request
func TestRevalidationHandler_NoSecretAllowsAnyRequest(t *testing.T) {
	r := setupRouter()
	c := cache.New()
	config := isr.Config{
		DefaultTTL:     60 * time.Second,
		OnDemandSecret: "", // no secret configured
		DocumentOpts: dom.DocumentOptions{
			Title: "Test App",
		},
	}
	h := isr.NewHandler(r, c, config)

	handler := h.RevalidationHandler()

	req := httptest.NewRequest(http.MethodPost, "/api/revalidate?path=/about", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200 (no secret required), got %d", resp.StatusCode)
	}
}

// Test 18: Multiple pages can be cached independently
func TestServeHTTP_MultiplePagesIndependent(t *testing.T) {
	h, c := setupHandler()

	// Request home page
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	body1, _ := io.ReadAll(w.Result().Body)
	if !strings.Contains(string(body1), "Home Page") {
		t.Error("expected home page content")
	}

	// Request about page
	req = httptest.NewRequest(http.MethodGet, "/about", nil)
	w = httptest.NewRecorder()
	h.ServeHTTP(w, req)

	body2, _ := io.ReadAll(w.Result().Body)
	if !strings.Contains(string(body2), "About Page") {
		t.Error("expected about page content")
	}

	// Both should be cached
	if c.Len() != 2 {
		t.Errorf("expected 2 cache entries, got %d", c.Len())
	}

	// Verify each is cached independently
	d1, found1, _ := c.Get("/")
	d2, found2, _ := c.Get("/about")
	if !found1 || !found2 {
		t.Fatal("expected both pages to be cached")
	}
	if !strings.Contains(string(d1), "Home Page") {
		t.Error("expected '/' cache to contain 'Home Page'")
	}
	if !strings.Contains(string(d2), "About Page") {
		t.Error("expected '/about' cache to contain 'About Page'")
	}
}

// Test 19: Duplicate revalidation requests are deduplicated
func TestServeHTTP_DeduplicatesRevalidation(t *testing.T) {
	h, c := setupHandler()

	// Populate with stale content
	c.Set("/", []byte("<html>Stale</html>"), 1*time.Nanosecond)
	time.Sleep(2 * time.Millisecond)

	// Issue multiple requests that should all trigger revalidation
	for i := 0; i < 5; i++ {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		w := httptest.NewRecorder()
		h.ServeHTTP(w, req)

		resp := w.Result()
		body, _ := io.ReadAll(resp.Body)

		// All should get the stale content
		if !strings.Contains(string(body), "Stale") {
			t.Errorf("request %d: expected stale content", i)
		}
		if resp.Header.Get("X-Cache") != "STALE" {
			t.Errorf("request %d: expected X-Cache STALE", i)
		}
	}

	// Wait for background revalidation to complete
	time.Sleep(100 * time.Millisecond)

	// After revalidation, content should be fresh
	data, found, stale := c.Get("/")
	if !found {
		t.Fatal("expected '/' to be cached after revalidation")
	}
	if stale {
		t.Error("expected fresh cache entry after revalidation")
	}
	if !strings.Contains(string(data), "Home Page") {
		t.Errorf("expected revalidated content to contain 'Home Page', got:\n%s", string(data))
	}
}

// Test 20: Rendered output is a valid HTML document
func TestServeHTTP_RendersValidHTMLDocument(t *testing.T) {
	h, _ := setupHandler()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	body, _ := io.ReadAll(w.Result().Body)
	html := string(body)

	checks := []struct {
		label    string
		contains string
	}{
		{"DOCTYPE", "<!DOCTYPE html>"},
		{"html tag", "<html"},
		{"head tag", "<head>"},
		{"body tag", "<body>"},
		{"closing html", "</html>"},
		{"title", "<title>Test App</title>"},
	}

	for _, check := range checks {
		if !strings.Contains(html, check.contains) {
			t.Errorf("expected HTML to contain %s (%q)", check.label, check.contains)
		}
	}
}
