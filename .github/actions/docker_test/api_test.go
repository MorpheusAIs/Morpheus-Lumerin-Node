package tests

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
)

const (
	baseURL  = "http://localhost:8082"
	username = "admin"
	password = "strongpassword"
)

func TestAuthentication(t *testing.T) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/auth-endpoint", baseURL), nil)
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

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}

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

			if resp.StatusCode != http.StatusOK {
				t.Errorf("Expected status 200, got %d", resp.StatusCode)
			}

			if resp.Header.Get("Content-Type") != "application/json" {
				t.Errorf("Expected content type application/json, got %s", resp.Header.Get("Content-Type"))
			}

			var result map[string]interface{}
			if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
				t.Fatalf("Failed to parse JSON: %v", err)
			}

			if _, exists := result["key"]; !exists {
				t.Errorf("Expected key 'key' in response, but it was missing")
			}
		})
	}
}
