package api

import (
	"github.com/gorilla/mux"
)

// endpoints
func SetupRoutes() *mux.Router {
	r := mux.NewRouter()

	r.HandleFunc("/test", TestHandler).Methods("GET")
	r.HandleFunc("/chat/completions", ChatCompletionsHandler).Methods("POST")

	return r
}
