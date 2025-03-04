package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"open-sonar/internal/cache"
)

// Helper function to create a test handler that records if it was called
func createTestHandler() (http.HandlerFunc, *bool) {
	called := false
	handler := func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}
	return handler, &called
}

func TestAuthMiddleware(t *testing.T) {
	// Clear existing API keys and add a test key
	validAPIKeyMutex.Lock()
	originalKeys := validAPIKeys
	validAPIKeys = make(map[string]bool)
	validAPIKeys["test-key"] = true
	validAPIKeyMutex.Unlock()

	// Restore original keys when done
	defer func() {
		validAPIKeyMutex.Lock()
		validAPIKeys = originalKeys
		validAPIKeyMutex.Unlock()
	}()

	tests := []struct {
		name       string
		path       string
		authHeader string
		wantStatus int
		wantCalled bool
	}{
		{
			name:       "Valid auth header",
			path:       "/chat",
			authHeader: "Bearer test-key",
			wantStatus: http.StatusOK,
			wantCalled: true,
		},
		{
			name:       "Invalid auth header",
			path:       "/chat",
			authHeader: "Bearer wrong-key",
			wantStatus: http.StatusUnauthorized,
			wantCalled: false,
		},
		{
			name:       "Missing auth header",
			path:       "/chat",
			authHeader: "",
			wantStatus: http.StatusUnauthorized,
			wantCalled: false,
		},
		{
			name:       "Test endpoint should skip auth",
			path:       "/test",
			authHeader: "",
			wantStatus: http.StatusOK,
			wantCalled: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test handler
			handler, called := createTestHandler()

			// Create request
			req := httptest.NewRequest("GET", tt.path, nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}

			// Create response recorder
			recorder := httptest.NewRecorder()

			// Apply middleware
			AuthMiddleware(handler).ServeHTTP(recorder, req)

			// Check status
			if recorder.Code != tt.wantStatus {
				t.Errorf("Expected status %d, got %d", tt.wantStatus, recorder.Code)
			}

			// Check if handler was called
			if *called != tt.wantCalled {
				t.Errorf("Expected handler called: %v, got: %v", tt.wantCalled, *called)
			}

			// parse the JSON error
			if recorder.Code != http.StatusOK {
				var errResp struct {
					Error struct {
						Code    int    `json:"code"`
						Message string `json:"message"`
					} `json:"error"`
				}
				if err := json.Unmarshal(recorder.Body.Bytes(), &errResp); err != nil {
					t.Errorf("Expected JSON error response, got: %s", recorder.Body.String())
				} else {
					if errResp.Error.Code != recorder.Code {
						t.Errorf("Expected error code %d, got %d", recorder.Code, errResp.Error.Code)
					}
					if errResp.Error.Message == "" {
						t.Error("Expected a non-empty error message")
					}
				}
			}
		})
	}
}

func TestRateLimitMiddleware(t *testing.T) {
	// Reset rate limiters for test
	apiLimiter = cache.NewRateLimiter(2, 1, 100*time.Millisecond)
	llmLimiter = cache.NewRateLimiter(1, 1, 100*time.Millisecond)

	tests := []struct {
		name       string
		path       string
		ip         string
		authHeader string
		runCount   int // How many times to run the request
		wantStatus int // Status of the last request
	}{
		{
			name:       "Regular endpoint not rate limited",
			path:       "/chat",
			ip:         "1.1.1.1",
			authHeader: "Bearer test-key",
			runCount:   2,
			wantStatus: http.StatusOK,
		},
		{
			name:       "Regular endpoint rate limited",
			path:       "/chat",
			ip:         "1.1.1.2",
			authHeader: "Bearer test-key-2",
			runCount:   3, // Over the limit of 2
			wantStatus: http.StatusTooManyRequests,
		},
		{
			name:       "LLM endpoint rate limited sooner",
			path:       "/chat/completions",
			ip:         "1.1.1.3",
			authHeader: "Bearer test-key-3",
			runCount:   2, // Over the limit of 1
			wantStatus: http.StatusTooManyRequests,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test handler
			handler, _ := createTestHandler()

			var recorder *httptest.ResponseRecorder

			for i := 0; i < tt.runCount; i++ {
				recorder = httptest.NewRecorder()

				req := httptest.NewRequest("GET", tt.path, nil)
				req.RemoteAddr = tt.ip + ":12345"
				if tt.authHeader != "" {
					req.Header.Set("Authorization", tt.authHeader)
				}

				RateLimitMiddleware(handler).ServeHTTP(recorder, req)
			}

			if recorder.Code != tt.wantStatus {
				t.Errorf("Expected status %d for last request, got %d", tt.wantStatus, recorder.Code)
			}

			// If rate-limited, parse the JSON error
			if recorder.Code == http.StatusTooManyRequests {
				var errResp struct {
					Error struct {
						Code    int    `json:"code"`
						Message string `json:"message"`
					} `json:"error"`
				}
				if err := json.Unmarshal(recorder.Body.Bytes(), &errResp); err != nil {
					t.Errorf("Expected JSON error response, got: %s", recorder.Body.String())
				} else {
					if errResp.Error.Code != http.StatusTooManyRequests {
						t.Errorf("Expected error code 429, got %d", errResp.Error.Code)
					}
					if errResp.Error.Message == "" {
						t.Error("Expected a non-empty error message")
					}
				}
			}
		})
	}
}

