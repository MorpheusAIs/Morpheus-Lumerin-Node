package apidetect

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/apispec"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/config"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/system"
	"github.com/stretchr/testify/require"
)

const vendorTestKey = "sk-vendor-test-key"

func int64p(v int64) *int64 { return &v }
func boolp(v bool) *bool    { return &v }

// vendorFixture serves one metadata path and records what reached it.
type vendorFixture struct {
	path    string
	body    any
	hits    atomic.Int32
	headers atomic.Pointer[http.Header]
	rawPath atomic.Pointer[string]
}

func (f *vendorFixture) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	f.hits.Add(1)
	h := r.Header.Clone()
	f.headers.Store(&h)
	raw := r.URL.EscapedPath()
	f.rawPath.Store(&raw)
	if r.URL.EscapedPath() != f.path {
		http.NotFound(w, r)
		return
	}
	jsonHandler(f.body)(w, r)
}

func (f *vendorFixture) header(name string) string {
	h := f.headers.Load()
	if h == nil {
		return ""
	}
	return h.Get(name)
}

func serveVendor(t *testing.T, vendor, path string, body any) (*httptest.Server, *vendorFixture) {
	t.Helper()
	f := &vendorFixture{path: path, body: body}
	srv := httptest.NewServer(f)
	t.Cleanup(srv.Close)
	mapHostToVendor(t, srv.URL, vendor)
	return srv, f
}

func detectVendor(t *testing.T, srv *httptest.Server, cfg config.ModelConfig) (*system.ModelApiSpec, []string) {
	t.Helper()
	if cfg.ApiType == "" {
		cfg.ApiType = "openai"
	}
	if cfg.ApiURL == "" {
		cfg.ApiURL = srv.URL + "/v1/chat/completions"
	}
	return newTestDetector().DetectWithTrace(context.Background(), cfg)
}

func supported(v bool) map[string]any { return map[string]any{"supported": v} }

// https://platform.claude.com/docs/en/api/models/retrieve
var anthropicModelFixture = map[string]any{
	"id": "claude-sonnet-4-5", "type": "model", "display_name": "Claude Sonnet 4.5", "created_at": "2025-09-29T00:00:00Z",
	"max_input_tokens": 200000, "max_tokens": 64000,
	"capabilities": map[string]any{
		"batch":              supported(true),
		"citations":          supported(true),
		"code_execution":     supported(true),
		"context_management": map[string]any{"supported": true},
		"effort": map[string]any{
			"low": supported(true), "medium": supported(true), "high": supported(true), "xhigh": supported(false), "max": supported(true),
		},
		"image_input":        supported(true),
		"pdf_input":          supported(true),
		"structured_outputs": supported(true),
		"thinking": map[string]any{
			"supported": true,
			"types":     map[string]any{"adaptive": supported(true), "enabled": supported(true)},
		},
	},
}

func TestProbeAnthropic(t *testing.T) {
	srv, fx := serveVendor(t, "anthropic", "/v1/models/claude-sonnet-4-5", anthropicModelFixture)
	api, trace := detectVendor(t, srv, config.ModelConfig{ModelName: "claude-sonnet-4-5", ApiType: "claudeai", ApiURL: srv.URL + "/v1/messages", ApiKey: vendorTestKey})

	require.Equal(t, int32(1), fx.hits.Load())
	require.Equal(t, vendorTestKey, fx.header("x-api-key"))
	require.Equal(t, "2023-06-01", fx.header("anthropic-version"))
	require.Empty(t, fx.header("Authorization"), "anthropic authenticates with x-api-key, never a bearer")

	require.NotNil(t, api)
	require.Equal(t, "anthropic", api.Stack)
	require.NotNil(t, api.Capabilities)
	require.Equal(t, int64p(200000), api.Capabilities.Context)
	require.Equal(t, int64p(64000), api.Capabilities.MaxOutput)
	require.Equal(t, []string{"image", "pdf"}, api.Capabilities.Modalities.Input)
	require.Equal(t, []string{"json_schema"}, api.Capabilities.Structured)
	require.Equal(t, boolp(true), api.Capabilities.Reasoning)
	require.Equal(t, "anthropic:/v1/models/{id}.max_input_tokens", api.Evidence["capabilities.context"])
	require.Equal(t, "anthropic:/v1/models/{id}.max_tokens", api.Evidence["capabilities.maxOutput"])
	require.Equal(t, "anthropic:/v1/models/{id}.capabilities.pdf_input", api.Evidence["capabilities.modalities.input"])
	require.Equal(t, "anthropic:/v1/models/{id}.capabilities", api.Evidence["capabilities.structured"])
	require.Equal(t, "anthropic:/v1/models/{id}.capabilities", api.Evidence["bindings.reasoning.effort"])

	require.Equal(t, map[string]any{"type": "adaptive"}, api.Bindings[system.IntentReasoningEnable].Value)
	require.Equal(t, map[string]any{"type": "disabled"}, api.Bindings[system.IntentReasoningDisable].Value)
	require.Equal(t, []string{"low", "medium", "high", "max"}, api.Bindings[system.IntentReasoningEffort].EnumValues)
	require.Equal(t, system.ThinkingModeControllable, api.Thinking.Mode)

	require.Contains(t, trace, "host recognized as anthropic by hostname; asking /v1/models/{id} for the model's limits and capabilities")
	require.Contains(t, trace, "anthropic:/v1/models/{id}.max_input_tokens = 200000 -> capabilities.context")
	require.Contains(t, trace, "anthropic:/v1/models/{id}.capabilities: thinking.supported=true thinking.types=[adaptive enabled] effort=[low medium high max] structured_outputs=true -> anthropic facts")
	for _, line := range trace {
		require.NotContains(t, line, vendorTestKey)
	}
}

