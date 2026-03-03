//go:build !js || !wasm

// Package isr implements Incremental Static Regeneration (ISR) for Golem
// applications. It provides an HTTP handler that serves pages with ISR
// behavior: fresh cache hits are served immediately, stale cache entries
// are served while triggering background revalidation, and cache misses
// cause on-demand SSR rendering with the result cached for future requests.
package isr

import (
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/Nu11ified/golem/dom"
	"github.com/Nu11ified/golem/internal/cache"
	"github.com/Nu11ified/golem/internal/ssr"
	"github.com/Nu11ified/golem/router"
)

// Config configures the ISR handler.
type Config struct {
	DefaultTTL     time.Duration       // default revalidation interval
	DocumentOpts   dom.DocumentOptions // document options for SSR rendering
	OnDemandSecret string              // secret for on-demand revalidation API
}

// Handler serves pages with ISR behavior.
type Handler struct {
	renderer     *ssr.SSRRenderer
	cache        *cache.Cache
	config       Config
	mu           sync.Mutex
	revalidating map[string]bool // tracks in-flight revalidations
}

// NewHandler creates a new ISR handler.
func NewHandler(r *router.Router, c *cache.Cache, config Config) *Handler {
	return &Handler{
		renderer:     ssr.NewSSRRenderer(r, config.DocumentOpts),
		cache:        c,
		config:       config,
		revalidating: make(map[string]bool),
	}
}

// ServeHTTP handles page requests with ISR caching.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	// Check cache
	data, found, stale := h.cache.Get(path)

	if found && !stale {
		// Fresh cache hit - serve immediately
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("X-Cache", "HIT")
		w.Write(data)
		return
	}

	if found && stale {
		// Stale cache hit - serve stale and trigger background revalidation
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("X-Cache", "STALE")
		w.Write(data)
		h.triggerRevalidation(path)
		return
	}

	// Cache miss - render, cache, and serve
	html, err := h.renderer.RenderPage(path)
	if err != nil {
		http.Error(w, "Page not found", http.StatusNotFound)
		return
	}

	ttl := h.config.DefaultTTL
	if ttl == 0 {
		ttl = 60 * time.Second
	}
	h.cache.Set(path, []byte(html), ttl)

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("X-Cache", "MISS")
	w.Write([]byte(html))
}

// Revalidate forces re-rendering of a specific path.
func (h *Handler) Revalidate(path string) error {
	html, err := h.renderer.RenderPage(path)
	if err != nil {
		return err
	}

	ttl := h.config.DefaultTTL
	if ttl == 0 {
		ttl = 60 * time.Second
	}
	h.cache.Set(path, []byte(html), ttl)
	return nil
}

// RevalidateByTag invalidates all pages with the given tag.
func (h *Handler) RevalidateByTag(tag string) {
	h.cache.InvalidateByTag(tag)
}

// triggerRevalidation starts a background revalidation for the path
// if one isn't already in progress.
func (h *Handler) triggerRevalidation(path string) {
	h.mu.Lock()
	if h.revalidating[path] {
		h.mu.Unlock()
		return
	}
	h.revalidating[path] = true
	h.mu.Unlock()

	go func() {
		defer func() {
			h.mu.Lock()
			delete(h.revalidating, path)
			h.mu.Unlock()
		}()

		if err := h.Revalidate(path); err != nil {
			log.Printf("[ISR] Revalidation failed for %s: %v", path, err)
		}
	}()
}

// RevalidationHandler returns an HTTP handler for on-demand revalidation API.
// POST /api/revalidate?path=/some/path&secret=<secret>
func (h *Handler) RevalidationHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		if h.config.OnDemandSecret != "" {
			if r.URL.Query().Get("secret") != h.config.OnDemandSecret {
				http.Error(w, "Invalid secret", http.StatusUnauthorized)
				return
			}
		}

		path := r.URL.Query().Get("path")
		tag := r.URL.Query().Get("tag")

		if path != "" {
			if err := h.Revalidate(path); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"revalidated":true}`))
			return
		}

		if tag != "" {
			h.RevalidateByTag(tag)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"revalidated":true}`))
			return
		}

		http.Error(w, "path or tag parameter required", http.StatusBadRequest)
	})
}
