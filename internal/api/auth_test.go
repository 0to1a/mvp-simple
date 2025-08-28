package api

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"project/internal/testutil"

	"github.com/gin-gonic/gin"
)

// Constants to eliminate duplicate literals
const (
	bearerPrefix = "Bearer "
	statusErrMsg = "Expected status %d, got %d"
)

func TestAuthRequired(t *testing.T) {
	gin.SetMode(gin.TestMode)
	secret := "test-secret-key"

	tests := []struct {
		name           string
		authHeader     string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "valid token",
			authHeader:     "",
			expectedStatus: http.StatusOK,
			expectedBody:   `{"message":"authorized"}`,
		},
		{
			name:           "missing authorization header",
			authHeader:     "",
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   `{"error":"missing bearer token"}`,
		},
		{
			name:           "invalid bearer format",
			authHeader:     "Invalid token-format",
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   `{"error":"missing bearer token"}`,
		},
		{
			name:           "bearer without token",
			authHeader:     bearerPrefix,
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   `{"error":"invalid token"}`,
		},
		{
			name:           "invalid token format",
			authHeader:     bearerPrefix + "invalid.token.format",
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   `{"error":"invalid token"}`,
		},
		{
			name:           "expired token",
			authHeader:     "",
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   `{"error":"invalid token"}`,
		},
		{
			name:           "token with missing sub claim",
			authHeader:     "",
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   `{"error":"missing sub in token"}`,
		},
		{
			name:           "token with invalid sub claim type",
			authHeader:     "",
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   `{"error":"invalid sub in token"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := createTestRouter(secret)
			req, _ := http.NewRequest("GET", "/test", nil)

			// Set up authorization header and get updated expected body
			updatedExpectedBody := setupAuthHeader(t, req, tt.name, secret)
			if updatedExpectedBody != "" {
				tt.expectedBody = updatedExpectedBody
			}
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}

			// Execute request and validate response
			recorder := httptest.NewRecorder()
			router.ServeHTTP(recorder, req)
			validateResponse(t, recorder, tt.expectedStatus, tt.expectedBody)
		})
	}
}

func TestAuthRequiredWithDifferentSecrets(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	correctSecret := "correct-secret"
	wrongSecret := "wrong-secret"

	// Create token with correct secret
	token, err := testutil.CreateValidJWTToken(correctSecret, 123, 456, false, "access")
	if err != nil {
		t.Fatalf("Failed to create token: %v", err)
	}

	// Test with wrong secret in middleware
	router := gin.New()
	router.Use(AuthRequired(wrongSecret))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "authorized"})
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", bearerPrefix+token)

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusUnauthorized {
		t.Errorf(statusErrMsg, http.StatusUnauthorized, recorder.Code)
	}

	responseBody := recorder.Body.String()
	if !contains(responseBody, "invalid token") {
		t.Errorf("Expected 'invalid token' error, got: %s", responseBody)
	}
}

func TestAuthRequiredSetsUserID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	secret := "test-secret"
	userID := int32(12345)

	token, err := testutil.CreateValidJWTToken(secret, userID, 456, false, "access")
	if err != nil {
		t.Fatalf("Failed to create token: %v", err)
	}

	var capturedUserID interface{}
	router := gin.New()
	router.Use(AuthRequired(secret))
	router.GET("/test", func(c *gin.Context) {
		capturedUserID, _ = c.Get("user_id")
		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", bearerPrefix+token)

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Errorf(statusErrMsg, http.StatusOK, recorder.Code)
	}

	if capturedUserID == nil {
		t.Error("Expected user_id to be set in context")
	}

	if capturedID, ok := capturedUserID.(int32); !ok || capturedID != userID {
		t.Errorf("Expected user_id %d, got %v (type: %T)", userID, capturedUserID, capturedUserID)
	}
}

// Helper functions for test setup and validation

// setupAuthHeader sets up the authorization header for a test case
func setupAuthHeader(t *testing.T, req *http.Request, testName, secret string) string {
	switch testName {
	case "valid token":
		token, err := testutil.CreateValidJWTToken(secret, 123, 456, false, "access")
		if err != nil {
			t.Fatalf("Failed to create valid token: %v", err)
		}
		req.Header.Set("Authorization", bearerPrefix+token)
		return `{"message":"authorized","user_id":123}`
	case "expired token":
		token, err := testutil.CreateExpiredJWTToken(secret, 123, 456)
		if err != nil {
			t.Fatalf("Failed to create expired token: %v", err)
		}
		req.Header.Set("Authorization", bearerPrefix+token)
	case "token with missing sub claim":
		token, err := testutil.CreateTokenWithMissingClaims(secret, "sub")
		if err != nil {
			t.Fatalf("Failed to create token with missing sub: %v", err)
		}
		req.Header.Set("Authorization", bearerPrefix+token)
	case "token with invalid sub claim type":
		token, err := testutil.CreateTokenWithInvalidClaims(secret, "sub")
		if err != nil {
			t.Fatalf("Failed to create token with invalid sub: %v", err)
		}
		req.Header.Set("Authorization", bearerPrefix+token)
	}
	return ""
}

// validateResponse validates the HTTP response against expected values
func validateResponse(t *testing.T, recorder *httptest.ResponseRecorder, expectedStatus int, expectedBody string) {
	// Check status code
	if recorder.Code != expectedStatus {
		t.Errorf(statusErrMsg, expectedStatus, recorder.Code)
	}

	// Check response body
	responseBody := recorder.Body.String()
	if expectedStatus == http.StatusOK {
		validateSuccessResponse(t, responseBody)
	} else {
		validateErrorResponse(t, responseBody, expectedBody)
	}
}

// validateSuccessResponse validates successful response structure
func validateSuccessResponse(t *testing.T, responseBody string) {
	if !contains(responseBody, "authorized") {
		t.Errorf("Expected response to contain 'authorized', got: %s", responseBody)
	}
	if !contains(responseBody, "user_id") {
		t.Errorf("Expected response to contain 'user_id', got: %s", responseBody)
	}
}

// validateErrorResponse validates error response content
func validateErrorResponse(t *testing.T, responseBody, expectedBody string) {
	if !contains(responseBody, expectedBody) {
		t.Errorf("Expected response to contain %q, got: %s", expectedBody, responseBody)
	}
}

// createTestRouter creates a router with auth middleware for testing
func createTestRouter(secret string) *gin.Engine {
	router := gin.New()
	router.Use(AuthRequired(secret))
	router.GET("/test", func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "user_id not set"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "authorized", "user_id": userID})
	})
	return router
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}
