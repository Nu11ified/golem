//go:build !js || !wasm

package e2e_test

import (
	"testing"
	"time"

	"github.com/Nu11ified/golem/dom"
	"github.com/Nu11ified/golem/internal/cache"
	"github.com/Nu11ified/golem/internal/ssr"
	"github.com/Nu11ified/golem/router"
)

func TestSSRResultCaching(t *testing.T) {
	r := router.NewRouter()
	callCount := 0
	r.AddSimpleRoute("/", func(params map[string]string) *dom.Element {
		callCount++
		return dom.Div(dom.Text("Page"))
	})

	renderer := ssr.NewSSRRenderer(r, dom.DocumentOptions{Title: "Test"})
	c := cache.New()

	// First render - cache miss
	html1, err := renderer.RenderPage("/")
	if err != nil {
		t.Fatal(err)
	}
	c.Set("/", []byte(html1), 10*time.Second)

	if callCount != 1 {
		t.Errorf("expected 1 render call, got %d", callCount)
	}

	// Second request - cache hit
	data, found, stale := c.Get("/")
	if !found {
		t.Fatal("expected cache hit")
	}
	if stale {
		t.Error("expected fresh entry, got stale")
	}
	if string(data) != html1 {
		t.Error("cached data does not match rendered HTML")
	}

	if callCount != 1 {
		t.Errorf("expected no additional render calls, got %d total", callCount)
	}
}

func TestCacheInvalidationByTag(t *testing.T) {
	c := cache.New()

	c.SetWithTags("/products", []byte("products page"), 10*time.Second, []string{"products", "catalog"})
	c.SetWithTags("/products/1", []byte("product 1"), 10*time.Second, []string{"products", "product-1"})
	c.SetWithTags("/products/2", []byte("product 2"), 10*time.Second, []string{"products", "product-2"})
	c.SetWithTags("/about", []byte("about page"), 10*time.Second, []string{"static"})

	if c.Len() != 4 {
		t.Fatalf("expected 4 entries, got %d", c.Len())
	}

	// Invalidate all product-tagged entries
	c.InvalidateByTag("products")

	_, found, _ := c.Get("/products")
	if found {
		t.Error("expected /products to be invalidated")
	}
	_, found, _ = c.Get("/products/1")
	if found {
		t.Error("expected /products/1 to be invalidated")
	}

	data, found, _ := c.Get("/about")
	if !found {
		t.Error("expected /about to still be cached")
	}
	if string(data) != "about page" {
		t.Error("about page data mismatch")
	}
}

func TestCacheStaleWhileRevalidate(t *testing.T) {
	c := cache.New()

	// Set entry with short TTL
	c.Set("/data", []byte("original data"), 50*time.Millisecond)

	// Immediately should be fresh
	data, found, stale := c.Get("/data")
	if !found {
		t.Fatal("expected cache hit")
	}
	if stale {
		t.Error("expected fresh entry immediately after set")
	}
	if string(data) != "original data" {
		t.Error("data mismatch")
	}

	// Wait for TTL to pass
	time.Sleep(60 * time.Millisecond)

	// Should be stale but still returned (stale-while-revalidate pattern)
	data, found, stale = c.Get("/data")
	if !found {
		t.Fatal("expected cache hit (stale)")
	}
	if !stale {
		t.Error("expected stale entry after TTL")
	}
	if string(data) != "original data" {
		t.Error("stale data should still be returned")
	}
}

func TestCacheWithTTLExpiry(t *testing.T) {
	c := cache.New()

	c.Set("/ephemeral", []byte("temp data"), 50*time.Millisecond)

	// Should be available immediately
	data, found, _ := c.Get("/ephemeral")
	if !found {
		t.Fatal("expected cache hit")
	}
	if string(data) != "temp data" {
		t.Error("data mismatch")
	}

	// Wait for expiry and cleanup
	time.Sleep(60 * time.Millisecond)

	cleaned := c.Cleanup()
	if cleaned != 1 {
		t.Errorf("expected 1 cleaned entry, got %d", cleaned)
	}

	if c.Len() != 0 {
		t.Errorf("expected 0 entries after cleanup, got %d", c.Len())
	}
}
