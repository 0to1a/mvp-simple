package internal

import (
	"os"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	// Save original environment
	origDatabaseURL := os.Getenv("DATABASE_URL")
	origJWTSecret := os.Getenv("JWT_SECRET")

	// Clean up after test
	defer func() {
		os.Setenv("DATABASE_URL", origDatabaseURL)
		os.Setenv("JWT_SECRET", origJWTSecret)
	}()

	tests := []struct {
		name        string
		envVars     map[string]string
		expectPanic bool
		expected    Config
	}{
		{
			name: "default values",
			envVars: map[string]string{
				"DATABASE_URL": "postgres://test",
			},
			expected: Config{
				DatabaseDSN: "postgres://test",
				JWTSecret:   "dev-secret-change-me",
			},
		},
		{
			name: "custom values",
			envVars: map[string]string{
				"DATABASE_URL":       "postgres://custom",
				"JWT_SECRET":         "custom-secret",
			},
			expected: Config{
				DatabaseDSN: "postgres://custom",
				JWTSecret:   "custom-secret",
			},
		},
		// Note: We can't easily test log.Fatal in unit tests as it calls os.Exit()
		// This test would terminate the test process
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear environment
			os.Unsetenv("DATABASE_URL")
			os.Unsetenv("JWT_SECRET")

			// Set test environment variables
			for key, value := range tt.envVars {
				if value != "" {
					os.Setenv(key, value)
				}
			}

			if tt.expectPanic {
				// Test panic case
				defer func() {
					if r := recover(); r == nil {
						t.Error("Expected panic but didn't get one")
					}
				}()
				LoadConfig()
				t.Error("Expected panic but function completed normally")
				return
			}

			config := LoadConfig()

			if config.DatabaseDSN != tt.expected.DatabaseDSN {
				t.Errorf("Expected DatabaseDSN %q, got %q", tt.expected.DatabaseDSN, config.DatabaseDSN)
			}
			if config.JWTSecret != tt.expected.JWTSecret {
				t.Errorf("Expected JWTSecret %q, got %q", tt.expected.JWTSecret, config.JWTSecret)
			}
		})
	}
}

func TestGetenv(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		def      string
		envValue string
		expected string
	}{
		{
			name:     "environment variable exists",
			key:      "TEST_KEY",
			def:      "default",
			envValue: "custom",
			expected: "custom",
		},
		{
			name:     "environment variable empty returns default",
			key:      "TEST_KEY",
			def:      "default",
			envValue: "",
			expected: "default",
		},
		{
			name:     "environment variable not set returns default",
			key:      "NONEXISTENT_KEY",
			def:      "default",
			envValue: "",
			expected: "default",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean up
			os.Unsetenv("TEST_KEY")

			if tt.envValue != "" {
				os.Setenv("TEST_KEY", tt.envValue)
			}

			result := getenv(tt.key, tt.def)
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}
