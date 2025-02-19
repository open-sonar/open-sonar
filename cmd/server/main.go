package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/joho/godotenv"
	"open-sonar/internal/api"
	"open-sonar/internal/utils"
)

func main() {
	// Load .env file
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found.")
	}

	utils.Info("Starting open-sonar server...")
	router := api.SetupRoutes()
	port := ":8080"
	utils.Info(fmt.Sprintf("Listening on port %s", port))
	err = http.ListenAndServe(port, router)
	if err != nil {
		utils.Error(fmt.Sprintf("Server failed: %v", err))
	}
}