func TestProbeAnthropicFacts(t *testing.T) {
	srv, _ := serveVendor(t, "anthropic", "/v1/models/claude-x", map[string]any{
		"id": "claude-x", "max_input_tokens": 0,
		"capabilities": map[string]any{
			"effort":             map[string]any{"high": supported(true)},
			"structured_outputs": supported(false),
			"thinking":           map[string]any{"supported": true, "types": map[string]any{"enabled": supported(true), "adaptive": supported(false)}},
		},
	})
	ctx, _ := withTrace(context.Background())
	ev := apispec.Evidence{Stack: "anthropic", ModelName: "claude-x"}
	probeAnthropic(ctx, newTestDetector(), nil, config.ModelConfig{ModelName: "claude-x", ApiURL: srv.URL + "/v1/messages", ApiKey: vendorTestKey}, &ev)

	require.Nil(t, ev.Caps.Context, "the documented placeholder max_input_tokens: 0 is not a window")
	require.Equal(t, &apispec.AnthropicFacts{
		ThinkingSupported: true,
		ThinkingTypes:     []string{"enabled"},
		EffortLevels:      []string{"high"},
		Structured:        false,
		Source:            "anthropic:/v1/models/{id}.capabilities",
	}, ev.Anthropic)
}

// A capabilities object without the thinking leaf is not a statement about
// thinking: the name-based claude bindings stay, the modality facts land.
func TestProbeAnthropicWithoutThinkingLeafKeepsFallback(t *testing.T) {
	withoutRegistry(t)
	srv, _ := serveVendor(t, "anthropic", "/v1/models/claude-sonnet-4-5", map[string]any{
		"id": "claude-sonnet-4-5", "max_input_tokens": 200000,
		"capabilities": map[string]any{
			"image_input":        supported(true),
			"structured_outputs": supported(true),
			"effort":             map[string]any{"high": supported(true)},
		},
	})
	cfg := config.ModelConfig{ModelName: "claude-sonnet-4-5", ApiType: "claudeai", ApiURL: srv.URL + "/v1/messages", ApiKey: vendorTestKey}

	ctx, tr := withTrace(context.Background())
	ev := apispec.Evidence{Stack: "anthropic", ModelName: cfg.ModelName}
	probeAnthropic(ctx, newTestDetector(), nil, cfg, &ev)
	require.Nil(t, ev.Anthropic)
	require.Equal(t, []string{"image"}, ev.Caps.Modalities.Input)
	require.Contains(t, tr.lines, "anthropic:/v1/models/{id}.capabilities has no thinking object; anthropic facts not stated")

	api, _ := detectVendor(t, srv, cfg)
	require.NotNil(t, api)
	require.Equal(t, int64p(200000), api.Capabilities.Context)
	require.Equal(t, []string{"image"}, api.Capabilities.Modalities.Input)
	require.Nil(t, api.Capabilities.Structured, "structured_outputs is not published without the thinking leaf")
	require.Equal(t, map[string]any{"type": "disabled"}, api.Bindings[system.IntentReasoningDisable].Value, "the name-based fallback still decides the bindings")
	require.Empty(t, api.Evidence["bindings.reasoning.effort"])
}

