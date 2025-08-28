package internal

import (
	"database/sql"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func TestNewApp(t *testing.T) {
	// Create a test config
	cfg := Config{
		DatabaseDSN: "postgres://test",
		JWTSecret:   "test-secret",
	}

	// Create a mock database connection
	db, err := sql.Open("pgx", "postgres://user:pass@localhost/nonexistent?sslmode=disable")
	if err != nil {
		t.Fatalf("Failed to create mock DB: %v", err)
	}
	defer db.Close()

	// Test NewApp
	app := NewApp(cfg, db)

	// Verify app fields
	if app.Cfg.DatabaseDSN != cfg.DatabaseDSN {
		t.Errorf("Expected DatabaseDSN %q, got %q", cfg.DatabaseDSN, app.Cfg.DatabaseDSN)
	}
	if app.Cfg.JWTSecret != cfg.JWTSecret {
		t.Errorf("Expected JWTSecret %q, got %q", cfg.JWTSecret, app.Cfg.JWTSecret)
	}

	// Verify DB connection is set
	if app.DB == nil {
		t.Error("Expected DB to be set, got nil")
	}

	// Verify Queries is initialized
	if app.Queries == nil {
		t.Error("Expected Queries to be initialized, got nil")
	}

	// Verify Cache is initialized
	if app.Cache == nil {
		t.Error("Expected Cache to be initialized, got nil")
	}
}

func TestAppCacheOperations(t *testing.T) {
	// Create test app
	cfg := Config{
		DatabaseDSN: "postgres://test",
		JWTSecret:   "test-secret",
	}

	db, err := sql.Open("pgx", "postgres://user:pass@localhost/nonexistent?sslmode=disable")
	if err != nil {
		t.Fatalf("Failed to create mock DB: %v", err)
	}
	defer db.Close()

	app := NewApp(cfg, db)

	t.Run("CacheSet and CacheGet", func(t *testing.T) {
		key := "test-key"
		value := "test-value"
		ttl := 5 * time.Second

		// Test CacheSet
		app.CacheSet(key, value, ttl)

		// Test CacheGet - should exist
		retrieved, exists := app.CacheGet(key)
		if !exists {
			t.Error("Expected key to exist in cache, but it doesn't")
		}
		if retrieved != value {
			t.Errorf("Expected value %q, got %q", value, retrieved)
		}
	})

	t.Run("CacheGet non-existent key", func(t *testing.T) {
		_, exists := app.CacheGet("non-existent-key")
		if exists {
			t.Error("Expected key to not exist in cache, but it does")
		}
	})

	t.Run("Cache expiration", func(t *testing.T) {
		key := "expire-key"
		value := "expire-value"
		ttl := 100 * time.Millisecond

		// Set cache with short TTL
		app.CacheSet(key, value, ttl)

		// Verify it exists initially
		_, exists := app.CacheGet(key)
		if !exists {
			t.Error("Expected key to exist immediately after setting")
		}

		// Wait for expiration
		time.Sleep(150 * time.Millisecond)

		// Verify it's expired
		_, exists = app.CacheGet(key)
		if exists {
			t.Error("Expected key to be expired, but it still exists")
		}
	})

	t.Run("Cache different data types", func(t *testing.T) {
		tests := []struct {
			name  string
			key   string
			value interface{}
		}{
			{"string", "string-key", "string-value"},
			{"int", "int-key", 42},
			{"bool", "bool-key", true},
			{"map", "map-key", map[string]interface{}{"nested": "value"}},
			{"slice", "slice-key", []string{"a", "b", "c"}},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				app.CacheSet(tt.key, tt.value, 5*time.Second)
				retrieved, exists := app.CacheGet(tt.key)
				if !exists {
					t.Errorf("Expected %s to exist in cache", tt.name)
				}
				
				// For complex types, we can't do direct comparison easily
				// but we can check that something was retrieved
				if retrieved == nil && tt.value != nil {
					t.Errorf("Expected non-nil value for %s", tt.name)
				}
			})
		}
	})
}

func TestAppCacheConfiguration(t *testing.T) {
	cfg := Config{
		DatabaseDSN: "postgres://test",
		JWTSecret:   "test-secret",
	}

	db, err := sql.Open("pgx", "postgres://user:pass@localhost/nonexistent?sslmode=disable")
	if err != nil {
		t.Fatalf("Failed to create mock DB: %v", err)
	}
	defer db.Close()

	app := NewApp(cfg, db)

	// Test that cache uses the hardcoded TTL (60 seconds)
	// We can't directly test the internal cache configuration,
	// but we can test that the cache works with a reasonable TTL
	key := "ttl-test-key"
	value := "ttl-test-value"
	ttl := 2 * time.Second

	app.CacheSet(key, value, ttl)
	
	// Should exist immediately
	_, exists := app.CacheGet(key)
	if !exists {
		t.Error("Expected key to exist with set TTL")
	}

	// Should still exist before TTL expires
	time.Sleep(1 * time.Second)
	_, exists = app.CacheGet(key)
	if !exists {
		t.Error("Expected key to still exist before TTL expiry")
	}
}
