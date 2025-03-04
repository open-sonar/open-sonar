package main

/*
#include <stdlib.h>
*/
import "C"
import (
	"log"
	"os"
	"strconv"

	"open-sonar/internal/utils"
	"open-sonar/sonar"
)

// reads the PORT and LOG_LEVEL environment variables, creates the server, and runs it (blocking)
//export StartServerFromEnv
func StartServerFromEnv() C.int {
	// default port 8080
	port := 8080
	if portStr := os.Getenv("PORT"); portStr != "" {
		if p, err := strconv.Atoi(portStr); err == nil {
			port = p
		}
	}

	// default to utils.InfoLevel
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

	// Create server with options from environment.
	server := sonar.NewServer(
		sonar.WithPort(port),
		sonar.WithLogLevel(logLevel),
	)

	// Run the server (blocking).
	if err := server.Run(); err != nil {
		log.Printf("Server error: %v", err)
		return -1
	}
	return 0
}

// creates the server using the given port and logLevel values, then runs the server (blocking).
//export StartServer
func StartServer(port C.int, logLevel C.int) C.int {
	server := sonar.NewServer(
		sonar.WithPort(int(port)),
		sonar.WithLogLevel(utils.LogLevel(int(logLevel))),
	)

	if err := server.Run(); err != nil {
		log.Printf("Server error: %v", err)
		return -1
	}
	return 0
}

// won't be executed, here for shared lib
func main() {}