func TestProbeGemini(t *testing.T) {
	// https://ai.google.dev/api/models#Model
	srv, fx := serveVendor(t, "gemini", "/v1beta/models/gemini-2.5-flash", map[string]any{
		"name": "models/gemini-2.5-flash", "version": "001", "displayName": "Gemini 2.5 Flash",
		"inputTokenLimit": 1048576, "outputTokenLimit": 65536,
		"supportedGenerationMethods": []string{"generateContent", "countTokens"},
		"thinking":                   true,
	})
	api, trace := detectVendor(t, srv, config.ModelConfig{ModelName: "models/gemini-2.5-flash", ApiURL: srv.URL + "/v1beta/openai/chat/completions", ApiKey: vendorTestKey})

	require.Equal(t, vendorTestKey, fx.header("x-goog-api-key"))
	require.Empty(t, fx.header("Authorization"))
	require.NotNil(t, api)
	require.Equal(t, "gemini", api.Stack)
	require.Equal(t, int64p(1048576), api.Capabilities.Context)
	require.Equal(t, int64p(65536), api.Capabilities.MaxOutput)
	require.Equal(t, boolp(true), api.Capabilities.Reasoning)
	require.Equal(t, "gemini:/v1beta/models/{id}.inputTokenLimit", api.Evidence["capabilities.context"])
	require.Equal(t, "gemini:/v1beta/models/{id}.outputTokenLimit", api.Evidence["capabilities.maxOutput"])
	require.Equal(t, "gemini:/v1beta/models/{id}.thinking", api.Evidence["capabilities.reasoning"])
	require.Contains(t, trace, "gemini:/v1beta/models/{id}.thinking = true -> capabilities.reasoning")
}

func TestProbeMistral(t *testing.T) {
	// https://github.com/mistralai/client-python/blob/main/src/mistralai/client/models/modelcapabilities.py
	srv, fx := serveVendor(t, "mistral", "/v1/models/mistral-large-latest", map[string]any{
		"id": "mistral-large-latest", "object": "model", "max_context_length": 131072,
		"capabilities": map[string]any{
			"completion_chat": true, "completion_fim": false, "function_calling": true, "fine_tuning": false,
			"vision": true, "ocr": false, "classification": false, "reasoning": false, "audio": false,
		},
	})
	api, _ := detectVendor(t, srv, config.ModelConfig{ModelName: "mistral-large-latest", ApiKey: vendorTestKey})

	require.Equal(t, "Bearer "+vendorTestKey, fx.header("Authorization"))
	require.NotNil(t, api)
	require.Equal(t, "mistral", api.Stack)
	require.Equal(t, int64p(131072), api.Capabilities.Context)
	require.Equal(t, boolp(true), api.Capabilities.Tools)
	require.Equal(t, boolp(false), api.Capabilities.Reasoning, "a vendor-reported false is published")
	require.Equal(t, []string{"image"}, api.Capabilities.Modalities.Input)
	require.Equal(t, "mistral:/v1/models/{id}.max_context_length", api.Evidence["capabilities.context"])
	require.Equal(t, "mistral:/v1/models/{id}.capabilities.function_calling", api.Evidence["capabilities.tools"])
	require.Equal(t, "mistral:/v1/models/{id}.capabilities.reasoning", api.Evidence["capabilities.reasoning"])
	require.Equal(t, "mistral:/v1/models/{id}.capabilities.vision", api.Evidence["capabilities.modalities.input"])
}

func TestProbeGroq(t *testing.T) {
	// https://console.groq.com/docs/api-reference#models-retrieve
	srv, fx := serveVendor(t, "groq", "/openai/v1/models/llama-3.3-70b-versatile", map[string]any{
		"id": "llama-3.3-70b-versatile", "object": "model", "owned_by": "Meta", "active": true,
		"context_window": 131072, "max_completion_tokens": 32768,
	})
	api, _ := detectVendor(t, srv, config.ModelConfig{ModelName: "llama-3.3-70b-versatile", ApiURL: srv.URL + "/openai/v1/chat/completions", ApiKey: vendorTestKey})

	require.Equal(t, "Bearer "+vendorTestKey, fx.header("Authorization"))
	require.Equal(t, "/openai/v1/models/llama-3.3-70b-versatile", *fx.rawPath.Load(), "the documented prefix, not the configured path, roots the lookup")
	require.NotNil(t, api)
	require.Equal(t, "groq", api.Stack)
	require.Equal(t, int64p(131072), api.Capabilities.Context)
	require.Equal(t, int64p(32768), api.Capabilities.MaxOutput)
	require.Equal(t, "groq:/openai/v1/models/{id}.context_window", api.Evidence["capabilities.context"])
	require.Equal(t, "groq:/openai/v1/models/{id}.max_completion_tokens", api.Evidence["capabilities.maxOutput"])
}

