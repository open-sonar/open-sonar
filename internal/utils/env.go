package utils

import "os"

// GetEnvWithDefault returns the value of an environment variable or a default value if not set
func GetEnvWithDefault(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
