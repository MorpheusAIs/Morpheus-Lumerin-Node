package modelhealth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/aiengine"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/apidetect"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/blockchainapi/structs"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/config"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/lib"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/system"
	"github.com/ethereum/go-ethereum/common"
	"github.com/sashabaranov/go-openai"
	"github.com/stretchr/testify/require"
)

// Contract §6 hygiene: the real Checker driving the real apidetect.Detector
// against an httptest engine, so the sweep → detect → compose → report path
// is exercised end to end rather than through the fakeDetector. The engine
// is llama.cpp-shaped (tools/server/README.md: /props, /v1/models) and also
// answers the health probe's chat completion, so a real aiengine.OpenAI
// adapter serves as the Adapters dep and every request the checker makes
// lands on one counted server.

// The subset of a Qwen3 chat template the family detector keys on:
// enable_thinking as a template kwarg (internal/apispec/family.go).
const integrationQwen3Template = `{%- if messages %}{% generation %}{% endgeneration %}{%- endif %}
{%- if enable_thinking is defined and enable_thinking is false %}<think></think>{%- endif %}`

const (
	integrationServedCtx  = int64(8192)
	integrationTrainedCtx = int64(32768)
)

// llamaCppBackend counts requests per "METHOD path" so a test can prove
// that detection reached the backend once across sweeps within the TTL
// while the health probe reached it on every sweep.
type llamaCppBackend struct {
	srv *httptest.Server

	mu     sync.Mutex
	counts map[string]int
}

func newLlamaCppBackend(t *testing.T) *llamaCppBackend {
	t.Helper()
	b := &llamaCppBackend{counts: map[string]int{}}

	mux := http.NewServeMux()
	mux.HandleFunc("/props", jsonResponse(map[string]any{
		"default_generation_settings": map[string]any{
			"id":    0,
			"n_ctx": integrationServedCtx,
		},
		"total_slots":   1,
		"chat_template": integrationQwen3Template,
		"modalities":    map[string]any{"vision": true, "video": false, "audio": false},
		"chat_template_caps": map[string]any{
			"supports_tool_calls":          true,
			"supports_parallel_tool_calls": false,
		},
		"build_info": "b11177-abcdef0",
	}))
	mux.HandleFunc("/v1/models", jsonResponse(map[string]any{
		"object": "list",
		"data": []map[string]any{{
			"id":       "qwen3-32b",
			"object":   "model",
			"owned_by": "llamacpp",
			"meta": map[string]any{
				"n_ctx_train": integrationTrainedCtx,
				"n_params":    32000000000,
			},
		}},
	}))
	mux.HandleFunc("/v1/chat/completions", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.NotFound(w, r)
			return
		}
		var req openai.ChatCompletionRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || len(req.Messages) == 0 {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		var x, y int
		if _, err := fmt.Sscanf(req.Messages[0].Content, "What is %d+%d?", &x, &y); err != nil {
			http.Error(w, "not the arithmetic probe", http.StatusBadRequest)
			return
		}
		jsonResponse(openai.ChatCompletionResponse{
			ID:     "chatcmpl-health",
			Object: "chat.completion",
			Model:  req.Model,
			Choices: []openai.ChatCompletionChoice{{
				Message:      openai.ChatCompletionMessage{Role: openai.ChatMessageRoleAssistant, Content: fmt.Sprintf("%d", x+y)},
				FinishReason: openai.FinishReasonStop,
			}},
		})(w, r)
	})

	b.srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b.mu.Lock()
		b.counts[r.Method+" "+r.URL.Path]++
		b.mu.Unlock()
		mux.ServeHTTP(w, r)
	}))
	t.Cleanup(b.srv.Close)
	return b
}

func jsonResponse(v any) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(v)
	}
}

func (b *llamaCppBackend) chatURL() string { return b.srv.URL + "/v1/chat/completions" }

func (b *llamaCppBackend) count(key string) int {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.counts[key]
}

