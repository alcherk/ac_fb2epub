// Package config provides configuration management for the application.
package config

import (
	"os"
)

// Config holds application configuration.
type Config struct {
	Port        string
	Environment string
	TempDir     string
	MaxFileSize int64 // in bytes
}

// Load reads configuration from environment variables and returns a Config instance.
func Load() *Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	env := os.Getenv("ENVIRONMENT")
	if env == "" {
		env = "development"
	}

	tempDir := os.Getenv("TEMP_DIR")
	if tempDir == "" {
		tempDir = "/tmp/fb2epub"
	}

	maxFileSize := int64(50 * 1024 * 1024) // 50MB default
	_ = os.Getenv("MAX_FILE_SIZE")         // Reserved for future use

	return &Config{
		Port:        port,
		Environment: env,
		TempDir:     tempDir,
		MaxFileSize: maxFileSize,
	}
}
