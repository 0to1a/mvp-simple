package api

import (
	"encoding/json"
	"testing"
)

func TestCompanyResponseJSONStructure(t *testing.T) {
	// Test that the CompanyResponse struct produces the expected JSON
	company := CompanyResponse{
		CompanyID:   1,
		CompanyName: "Test Company",
		IsAdmin:     true,
	}

	jsonBytes, err := json.Marshal(company)
	if err != nil {
		t.Fatalf("Failed to marshal CompanyResponse: %v", err)
	}

	expectedJSON := `{"CompanyID":1,"CompanyName":"Test Company","IsAdmin":true}`
	actualJSON := string(jsonBytes)

	if actualJSON != expectedJSON {
		t.Errorf("Expected JSON: %s, got: %s", expectedJSON, actualJSON)
	}
}

func TestCompaniesListResponseStructure(t *testing.T) {
	// Test that the response structure matches the expected format
	companies := []CompanyResponse{
		{
			CompanyID:   1,
			CompanyName: "Test Company",
			IsAdmin:     true,
		},
		{
			CompanyID:   2,
			CompanyName: "Another Company",
			IsAdmin:     false,
		},
	}

	response := map[string]interface{}{
		"data": companies,
	}

	jsonBytes, err := json.Marshal(response)
	if err != nil {
		t.Fatalf("Failed to marshal response: %v", err)
	}

	// Parse it back to verify structure
	var parsed map[string]interface{}
	if err := json.Unmarshal(jsonBytes, &parsed); err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	// Check that "data" key exists
	data, exists := parsed["data"]
	if !exists {
		t.Error("Expected 'data' key in response")
	}

	// Check that data is an array
	dataArray, ok := data.([]interface{})
	if !ok {
		t.Error("Expected 'data' to be an array")
	}

	// Check that we have the right number of companies
	if len(dataArray) != 2 {
		t.Errorf("Expected 2 companies, got %d", len(dataArray))
	}

	// Check the first company structure
	if len(dataArray) > 0 {
		firstCompany, ok := dataArray[0].(map[string]interface{})
		if !ok {
			t.Error("Expected company to be an object")
		} else {
			// Check required fields exist and have correct types
			if companyID, exists := firstCompany["CompanyID"]; !exists {
				t.Error("Expected CompanyID field")
			} else if _, ok := companyID.(float64); !ok { // JSON numbers are float64
				t.Error("Expected CompanyID to be a number")
			}

			if companyName, exists := firstCompany["CompanyName"]; !exists {
				t.Error("Expected CompanyName field")
			} else if _, ok := companyName.(string); !ok {
				t.Error("Expected CompanyName to be a string")
			}

			if isAdmin, exists := firstCompany["IsAdmin"]; !exists {
				t.Error("Expected IsAdmin field")
			} else if _, ok := isAdmin.(bool); !ok {
				t.Error("Expected IsAdmin to be a boolean")
			}

			// Ensure no nested objects (like the old sql.NullBool structure)
			for key, value := range firstCompany {
				if nestedObj, isObj := value.(map[string]interface{}); isObj {
					if _, hasBool := nestedObj["Bool"]; hasBool {
						if _, hasValid := nestedObj["Valid"]; hasValid {
							t.Errorf("Field %s still has the old sql.NullBool structure", key)
						}
					}
				}
			}
		}
	}
}
