package api

import (
    "github.com/gorilla/mux"
)

func SetupRoutes() *mux.Router {
    r := mux.NewRouter()

    r.HandleFunc("/test", TestHandler).Methods("GET")

    r.HandleFunc("/chat", ChatHandler).Methods("POST")

    return r
}
