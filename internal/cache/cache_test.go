package cache

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestCacheSetGet(t *testing.T) {
	cache := New()

	// Test setting and getting a value
	cache.Set("key1", "value1", 0) // No expiration
	val, found := cache.Get("key1")
	if !found {
		t.Error("Expected to find key1 in cache, but it wasn't found")
	}
	if val != "value1" {
		t.Errorf("Expected value1, got %v", val)
	}

	// Test getting a non-existent key
	_, found = cache.Get("nonexistent")
	if found {
		t.Error("Did not expect to find nonexistent key")
	}
}

func TestCacheExpiration(t *testing.T) {
	cache := New()

	cache.Set("short", "value", 20*time.Millisecond)

	// Set item with longer expiration
	cache.Set("long", "value", 1*time.Hour)

	// Set item with no expiration
	cache.Set("forever", "value", 0)

	// Verify all items are initially present
	_, found := cache.Get("short")
	if !found {
		t.Error("Expected to find item with short expiration")
	}

	time.Sleep(30 * time.Millisecond)

	_, found = cache.Get("short")
	if found {
		t.Error("Did not expect to find expired item")
	}

	_, found = cache.Get("long")
	if !found {
		t.Error("Expected to find item with long expiration")
	}

	_, found = cache.Get("forever")
	if !found {
		t.Error("Expected to find item with no expiration")
	}
}

func TestCacheDelete(t *testing.T) {
	cache := New()

	// Set some items
	cache.Set("key1", "value1", 0)
	cache.Set("key2", "value2", 0)

	// Delete one item
	cache.Delete("key1")

	// Verify it's gone
	_, found := cache.Get("key1")
	if found {
		t.Error("Did not expect to find deleted item")
	}

	// Verify other item still exists
	_, found = cache.Get("key2")
	if !found {
		t.Error("Expected to find non-deleted item")
	}
}

func TestCacheClear(t *testing.T) {
	cache := New()

	// Set some items
	cache.Set("key1", "value1", 0)
	cache.Set("key2", "value2", 0)

	// Clear the cache
	cache.Clear()

	// Verify all items are gone
	_, found1 := cache.Get("key1")
	_, found2 := cache.Get("key2")
	if found1 || found2 {
		t.Error("Did not expect to find any items after clearing cache")
	}
}

func TestCacheJanitor(t *testing.T) {
	cache := New()

	// Set item with short expiration
	cache.Set("temp", "value", 10*time.Millisecond)

	time.Sleep(20 * time.Millisecond)

	// The item should be expired but still present in the internal map until the janitor cleans it up
	cache.mu.RLock()
	item, found := cache.items["temp"]
	cache.mu.RUnlock()

	if !found {
		t.Log("Item was removed from the map, possibly by the janitor")
	} else if !item.Expired() {
		t.Error("Expected item to be expired")
	}
}

// Distributed Tests

// ensures that when distribution is enabled,
func TestDistributedBasic(t *testing.T) {
	cache1 := New()
	cache2 := New()

	ts1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cache1.RegisterHandlers(http.DefaultServeMux)
		http.DefaultServeMux.ServeHTTP(w, r)
	}))
	defer ts1.Close()

	ts2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cache2.RegisterHandlers(http.DefaultServeMux)
		http.DefaultServeMux.ServeHTTP(w, r)
	}))
	defer ts2.Close()

	cache1.EnableDistributed([]string{ts2.URL})
	cache2.EnableDistributed([]string{ts1.URL})

	// If we Set on cache1, it should replicate to cache2
	cache1.Set("distributedKey", "distributedValue", 0)

	// Small sleep to let replication HTTP calls complete
	time.Sleep(50 * time.Millisecond)

	val2, found2 := cache2.Get("distributedKey")
	if !found2 {
		t.Error("Expected to find 'distributedKey' on cache2 after replication")
	}
	if val2 != "distributedValue" {
		t.Errorf("Expected 'distributedValue' on cache2, got %v", val2)
	}
}

// ensures that if a key is absent locally,
// but present on a peer, calling Get will retrieve it and store it locally.
func TestDistributedGetFallback(t *testing.T) {
    cache1 := New()
    cache2 := New()

    // Create a dedicated mux for cache1
    mux1 := http.NewServeMux()
    cache1.RegisterHandlers(mux1)
    ts1 := httptest.NewServer(mux1)
    defer ts1.Close()

    // Create a dedicated mux for cache2
    mux2 := http.NewServeMux()
    cache2.RegisterHandlers(mux2)
    ts2 := httptest.NewServer(mux2)
    defer ts2.Close()

    cache1.EnableDistributed([]string{ts2.URL})
    cache2.EnableDistributed([]string{ts1.URL})

    // Put a key in cache2
    cache2.Set("fallbackKey", "fallbackValue", 0)

    // Let cache1 attempt to Get it (not found locally).
    val, found := cache1.Get("fallbackKey")
    if !found {
        t.Error("Expected to retrieve 'fallbackKey' from cache2 via fallback")
    }
    if val != "fallbackValue" {
        t.Errorf("Expected 'fallbackValue', got %v", val)
    }

    // It should now be stored in cache1
    val1, found1 := cache1.Get("fallbackKey")
    if !found1 || val1 != "fallbackValue" {
        t.Error("Expected 'fallbackKey' to be cached locally after fallback retrieval")
    }

}
