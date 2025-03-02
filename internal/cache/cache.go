package cache

import (
	"sync"
	"time"
)

// Item represents a cached item with expiration
type Item struct {
	Value      interface{}
	Expiration int64
}

// Expired returns true if the item has expired
func (item Item) Expired() bool {
	if item.Expiration == 0 {
		return false
	}
	return time.Now().UnixNano() > item.Expiration
}

// Cache is an in-memory key:value store with expiration
type Cache struct {
	items map[string]Item
	mu    sync.RWMutex
}

// New creates a new Cache
func New() *Cache {
	cache := &Cache{
		items: make(map[string]Item),
	}

	// Start the janitor to clean up expired items
	go cache.janitor()

	return cache
}

// Set adds an item to the cache with the given expiration duration
func (c *Cache) Set(key string, value interface{}, duration time.Duration) {
	var expiration int64
	if duration > 0 {
		expiration = time.Now().Add(duration).UnixNano()
	}

	c.mu.Lock()
	c.items[key] = Item{
		Value:      value,
		Expiration: expiration,
	}
	c.mu.Unlock()
}

// Get retrieves an item from the cache
func (c *Cache) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	item, found := c.items[key]
	c.mu.RUnlock()

	// Not found or expired
	if !found || item.Expired() {
		return nil, false
	}

	return item.Value, true
}

// Delete removes an item from the cache
func (c *Cache) Delete(key string) {
	c.mu.Lock()
	delete(c.items, key)
	c.mu.Unlock()
}

// Clear removes all items from the cache
func (c *Cache) Clear() {
	c.mu.Lock()
	c.items = make(map[string]Item)
	c.mu.Unlock()
}

// janitor cleans up expired items every minute
func (c *Cache) janitor() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		<-ticker.C
		c.deleteExpired()
	}
}

// deleteExpired removes all expired items from the cache
func (c *Cache) deleteExpired() {
	now := time.Now().UnixNano()

	c.mu.Lock()
	for k, v := range c.items {
		if v.Expiration > 0 && now > v.Expiration {
			delete(c.items, k)
		}
	}
	c.mu.Unlock()
}
