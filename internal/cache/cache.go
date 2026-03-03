package cache

import (
	"sync"
	"sync/atomic"
	"time"
)

// Entry represents a cached item
type Entry struct {
	Data      []byte
	CreatedAt time.Time
	TTL       time.Duration
	Tags      []string
}

// IsStale returns true if the entry has exceeded its TTL
func (e *Entry) IsStale() bool {
	if e.TTL == 0 {
		return false // no TTL means never stale
	}
	return time.Since(e.CreatedAt) > e.TTL
}

// Cache implements an in-memory cache with TTL and tag-based invalidation
type Cache struct {
	entries map[string]*Entry
	tags    map[string]map[string]bool // tag -> set of keys
	mu      sync.RWMutex

	// Stats
	hits   int64
	misses int64
}

// New creates a new cache
func New() *Cache {
	return &Cache{
		entries: make(map[string]*Entry),
		tags:    make(map[string]map[string]bool),
	}
}

// Get retrieves an item from the cache.
// Returns (data, found, stale). Stale entries are still returned
// to support the stale-while-revalidate pattern.
func (c *Cache) Get(key string) ([]byte, bool, bool) {
	c.mu.RLock()
	entry, ok := c.entries[key]
	c.mu.RUnlock()

	if !ok {
		atomic.AddInt64(&c.misses, 1)
		return nil, false, false
	}

	atomic.AddInt64(&c.hits, 1)
	return entry.Data, true, entry.IsStale()
}

// Set stores an item in the cache with the given TTL.
// A TTL of 0 means the entry never expires.
func (c *Cache) Set(key string, data []byte, ttl time.Duration) {
	c.SetWithTags(key, data, ttl, nil)
}

// SetWithTags stores an item with associated tags for tag-based invalidation.
func (c *Cache) SetWithTags(key string, data []byte, ttl time.Duration, tags []string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// If the key already exists, remove its old tag associations
	if old, ok := c.entries[key]; ok {
		c.removeTagAssociations(key, old.Tags)
	}

	c.entries[key] = &Entry{
		Data:      data,
		CreatedAt: time.Now(),
		TTL:       ttl,
		Tags:      tags,
	}

	// Register tag associations
	for _, tag := range tags {
		if c.tags[tag] == nil {
			c.tags[tag] = make(map[string]bool)
		}
		c.tags[tag][key] = true
	}
}

// Invalidate removes a specific key from the cache.
func (c *Cache) Invalidate(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if entry, ok := c.entries[key]; ok {
		c.removeTagAssociations(key, entry.Tags)
		delete(c.entries, key)
	}
}

// InvalidateByTag removes all entries associated with the given tag.
func (c *Cache) InvalidateByTag(tag string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	keys, ok := c.tags[tag]
	if !ok {
		return
	}

	for key := range keys {
		if entry, exists := c.entries[key]; exists {
			// Remove this key from all of its tags (not just the target tag)
			c.removeTagAssociations(key, entry.Tags)
			delete(c.entries, key)
		}
	}

	// The tag set itself is cleaned up by removeTagAssociations,
	// but ensure it is removed in case of empty leftovers.
	delete(c.tags, tag)
}

// InvalidateAll clears the entire cache.
func (c *Cache) InvalidateAll() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.entries = make(map[string]*Entry)
	c.tags = make(map[string]map[string]bool)
}

// Len returns the number of entries in the cache.
func (c *Cache) Len() int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return len(c.entries)
}

// Stats returns cache hit/miss statistics.
func (c *Cache) Stats() (hits, misses int64) {
	return atomic.LoadInt64(&c.hits), atomic.LoadInt64(&c.misses)
}

// Cleanup removes expired entries and returns the number of entries removed.
// This can be called periodically to reclaim memory.
func (c *Cache) Cleanup() int {
	c.mu.Lock()
	defer c.mu.Unlock()

	removed := 0
	for key, entry := range c.entries {
		if entry.IsStale() {
			c.removeTagAssociations(key, entry.Tags)
			delete(c.entries, key)
			removed++
		}
	}
	return removed
}

// StartCleanup starts a background goroutine that periodically cleans expired
// entries. It returns a stop function that should be called to terminate the
// cleanup goroutine.
func (c *Cache) StartCleanup(interval time.Duration) func() {
	ticker := time.NewTicker(interval)
	done := make(chan struct{})

	go func() {
		for {
			select {
			case <-ticker.C:
				c.Cleanup()
			case <-done:
				ticker.Stop()
				return
			}
		}
	}()

	return func() {
		close(done)
	}
}

// removeTagAssociations removes the given key from all specified tag sets.
// Must be called with c.mu held (write lock).
func (c *Cache) removeTagAssociations(key string, tags []string) {
	for _, tag := range tags {
		if keySet, ok := c.tags[tag]; ok {
			delete(keySet, key)
			if len(keySet) == 0 {
				delete(c.tags, tag)
			}
		}
	}
}