// detectionRequests is every request that is not the health probe: the
// engine fingerprints (including the 404s of probes that run before the
// llama.cpp one) and the two llama.cpp fact endpoints.
func (b *llamaCppBackend) detectionRequests() int {
	b.mu.Lock()
	defer b.mu.Unlock()
	n := 0
	for key, c := range b.counts {
		if key != "POST /v1/chat/completions" {
			n += c
		}
	}
	return n
}

func (b *llamaCppBackend) adapter() *aiengine.OpenAI {
	return aiengine.NewOpenAIEngine("qwen3-32b", b.chatURL(), "", 0, lib.NewTestLogger(), nil)
}

// newIntegrationChecker wires the real detector (or none) in front of the
// backend for one model; the TTL is an hour so the second sweep of every
// test is inside it.
func newIntegrationChecker(backend *llamaCppBackend, cfg config.ModelConfig, hasBid bool, detect bool) *Checker {
	deps := &mockDeps{
		tags:     map[common.Hash][]string{modelLLM: {"llm"}},
		names:    map[common.Hash]string{modelLLM: "public-qwen3"},
		modelIDs: []common.Hash{modelLLM},
		adapter:  backend.adapter(),
		configs:  map[common.Hash]config.ModelConfig{modelLLM: cfg},
	}
	if hasBid {
		deps.bids = []*structs.Bid{bidFor(modelLLM)}
	}
	var det ApiDetector
	if detect {
		opts := apidetect.DefaultOptions()
		opts.CacheTTL = time.Hour
		det = apidetect.NewDetector(lib.NewTestLogger(), opts)
	}
	return NewChecker(Deps{Adapters: deps, Bids: deps, Models: deps, ModelConfigs: deps, ApiDetect: det}, time.Hour, 5*time.Second, 0, 0, lib.NewTestLogger())
}

func requireLlamaCppEvidence(t *testing.T, api *system.ModelApiSpec) {
	t.Helper()
	require.Equal(t, "llamacpp:/props.default_generation_settings.n_ctx", api.Evidence[system.EvidenceKeyCapability("context")])
	require.Equal(t, "llamacpp:/v1/models.meta.n_ctx_train", api.Evidence[system.EvidenceKeyCapability("contextTrained")])
	require.Equal(t, "llamacpp:/props.modalities.vision", api.Evidence[system.EvidenceKeyCapability("modalities.input")])
	require.Equal(t, "llamacpp:/props.chat_template_caps.supports_tool_calls", api.Evidence[system.EvidenceKeyCapability("tools")])
}

