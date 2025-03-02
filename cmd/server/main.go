package main

import (
	"log"
	"os"
	"strconv"

	"open-sonar/internal/utils"
	"open-sonar/sonar"
)

func main() {
	// Parse port from environment
	port := 8080
	if portStr := os.Getenv("PORT"); portStr != "" {
		if p, err := strconv.Atoi(portStr); err == nil {
			port = p
		}
	}

	// Parse log level from environment
	logLevel := utils.InfoLevel
	if logLevelStr := os.Getenv("LOG_LEVEL"); logLevelStr != "" {
		switch logLevelStr {
		case "DEBUG":
			logLevel = utils.DebugLevel
		case "INFO":
			logLevel = utils.InfoLevel
		case "WARN":
			logLevel = utils.WarnLevel
		case "ERROR":
			logLevel = utils.ErrorLevel
		}
	}

	// Create server with options from environment
	server := sonar.NewServer(
		sonar.WithPort(port),
		sonar.WithLogLevel(logLevel), // Now correctly passing utils.LogLevel
	)

	// Run the server (blocking)
	if err := server.Run(); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
