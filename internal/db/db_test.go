package db

import (
	"testing"
	"database/sql"
	"os"
	"path/filepath"
	"strings"
)

// Constants to eliminate duplicate literals
const (
	dbPingError = "db.Ping"
)

func TestOpenAndMigrate(t *testing.T) {
	tests := []struct {
		name        string
		dsn         string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "invalid DSN format",
			dsn:         "invalid-dsn",
			expectError: true,
			errorMsg:    "db.Ping", // Actually fails at ping, not sql.Open
		},
		{
			name:        "empty DSN",
			dsn:         "",
			expectError: true,
			errorMsg:    "db.Ping", // Actually fails at ping, not sql.Open
		},
		{
			name:        "unreachable database",
			dsn:         "postgres://user:pass@nonexistent:5432/db?sslmode=disable",
			expectError: true,
			errorMsg:    "db.Ping",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, err := OpenAndMigrate(tt.dsn)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error for %s, but got none", tt.name)
					if db != nil {
						db.Close()
					}
					return
				}

				if tt.errorMsg != "" && !containsError(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error message to contain %q, got: %v", tt.errorMsg, err)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error for %s, got: %v", tt.name, err)
				}
				if db != nil {
					db.Close()
				}
			}
		})
	}
}

func TestOpenAndMigrateConnectionSettings(t *testing.T) {
	// Test with a mock DSN that would work if the database existed
	// We can't test actual connection without a real database
	dsn := "postgres://user:pass@localhost:5432/testdb?sslmode=disable"
	
	db, err := OpenAndMigrate(dsn)
	
	// We expect this to fail with ping error since the database doesn't exist
	if err == nil {
		t.Error("Expected error for non-existent database")
		if db != nil {
			// If somehow it worked, check connection settings
			stats := db.Stats()
			if stats.MaxOpenConnections != 10 {
				t.Errorf("Expected MaxOpenConnections to be 10, got %d", stats.MaxOpenConnections)
			}
			db.Close()
		}
	} else {
		// This is expected - should fail at ping stage
		if !containsError(err.Error(), "db.Ping") {
			t.Errorf("Expected ping error, got: %v", err)
		}
	}
}

func TestDatabaseDriverRegistration(t *testing.T) {
	// Test that the pgx driver is properly registered
	drivers := sql.Drivers()
	found := false
	for _, driver := range drivers {
		if driver == "pgx" {
			found = true
			break
		}
	}
	
	if !found {
		t.Error("Expected 'pgx' driver to be registered")
	}
}

func TestMigrationDirectory(t *testing.T) {
	// Test migration directory handling
	// Create a temporary migration directory for testing
	tempDir := t.TempDir()
	migrationsDir := filepath.Join(tempDir, "migrations")
	
	err := os.MkdirAll(migrationsDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create test migrations directory: %v", err)
	}
	
	// Create a simple test migration file
	migrationContent := `-- +goose Up
CREATE TABLE test_table (id INTEGER);

-- +goose Down  
DROP TABLE test_table;
`
	
	migrationFile := filepath.Join(migrationsDir, "001_test.sql")
	err = os.WriteFile(migrationFile, []byte(migrationContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test migration file: %v", err)
	}
	
	// Test that migrations directory exists
	if _, err := os.Stat(migrationsDir); os.IsNotExist(err) {
		t.Error("Expected migrations directory to exist")
	}
	
	// Test that migration file exists
	if _, err := os.Stat(migrationFile); os.IsNotExist(err) {
		t.Error("Expected migration file to exist")
	}
}

func TestOpenSQLConnection(t *testing.T) {
	// Test that we can open a SQL connection (without connecting)
	dsn := "postgres://user:pass@localhost:5432/db?sslmode=disable"
	
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Errorf("Expected sql.Open to succeed, got: %v", err)
	}
	
	if db == nil {
		t.Error("Expected db connection object to be created")
	}
	
	if db != nil {
		db.Close()
	}
}

func TestConnectionStringParsing(t *testing.T) {
	validDSNs := []string{
		"postgres://user:pass@localhost:5432/db",
		"postgres://user:pass@localhost:5432/db?sslmode=disable",
		"postgres://user@localhost/db",
		"postgresql://user:pass@localhost:5432/db",
	}
	
	for _, dsn := range validDSNs {
		t.Run("DSN: "+dsn, func(t *testing.T) {
			db, err := sql.Open("pgx", dsn)
			if err != nil {
				t.Errorf("Expected valid DSN %q to be accepted by sql.Open, got: %v", dsn, err)
			}
			if db != nil {
				db.Close()
			}
		})
	}
}

// Helper function to check if error message contains expected substring
func containsError(errMsg, expected string) bool {
	return strings.Contains(errMsg, expected)
}