func TestIntegrationCheckerDetectsLlamaCppBackend(t *testing.T) {
	backend := newLlamaCppBackend(t)
	cfg := config.ModelConfig{ModelName: "qwen3-32b", ApiType: "openai", ApiURL: backend.chatURL()}
	checker := newIntegrationChecker(backend, cfg, true, true)

	checker.checkAll(context.Background(), common.Address{})

	report := reportByID(t, checker.GetReports(), modelLLM)
	require.Equal(t, system.ModelHealthStatusHealthy, report.Status, "the real adapter must reach the counted backend")
	require.NotNil(t, report.PromptCorrect)
	require.True(t, *report.PromptCorrect)
	require.Equal(t, "public-qwen3", report.ModelName)

	api := report.Api
	require.NotNil(t, api)
	require.Equal(t, system.ApiSpecSourceDetected, api.Source)
	require.Equal(t, "llamacpp", api.Stack)
	require.Empty(t, api.Via)
	require.Equal(t, "qwen3", api.ModelFamily, "family comes from the configured name")
	require.NotZero(t, api.DeclaredAt)

	require.NotNil(t, api.Thinking)
	require.Equal(t, system.ThinkingModeControllable, api.Thinking.Mode)
	for intent, value := range map[string]bool{system.IntentReasoningDisable: false, system.IntentReasoningEnable: true} {
		b := api.Bindings[intent]
		require.NotNil(t, b, intent)
		require.Equal(t, system.BindingKindTemplateKwarg, b.Kind, intent)
		require.Equal(t, "chat_template_kwargs.enable_thinking", b.Param, intent)
		require.Equal(t, "boolean", b.ParamType, intent)
		require.Equal(t, value, b.Value, intent)
	}

	caps := api.Capabilities
	require.NotNil(t, caps)
	require.NotNil(t, caps.Context)
	require.Equal(t, integrationServedCtx, *caps.Context)
	require.NotNil(t, caps.ContextTrained)
	require.Equal(t, integrationTrainedCtx, *caps.ContextTrained)
	require.NotNil(t, caps.Modalities)
	require.Equal(t, []string{"text", "image"}, caps.Modalities.Input, "vision true lists image; text is implied for a chat model (amendment b)")
	require.NotNil(t, caps.Tools)
	require.True(t, *caps.Tools)
	require.NotNil(t, caps.ToolsParallel)
	require.False(t, *caps.ToolsParallel)
	requireLlamaCppEvidence(t, api)
	for key, source := range api.Evidence {
		require.False(t, strings.HasPrefix(key, "bindings."), "template-table bindings carry no backend provenance: %s=%s", key, source)
	}

	// Detection reached the backend: both llama.cpp fact endpoints once.
	require.Equal(t, 1, backend.count("GET /props"))
	require.Equal(t, 1, backend.count("GET /v1/models"))
	require.Equal(t, 1, backend.count("POST /v1/chat/completions"))
	detectionAfterFirst := backend.detectionRequests()
	require.Positive(t, detectionAfterFirst)

	// Second sweep within the TTL: the detector answers from its cache, so
	// no detection request is repeated; only the health probe goes out again.
	checker.checkAll(context.Background(), common.Address{})
	second := reportByID(t, checker.GetReports(), modelLLM)
	require.Equal(t, detectionAfterFirst, backend.detectionRequests(), "cached detection must not touch the backend again")
	require.Equal(t, 1, backend.count("GET /props"))
	require.Equal(t, 1, backend.count("GET /v1/models"))
	require.Equal(t, 2, backend.count("POST /v1/chat/completions"))
	require.NotNil(t, second.Api)
	require.Equal(t, api.DeclaredAt, second.Api.DeclaredAt, "an unchanged spec keeps its DeclaredAt")
	require.Equal(t, "llamacpp", second.Api.Stack)
	require.Equal(t, integrationServedCtx, *second.Api.Capabilities.Context)
}

func TestIntegrationDeclaredContextWinsOverDetected(t *testing.T) {
	backend := newLlamaCppBackend(t)
	cfg := config.ModelConfig{ModelName: "qwen3-32b", ApiType: "openai", ApiURL: backend.chatURL(), Capabilities: declaredContext(4096)}
	checker := newIntegrationChecker(backend, cfg, true, true)

	checker.checkAll(context.Background(), common.Address{})

	api := reportByID(t, checker.GetReports(), modelLLM).Api
	require.NotNil(t, api)
	require.Equal(t, system.ApiSpecSourceDetected, api.Source, "a declared block does not turn detection off")
	require.Equal(t, "llamacpp", api.Stack)

	caps := api.Capabilities
	require.NotNil(t, caps)
	require.NotNil(t, caps.Context)
	require.Equal(t, int64(4096), *caps.Context, "declared wins field by field")
	require.Equal(t, "declared", api.Evidence[system.EvidenceKeyCapability("context")])
	require.NotNil(t, caps.ContextTrained)
	require.Equal(t, integrationTrainedCtx, *caps.ContextTrained, "a field the block does not name keeps the backend fact")
	require.Equal(t, "llamacpp:/v1/models.meta.n_ctx_train", api.Evidence[system.EvidenceKeyCapability("contextTrained")])
	require.Equal(t, []string{"text", "image"}, caps.Modalities.Input)
	require.True(t, *caps.Tools)
	require.Equal(t, 1, backend.count("GET /props"))
}

