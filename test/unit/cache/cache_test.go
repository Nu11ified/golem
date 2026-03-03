package cache_test

import (
	"testing"
	"time"

	"github.com/Nu11ified/golem/internal/cache"
)

func TestCache_SetAndGet(t *testing.T) {
	c := cache.New()
	c.Set("key1", []byte("value1"), 0) // no TTL

	data, found, stale := c.Get("key1")
	if !found {
		t.Fatal("expected to find key1")
	}
	if stale {
		t.Error("should not be stale with no TTL")
	}
	if string(data) != "value1" {
		t.Errorf("expected 'value1', got '%s'", string(data))
	}
}

func TestCache_GetMissing(t *testing.T) {
	c := cache.New()
	_, found, _ := c.Get("nonexistent")
	if found {
		t.Error("should not find nonexistent key")
	}
}

func TestCache_TTLStale(t *testing.T) {
	c := cache.New()
	c.Set("key1", []byte("value1"), 50*time.Millisecond)

	// Should be fresh immediately
	_, found, stale := c.Get("key1")
	if !found {
		t.Fatal("expected to find key1")
	}
	if stale {
		t.Error("should not be stale yet")
	}

	// Wait for TTL to expire
	time.Sleep(100 * time.Millisecond)

	data, found, stale := c.Get("key1")
	if !found {
		t.Fatal("stale entries should still be found")
	}
	if !stale {
		t.Error("should be stale after TTL")
	}
	if string(data) != "value1" {
		t.Error("stale data should still be returned")
	}
}

func TestCache_Invalidate(t *testing.T) {
	c := cache.New()
	c.Set("key1", []byte("value1"), 0)
	c.Invalidate("key1")

	_, found, _ := c.Get("key1")
	if found {
		t.Error("should not find invalidated key")
	}
}

func TestCache_InvalidateByTag(t *testing.T) {
	c := cache.New()
	c.SetWithTags("post-1", []byte("content1"), 0, []string{"posts", "blog"})
	c.SetWithTags("post-2", []byte("content2"), 0, []string{"posts"})
	c.SetWithTags("page-1", []byte("about"), 0, []string{"pages"})

	c.InvalidateByTag("posts")

	_, found1, _ := c.Get("post-1")
	_, found2, _ := c.Get("post-2")
	_, found3, _ := c.Get("page-1")

	if found1 {
		t.Error("post-1 should be invalidated")
	}
	if found2 {
		t.Error("post-2 should be invalidated")
	}
	if !found3 {
		t.Error("page-1 should NOT be invalidated")
	}
}

func TestCache_InvalidateAll(t *testing.T) {
	c := cache.New()
	c.Set("a", []byte("1"), 0)
	c.Set("b", []byte("2"), 0)
	c.Set("c", []byte("3"), 0)

	c.InvalidateAll()

	if c.Len() != 0 {
		t.Errorf("expected 0 entries, got %d", c.Len())
	}
}

func TestCache_Len(t *testing.T) {
	c := cache.New()
	if c.Len() != 0 {
		t.Error("expected 0 entries")
	}
	c.Set("a", []byte("1"), 0)
	c.Set("b", []byte("2"), 0)
	if c.Len() != 2 {
		t.Errorf("expected 2 entries, got %d", c.Len())
	}
}

func TestCache_Stats(t *testing.T) {
	c := cache.New()
	c.Set("key1", []byte("value1"), 0)

	c.Get("key1")    // hit
	c.Get("key1")    // hit
	c.Get("missing") // miss

	hits, misses := c.Stats()
	if hits != 2 {
		t.Errorf("expected 2 hits, got %d", hits)
	}
	if misses != 1 {
		t.Errorf("expected 1 miss, got %d", misses)
	}
}

func TestCache_Cleanup(t *testing.T) {
	c := cache.New()
	c.Set("fresh", []byte("data"), 1*time.Hour)
	c.Set("expired1", []byte("data"), 10*time.Millisecond)
	c.Set("expired2", []byte("data"), 10*time.Millisecond)

	time.Sleep(50 * time.Millisecond)

	removed := c.Cleanup()
	if removed != 2 {
		t.Errorf("expected 2 removed, got %d", removed)
	}
	if c.Len() != 1 {
		t.Errorf("expected 1 remaining, got %d", c.Len())
	}
}

func TestCache_Overwrite(t *testing.T) {
	c := cache.New()
	c.Set("key", []byte("old"), 0)
	c.Set("key", []byte("new"), 0)

	data, found, _ := c.Get("key")
	if !found {
		t.Fatal("expected to find key")
	}
	if string(data) != "new" {
		t.Errorf("expected 'new', got '%s'", string(data))
	}
}

func TestCache_ConcurrentAccess(t *testing.T) {
	c := cache.New()
	done := make(chan bool)

	// Writer
	go func() {
		for i := 0; i < 1000; i++ {
			c.Set("key", []byte("value"), time.Second)
		}
		done <- true
	}()

	// Reader
	go func() {
		for i := 0; i < 1000; i++ {
			c.Get("key")
		}
		done <- true
	}()

	// Invalidator
	go func() {
		for i := 0; i < 100; i++ {
			c.Invalidate("key")
		}
		done <- true
	}()

	<-done
	<-done
	<-done
	// No panic = success
}

func TestCache_StartCleanup(t *testing.T) {
	c := cache.New()
	c.Set("short", []byte("data"), 50*time.Millisecond)
	c.Set("long", []byte("data"), 10*time.Second)

	stop := c.StartCleanup(30 * time.Millisecond)
	defer stop()

	time.Sleep(200 * time.Millisecond)

	if c.Len() != 1 {
		t.Errorf("expected 1 remaining after cleanup, got %d", c.Len())
	}
}
