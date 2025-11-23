// Package config provides configuration management for the application.
package config

import (
	"os"
	"strconv"
)

// Config holds application configuration.
type Config struct {
	Port                string
	Environment         string
	TempDir             string
	MaxFileSize         int64 // in bytes
	CleanupTriggerCount int   // Number of completed conversions before cleanup
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
	if sizeStr := os.Getenv("MAX_FILE_SIZE"); sizeStr != "" {
		if parsedSize, err := strconv.ParseInt(sizeStr, 10, 64); err == nil && parsedSize > 0 {
			maxFileSize = parsedSize
		}
	}

	cleanupTriggerCount := 10 // Default: cleanup after 10 completed conversions
	if countStr := os.Getenv("CLEANUP_TRIGGER_COUNT"); countStr != "" {
		if parsedCount, err := strconv.Atoi(countStr); err == nil && parsedCount > 0 {
			cleanupTriggerCount = parsedCount
		}
	}

	return &Config{
		Port:                port,
		Environment:         env,
		TempDir:             tempDir,
		MaxFileSize:         maxFileSize,
		CleanupTriggerCount: cleanupTriggerCount,
	}
}