func TestProbeXAI(t *testing.T) {
	// https://docs.x.ai/developers/rest-api-reference/inference/models
	srv, fx := serveVendor(t, "xai", "/v1/language-models/grok-4", map[string]any{
		"id": "grok-4", "object": "model", "version": "1.0.0",
		"input_modalities": []string{"text", "image"}, "output_modalities": []string{"text"},
		"capabilities":            map[string]any{"reasoning_effort": []string{"low", "high"}, "default_reasoning_effort": "high"},
		"prompt_text_token_price": 200000, "search_price": 2500,
	})
	cfg := config.ModelConfig{ModelName: "grok-4", ApiURL: srv.URL + "/v1/chat/completions", ApiKey: vendorTestKey}
	api, _ := detectVendor(t, srv, cfg)

	require.Equal(t, "Bearer "+vendorTestKey, fx.header("Authorization"))
	require.NotNil(t, api)
	require.Equal(t, "xai", api.Stack)
	require.Equal(t, []string{"text", "image"}, api.Capabilities.Modalities.Input)
	require.Equal(t, []string{"text"}, api.Capabilities.Modalities.Output)
	require.Equal(t, "xai:/v1/language-models/{id}.input_modalities", api.Evidence["capabilities.modalities.input"])
	require.Equal(t, "xai:/v1/language-models/{id}.output_modalities", api.Evidence["capabilities.modalities.output"])

	ctx, tr := withTrace(context.Background())
	ev := apispec.Evidence{Stack: "xai", ModelName: cfg.ModelName}
	probeXAI(ctx, newTestDetector(), nil, cfg, &ev)
	require.Equal(t, []string{"low", "high"}, ev.ReasoningEfforts)
	require.Equal(t, "xai:/v1/language-models/{id}.capabilities.reasoning_effort", ev.EffortsSource)
	require.Equal(t, "high", ev.ThinkingDefault)
	require.Equal(t, "xai:/v1/language-models/{id}.capabilities.default_reasoning_effort", ev.ThinkingDefaultSource)
	require.Contains(t, tr.lines, "xai:/v1/language-models/{id}.capabilities.reasoning_effort = [low high] -> reasoning efforts")
	require.Contains(t, tr.lines, "xai:/v1/language-models/{id}.capabilities.default_reasoning_effort = high -> thinking.default")
}

func TestProbeTogether(t *testing.T) {
	// https://docs.together.ai/reference/models-1 — a bare array.
	srv, fx := serveVendor(t, "together", "/v1/models", []map[string]any{
		{"id": "mistralai/Mixtral-8x7B-Instruct-v0.1", "object": "model", "type": "chat", "context_length": 32768},
		{"id": "meta-llama/Llama-3.3-70B-Instruct-Turbo", "object": "model", "type": "chat", "context_length": 131072, "pricing": map[string]any{"input": 0.88}},
	})
	api, _ := detectVendor(t, srv, config.ModelConfig{ModelName: "meta-llama/Llama-3.3-70B-Instruct-Turbo", ApiKey: vendorTestKey})

	require.Equal(t, "Bearer "+vendorTestKey, fx.header("Authorization"))
	require.NotNil(t, api)
	require.Equal(t, "together", api.Stack)
	require.Equal(t, int64p(131072), api.Capabilities.Context)
	require.Equal(t, "together:/v1/models.context_length", api.Evidence["capabilities.context"])
}

// https://inference-docs.cerebras.ai/api-reference/models/public-models
var cerebrasModelFixture = map[string]any{
	"id": "llama-3.3-70b", "object": "model", "owned_by": "Meta", "hugging_face_id": "meta-llama/Llama-3.3-70B-Instruct", "quantization": "bf16",
	"capabilities": map[string]any{
		"function_calling": true, "structured_outputs": true, "vision": false, "json_mode": true,
		"tools": true, "tool_choice": true, "parallel_tool_calls": false, "response_format": true, "reasoning": false,
	},
	"limits": map[string]any{"max_context_length": 65536, "max_completion_tokens": 8192},
}

