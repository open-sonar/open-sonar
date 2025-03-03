package cache

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"sync"
	"time"
)

// represents a cached item with expiration
type Item struct {
	Value      interface{}
	Expiration int64
}

// returns true if the item has expired
func (item Item) Expired() bool {
	if item.Expiration == 0 {
		return false
	}
	return time.Now().UnixNano() > item.Expiration
}

type Cache struct {
	items map[string]Item
	mu    sync.RWMutex

	distributed bool     // whether to replicate changes to peers
	peers       []string // list of peer endpoints (e.g. "http://host:port")
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

// enables distributed caching and sets the list of peers.
func (c *Cache) EnableDistributed(peers []string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.distributed = true
	c.peers = peers
}

// disables distributed caching.
func (c *Cache) DisableDistributed() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.distributed = false
	c.peers = nil
}

// Set adds an item to the cache with the given expiration duration.
// If distributed caching is enabled, this also replicates the set operation to all peers.
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
	distributed := c.distributed
	peers := c.peers
	c.mu.Unlock()

	// Replicate to peers if enabled
	if distributed {
		c.replicateSet(key, value, expiration, peers)
	}
}

// Get retrieves an item from the cache. If not found (or expired) locally
func (c *Cache) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	item, found := c.items[key]
	distributed := c.distributed
	peers := c.peers
	c.mu.RUnlock()

	// Not found or expired locally
	if !found || item.Expired() {
		if distributed {
			val, ok := c.fetchFromPeers(key, peers)
			if ok {
				return val, true
			}
		}
		return nil, false
	}

	return item.Value, true
}

// Delete removes an item from the cache.
func (c *Cache) Delete(key string) {
	c.mu.Lock()
	delete(c.items, key)
	distributed := c.distributed
	peers := c.peers
	c.mu.Unlock()

	// Replicate if enabled
	if distributed {
		c.replicateDelete(key, peers)
	}
}

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

// removes all expired items from the cache
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

// attempts to replicate a Set operation to each peer (best-effort).
func (c *Cache) replicateSet(key string, value interface{}, expiration int64, peers []string) {
	reqBody := struct {
		Key        string      `json:"key"`
		Value      interface{} `json:"value"`
		Expiration int64       `json:"expiration"`
	}{
		Key:        key,
		Value:      value,
		Expiration: expiration,
	}

	data, _ := json.Marshal(reqBody)

	for _, peer := range peers {
		go func(p string) {
			endpoint := p + "/cache/set"
			req, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewReader(data))
			if err != nil {
				log.Printf("replicateSet: error creating request for %s: %v", p, err)
				return
			}
			req.Header.Set("Content-Type", "application/json")
			client := &http.Client{Timeout: 2 * time.Second}
			if _, err = client.Do(req); err != nil {
				log.Printf("replicateSet: error sending to %s: %v", p, err)
			}
		}(peer)
	}
}

// attempts to replicate a Delete operation to each peer (best-effort).
func (c *Cache) replicateDelete(key string, peers []string) {
	reqBody := struct {
		Key string `json:"key"`
	}{
		Key: key,
	}
	data, _ := json.Marshal(reqBody)

	for _, peer := range peers {
		go func(p string) {
			endpoint := p + "/cache/delete"
			req, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewReader(data))
			if err != nil {
				log.Printf("replicateDelete: error creating request for %s: %v", p, err)
				return
			}
			req.Header.Set("Content-Type", "application/json")
			client := &http.Client{Timeout: 2 * time.Second}
			if _, err = client.Do(req); err != nil {
				log.Printf("replicateDelete: error sending to %s: %v", p, err)
			}
		}(peer)
	}
}

// queries each peer (in order) to see if it has the item.
func (c *Cache) fetchFromPeers(key string, peers []string) (interface{}, bool) {
	for _, peer := range peers {
		val, ok := c.fetchFromPeer(key, peer)
		if ok {
			// store in local cache without re-distributing
			c.mu.Lock()
			c.items[key] = val
			c.mu.Unlock()
			return val.Value, true
		}
	}
	return nil, false
}

// attempts to fetch a single item from a specific peer
func (c *Cache) fetchFromPeer(key string, peer string) (Item, bool) {
	endpoint := peer + "/cache/get?key=" + key

	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get(endpoint)
	if err != nil {
		log.Printf("fetchFromPeer: error GET from %s: %v", peer, err)
		return Item{}, false
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return Item{}, false
	}

	body, _ := io.ReadAll(resp.Body)
	var result struct {
		Key        string      `json:"key"`
		Value      interface{} `json:"value"`
		Expiration int64       `json:"expiration"`
		Found      bool        `json:"found"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		log.Printf("fetchFromPeer: invalid JSON from %s: %v", peer, err)
		return Item{}, false
	}

	if !result.Found {
		return Item{}, false
	}

	item := Item{
		Value:      result.Value,
		Expiration: result.Expiration,
	}
	// Even if the item is past expiration on the peer, consider it invalid
	if item.Expired() {
		return Item{}, false
	}
	return item, true
}

// creates and returns a fresh *http.ServeMux with the /cache handler
func (c *Cache) NewMux() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/cache/set", c.handleSet)
	mux.HandleFunc("/cache/delete", c.handleDelete)
	mux.HandleFunc("/cache/get", c.handleGet)
	return mux
}

// registers /cache routes on an existing *http.ServeMux
func (c *Cache) RegisterHandlers(mux *http.ServeMux) {
	mux.HandleFunc("/cache/set", c.handleSet)
	mux.HandleFunc("/cache/delete", c.handleDelete)
	mux.HandleFunc("/cache/get", c.handleGet)
}

// handles POST /cache/set
func (c *Cache) handleSet(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	defer r.Body.Close()

	var req struct {
		Key        string      `json:"key"`
		Value      interface{} `json:"value"`
		Expiration int64       `json:"expiration"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	c.mu.Lock()
	c.items[req.Key] = Item{
		Value:      req.Value,
		Expiration: req.Expiration,
	}
	c.mu.Unlock()

	w.WriteHeader(http.StatusOK)
}

// handleDelete handles POST /cache/delete
func (c *Cache) handleDelete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	defer r.Body.Close()

	var req struct {
		Key string `json:"key"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	c.mu.Lock()
	delete(c.items, req.Key)
	c.mu.Unlock()

	w.WriteHeader(http.StatusOK)
}

// handleGet handles GET /cache/get?key=...
func (c *Cache) handleGet(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	key := r.URL.Query().Get("key")
	if key == "" {
		http.Error(w, "Missing 'key' query param", http.StatusBadRequest)
		return
	}

	c.mu.RLock()
	item, found := c.items[key]
	c.mu.RUnlock()

	type response struct {
		Key        string      `json:"key"`
		Value      interface{} `json:"value"`
		Expiration int64       `json:"expiration"`
		Found      bool        `json:"found"`
	}
	resp := response{
		Key:        key,
		Value:      nil,
		Expiration: 0,
		Found:      false,
	}

	if found && !item.Expired() {
		resp.Value = item.Value
		resp.Expiration = item.Expiration
		resp.Found = true
	}

	data, _ := json.Marshal(resp)
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}
