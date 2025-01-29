package tests

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"
)

const (
	baseURL  = "http://localhost:8082"
	username = "admin"
	password = "strongpassword"
)

func TestGetOperations(t *testing.T) {
	endpoints := []string{"/blockchain/balance", "/blockchain/latestBlock", "/config", "/wallet"}

	for _, endpoint := range endpoints {
		t.Run(endpoint, func(t *testing.T) {
			req, err := http.NewRequest("GET", fmt.Sprintf("%s%s", baseURL, endpoint), nil)
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}

			req.SetBasicAuth(username, password)

			client := &http.Client{}
			resp, err := client.Do(req)
			if err != nil {
				t.Fatalf("Failed to make request: %v", err)
			}
			defer resp.Body.Close()

			// Debug HTTP response status
			if resp.StatusCode != http.StatusOK {
				t.Errorf("Unexpected status %d for %s", resp.StatusCode, endpoint)
				body, _ := io.ReadAll(resp.Body)
				t.Logf("Response body: %s", string(body))
				return
			}

			// Adjust content type check to allow variations
			contentType := resp.Header.Get("Content-Type")
			if contentType != "application/json" && contentType != "application/json; charset=utf-8" {
				t.Errorf("Expected content type application/json, got %s", contentType)
			}

			// Read response body
			body, _ := io.ReadAll(resp.Body)

			// Parse JSON into a map
			var result map[string]interface{}
			if err := json.Unmarshal(body, &result); err != nil {
				t.Fatalf("Failed to parse JSON for %s: %v", endpoint, err)
			}

			// Pretty-print JSON response
			prettyJSON, err := json.MarshalIndent(result, "", "  ")
			if err != nil {
				t.Fatalf("Failed to format JSON for %s: %v", endpoint, err)
			}

			// Print nicely formatted JSON
			fmt.Printf("\nResponse for %s:\n%s\n\n", endpoint, string(prettyJSON))

			// Log formatted JSON for CI/CD debugging
			t.Logf("Parsed JSON Response for %s:\n%s", endpoint, string(prettyJSON))
		})
	}
}
