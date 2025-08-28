package testutil

import (
	"time"
	"database/sql"
	core "project/internal"

	"github.com/golang-jwt/jwt/v5"
	_ "github.com/jackc/pgx/v5/stdlib"
)

// CreateTestApp creates a test application instance with mock database
func CreateTestApp() *core.App {
	cfg := core.Config{
		DatabaseDSN: "postgres://test",
		JWTSecret:   "test-secret-key",
	}

	// Create a mock database connection (won't actually connect)
	db, _ := sql.Open("pgx", "postgres://user:pass@localhost/nonexistent?sslmode=disable")
	
	return core.NewApp(cfg, db)
}

// CreateTestConfig creates a test configuration
func CreateTestConfig() core.Config {
	return core.Config{
		DatabaseDSN: "postgres://test",
		JWTSecret:   "test-secret-key",
	}
}

// CreateValidJWTToken creates a valid JWT token for testing
func CreateValidJWTToken(secret string, userID, companyID int32, isAdmin bool, tokenType string) (string, error) {
	var expiration time.Duration
	if tokenType == "refresh" {
		expiration = 7 * 24 * time.Hour // 7 days
	} else {
		expiration = 24 * time.Hour // 24 hours
	}

	claims := jwt.MapClaims{
		"sub":        float64(userID), // JWT lib converts numbers to float64
		"company_id": float64(companyID),
		"is_admin":   isAdmin,
		"type":       tokenType,
		"exp":        time.Now().Add(expiration).Unix(),
		"iat":        time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// CreateExpiredJWTToken creates an expired JWT token for testing
func CreateExpiredJWTToken(secret string, userID, companyID int32) (string, error) {
	claims := jwt.MapClaims{
		"sub":        float64(userID),
		"company_id": float64(companyID),
		"is_admin":   false,
		"type":       "access",
		"exp":        time.Now().Add(-1 * time.Hour).Unix(), // Expired 1 hour ago
		"iat":        time.Now().Add(-2 * time.Hour).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// CreateInvalidJWTToken creates an invalid JWT token for testing
func CreateInvalidJWTToken() string {
	return "invalid.jwt.token"
}

// CreateTokenWithMissingClaims creates a JWT token with missing required claims
func CreateTokenWithMissingClaims(secret string, missingClaim string) (string, error) {
	claims := jwt.MapClaims{
		"exp": time.Now().Add(24 * time.Hour).Unix(),
		"iat": time.Now().Unix(),
	}

	// Add claims except the missing one
	if missingClaim != "sub" {
		claims["sub"] = float64(123)
	}
	if missingClaim != "company_id" {
		claims["company_id"] = float64(456)
	}
	if missingClaim != "type" {
		claims["type"] = "access"
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// CreateTokenWithInvalidClaims creates a JWT token with invalid claim types
func CreateTokenWithInvalidClaims(secret string, invalidClaim string) (string, error) {
	claims := jwt.MapClaims{
		"sub":        float64(123),
		"company_id": float64(456),
		"is_admin":   false,
		"type":       "access",
		"exp":        time.Now().Add(24 * time.Hour).Unix(),
		"iat":        time.Now().Unix(),
	}

	// Make specific claim invalid
	switch invalidClaim {
	case "sub":
		claims["sub"] = "invalid-string-id"
	case "company_id":
		claims["company_id"] = "invalid-string-id"
	case "type":
		claims["type"] = 12345 // Should be string
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}