func TestIntegrationDetectionOffPublishesDeclaredOnly(t *testing.T) {
	backend := newLlamaCppBackend(t)
	cfg := config.ModelConfig{ModelName: "qwen3-32b", ApiType: "openai", ApiURL: backend.chatURL(), Capabilities: declaredContext(4096)}
	// MODEL_API_DETECT_ENABLED=false: cmd/main.go leaves Deps.ApiDetect nil.
	checker := newIntegrationChecker(backend, cfg, true, false)

	checker.checkAll(context.Background(), common.Address{})
	checker.checkAll(context.Background(), common.Address{})

	report := reportByID(t, checker.GetReports(), modelLLM)
	require.Equal(t, system.ModelHealthStatusHealthy, report.Status)
	api := report.Api
	require.NotNil(t, api, "the declared block is published without a detector")
	require.Equal(t, system.ApiSpecSourceDeclared, api.Source)
	require.Empty(t, api.Stack)
	require.Nil(t, api.Bindings)
	require.Nil(t, api.Thinking)
	require.Equal(t, "qwen3", api.ModelFamily, "the family is read off the configured name, not the backend; without a stack it carries no bindings")
	require.NotNil(t, api.Capabilities)
	require.Equal(t, int64(4096), *api.Capabilities.Context)
	require.Nil(t, api.Capabilities.ContextTrained)
	require.Nil(t, api.Capabilities.Modalities)
	require.Nil(t, api.Capabilities.Tools)
	require.Equal(t, map[string]string{system.EvidenceKeyCapability("context"): "declared"}, api.Evidence)

	require.Zero(t, backend.detectionRequests(), "no detector, no detection request")
	require.Equal(t, 2, backend.count("POST /v1/chat/completions"), "the health probe still runs every sweep")
}

func TestIntegrationTransportMismatchKeepsDeclaredOnly(t *testing.T) {
	backend := newLlamaCppBackend(t)
	// apiType claudeai on an OpenAI-transport engine: the stack the probe
	// fingerprinted cannot be what this model is spoken to with.
	cfg := config.ModelConfig{ModelName: "qwen3-32b", ApiType: "claudeai", ApiURL: backend.chatURL(), Capabilities: declaredContext(4096)}
	checker := newIntegrationChecker(backend, cfg, false, true)

	checker.checkAll(context.Background(), common.Address{})

	report := reportByID(t, checker.GetReports(), modelLLM)
	require.Equal(t, system.ModelHealthStatusNoBid, report.Status)
	api := report.Api
	require.NotNil(t, api)
	require.Equal(t, 1, backend.count("GET /props"), "the engine was fingerprinted before the drop")

	require.Empty(t, api.Stack)
	require.Empty(t, api.Via)
	require.Nil(t, api.Bindings)
	require.Nil(t, api.Parameters)
	require.Nil(t, api.Thinking, "controllable thinking goes with the bindings that control it")
	require.Equal(t, "qwen3", api.ModelFamily, "family describes the model, not the wire")

	caps := api.Capabilities
	require.NotNil(t, caps)
	require.NotNil(t, caps.Context)
	require.Equal(t, int64(4096), *caps.Context)
	require.Nil(t, caps.ContextTrained, "read from the dropped stack's endpoint")
	require.Nil(t, caps.Modalities)
	require.Nil(t, caps.Tools)
	require.Nil(t, caps.ToolsParallel)
	require.Nil(t, caps.Structured)
	require.Equal(t, "declared", api.Evidence[system.EvidenceKeyCapability("context")])
	for key, source := range api.Evidence {
		require.False(t, strings.HasPrefix(source, "llamacpp:"), "no evidence may survive from the dropped stack: %s=%s", key, source)
		require.False(t, strings.HasPrefix(key, "bindings."), "binding evidence goes with the bindings: %s=%s", key, source)
	}

	// The drop is cached like any result: the second sweep repeats nothing.
	detectionAfterFirst := backend.detectionRequests()
	checker.checkAll(context.Background(), common.Address{})
	require.Equal(t, detectionAfterFirst, backend.detectionRequests())
	require.Zero(t, backend.count("POST /v1/chat/completions"), "no bid, no health probe")
}
