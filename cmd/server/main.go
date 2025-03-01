package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"open-sonar/internal/api"
	"open-sonar/internal/utils"

	"github.com/joho/godotenv"
)

func main() {
	// Load .env file
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, using environment variables.")
	}

	// Set log level
	logLevelStr := os.Getenv("LOG_LEVEL")
	if logLevelStr != "" {
		setLogLevelFromString(logLevelStr)
	}

	utils.Info("Starting open-sonar server...")
	router := api.SetupRoutes()

	// Get port from environment or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	utils.Info(fmt.Sprintf("Listening on port %s", port))
	err = http.ListenAndServe(":"+port, router)
	if err != nil {
		utils.Error(fmt.Sprintf("Server failed: %v", err))
	}
}

// setLogLevelFromString sets the log level from a string
func setLogLevelFromString(level string) {
	switch level {
	case "DEBUG":
		utils.SetLogLevel(utils.DebugLevel)
	case "INFO":
		utils.SetLogLevel(utils.InfoLevel)
	case "WARN":
		utils.SetLogLevel(utils.WarnLevel)
	case "ERROR":
		utils.SetLogLevel(utils.ErrorLevel)
	default:
		utils.SetLogLevel(utils.InfoLevel)
	}
}
