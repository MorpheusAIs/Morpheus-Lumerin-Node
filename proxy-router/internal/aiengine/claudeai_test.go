package aiengine

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	gcs "github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/chatstorage/genericchatstorage"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/lib"
	"github.com/stretchr/testify/require"
)

const minimalClaudeAIResponse = `{"id":"msg_1","type":"message","role":"assistant","model":"claude-x",` +
	`"content":[{"type":"text","text":"ok"}],"usage":{"input_tokens":1,"output_tokens":1}}`

// promptClaudeAIViaMockBackend drives ClaudeAI.Prompt against a fake
// Anthropic Messages backend and returns the request body it received, so
// tests can assert on the max_tokens the adapter actually forwarded.
func promptClaudeAIViaMockBackend(t *testing.T, reqMaxTokens int) map[string]any {
	t.Helper()

	var captured map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		require.NoError(t, json.Unmarshal(body, &captured))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(minimalClaudeAIResponse))
	}))
	defer server.Close()

	engine := NewClaudeAIEngine("claude-x", server.URL, "", time.Minute, lib.NewTestLogger(), nil)

	req := &gcs.OpenAICompletionRequestExtra{}
	req.MaxTokens = reqMaxTokens

	err := engine.Prompt(context.Background(), req, func(ctx context.Context, chunk gcs.Chunk, errResp *gcs.AiEngineErrorResponse) error {
		return nil
	})
	require.NoError(t, err)
	return captured
}

func TestPromptForwardsCallerMaxTokens(t *testing.T) {
	captured := promptClaudeAIViaMockBackend(t, 4096)
	require.Equal(t, float64(4096), captured["max_tokens"])
}

func TestPromptDefaultsMaxTokensWhenUnset(t *testing.T) {
	captured := promptClaudeAIViaMockBackend(t, 0)
	require.Equal(t, float64(1024), captured["max_tokens"])
}
