package api

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"open-sonar/internal/cache"
	"open-sonar/internal/utils"
)

var (
	// Global rate limiter for API endpoints
	// 60 requests per min
	apiLimiter = cache.NewRateLimiter(60, 60, time.Minute)

	// More restrictive limiter for LLM providers
	// 20 requests per min
	llmLimiter = cache.NewRateLimiter(20, 20, time.Minute)

	// IP-based limiter for unauthenticated requests
	// 30 requests per min
	ipLimiter = cache.NewRateLimiter(30, 30, time.Minute)

	// Request cache with 5-minute TTL
	responseCache = cache.New()

	// Set of valid API keys (for simple auth)
	validAPIKeys     = make(map[string]bool)
	validAPIKeyMutex sync.RWMutex
)

func init() {
	apiKey := os.Getenv("AUTH_TOKEN")
	if apiKey != "" {
		validAPIKeyMutex.Lock()
		validAPIKeys[apiKey] = true
		validAPIKeyMutex.Unlock()
	}
}

// adds a new API key at runtime
func AddAPIKey(key string) {
	validAPIKeyMutex.Lock()
	validAPIKeys[key] = true
	validAPIKeyMutex.Unlock()
}

// checks if the API key is valid
func IsValidAPIKey(key string) bool {
	validAPIKeyMutex.RLock()
	defer validAPIKeyMutex.RUnlock()

	if len(validAPIKeys) == 0 {
		return true
	}

	return validAPIKeys[key]
}

// limits the number of requests
func RateLimitMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract API key or IP
		apiKey := extractAPIKey(r)
		var limitKey string

		if apiKey != "" {
			limitKey = "key:" + apiKey
		} else {
			// Fall back
			limitKey = "ip:" + getClientIP(r)
		}

		// Use different limiters based on endpoint
		var allowed bool
		if strings.HasPrefix(r.URL.Path, "/chat/completions") {
			allowed = llmLimiter.Allow(limitKey)
		} else {
			allowed = apiLimiter.Allow(limitKey)
		}

		if !allowed {
			utils.Warn(fmt.Sprintf("Rate limit exceeded for %s", limitKey))
			WriteJSONError(w, http.StatusTooManyRequests, "Rate limit exceeded. Try again later.")
			return
		}

		next.ServeHTTP(w, r)
	})
}

// authenticates requests
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip auth for certain endpoints
		if r.URL.Path == "/test" || r.Method == "OPTIONS" {
			next.ServeHTTP(w, r)
			return
		}

		// Extract bearer token
		apiKey := extractAPIKey(r)

		// Check if valid
		if !IsValidAPIKey(apiKey) {
			utils.Warn(fmt.Sprintf("Invalid API key from %s", getClientIP(r)))
			WriteJSONError(w, http.StatusUnauthorized, "Unauthorized: Invalid API key")
			return
		}

		next.ServeHTTP(w, r)
	})
}

// caches responses for GET requests
func CacheMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Only cache GET requests
		if r.Method != "GET" {
			next.ServeHTTP(w, r)
			return
		}

		cacheKey := r.URL.String()

		if cached, found := responseCache.Get(cacheKey); found {
			cachedResp := cached.(string)
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("X-Cache", "HIT")
			w.Write([]byte(cachedResp))
			return
		}

		recorder := &responseRecorder{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
			body:           &strings.Builder{},
		}

		// Call the next handler
		next.ServeHTTP(recorder, r)

		// Only cache successful responses
		if recorder.statusCode == http.StatusOK {
			responseCache.Set(cacheKey, recorder.body.String(), 5*time.Minute)
		}
	})
}

// Helper to extract API key from request
func extractAPIKey(r *http.Request) string {
	// Try from Authorization header
	auth := r.Header.Get("Authorization")
	if auth != "" && strings.HasPrefix(auth, "Bearer ") {
		return strings.TrimPrefix(auth, "Bearer ")
	}

	// Try from query parameter
	return r.URL.Query().Get("api_key")
}

// Helper to get client IP address
func getClientIP(r *http.Request) string {
	// Check for X-Forwarded-For header
	if xForwardedFor := r.Header.Get("X-Forwarded-For"); xForwardedFor != "" {
		ips := strings.Split(xForwardedFor, ",")
		return strings.TrimSpace(ips[0])
	}

	// Fall back to RemoteAddr
	ip := r.RemoteAddr

	// Remove port if present
	if colon := strings.LastIndex(ip, ":"); colon != -1 {
		ip = ip[:colon]
	}

	return ip
}

// custom ResponseWriter to capture the response
type responseRecorder struct {
	http.ResponseWriter
	statusCode int
	body       *strings.Builder
}

func (r *responseRecorder) WriteHeader(statusCode int) {
	r.statusCode = statusCode
	r.ResponseWriter.WriteHeader(statusCode)
}

func (r *responseRecorder) Write(b []byte) (int, error) {
	r.body.Write(b)
	return r.ResponseWriter.Write(b)
}

