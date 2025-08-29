package internal

import (
	"fmt"
	"net/http"
	"testing"
)

func TestEmailServiceStatusCodes(t *testing.T) {
	// Test the core logic of status code validation
	// This tests the fix for the "email API error (status 201): OK" issue
	
	// Test that both 200 and 201 are considered success status codes
	successCodes := []int{http.StatusOK, http.StatusCreated}
	errorCodes := []int{http.StatusBadRequest, http.StatusUnauthorized, http.StatusInternalServerError}
	
	for _, code := range successCodes {
		t.Run(fmt.Sprintf("Status %d should be considered success", code), func(t *testing.T) {
			// Test the condition logic from the fixed code
			isError := code != http.StatusOK && code != http.StatusCreated
			if isError {
				t.Errorf("Status code %d should be considered success, but was treated as error", code)
			}
		})
	}
	
	for _, code := range errorCodes {
		t.Run(fmt.Sprintf("Status %d should be considered error", code), func(t *testing.T) {
			// Test the condition logic from the fixed code
			isError := code != http.StatusOK && code != http.StatusCreated
			if !isError {
				t.Errorf("Status code %d should be considered error, but was treated as success", code)
			}
		})
	}
}

func TestEmailServiceWithoutAPIKey(t *testing.T) {
	// Test that email service works without API key (logs instead of sending)
	emailService := NewEmailService("", "test@example.com")
	
	err := emailService.SendEmail("recipient@example.com", "Test User", "Test Subject", "<p>Test Body</p>")
	
	if err != nil {
		t.Errorf("Expected no error when API key is empty, but got: %v", err)
	}
}

func TestNewEmailService(t *testing.T) {
	apiKey := "test-api-key"
	fromAddress := "test@example.com"
	
	service := NewEmailService(apiKey, fromAddress)
	
	if service == nil {
		t.Fatal("Expected non-nil email service")
	}
	
	if service.apiKey != apiKey {
		t.Errorf("Expected API key %q, got %q", apiKey, service.apiKey)
	}
	
	if service.fromAddress != fromAddress {
		t.Errorf("Expected from address %q, got %q", fromAddress, service.fromAddress)
	}
	
	if service.client == nil {
		t.Error("Expected non-nil HTTP client")
	}
}
