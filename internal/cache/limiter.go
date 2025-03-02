package cache

import (
	"sync"
	"time"
)

// RateLimiter provides rate limiting functionality
type RateLimiter struct {
	tokens      map[string]int
	maxTokens   int
	refillRate  int
	refillEvery time.Duration
	mu          sync.Mutex
}

// NewRateLimiter creates a new rate limiter
// maxTokens: maximum tokens allowed
// refillRate: how many tokens to refill per interval
// refillEvery: how often to refill tokens
func NewRateLimiter(maxTokens, refillRate int, refillEvery time.Duration) *RateLimiter {
	limiter := &RateLimiter{
		tokens:      make(map[string]int),
		maxTokens:   maxTokens,
		refillRate:  refillRate,
		refillEvery: refillEvery,
	}

	// Start token refiller
	go limiter.refiller()

	return limiter
}

// Allow checks if the operation is allowed for the given key
// Returns true if the operation is allowed, false otherwise
func (r *RateLimiter) Allow(key string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Initialize key if it doesn't exist
	if _, exists := r.tokens[key]; !exists {
		r.tokens[key] = r.maxTokens
	}

	// Check if tokens are available
	if r.tokens[key] > 0 {
		r.tokens[key]--
		return true
	}

	return false
}

// GetRemainingTokens returns the number of remaining tokens for the key
func (r *RateLimiter) GetRemainingTokens(key string) int {
	r.mu.Lock()
	defer r.mu.Unlock()

	if tokens, exists := r.tokens[key]; exists {
		return tokens
	}
	return r.maxTokens
}

// refiller periodically refills tokens
func (r *RateLimiter) refiller() {
	ticker := time.NewTicker(r.refillEvery)
	defer ticker.Stop()

	for {
		<-ticker.C
		r.refillTokens()
	}
}

// refillTokens adds tokens up to the maximum
func (r *RateLimiter) refillTokens() {
	r.mu.Lock()
	defer r.mu.Unlock()

	for key, tokens := range r.tokens {
		newTokens := tokens + r.refillRate
		if newTokens > r.maxTokens {
			newTokens = r.maxTokens
		}
		r.tokens[key] = newTokens
	}
}
