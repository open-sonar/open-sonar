package cache

import (
	"testing"
	"time"
)

func TestRateLimiterBasic(t *testing.T) {
	// Create a limiter with 3 tokens that refills 1 token every 50ms
	limiter := NewRateLimiter(3, 1, 50*time.Millisecond)

	// Should allow 3 initial requests
	if !limiter.Allow("test") {
		t.Error("Expected first request to be allowed")
	}
	if !limiter.Allow("test") {
		t.Error("Expected second request to be allowed")
	}
	if !limiter.Allow("test") {
		t.Error("Expected third request to be allowed")
	}

	// Fourth request should be denied
	if limiter.Allow("test") {
		t.Error("Expected fourth request to be denied")
	}
}

func TestRateLimiterRefill(t *testing.T) {
	// Create a limiter with 1 token that refills 1 token every 50ms
	limiter := NewRateLimiter(1, 1, 50*time.Millisecond)

	// Use the token
	if !limiter.Allow("test") {
		t.Error("Expected first request to be allowed")
	}

	// Next request should be denied
	if limiter.Allow("test") {
		t.Error("Expected second request to be denied")
	}

	// Wait for a refill
	time.Sleep(60 * time.Millisecond)

	// Now should be allowed again
	if !limiter.Allow("test") {
		t.Error("Expected request after refill to be allowed")
	}
}

func TestRateLimiterMultipleKeys(t *testing.T) {
	// Create a limiter with 2 tokens that refills 1 token every 50ms
	limiter := NewRateLimiter(2, 1, 50*time.Millisecond)

	// Different keys should have separate token buckets
	if !limiter.Allow("key1") {
		t.Error("Expected first request for key1 to be allowed")
	}
	if !limiter.Allow("key1") {
		t.Error("Expected second request for key1 to be allowed")
	}
	if limiter.Allow("key1") {
		t.Error("Expected third request for key1 to be denied")
	}

	// key2 should still have all tokens
	if !limiter.Allow("key2") {
		t.Error("Expected first request for key2 to be allowed")
	}
	if !limiter.Allow("key2") {
		t.Error("Expected second request for key2 to be allowed")
	}
	if limiter.Allow("key2") {
		t.Error("Expected third request for key2 to be denied")
	}
}

func TestRateLimiterGetRemainingTokens(t *testing.T) {
	// Create a limiter with 5 tokens
	limiter := NewRateLimiter(5, 1, 50*time.Millisecond)

	// Initially should have max tokens
	if tokens := limiter.GetRemainingTokens("test"); tokens != 5 {
		t.Errorf("Expected 5 initial tokens, got %d", tokens)
	}

	// Use 3 tokens
	limiter.Allow("test")
	limiter.Allow("test")
	limiter.Allow("test")

	// Should have 2 tokens left
	if tokens := limiter.GetRemainingTokens("test"); tokens != 2 {
		t.Errorf("Expected 2 remaining tokens, got %d", tokens)
	}

	// Non-existent key should return max tokens
	if tokens := limiter.GetRemainingTokens("nonexistent"); tokens != 5 {
		t.Errorf("Expected 5 tokens for new key, got %d", tokens)
	}
}