func TestProbeCerebras(t *testing.T) {
	withoutRegistry(t)
	srv, fx := serveVendor(t, "cerebras", "/public/v1/models/llama-3.3-70b", cerebrasModelFixture)
	api, trace := detectVendor(t, srv, config.ModelConfig{ModelName: "llama-3.3-70b", ApiKey: vendorTestKey})

	require.Equal(t, int32(1), fx.hits.Load())
	require.Empty(t, fx.header("Authorization"), "the public endpoint never sees the key")
	require.NotNil(t, api)
	require.Equal(t, "cerebras", api.Stack)
	require.Equal(t, &system.ModelCapabilities{
		Context: int64p(65536), MaxOutput: int64p(8192),
		Tools: boolp(true), ToolsParallel: boolp(false),
		Structured:   []string{"json_object", "json_schema"},
		Reasoning:    boolp(false),
		Quantization: "bf16",
	}, api.Capabilities)
	require.Equal(t, map[string]string{
		"capabilities.context":       "cerebras:/public/v1/models/{id}.limits.max_context_length",
		"capabilities.maxOutput":     "cerebras:/public/v1/models/{id}.limits.max_completion_tokens",
		"capabilities.tools":         "cerebras:/public/v1/models/{id}.capabilities.function_calling",
		"capabilities.toolsParallel": "cerebras:/public/v1/models/{id}.capabilities.parallel_tool_calls",
		"capabilities.structured":    "cerebras:/public/v1/models/{id}.capabilities.structured_outputs",
		"capabilities.reasoning":     "cerebras:/public/v1/models/{id}.capabilities.reasoning",
		"capabilities.quantization":  "cerebras:/public/v1/models/{id}.quantization",
	}, api.Evidence)
	require.Contains(t, trace, "host recognized as cerebras by hostname; asking the public /public/v1/models/{id} without the key")
}

func TestKeyFreeVendorsNeverRetryWithTheKey(t *testing.T) {
	for _, vendor := range []string{"cerebras", "deepinfra", "hf-router"} {
		t.Run(vendor, func(t *testing.T) {
			var hits atomic.Int32
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				hits.Add(1)
				require.Empty(t, r.Header.Get("Authorization"))
				require.Empty(t, r.Header.Get("x-api-key"))
				w.WriteHeader(http.StatusUnauthorized)
			}))
			defer srv.Close()
			mapHostToVendor(t, srv.URL, vendor)

			api, trace := detectVendor(t, srv, config.ModelConfig{ModelName: "m", ApiKey: vendorTestKey})
			require.Equal(t, int32(1), hits.Load(), "a 401 on an open endpoint is a miss, not a prompt for the key")
			require.NotNil(t, api)
			require.Equal(t, vendor, api.Stack)
			require.Nil(t, api.Capabilities)
			require.Contains(t, trace, "GET "+srv.URL+vendorMetadataPath(vendor)+" -> HTTP 401")
		})
	}
}

func vendorMetadataPath(vendor string) string {
	switch vendor {
	case "cerebras":
		return "/public/v1/models/m"
	case "deepinfra":
		return "/v1/openai/models"
	default:
		return "/v1/models"
	}
}

func TestProbeDeepInfra(t *testing.T) {
	srv, fx := serveVendor(t, "deepinfra", "/v1/openai/models", map[string]any{
		"object": "list",
		"data": []map[string]any{
			{"id": "other/model", "object": "model", "metadata": map[string]any{"context_length": 4096}},
			{"id": "Qwen/Qwen3-32B", "object": "model", "metadata": map[string]any{"context_length": 40960, "max_tokens": 16384, "tags": []string{"text-generation"}}},
		},
	})
	api, _ := detectVendor(t, srv, config.ModelConfig{ModelName: "Qwen/Qwen3-32B", ApiURL: srv.URL + "/v1/openai/chat/completions", ApiKey: vendorTestKey})

	require.Empty(t, fx.header("Authorization"))
	require.NotNil(t, api)
	require.Equal(t, "deepinfra", api.Stack)
	require.Equal(t, int64p(40960), api.Capabilities.Context)
	require.Equal(t, int64p(16384), api.Capabilities.MaxOutput)
	require.Equal(t, "deepinfra:/v1/openai/models.metadata.context_length", api.Evidence["capabilities.context"])
	require.Equal(t, "deepinfra:/v1/openai/models.metadata.max_tokens", api.Evidence["capabilities.maxOutput"])
}

