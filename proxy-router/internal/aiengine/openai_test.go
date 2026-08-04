package aiengine

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	gcs "github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/chatstorage/genericchatstorage"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/lib"
)

func promptViaMockBackend(t *testing.T, statusCode int, responseBody string) *gcs.AiEngineErrorResponse {
	t.Helper()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)
		_, _ = w.Write([]byte(responseBody))
	}))
	defer server.Close()

	engine := NewOpenAIEngine("test-model", server.URL, "", time.Minute, &lib.LoggerMock{}, nil)

	var captured *gcs.AiEngineErrorResponse
	err := engine.Prompt(context.Background(), &gcs.OpenAICompletionRequestExtra{}, func(ctx context.Context, completion gcs.Chunk, errResp *gcs.AiEngineErrorResponse) error {
		captured = errResp
		return nil
	})
	if err != nil {
		t.Fatalf("Prompt returned unexpected error: %v", err)
	}
	if captured == nil {
		t.Fatal("expected error callback with AiEngineErrorResponse, got none")
	}
	return captured
}

func TestReadErrorPropagatesUpstreamStatusCode(t *testing.T) {
	resp := promptViaMockBackend(t, http.StatusTooManyRequests, `{"error":{"message":"Rate limit exceeded","code":"rate_limit_exceeded"}}`)

	if resp.StatusCode != http.StatusTooManyRequests {
		t.Errorf("StatusCode = %d, want %d", resp.StatusCode, http.StatusTooManyRequests)
	}
	if resp.ProviderModelError == nil {
		t.Error("ProviderModelError should contain the parsed upstream body")
	}
}

func TestReadErrorPropagatesStatusForNonJSONBody(t *testing.T) {
	resp := promptViaMockBackend(t, http.StatusServiceUnavailable, "upstream capacity exhausted")

	if resp.StatusCode != http.StatusServiceUnavailable {
		t.Errorf("StatusCode = %d, want %d", resp.StatusCode, http.StatusServiceUnavailable)
	}
	wrapped, ok := resp.ProviderModelError.(map[string]interface{})
	if !ok {
		t.Fatalf("ProviderModelError = %T, want wrapped map", resp.ProviderModelError)
	}
	inner, ok := wrapped["error"].(map[string]interface{})
	if !ok {
		t.Fatalf("wrapped error missing: %v", wrapped)
	}
	if inner["message"] != "upstream capacity exhausted" {
		t.Errorf("message = %v, want raw upstream text", inner["message"])
	}
}

func TestAiEngineErrorResponseHTTPStatusCode(t *testing.T) {
	cases := []struct {
		name     string
		upstream int
		want     int
	}{
		{"unknown defaults to 400", 0, 400},
		{"429 preserved", 429, 429},
		{"503 preserved", 503, 503},
		{"non-error code falls back to 400", 200, 400},
		{"out-of-range falls back to 400", 700, 400},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			resp := gcs.NewAiEngineErrorResponse(tc.upstream, map[string]interface{}{})
			if got := resp.HTTPStatusCode(); got != tc.want {
				t.Errorf("HTTPStatusCode() = %d, want %d", got, tc.want)
			}
		})
	}
}
