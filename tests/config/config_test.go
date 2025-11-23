package config_test

import (
	"os"
	"testing"

	"github.com/lex/fb2epub/config"
)

func TestLoad_DefaultValues(t *testing.T) {
	// Clear environment variables
	os.Clearenv()

	cfg := config.Load()

	if cfg.Port != "8080" {
		t.Errorf("Expected default port 8080, got %s", cfg.Port)
	}

	if cfg.Environment != "development" {
		t.Errorf("Expected default environment 'development', got %s", cfg.Environment)
	}

	if cfg.TempDir != "/tmp/fb2epub" {
		t.Errorf("Expected default temp dir '/tmp/fb2epub', got %s", cfg.TempDir)
	}

	expectedMaxFileSize := int64(50 * 1024 * 1024) // 50MB
	if cfg.MaxFileSize != expectedMaxFileSize {
		t.Errorf("Expected default max file size %d, got %d", expectedMaxFileSize, cfg.MaxFileSize)
	}

	if cfg.CleanupTriggerCount != 10 {
		t.Errorf("Expected default cleanup trigger count 10, got %d", cfg.CleanupTriggerCount)
	}
}

func TestLoad_EnvironmentVariables(t *testing.T) {
	tests := []struct {
		name     string
		envVars  map[string]string
		validate func(*testing.T, *config.Config)
	}{
		{
			name: "custom port",
			envVars: map[string]string{
				"PORT": "9090",
			},
			validate: func(t *testing.T, cfg *config.Config) {
				if cfg.Port != "9090" {
					t.Errorf("Expected port 9090, got %s", cfg.Port)
				}
			},
		},
		{
			name: "production environment",
			envVars: map[string]string{
				"ENVIRONMENT": "production",
			},
			validate: func(t *testing.T, cfg *config.Config) {
				if cfg.Environment != "production" {
					t.Errorf("Expected environment 'production', got %s", cfg.Environment)
				}
			},
		},
		{
			name: "custom temp dir",
			envVars: map[string]string{
				"TEMP_DIR": "/custom/temp",
			},
			validate: func(t *testing.T, cfg *config.Config) {
				if cfg.TempDir != "/custom/temp" {
					t.Errorf("Expected temp dir '/custom/temp', got %s", cfg.TempDir)
				}
			},
		},
		{
			name: "custom max file size",
			envVars: map[string]string{
				"MAX_FILE_SIZE": "104857600", // 100MB
			},
			validate: func(t *testing.T, cfg *config.Config) {
				expected := int64(104857600)
				if cfg.MaxFileSize != expected {
					t.Errorf("Expected max file size %d, got %d", expected, cfg.MaxFileSize)
				}
			},
		},
		{
			name: "custom cleanup trigger count",
			envVars: map[string]string{
				"CLEANUP_TRIGGER_COUNT": "5",
			},
			validate: func(t *testing.T, cfg *config.Config) {
				if cfg.CleanupTriggerCount != 5 {
					t.Errorf("Expected cleanup trigger count 5, got %d", cfg.CleanupTriggerCount)
				}
			},
		},
		{
			name: "all variables",
			envVars: map[string]string{
				"PORT":                "3000",
				"ENVIRONMENT":         "production",
				"TEMP_DIR":            "/app/temp",
				"MAX_FILE_SIZE":       "52428800",
				"CLEANUP_TRIGGER_COUNT": "20",
			},
			validate: func(t *testing.T, cfg *config.Config) {
				if cfg.Port != "3000" {
					t.Errorf("Expected port 3000, got %s", cfg.Port)
				}
				if cfg.Environment != "production" {
					t.Errorf("Expected environment 'production', got %s", cfg.Environment)
				}
				if cfg.TempDir != "/app/temp" {
					t.Errorf("Expected temp dir '/app/temp', got %s", cfg.TempDir)
				}
				if cfg.MaxFileSize != 52428800 {
					t.Errorf("Expected max file size 52428800, got %d", cfg.MaxFileSize)
				}
				if cfg.CleanupTriggerCount != 20 {
					t.Errorf("Expected cleanup trigger count 20, got %d", cfg.CleanupTriggerCount)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Clearenv()
			for k, v := range tt.envVars {
				os.Setenv(k, v)
			}
			defer os.Clearenv()

			cfg := config.Load()
			tt.validate(t, cfg)
		})
	}
}

func TestLoad_InvalidValues(t *testing.T) {
	tests := []struct {
		name    string
		envVars  map[string]string
		validate func(*testing.T, *config.Config)
	}{
		{
			name: "negative max file size",
			envVars: map[string]string{
				"MAX_FILE_SIZE": "-100",
			},
			validate: func(t *testing.T, cfg *config.Config) {
				// Should use default value
				expected := int64(50 * 1024 * 1024)
				if cfg.MaxFileSize != expected {
					t.Errorf("Expected default max file size %d, got %d", expected, cfg.MaxFileSize)
				}
			},
		},
		{
			name: "zero max file size",
			envVars: map[string]string{
				"MAX_FILE_SIZE": "0",
			},
			validate: func(t *testing.T, cfg *config.Config) {
				// Should use default value
				expected := int64(50 * 1024 * 1024)
				if cfg.MaxFileSize != expected {
					t.Errorf("Expected default max file size %d, got %d", expected, cfg.MaxFileSize)
				}
			},
		},
		{
			name: "non-numeric max file size",
			envVars: map[string]string{
				"MAX_FILE_SIZE": "not-a-number",
			},
			validate: func(t *testing.T, cfg *config.Config) {
				// Should use default value
				expected := int64(50 * 1024 * 1024)
				if cfg.MaxFileSize != expected {
					t.Errorf("Expected default max file size %d, got %d", expected, cfg.MaxFileSize)
				}
			},
		},
		{
			name: "negative cleanup trigger count",
			envVars: map[string]string{
				"CLEANUP_TRIGGER_COUNT": "-5",
			},
			validate: func(t *testing.T, cfg *config.Config) {
				// Should use default value
				if cfg.CleanupTriggerCount != 10 {
					t.Errorf("Expected default cleanup trigger count 10, got %d", cfg.CleanupTriggerCount)
				}
			},
		},
		{
			name: "zero cleanup trigger count",
			envVars: map[string]string{
				"CLEANUP_TRIGGER_COUNT": "0",
			},
			validate: func(t *testing.T, cfg *config.Config) {
				// Should use default value
				if cfg.CleanupTriggerCount != 10 {
					t.Errorf("Expected default cleanup trigger count 10, got %d", cfg.CleanupTriggerCount)
				}
			},
		},
		{
			name: "non-numeric cleanup trigger count",
			envVars: map[string]string{
				"CLEANUP_TRIGGER_COUNT": "not-a-number",
			},
			validate: func(t *testing.T, cfg *config.Config) {
				// Should use default value
				if cfg.CleanupTriggerCount != 10 {
					t.Errorf("Expected default cleanup trigger count 10, got %d", cfg.CleanupTriggerCount)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Clearenv()
			for k, v := range tt.envVars {
				os.Setenv(k, v)
			}
			defer os.Clearenv()

			cfg := config.Load()
			tt.validate(t, cfg)
		})
	}
}