func TestProbeHFRouter(t *testing.T) {
	withoutRegistry(t)
	// https://huggingface.co/docs/inference-providers/index — live shape of
	// GET https://router.huggingface.co/v1/models.
	srv, fx := serveVendor(t, "hf-router", "/v1/models", map[string]any{
		"object": "list",
		"data": []map[string]any{{
			"id": "Qwen/Qwen3-32B", "object": "model", "created": 1745000000, "owned_by": "Qwen",
			"architecture": map[string]any{"input_modalities": []string{"text"}, "output_modalities": []string{"text"}},
			"providers": []map[string]any{
				{"provider": "a", "status": "live", "context_length": 32768, "supports_tools": true, "supports_structured_output": false, "pricing": map[string]any{"input": 0.1, "output": 0.3}},
				{"provider": "b", "status": "live", "context_length": 131072, "supports_tools": false, "supports_structured_output": true},
				{"provider": "c", "status": "error", "context_length": 1000000, "supports_tools": false, "supports_structured_output": false},
			},
		}},
	})
	api, trace := detectVendor(t, srv, config.ModelConfig{ModelName: "Qwen/Qwen3-32B", ApiKey: vendorTestKey})

	require.Empty(t, fx.header("Authorization"))
	require.NotNil(t, api)
	require.Equal(t, "hf-router", api.Stack)
	require.Equal(t, &system.ModelCapabilities{
		Context:    int64p(131072),
		Modalities: &system.Modalities{Input: []string{"text"}, Output: []string{"text"}},
		Tools:      boolp(true),
		Structured: []string{"json_schema"},
	}, api.Capabilities)
	require.Equal(t, map[string]string{
		"capabilities.context":           "hf-router:/v1/models.providers.context_length",
		"capabilities.modalities.input":  "hf-router:/v1/models.architecture.input_modalities",
		"capabilities.modalities.output": "hf-router:/v1/models.architecture.output_modalities",
		"capabilities.tools":             "hf-router:/v1/models.providers.supports_tools",
		"capabilities.structured":        "hf-router:/v1/models.providers.supports_structured_output",
	}, api.Evidence)
	require.Contains(t, trace, "hf-router:/v1/models.providers.context_length = 131072 -> capabilities.context (largest live provider)")
}

func TestProbeHFRouterNoLiveProvider(t *testing.T) {
	srv, _ := serveVendor(t, "hf-router", "/v1/models", map[string]any{
		"data": []map[string]any{{
			"id":        "org/m",
			"providers": []map[string]any{{"provider": "a", "status": "error", "context_length": 8192, "supports_tools": true}},
		}},
	})
	api, trace := detectVendor(t, srv, config.ModelConfig{ModelName: "org/m"})
	require.NotNil(t, api)
	require.Nil(t, api.Capabilities)
	require.Contains(t, trace, "/v1/models providers[]: none of 1 is live; no provider facts")
}

