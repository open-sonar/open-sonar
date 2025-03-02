package cache

import (
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

	// Set item with short expiration
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

	// Wait for the short expiration item to expire
	time.Sleep(30 * time.Millisecond)

	// Verify expired item is gone
	_, found = cache.Get("short")
	if found {
		t.Error("Did not expect to find expired item")
	}

	// Verify other items are still there
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
	// This test is more difficult to verify automatically
	// as the janitor runs asynchronously, but we can set up
	// a scenario and check if it works as expected
	cache := New()

	// Set item with short expiration
	cache.Set("temp", "value", 10*time.Millisecond)

	// Wait for expiration but less than janitor interval
	time.Sleep(20 * time.Millisecond)

	// The item should be expired but still present in the internal map
	// until the janitor cleans it up
	// This is an implementation detail, so we directly check the item
	// is expired but still in the map
	cache.mu.RLock()
	item, found := cache.items["temp"]
	cache.mu.RUnlock()

	if !found {
		// This could fail if the janitor happened to run, so it's not a strict test
		t.Log("Item was removed from the map, possibly by the janitor")
	} else if !item.Expired() {
		t.Error("Expected item to be expired")
	}
}
