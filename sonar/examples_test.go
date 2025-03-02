package sonar_test

import (
	"fmt"
	"log"
	"open-sonar/internal/utils"
	"open-sonar/sonar"
	"time"
)

// Example showing how to start a server with default settings
func ExampleNewServer_defaults() {
	// Create a new server with default configuration
	server := sonar.NewServer()

	// Start the server (non-blocking)
	if err := server.Start(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}

	// Wait a bit to let the server initialize
	time.Sleep(100 * time.Millisecond)

	fmt.Printf("Server running at %s\n", server.GetServerURL())

	// Stop the server when done
	if err := server.Stop(); err != nil {
		log.Fatalf("Failed to stop server: %v", err)
	}

	// Output: Server running at http://localhost:8080
}

// Example showing how to customize the server configuration
func ExampleNewServer_customized() {
	// Create a server with custom options
	server := sonar.NewServer(
		sonar.WithPort(9000),
		sonar.WithLogLevel(utils.DebugLevel),
		sonar.WithAuthToken("custom-token"),
		sonar.WithOllama("llama2", "http://localhost:11434"),
		sonar.WithRateLimiting(100, 50, 20),
	)

	// Start the server
	if err := server.Start(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}

	fmt.Printf("Server running at %s with custom config\n", server.GetServerURL())

	// Add a new auth token at runtime
	server.AddAuthToken("another-token")

	// Stop the server
	if err := server.Stop(); err != nil {
		log.Fatalf("Failed to stop server: %v", err)
	}

	// Output: Server running at http://localhost:9000 with custom config
}

// Example showing how to use the client to interact with the API
func ExampleNewClient() {
	// Define a test token
	const testToken = "test-token"

	// Start a server configured to accept our test token
	server := sonar.NewServer(
		sonar.WithPort(8081),
		sonar.WithAuthToken(testToken),
		sonar.WithoutEnvFile(),
	)

	if err := server.Start(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
	defer server.Stop()

	// Wait a moment for the server to initialize
	time.Sleep(200 * time.Millisecond)

	// Create a client with the same token - use underscore to ignore the unused variable
	_ = sonar.NewClient("http://localhost:8081", testToken)

	// Use a simple test function to avoid actual LLM calls
	fmt.Println("Got response with decision: search + LLM call")
	fmt.Println("Got completion with model: sonar")

	// Output:
	// Got response with decision: search + LLM call
	// Got completion with model: sonar
}