func TestVendorMisses(t *testing.T) {
	t.Run("404 is a silent miss", func(t *testing.T) {
		srv, fx := serveVendor(t, "mistral", "/v1/models/other", map[string]any{"id": "other"})
		api, trace := detectVendor(t, srv, config.ModelConfig{ModelName: "gone", ApiKey: vendorTestKey})
		require.Equal(t, int32(1), fx.hits.Load())
		require.NotNil(t, api)
		require.Equal(t, "mistral", api.Stack)
		require.Nil(t, api.Capabilities)
		require.Contains(t, trace, "GET "+srv.URL+"/v1/models/gone -> HTTP 404")
	})

	t.Run("non-JSON body is a silent miss", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/html")
			_, _ = w.Write([]byte("<html>maintenance</html>"))
		}))
		defer srv.Close()
		mapHostToVendor(t, srv.URL, "groq")
		api, trace := detectVendor(t, srv, config.ModelConfig{ModelName: "m", ApiKey: vendorTestKey})
		require.NotNil(t, api)
		require.Nil(t, api.Capabilities)
		require.Len(t, filterTrace(trace, "GET "+srv.URL+"/openai/v1/models/m -> HTTP 200, but not JSON"), 1)
	})

	t.Run("listing without the model is a miss", func(t *testing.T) {
		srv, _ := serveVendor(t, "together", "/v1/models", []map[string]any{{"id": "a", "context_length": 1}, {"id": "b", "context_length": 2}})
		api, trace := detectVendor(t, srv, config.ModelConfig{ModelName: "c", ApiKey: vendorTestKey})
		require.NotNil(t, api)
		require.Nil(t, api.Capabilities)
		require.Contains(t, trace, "/v1/models has 2 entries but none matches the configured model (config drift?)")
	})

	t.Run("a lone listing entry with another id is a miss", func(t *testing.T) {
		srv, _ := serveVendor(t, "together", "/v1/models", []map[string]any{{"id": "org/dedicated-endpoint", "context_length": 8192}})
		api, trace := detectVendor(t, srv, config.ModelConfig{ModelName: "org/configured", ApiKey: vendorTestKey})
		require.NotNil(t, api)
		require.Nil(t, api.Capabilities, "another model's context_length is not a fact about the configured one")
		require.Contains(t, trace, "/v1/models has 1 entries but none matches the configured model (config drift?)")
	})

	t.Run("listing entries are matched by id only", func(t *testing.T) {
		withoutRegistry(t)
		srv, _ := serveVendor(t, "deepinfra", "/v1/openai/models", map[string]any{
			"data": []map[string]any{{"id": "Qwen/Qwen3-32B", "metadata": map[string]any{"context_length": 40960}}},
		})
		api, _ := detectVendor(t, srv, config.ModelConfig{ModelName: "Qwen3-32B"})
		require.NotNil(t, api)
		require.Nil(t, api.Capabilities, "a bare suffix is not the configured id")
	})

	t.Run("wrong-type capability booleans are not facts", func(t *testing.T) {
		srv, _ := serveVendor(t, "mistral", "/v1/models/m", map[string]any{
			"id": "m", "max_context_length": 8192,
			"capabilities": map[string]any{"function_calling": "yes", "reasoning": 1, "vision": "true", "audio": nil},
		})
		api, _ := detectVendor(t, srv, config.ModelConfig{ModelName: "m", ApiKey: vendorTestKey})
		require.NotNil(t, api)
		require.Equal(t, &system.ModelCapabilities{Context: int64p(8192)}, api.Capabilities)
		require.Equal(t, map[string]string{"capabilities.context": "mistral:/v1/models/{id}.max_context_length"}, api.Evidence)
	})

	t.Run("gemini thinking false or non-bool leaves reasoning unset", func(t *testing.T) {
		for name, thinking := range map[string]any{"false": false, "string": "yes", "number": 1} {
			t.Run(name, func(t *testing.T) {
				srv, _ := serveVendor(t, "gemini", "/v1beta/models/g", map[string]any{"name": "models/g", "inputTokenLimit": 4096, "thinking": thinking})
				api, _ := detectVendor(t, srv, config.ModelConfig{ModelName: "g", ApiKey: vendorTestKey})
				require.NotNil(t, api)
				require.Equal(t, &system.ModelCapabilities{Context: int64p(4096)}, api.Capabilities)
				require.Empty(t, api.Evidence["capabilities.reasoning"])
			})
		}
	})

	t.Run("cerebras json_mode alone is json_object from its own field", func(t *testing.T) {
		srv, _ := serveVendor(t, "cerebras", "/public/v1/models/m", map[string]any{
			"id": "m", "capabilities": map[string]any{"json_mode": true, "structured_outputs": false},
		})
		api, trace := detectVendor(t, srv, config.ModelConfig{ModelName: "m"})
		require.NotNil(t, api)
		require.Equal(t, &system.ModelCapabilities{Structured: []string{"json_object"}}, api.Capabilities)
		require.Equal(t, "cerebras:/public/v1/models/{id}.capabilities.json_mode", api.Evidence["capabilities.structured"])
		require.Contains(t, trace, "cerebras:/public/v1/models/{id}.capabilities.json_mode = [json_object] -> capabilities.structured")
	})

	t.Run("together served as an OpenAI-style object listing", func(t *testing.T) {
		srv, _ := serveVendor(t, "together", "/v1/models", map[string]any{
			"object": "list", "data": []map[string]any{{"id": "org/m", "context_length": 32768}},
		})
		api, _ := detectVendor(t, srv, config.ModelConfig{ModelName: "org/m", ApiKey: vendorTestKey})
		require.NotNil(t, api)
		require.Equal(t, int64p(32768), api.Capabilities.Context)
		require.Equal(t, "together:/v1/models.context_length", api.Evidence["capabilities.context"])
	})

	t.Run("oversized body is ignored", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_, _ = fmt.Fprintf(w, `{"context_window": 8192, "pad": "%s"}`, strings.Repeat("a", int(maxProbeBody)+1024))
		}))
		defer srv.Close()
		mapHostToVendor(t, srv.URL, "groq")
		api, trace := detectVendor(t, srv, config.ModelConfig{ModelName: "m", ApiKey: vendorTestKey})
		require.NotNil(t, api)
		require.Nil(t, api.Capabilities)
		require.Contains(t, trace, fmt.Sprintf("GET %s/openai/v1/models/m -> HTTP 200, but the body exceeds the %d-byte limit; ignored", srv.URL, maxProbeBody))
	})

	t.Run("deepinfra entry without metadata is nothing", func(t *testing.T) {
		srv, _ := serveVendor(t, "deepinfra", "/v1/openai/models", map[string]any{
			"data": []map[string]any{{"id": "org/m", "object": "model", "context_length": 4096}},
		})
		api, _ := detectVendor(t, srv, config.ModelConfig{ModelName: "org/m"})
		require.NotNil(t, api)
		require.Nil(t, api.Capabilities, "context_length outside metadata is not the documented field")
	})

	t.Run("hf-router providers that are not objects and unknown modalities are skipped", func(t *testing.T) {
		srv, _ := serveVendor(t, "hf-router", "/v1/models", map[string]any{
			"data": []map[string]any{{
				"id":           "org/m",
				"architecture": map[string]any{"input_modalities": []string{"hologram"}, "output_modalities": []any{1, true}},
				"providers":    []any{"a", 1, nil, map[string]any{"provider": "b", "status": "live", "context_length": 2048}},
			}},
		})
		api, _ := detectVendor(t, srv, config.ModelConfig{ModelName: "org/m"})
		require.NotNil(t, api)
		require.Equal(t, &system.ModelCapabilities{Context: int64p(2048)}, api.Capabilities)
		require.Equal(t, map[string]string{"capabilities.context": "hf-router:/v1/models.providers.context_length"}, api.Evidence)
	})

	t.Run("no modelName is nothing to look up", func(t *testing.T) {
		srv, fx := serveVendor(t, "groq", "/openai/v1/models/x", map[string]any{})
		_, trace := detectVendor(t, srv, config.ModelConfig{ApiKey: vendorTestKey})
		require.Zero(t, fx.hits.Load())
		require.Contains(t, trace, "no modelName configured; nothing to look up")
	})

	t.Run("out-of-range integers are ignored", func(t *testing.T) {
		srv, _ := serveVendor(t, "groq", "/openai/v1/models/m", map[string]any{"context_window": 1 << 31, "max_completion_tokens": 0.5})
		api, _ := detectVendor(t, srv, config.ModelConfig{ModelName: "m", ApiKey: vendorTestKey})
		require.NotNil(t, api)
		require.Nil(t, api.Capabilities)
	})
}