func TestCacheMiddleware(t *testing.T) {
	requestCount := 0
	handler := func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Response " + string(rune(requestCount+'0'))))
	}

	// Apply the cache middleware
	cachedHandler := CacheMiddleware(http.HandlerFunc(handler))

	// First request should miss cache
	req1 := httptest.NewRequest("GET", "/test?param=1", nil)
	recorder1 := httptest.NewRecorder()
	cachedHandler.ServeHTTP(recorder1, req1)

	// Second identical request should hit cache
	req2 := httptest.NewRequest("GET", "/test?param=1", nil)
	recorder2 := httptest.NewRecorder()
	cachedHandler.ServeHTTP(recorder2, req2)

	// Different request should miss cache
	req3 := httptest.NewRequest("GET", "/test?param=2", nil)
	recorder3 := httptest.NewRecorder()
	cachedHandler.ServeHTTP(recorder3, req3)

	// Check responses
	if recorder1.Body.String() != "Response 1" {
		t.Errorf("Expected 'Response 1', got '%s'", recorder1.Body.String())
	}
	if recorder2.Body.String() != "Response 1" {
		t.Errorf("Expected cached 'Response 1', got '%s'", recorder2.Body.String())
	}
	if recorder3.Body.String() != "Response 2" {
		t.Errorf("Expected 'Response 2', got '%s'", recorder3.Body.String())
	}

	if recorder2.Header().Get("X-Cache") != "HIT" {
		t.Error("Expected cache HIT for second request")
	}

	// POST requests should not be cached
	requestCount = 0 // Reset counter
	postReq1 := httptest.NewRequest("POST", "/test", strings.NewReader("{}"))
	postRecorder1 := httptest.NewRecorder()
	cachedHandler.ServeHTTP(postRecorder1, postReq1)

	postReq2 := httptest.NewRequest("POST", "/test", strings.NewReader("{}"))
	postRecorder2 := httptest.NewRecorder()
	cachedHandler.ServeHTTP(postRecorder2, postReq2)

	// Both POST requests should execute handler
	if postRecorder1.Body.String() != "Response 1" {
		t.Errorf("Expected 'Response 1', got '%s'", postRecorder1.Body.String())
	}
	if postRecorder2.Body.String() != "Response 2" {
		t.Errorf("Expected 'Response 2', got '%s'", postRecorder2.Body.String())
	}
}

func TestExtractAPIKey(t *testing.T) {
	tests := []struct {
		name        string
		authHeader  string
		queryParams string
		want        string
	}{
		{
			name:       "Bearer token in header",
			authHeader: "Bearer my-token",
			want:       "my-token",
		},
		{
			name:        "API key in query parameter",
			queryParams: "api_key=query-token",
			want:        "query-token",
		},
		{
			name:        "Header takes precedence over query",
			authHeader:  "Bearer header-token",
			queryParams: "api_key=query-token",
			want:        "header-token",
		},
		{
			name: "No token",
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/test", nil)
			if tt.queryParams != "" {
				req = httptest.NewRequest("GET", "/test?"+tt.queryParams, nil)
			}
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}

			got := extractAPIKey(req)
			if got != tt.want {
				t.Errorf("extractAPIKey() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetClientIP(t *testing.T) {
	tests := []struct {
		name           string
		remoteAddr     string
		forwardedForIP string
		want           string
	}{
		{
			name:       "Remote addr with port",
			remoteAddr: "192.168.1.1:12345",
			want:       "192.168.1.1",
		},
		{
			name:           "X-Forwarded-For header",
			remoteAddr:     "10.0.0.1:12345",
			forwardedForIP: "203.0.113.195, 70.41.3.18",
			want:           "203.0.113.195",
		},
		{
			name:       "IPv6 remote addr",
			remoteAddr: "[2001:db8::1]:12345",
			want:       "[2001:db8::1]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/test", nil)
			req.RemoteAddr = tt.remoteAddr
			if tt.forwardedForIP != "" {
				req.Header.Set("X-Forwarded-For", tt.forwardedForIP)
			}

			got := getClientIP(req)
			if got != tt.want {
				t.Errorf("getClientIP() = %v, want %v", got, tt.want)
			}
		})
	}
}
