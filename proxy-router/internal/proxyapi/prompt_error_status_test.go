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
	if s.errResp == nil {
		return nil
	}
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
	adapter         aiengine.AIEngineStream
	getAdapterCalls int
	storeContext    bool
	forwardContext  bool
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
	s.getAdapterCalls++
	s.storeContext = storeContext
	s.forwardContext = forwardContext
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

func TestPromptHistoryHeaderControlsRequestContext(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name            string
		header          string
		storeDefault    bool
		forwardDefault  bool
		wantStore       bool
		wantForward     bool
		wantStatus      int
		wantAdapterCall bool
	}{
		{name: "missing preserves defaults", storeDefault: true, forwardDefault: true, wantStore: true, wantForward: true, wantStatus: http.StatusOK, wantAdapterCall: true},
		{name: "default preserves defaults", header: "default", storeDefault: true, forwardDefault: false, wantStore: true, wantForward: false, wantStatus: http.StatusOK, wantAdapterCall: true},
		{name: "on cannot override disabled server policy", header: "on", storeDefault: false, forwardDefault: false, wantStore: false, wantForward: false, wantStatus: http.StatusOK, wantAdapterCall: true},
		{name: "off disables storage and forwarding", header: "off", storeDefault: true, forwardDefault: true, wantStore: false, wantForward: false, wantStatus: http.StatusOK, wantAdapterCall: true},
		{name: "off is case insensitive", header: " OFF ", storeDefault: true, forwardDefault: true, wantStore: false, wantForward: false, wantStatus: http.StatusOK, wantAdapterCall: true},
		{name: "invalid value is rejected", header: "false", storeDefault: true, forwardDefault: true, wantStatus: http.StatusBadRequest, wantAdapterCall: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine := &stubAIEngine{adapter: &stubAdapter{}}
			controller := &ProxyController{
				aiEngine:           engine,
				log:                &lib.LoggerMock{},
				storeChatContext:   tt.storeDefault,
				forwardChatContext: tt.forwardDefault,
			}

			w := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(w)
			ctx.Request = httptest.NewRequest(http.MethodPost, "/v1/chat/completions",
				strings.NewReader(`{"messages":[{"role":"user","content":"hello"}]}`))
			ctx.Request.Header.Set("Content-Type", "application/json")
			if tt.header != "" {
				ctx.Request.Header.Set("x-morpheus-history", tt.header)
			}

			controller.Prompt(ctx)

			if w.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d; body = %s", w.Code, tt.wantStatus, w.Body.String())
			}
			if got := engine.getAdapterCalls > 0; got != tt.wantAdapterCall {
				t.Errorf("adapter called = %t, want %t", got, tt.wantAdapterCall)
			}
			if tt.wantAdapterCall {
				if engine.storeContext != tt.wantStore {
					t.Errorf("storeContext = %t, want %t", engine.storeContext, tt.wantStore)
				}
				if engine.forwardContext != tt.wantForward {
					t.Errorf("forwardContext = %t, want %t", engine.forwardContext, tt.wantForward)
				}
			}
		})
	}
}
