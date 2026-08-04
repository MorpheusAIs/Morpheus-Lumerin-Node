package proxyapi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/aiengine"
	gsc "github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/chatstorage/genericchatstorage"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/lib"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gin-gonic/gin"
)

type stubAdapter struct {
	errResp *gsc.AiEngineErrorResponse
}

func (s *stubAdapter) Prompt(ctx context.Context, prompt *gsc.OpenAICompletionRequestExtra, cb gsc.CompletionCallback) error {
	return cb(ctx, nil, s.errResp)
}

func (s *stubAdapter) AudioTranscription(ctx context.Context, prompt *gsc.AudioTranscriptionRequest, cb gsc.CompletionCallback) error {
	return nil
}

func (s *stubAdapter) AudioSpeech(ctx context.Context, prompt *gsc.AudioSpeechRequest, cb gsc.CompletionCallback) error {
	return nil
}

func (s *stubAdapter) Embeddings(ctx context.Context, prompt *gsc.EmbeddingsRequest, cb gsc.CompletionCallback) error {
	return nil
}

func (s *stubAdapter) ApiType() string { return "openai" }

type stubAIEngine struct {
	adapter aiengine.AIEngineStream
}

func (s *stubAIEngine) GetLocalModels() ([]aiengine.LocalModel, error) { return nil, nil }
func (s *stubAIEngine) GetLocalAgents() ([]aiengine.LocalAgent, error) { return nil, nil }
func (s *stubAIEngine) CallAgentTool(ctx context.Context, sessionID, agentID common.Hash, toolName string, input map[string]interface{}) (interface{}, error) {
	return nil, nil
}
func (s *stubAIEngine) GetAgentTools(ctx context.Context, sessionID, agentID common.Hash) ([]aiengine.AgentTool, error) {
	return nil, nil
}
func (s *stubAIEngine) GetAdapter(ctx context.Context, chatID, modelID, sessionID common.Hash, storeContext, forwardContext bool) (aiengine.AIEngineStream, error) {
	return s.adapter, nil
}

func promptWithModelError(t *testing.T, errResp *gsc.AiEngineErrorResponse) *httptest.ResponseRecorder {
	t.Helper()
	gin.SetMode(gin.TestMode)

	controller := &ProxyController{
		aiEngine: &stubAIEngine{adapter: &stubAdapter{errResp: errResp}},
		log:      &lib.LoggerMock{},
	}

	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = httptest.NewRequest("POST", "/v1/chat/completions",
		strings.NewReader(`{"messages":[{"role":"user","content":"hello"}]}`))
	ctx.Request.Header.Set("Content-Type", "application/json")

	controller.Prompt(ctx)
	return w
}

func TestPromptReturnsUpstreamStatusForModelError(t *testing.T) {
	errResp := gsc.NewAiEngineErrorResponse(
		http.StatusTooManyRequests,
		map[string]interface{}{"error": map[string]interface{}{"message": "Rate limit exceeded"}},
	)

	w := promptWithModelError(t, errResp)

	if w.Code != http.StatusTooManyRequests {
		t.Errorf("status = %d, want %d", w.Code, http.StatusTooManyRequests)
	}
	var body gsc.AiEngineErrorResponse
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("response body is not an AiEngineErrorResponse: %v", err)
	}
	if body.ProviderModelError == nil {
		t.Error("providerModelError missing from response body")
	}
	if body.StatusCode != http.StatusTooManyRequests {
		t.Errorf("statusCode = %d, want %d", body.StatusCode, http.StatusTooManyRequests)
	}
}

func TestPromptDefaultsTo400WhenStatusUnknown(t *testing.T) {
	// Older providers don't send statusCode; keep the historical 400.
	errResp := gsc.NewAiEngineErrorResponse(0, map[string]interface{}{"error": "Authentication failed"})

	w := promptWithModelError(t, errResp)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}
