package api

import (
	"github.com/gorilla/mux"
)

// SetupRoutes configures all API endpoints
func SetupRoutes() *mux.Router {
	r := mux.NewRouter()

	// Basic endpoints
	r.HandleFunc("/test", TestHandler).Methods("GET")

	// Chat endpoints
	r.HandleFunc("/chat", ChatHandler).Methods("POST")
	r.HandleFunc("/chat/completions", ChatCompletionsHandler).Methods("POST")

	return r
}
