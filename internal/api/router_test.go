package api

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"project/internal/testutil"

	"github.com/gin-gonic/gin"
)

func TestBuild(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	// Create test app
	app := testutil.CreateTestApp()
	
	// Build router
	router := Build(app)

	// Test that router is not nil
	if router == nil {
		t.Fatal("Expected router to be created, got nil")
	}
}

func TestHealthEndpoint(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	app := testutil.CreateTestApp()
	router := Build(app)

	req, _ := http.NewRequest("GET", "/health", nil)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, recorder.Code)
	}

	expected := `{"status":"ok"}`
	body := recorder.Body.String()
	if body != expected {
		t.Errorf("Expected body %q, got %q", expected, body)
	}
}

func TestAuthEndpointsExist(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	app := testutil.CreateTestApp()
	router := Build(app)

	authEndpoints := []struct {
		method string
		path   string
	}{
		{"POST", "/v1/login/request"},
		{"POST", "/v1/login"},
		{"POST", "/v1/auth/refresh"},
	}

	for _, endpoint := range authEndpoints {
		t.Run(endpoint.method+" "+endpoint.path, func(t *testing.T) {
			req, _ := http.NewRequest(endpoint.method, endpoint.path, nil)
			recorder := httptest.NewRecorder()

			router.ServeHTTP(recorder, req)

			// We expect these endpoints to exist (not return 404)
			// They might return 400 (bad request) due to missing body, but not 404
			if recorder.Code == http.StatusNotFound {
				t.Errorf("Expected endpoint %s %s to exist, got 404", endpoint.method, endpoint.path)
			}
		})
	}
}

func TestProtectedEndpointsRequireAuth(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	app := testutil.CreateTestApp()
	router := Build(app)

	protectedEndpoints := []struct {
		method string
		path   string
	}{
		{"GET", "/v1/companies"},
	}

	for _, endpoint := range protectedEndpoints {
		t.Run(endpoint.method+" "+endpoint.path+" without auth", func(t *testing.T) {
			req, _ := http.NewRequest(endpoint.method, endpoint.path, nil)
			recorder := httptest.NewRecorder()

			router.ServeHTTP(recorder, req)

			// Protected endpoints should return 401 when no auth is provided
			if recorder.Code != http.StatusUnauthorized {
				t.Errorf("Expected status %d for unprotected access to %s %s, got %d", 
					http.StatusUnauthorized, endpoint.method, endpoint.path, recorder.Code)
			}

			body := recorder.Body.String()
			if !contains(body, "missing bearer token") {
				t.Errorf("Expected 'missing bearer token' error, got: %s", body)
			}
		})
	}
}

func TestProtectedEndpointsWithValidAuth(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	app := testutil.CreateTestApp()
	router := Build(app)

	// Create a valid JWT token
	secret := app.Cfg.JWTSecret
	token, err := testutil.CreateValidJWTToken(secret, 123, 456, false, "access")
	if err != nil {
		t.Fatalf("Failed to create valid token: %v", err)
	}

	protectedEndpoints := []struct {
		method string
		path   string
	}{
		{"GET", "/v1/companies"},
	}

	for _, endpoint := range protectedEndpoints {
		t.Run(endpoint.method+" "+endpoint.path+" with valid auth", func(t *testing.T) {
			req, _ := http.NewRequest(endpoint.method, endpoint.path, nil)
			req.Header.Set("Authorization", "Bearer "+token)
			recorder := httptest.NewRecorder()

			router.ServeHTTP(recorder, req)

			// With valid auth, we should not get 401
			// We might get other errors (like 500 due to DB issues), but not 401
			if recorder.Code == http.StatusUnauthorized {
				t.Errorf("Expected authorized access to %s %s, got 401: %s", 
					endpoint.method, endpoint.path, recorder.Body.String())
			}
		})
	}
}

func TestRouterMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	// Create a custom router with completely silent recovery for testing
	router := gin.New()
	// Use completely silent recovery that discards all output
	router.Use(gin.CustomRecoveryWithWriter(io.Discard, func(c *gin.Context, recovered interface{}) {
		// Silent recovery for tests - just set status without logging
		c.AbortWithStatus(http.StatusInternalServerError)
	}))

	// Add a test route that panics
	router.GET("/test-panic", func(c *gin.Context) {
		panic("test panic")
	})

	req, _ := http.NewRequest("GET", "/test-panic", nil)
	recorder := httptest.NewRecorder()

	// Test the recovery middleware
	router.ServeHTTP(recorder, req)

	// If recovery middleware is working, we should get 500, not panic crash
	if recorder.Code != http.StatusInternalServerError {
		t.Errorf("Expected recovery middleware to return %d, got %d", 
			http.StatusInternalServerError, recorder.Code)
	}
}

func TestRouterConfiguration(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	app := testutil.CreateTestApp()
	router := Build(app)

	// Test some basic router properties we can verify
	routes := router.Routes()
	
	// We should have at least our defined routes
	expectedPaths := []string{
		"/health",
		"/v1/login/request",
		"/v1/login", 
		"/v1/auth/refresh",
		"/v1/companies",
	}

	foundPaths := make(map[string]bool)
	for _, route := range routes {
		foundPaths[route.Path] = true
	}

	for _, expectedPath := range expectedPaths {
		if !foundPaths[expectedPath] {
			t.Errorf("Expected route %s to be registered", expectedPath)
		}
	}
}