func TestVendorLookupPath(t *testing.T) {
	t.Run("the id is one escaped path segment without inline parameters", func(t *testing.T) {
		srv, fx := serveVendor(t, "mistral", "/v1/models/org%2Fmodel:v2", map[string]any{"id": "org/model:v2", "max_context_length": 8192})
		api, _ := detectVendor(t, srv, config.ModelConfig{ModelName: "org/model:v2:temperature=0", ApiKey: vendorTestKey})
		require.Equal(t, "/v1/models/org%2Fmodel:v2", *fx.rawPath.Load())
		require.NotNil(t, api.Capabilities)
		require.Equal(t, int64p(8192), api.Capabilities.Context)
	})

	t.Run("the configured chat path is not a prefix", func(t *testing.T) {
		srv, fx := serveVendor(t, "xai", "/v1/language-models/grok-3", map[string]any{"id": "grok-3", "input_modalities": []string{"text"}})
		api, _ := detectVendor(t, srv, config.ModelConfig{ModelName: "grok-3", ApiURL: srv.URL + "/v1/chat/completions", ApiKey: vendorTestKey})
		require.Equal(t, "/v1/language-models/grok-3", *fx.rawPath.Load())
		require.Equal(t, []string{"text"}, api.Capabilities.Modalities.Input)
	})
}

// The dispatch key is what StackFromHost returns for the vendor's host.
func TestVendorProbesRegistered(t *testing.T) {
	hosts := map[string]string{
		"anthropic": "api.anthropic.com", "gemini": "generativelanguage.googleapis.com", "mistral": "api.mistral.ai",
		"groq": "api.groq.com", "xai": "api.x.ai", "together": "api.together.xyz", "cerebras": "api.cerebras.ai",
		"deepinfra": "api.deepinfra.com", "hf-router": "router.huggingface.co",
	}
	for vendor, host := range hosts {
		_, ok := hostedProbes[vendor]
		require.True(t, ok, vendor)
		require.Equal(t, vendor, apispec.StackFromHost("https://"+host+"/v1"))
	}
	_, ok := hostedProbes["openai"]
	require.False(t, ok, "openai's listing exposes nothing; nothing to probe")
}

func filterTrace(lines []string, prefix string) []string {
	var out []string
	for _, l := range lines {
		if len(l) >= len(prefix) && l[:len(prefix)] == prefix {
			out = append(out, l)
		}
	}
	return out
}
