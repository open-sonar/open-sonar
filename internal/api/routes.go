package api

import (
	"net/http"

	"github.com/gorilla/mux"
)

// SetupRoutes configures all API endpoints
func SetupRoutes() *mux.Router {
	r := mux.NewRouter()

	// Apply global middleware
	r.Use(RateLimitMiddleware)
	r.Use(AuthMiddleware)
	r.Use(CacheMiddleware)

	// Basic endpoints
	r.HandleFunc("/test", TestHandler).Methods("GET")

	// Chat endpoints
	r.HandleFunc("/chat", ChatHandler).Methods("POST")
	r.HandleFunc("/chat/completions", ChatCompletionsHandler).Methods("POST")

	// Add OPTIONS methods for CORS preflight requests
	r.HandleFunc("/chat", OptionsHandler).Methods("OPTIONS")
	r.HandleFunc("/chat/completions", OptionsHandler).Methods("OPTIONS")

	return r
}

// OptionsHandler handles CORS preflight requests
func OptionsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	w.WriteHeader(http.StatusOK)
}
