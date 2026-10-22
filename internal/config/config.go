package config

import (
	"os"
)

// Config holds environment variables to be used throughout the service
type Config struct {
	Port         string
	DBPath       string
	OpenAIAPIKey string
	OpenAIModel  string
}

// Load returns a Config struct with environment variables, or their defaults
func Load() Config {
	return Config{
		Port:         getEnv("PORT", "8080"),
		DBPath:       getEnv("DB_PATH", "triage.db"),
		OpenAIAPIKey: getEnv("OPENAI_API_KEY", ""),
		OpenAIModel:  getEnv("OPENAI_MODEL", "gpt-5.4-mini"),
	}
}

// getEnv returns the environment variable corresponding to the key, or the fallback if unset or empty
func getEnv(key, fallback string) string {
	value, ok := os.LookupEnv(key)
	if !ok || value == "" {
		return fallback
	}
	return value
}
