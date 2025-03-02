package sonar

import (
	"open-sonar/internal/utils"
)

// Config holds all configuration options for the Open Sonar server
type Config struct {
	// Server configuration
	Port        int
	LogLevel    utils.LogLevel // Changed from int to utils.LogLevel
	TLS         bool
	CertFile    string
	KeyFile     string
	LoadEnvFile bool
	EnvFilePath string

	// Authentication
	AuthToken string

	// LLM configuration
	DefaultProvider string
	OllamaModel     string
	OllamaHost      string
	OpenAIAPIKey    string
	OpenAIModel     string
	AnthropicAPIKey string
	AnthropicModel  string

	// Rate limiting
	MaxRequestsPerMinute       int
	MaxLLMRequestsPerMinute    int
	MaxUnauthRequestsPerMinute int
}

// Option is a function that modifies the Config
type Option func(*Config)

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	return &Config{
		// Server defaults
		Port:        8080,
		LogLevel:    utils.InfoLevel,
		TLS:         false,
		LoadEnvFile: true,
		EnvFilePath: ".env",

		// LLM defaults
		DefaultProvider: "ollama",
		OllamaModel:     "deepseek-r1:1.5b",
		OllamaHost:      "http://localhost:11434",
		OpenAIModel:     "gpt-3.5-turbo",
		AnthropicModel:  "claude-3-opus-20240229",

		// Rate limiting defaults
		MaxRequestsPerMinute:       60,
		MaxLLMRequestsPerMinute:    20,
		MaxUnauthRequestsPerMinute: 30,
	}
}

// WithPort sets the server port
func WithPort(port int) Option {
	return func(c *Config) {
		c.Port = port
	}
}

// WithLogLevel sets the log level
func WithLogLevel(level utils.LogLevel) Option { // Changed parameter type
	return func(c *Config) {
		c.LogLevel = level
	}
}

// WithTLS enables TLS with the given cert and key files
func WithTLS(certFile, keyFile string) Option {
	return func(c *Config) {
		c.TLS = true
		c.CertFile = certFile
		c.KeyFile = keyFile
	}
}

// WithEnvFile sets the environment file path
func WithEnvFile(path string) Option {
	return func(c *Config) {
		c.LoadEnvFile = true
		c.EnvFilePath = path
	}
}

// WithoutEnvFile disables loading environment from file
func WithoutEnvFile() Option {
	return func(c *Config) {
		c.LoadEnvFile = false
	}
}

// WithAuthToken sets the authentication token
func WithAuthToken(token string) Option {
	return func(c *Config) {
		c.AuthToken = token
	}
}

// WithOllama configures the Ollama LLM provider
func WithOllama(model string, host string) Option {
	return func(c *Config) {
		c.DefaultProvider = "ollama"
		c.OllamaModel = model
		c.OllamaHost = host
	}
}

// WithOpenAI configures the OpenAI LLM provider
func WithOpenAI(apiKey string, model string) Option {
	return func(c *Config) {
		c.DefaultProvider = "openai"
		c.OpenAIAPIKey = apiKey
		c.OpenAIModel = model
	}
}

// WithAnthropic configures the Anthropic LLM provider
func WithAnthropic(apiKey string, model string) Option {
	return func(c *Config) {
		c.DefaultProvider = "anthropic"
		c.AnthropicAPIKey = apiKey
		c.AnthropicModel = model
	}
}

// WithRateLimiting configures rate limiting
func WithRateLimiting(maxRequests, maxLLMRequests, maxUnauthRequests int) Option {
	return func(c *Config) {
		c.MaxRequestsPerMinute = maxRequests
		c.MaxLLMRequestsPerMinute = maxLLMRequests
		c.MaxUnauthRequestsPerMinute = maxUnauthRequests
	}
}
