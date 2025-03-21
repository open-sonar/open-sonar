package sonar

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"open-sonar/internal/api"
	"open-sonar/internal/llm"
	"open-sonar/internal/utils"

	"github.com/joho/godotenv"
)

// Server represents an Open Sonar server instance
type Server struct {
	Config     *Config
	httpServer *http.Server
	running    bool
	stopCh     chan struct{}
}

// NewServer creates a new Open Sonar server with the given options
func NewServer(options ...Option) *Server {
	// Create default config
	config := DefaultConfig()

	// Apply provided options
	for _, option := range options {
		option(config)
	}

	return &Server{
		Config: config,
		stopCh: make(chan struct{}),
	}
}

// Start starts the server in a non-blocking way
func (s *Server) Start() error {
	if s.running {
		return fmt.Errorf("server already running")
	}

	// Apply environment variables if specified
	if s.Config.LoadEnvFile && s.Config.EnvFilePath != "" {
		err := godotenv.Load(s.Config.EnvFilePath)
		if err != nil {
			utils.Warn(fmt.Sprintf("Failed to load env file: %v", err))
		}
	}

	// Set log level
	utils.SetLogLevel(s.Config.LogLevel)

	// Set environment variables from config
	setEnvironmentVariables(s.Config)

	// Add the auth token directly to the API key registry
	// This ensures the token is immediately available without waiting for env var processing
	if s.Config.AuthToken != "" {
		api.AddAPIKey(s.Config.AuthToken)
	}

	// Set up router
	router := api.SetupRoutes()

	// Create HTTP server
	s.httpServer = &http.Server{
		Addr:    fmt.Sprintf(":%d", s.Config.Port),
		Handler: router,
	}

	// Start server in a goroutine
	go func() {
		utils.Info(fmt.Sprintf("Starting Open Sonar server on port %d", s.Config.Port))
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			utils.Error(fmt.Sprintf("Server failed: %v", err))
		}
	}()

	s.running = true

	return nil
}

// Run starts the server and blocks until it's stopped
func (s *Server) Run() error {
	if err := s.Start(); err != nil {
		return err
	}

	// Handle graceful shutdown on signals
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	// Wait for either a signal or stopCh
	select {
	case sig := <-sigCh:
		utils.Info(fmt.Sprintf("Received signal %v, shutting down...", sig))
	case <-s.stopCh:
		utils.Info("Shutdown requested")
	}

	return s.Stop()
}

// Stop gracefully stops the server
func (s *Server) Stop() error {
	if !s.running {
		return nil
	}

	utils.Info("Stopping Open Sonar server...")

	// Create a context with timeout for graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Shutdown the server
	if err := s.httpServer.Shutdown(ctx); err != nil {
		utils.Error(fmt.Sprintf("Server shutdown failed: %v", err))
		return err
	}

	close(s.stopCh)
	s.running = false
	utils.Info("Server stopped")
	return nil
}

// AddAuthToken adds a new authentication token at runtime
func (s *Server) AddAuthToken(token string) {
	api.AddAPIKey(token)
}

// IsRunning returns whether the server is currently running
func (s *Server) IsRunning() bool {
	return s.running
}

// SetLLMProvider changes the LLM provider at runtime
func (s *Server) SetLLMProvider(providerName string) error {
	_, err := llm.NewLLMProvider(providerName)
	if err != nil {
		return err
	}
	os.Setenv("DEFAULT_LLM_PROVIDER", providerName)
	return nil
}

// GetServerURL returns the full URL where the server is running
func (s *Server) GetServerURL() string {
	protocol := "http"
	if s.Config.TLS {
		protocol = "https"
	}
	return fmt.Sprintf("%s://localhost:%d", protocol, s.Config.Port)
}

// OllamaStatus represents the status of an Ollama server
type OllamaStatus struct {
	Available   bool     `json:"available"`
	Version     string   `json:"version"`
	Models      []string `json:"models"`
	ServerTime  string   `json:"server_time"`
	Initialized bool     `json:"initialized"`
}

// CheckOllama verifies if an Ollama server is available at the given host
func (s *Server) CheckOllama() (OllamaStatus, error) {
	host := s.Config.OllamaHost
	if host == "" {
		host = "http://localhost:11434"
	}

	status := OllamaStatus{
		Available:   false,
		ServerTime:  time.Now().Format(time.RFC3339),
		Models:      []string{},
		Initialized: false,
	}

	// Check if Ollama server is running
	versionResp, err := http.Get(host + "/api/version")
	if err != nil {
		return status, fmt.Errorf("failed to connect to Ollama server: %w", err)
	}
	defer versionResp.Body.Close()

	if versionResp.StatusCode != http.StatusOK {
		return status, fmt.Errorf("Ollama server returned non-OK status: %d", versionResp.StatusCode)
	}

	// Parse version info
	var versionInfo struct {
		Version string `json:"version"`
	}
	if err := json.NewDecoder(versionResp.Body).Decode(&versionInfo); err == nil {
		status.Version = versionInfo.Version
	}

	// Mark as available
	status.Available = true

	// Get available models
	modelResp, err := http.Get(host + "/api/tags")
	if err == nil && modelResp.StatusCode == http.StatusOK {
		defer modelResp.Body.Close()

		var modelList struct {
			Models []struct {
				Name string `json:"name"`
			} `json:"models"`
		}

		if err := json.NewDecoder(modelResp.Body).Decode(&modelList); err == nil {
			for _, model := range modelList.Models {
				status.Models = append(status.Models, model.Name)
			}

			// Check if our configured model is available
			for _, model := range status.Models {
				if model == s.Config.OllamaModel {
					status.Initialized = true
					break
				}
			}
		}
	}

	utils.Info(fmt.Sprintf("Ollama server check: available=%v, version=%s, models=%d",
		status.Available, status.Version, len(status.Models)))
	return status, nil
}

// setEnvironmentVariables sets environment variables from config
func setEnvironmentVariables(config *Config) {
	// Server config
	os.Setenv("PORT", fmt.Sprintf("%d", config.Port))
	os.Setenv("LOG_LEVEL", logLevelToString(config.LogLevel))

	// Authentication
	if config.AuthToken != "" {
		os.Setenv("AUTH_TOKEN", config.AuthToken)
	}

	// LLM configuration
	if config.OllamaModel != "" {
		os.Setenv("OLLAMA_MODEL", config.OllamaModel)
	}
	if config.OllamaHost != "" {
		os.Setenv("OLLAMA_HOST", config.OllamaHost)
	}
	if config.OpenAIAPIKey != "" {
		os.Setenv("OPENAI_API_KEY", config.OpenAIAPIKey)
	}
	if config.OpenAIModel != "" {
		os.Setenv("OPENAI_MODEL", config.OpenAIModel)
	}
	if config.AnthropicAPIKey != "" {
		os.Setenv("ANTHROPIC_API_KEY", config.AnthropicAPIKey)
	}
	if config.AnthropicModel != "" {
		os.Setenv("ANTHROPIC_MODEL", config.AnthropicModel)
	}
}

// logLevelToString converts a log level to its string representation
func logLevelToString(level utils.LogLevel) string { // Changed parameter type
	switch level {
	case utils.DebugLevel:
		return "DEBUG"
	case utils.InfoLevel:
		return "INFO"
	case utils.WarnLevel:
		return "WARN"
	case utils.ErrorLevel:
		return "ERROR"
	default:
		return "INFO"
	}
}
