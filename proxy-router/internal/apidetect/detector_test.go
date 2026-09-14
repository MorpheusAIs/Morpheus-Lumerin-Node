package apidetect

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/config"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/lib"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/system"
	"github.com/stretchr/testify/require"
)

func newTestDetector() *Detector {
	return NewDetector(lib.NewTestLogger(), DefaultOptions())
}

func detect(t *testing.T, url, apiType, modelName string) *system.ModelApiSpec {
	t.Helper()
	return newTestDetector().Detect(context.Background(), config.ModelConfig{ModelName: modelName, ApiType: apiType, ApiURL: url})
}

func jsonHandler(v any) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(v)
	}
}

// --- self-hosted engine fingerprinting ---

const qwen3Template = `{%- if messages %}{% generation %}{% endgeneration %}{%- endif %}
{%- if enable_thinking is defined and enable_thinking is false %}<think></think>{%- endif %}`

const deepseekV31Template = `{% if not thinking is defined %}{% set thinking = false %}{% endif %}
{% if thinking %}<think>{% endif %}`

func TestDetectLlamaCppQwen3Template(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/props", jsonHandler(map[string]any{
		"total_slots":   4,
		"chat_template": qwen3Template,
	}))
	srv := httptest.NewServer(mux)
	defer srv.Close()

	api := detect(t, srv.URL+"/v1/chat/completions", "openai", "qwen3-32b")
	require.NotNil(t, api)
	require.Equal(t, "llamacpp", api.Stack)
	require.Equal(t, "qwen3", api.ModelFamily)
	require.NotNil(t, api.Thinking)
	require.Equal(t, system.ThinkingModeControllable, api.Thinking.Mode)
	disable := api.Bindings[system.IntentReasoningDisable]
	require.NotNil(t, disable)
	require.Equal(t, system.BindingKindTemplateKwarg, disable.Kind)
	require.Equal(t, "chat_template_kwargs.enable_thinking", disable.Param)
	require.Equal(t, "boolean", disable.ParamType)
	require.Equal(t, false, disable.Value)
	enable := api.Bindings[system.IntentReasoningEnable]
	require.NotNil(t, enable)
	require.Equal(t, true, enable.Value)
}

func TestDetectLlamaCppDeepSeekTemplateUsesThinkingKwarg(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/props", jsonHandler(map[string]any{
		"total_slots":   1,
		"chat_template": deepseekV31Template,
	}))
	srv := httptest.NewServer(mux)
	defer srv.Close()

	api := detect(t, srv.URL+"/v1/chat/completions", "openai", "deepseek-v3.1")
	require.NotNil(t, api)
	require.Equal(t, "llamacpp", api.Stack)
	disable := api.Bindings[system.IntentReasoningDisable]
	require.NotNil(t, disable)
	require.Equal(t, "chat_template_kwargs.thinking", disable.Param)
	require.Equal(t, false, disable.Value)
}

func TestDetectOllama(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/tags", jsonHandler(map[string]any{
		"models": []map[string]any{{"name": "qwen3:32b"}},
	}))
	mux.HandleFunc("/api/show", jsonHandler(map[string]any{
		"template":     "{{ .Prompt }}",
		"capabilities": []string{"completion", "tools", "thinking"},
		"model_info":   map[string]any{"general.architecture": "qwen3"},
	}))
	srv := httptest.NewServer(mux)
	defer srv.Close()

	api := detect(t, srv.URL+"/v1/chat/completions", "openai", "qwen3:32b")
	require.NotNil(t, api)
	require.Equal(t, "ollama", api.Stack)
	require.Equal(t, "qwen3", api.ModelFamily)
	require.NotNil(t, api.Thinking)
	require.Equal(t, system.ThinkingModeControllable, api.Thinking.Mode)
	// Ollama's /v1 endpoint documents reasoning_effort, with "none" as off
	// and any other level (medium) as on.
	disable := api.Bindings[system.IntentReasoningDisable]
	require.NotNil(t, disable)
	require.Equal(t, system.BindingKindBodyParam, disable.Kind)
	require.Equal(t, "reasoning_effort", disable.Param)
	require.Equal(t, "none", disable.Value)
	enable := api.Bindings[system.IntentReasoningEnable]
	require.NotNil(t, enable)
	require.Equal(t, system.BindingKindBodyParam, enable.Kind)
	require.Equal(t, "reasoning_effort", enable.Param)
	require.Equal(t, "medium", enable.Value)
	// stack table adds the /v1 effort levels and the documented native knobs
	require.Equal(t, "reasoning_effort", api.Bindings[system.IntentReasoningEffort].Param)
	require.Equal(t, "options.num_ctx", api.Bindings[system.IntentContextNumCtx].Param)
	require.Equal(t, system.BindingKindNativeBodyParam, api.Bindings[system.IntentContextNumCtx].Kind)
}

func TestDetectSGLangAlwaysOnReasoner(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/get_model_info", jsonHandler(map[string]any{
		"model_path":    "deepseek-ai/DeepSeek-R1",
		"is_generation": true,
	}))
	srv := httptest.NewServer(mux)
	defer srv.Close()

	api := detect(t, srv.URL+"/v1/chat/completions", "openai", "whatever-local-alias")
	require.NotNil(t, api)
	require.Equal(t, "sglang", api.Stack)
	require.Equal(t, "deepseek-r1", api.ModelFamily)
	require.NotNil(t, api.Thinking)
	require.Equal(t, system.ThinkingModeAlwaysOn, api.Thinking.Mode)
	// no toggle exists, but the stack's model-independent knobs still apply
	require.Nil(t, api.Bindings[system.IntentReasoningDisable])
	require.Nil(t, api.Bindings[system.IntentReasoningEnable])
	require.Equal(t, "separate_reasoning", api.Bindings[system.IntentReasoningFormat].Param)
	require.Equal(t, "top_k", api.Bindings[system.IntentSamplingTopK].Param)
}

func TestDetectVLLMFamilyDefaultFromModelsList(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/version", jsonHandler(map[string]any{"version": "0.8.4"}))
	mux.HandleFunc("/v1/models", jsonHandler(map[string]any{
		"object": "list",
		"data": []map[string]any{
			{"id": "Qwen/Qwen3-235B-A22B", "object": "model", "owned_by": "vllm", "max_model_len": 40960},
		},
	}))
	srv := httptest.NewServer(mux)
	defer srv.Close()

	api := detect(t, srv.URL+"/v1/chat/completions", "openai", "Qwen/Qwen3-235B-A22B")
	require.NotNil(t, api)
	require.Equal(t, "vllm", api.Stack)
	require.Equal(t, "qwen3", api.ModelFamily)
	require.NotNil(t, api.Thinking)
	require.Equal(t, "chat_template_kwargs.enable_thinking", api.Bindings[system.IntentReasoningDisable].Param)
}

func TestDetectTGI(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/info", jsonHandler(map[string]any{
		"model_id": "mistralai/Mistral-7B-Instruct-v0.3",
		"version":  "3.0.1",
	}))
	srv := httptest.NewServer(mux)
	defer srv.Close()

	api := detect(t, srv.URL+"/v1/chat/completions", "openai", "mistral-7b")
	require.NotNil(t, api)
	require.Equal(t, "tgi", api.Stack)
	require.Equal(t, "mistral", api.ModelFamily)
	require.Nil(t, api.Thinking)
}

func TestDetectLMStudio(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v0/models", jsonHandler(map[string]any{
		"object": "list",
		"data": []map[string]any{
			{"id": "granite-3.2-8b", "arch": "granite", "type": "llm"},
		},
	}))
	srv := httptest.NewServer(mux)
	defer srv.Close()

	api := detect(t, srv.URL+"/v1/chat/completions", "openai", "granite-3.2-8b")
	require.NotNil(t, api)
	require.Equal(t, "lmstudio", api.Stack)
	require.Equal(t, "granite", api.ModelFamily)
	require.NotNil(t, api.Thinking)
	require.Equal(t, "chat_template_kwargs.thinking", api.Bindings[system.IntentReasoningDisable].Param)
}

func TestDetectKoboldCpp(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/extra/version", jsonHandler(map[string]any{
		"result": "KoboldCpp", "version": "1.90",
	}))
	srv := httptest.NewServer(mux)
	defer srv.Close()

	api := detect(t, srv.URL+"/v1/chat/completions", "openai", "llama-3.3-70b")
	require.NotNil(t, api)
	require.Equal(t, "koboldcpp", api.Stack)
	require.Equal(t, "llama", api.ModelFamily)
}

// --- registry-style hosted APIs detected by response shape ---

func TestDetectOpenRouterByShape(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/models", jsonHandler(map[string]any{
		"data": []map[string]any{
			{
				"id": "deepseek/deepseek-chat-v3.1",
				"supported_parameters": []string{
					"temperature", "top_p", "reasoning", "include_reasoning", "tools", "response_format",
				},
			},
		},
	}))
	srv := httptest.NewServer(mux)
	defer srv.Close()

	api := detect(t, srv.URL+"/api/v1/chat/completions", "openai", "deepseek/deepseek-chat-v3.1")
	require.NotNil(t, api)
	require.Equal(t, "openrouter", api.Stack)
	require.Contains(t, api.Parameters, "reasoning")
	require.Contains(t, api.Parameters, "tools")
	require.NotNil(t, api.Thinking)
	require.Equal(t, system.ThinkingModeControllable, api.Thinking.Mode)
	disable := api.Bindings[system.IntentReasoningDisable]
	require.NotNil(t, disable)
	require.Equal(t, system.BindingKindBodyParam, disable.Kind)
	require.Equal(t, "reasoning.effort", disable.Param)
	require.Equal(t, "none", disable.Value)
	effort := api.Bindings[system.IntentReasoningEffort]
	require.NotNil(t, effort)
	require.Equal(t, "reasoning.effort", effort.Param)
	require.Contains(t, effort.EnumValues, "xhigh")
}

func TestDetectVeniceByShape(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/models", jsonHandler(map[string]any{
		"data": []map[string]any{
			{
				"id": "qwen3-235b",
				"model_spec": map[string]any{
					"capabilities": map[string]any{
						"supportsReasoning":       true,
						"supportsFunctionCalling": true,
						"supportsVision":          false,
					},
				},
			},
		},
	}))
	srv := httptest.NewServer(mux)
	defer srv.Close()

	api := detect(t, srv.URL+"/api/v1/chat/completions", "openai", "qwen3-235b")
	require.NotNil(t, api)
	require.Equal(t, "venice", api.Stack)
	require.Contains(t, api.Parameters, "reasoning")
	require.Contains(t, api.Parameters, "tools")
	require.NotContains(t, api.Parameters, "vision")
}

// --- hosted vendors recognized by hostname, no probing ---

func TestDetectHostedVendorSkipsProbing(t *testing.T) {
	d := newTestDetector()
	api := d.Detect(context.Background(), config.ModelConfig{
		ModelName: "claude-sonnet-4-5",
		ApiType:   "claudeai",
		ApiURL:    "https://api.anthropic.com/v1/messages",
	})
	require.NotNil(t, api)
	require.Equal(t, "anthropic", api.Stack)
	require.Equal(t, "claude", api.ModelFamily)
	require.NotNil(t, api.Thinking)
	disable := api.Bindings[system.IntentReasoningDisable]
	require.NotNil(t, disable)
	require.Equal(t, "thinking", disable.Param)
	require.Equal(t, map[string]any{"type": "disabled"}, disable.Value)
	require.Equal(t, "output_config.effort", api.Bindings[system.IntentReasoningEffort].Param)
	// stack table: Messages API extras
	require.Equal(t, "output_config.format", api.Bindings[system.IntentResponseFormatSchema].Param)
	require.Equal(t, "thinking.display", api.Bindings[system.IntentReasoningFormat].Param)
	require.Contains(t, api.Parameters, "max_tokens")
	require.Contains(t, api.Parameters, "stop_sequences")
	require.NotContains(t, api.Parameters, "stop")
}

// --- caching ---

func TestDetectCachesResult(t *testing.T) {
	var calls int32
	mux := http.NewServeMux()
	mux.HandleFunc("/props", func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		jsonHandler(map[string]any{"total_slots": 1, "chat_template": qwen3Template})(w, r)
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	d := newTestDetector()
	cfg := config.ModelConfig{ModelName: "qwen3-32b", ApiType: "openai", ApiURL: srv.URL + "/v1/chat/completions"}

	first := d.Detect(context.Background(), cfg)
	second := d.Detect(context.Background(), cfg)
	require.NotNil(t, first)
	require.Equal(t, first.Stack, second.Stack)
	require.Equal(t, int32(1), atomic.LoadInt32(&calls))
}

// --- tracing: DetectWithTrace explains every step and bypasses the cache ---

func TestDetectWithTraceExplainsSteps(t *testing.T) {
	var calls int32
	mux := http.NewServeMux()
	mux.HandleFunc("/props", func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		jsonHandler(map[string]any{"total_slots": 1, "chat_template": qwen3Template})(w, r)
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	d := newTestDetector()
	cfg := config.ModelConfig{ModelName: "qwen3-32b", ApiType: "openai", ApiURL: srv.URL + "/v1/chat/completions"}

	api, trace := d.DetectWithTrace(context.Background(), cfg)
	require.NotNil(t, api)
	require.Equal(t, "llamacpp", api.Stack)
	require.NotEmpty(t, trace)

	joined := strings.Join(trace, "\n")
	require.Contains(t, joined, "/props")
	require.Contains(t, joined, "llamacpp")
	require.Contains(t, joined, "chat template")
	require.Contains(t, joined, "enable_thinking")
	require.Contains(t, joined, "qwen3")
	// misses are traced too: the ollama probe ran before /props matched
	require.Contains(t, joined, "/api/tags")

	// trace runs bypass the cache: a second call probes again
	_, _ = d.DetectWithTrace(context.Background(), cfg)
	require.Equal(t, int32(2), atomic.LoadInt32(&calls))
}

func TestDetectWithTraceHostedVendor(t *testing.T) {
	d := newTestDetector()
	_, trace := d.DetectWithTrace(context.Background(), config.ModelConfig{
		ModelName: "claude-opus-4-6",
		ApiType:   "claudeai",
		ApiURL:    "https://api.anthropic.com/v1/messages",
	})
	joined := strings.Join(trace, "\n")
	require.Contains(t, joined, "anthropic")
	require.Contains(t, joined, "family")
	require.Contains(t, joined, "claude")
}

// --- nothing detectable: fall back to name heuristics only ---

func TestDetectNothingReachableFallsBackToName(t *testing.T) {
	srv := httptest.NewServer(http.NotFoundHandler())
	defer srv.Close()

	api := detect(t, srv.URL+"/v1/chat/completions", "openai", "deepseek-r1-distill")
	require.NotNil(t, api)
	require.Equal(t, "", api.Stack)
	require.Equal(t, "deepseek-r1", api.ModelFamily)
	require.Equal(t, system.ThinkingModeAlwaysOn, api.Thinking.Mode)
}

// --- review regressions: family resolution, hosted gating, ollama rewrite, secrets ---

func TestDetectRawArchitectureFallsThroughToName(t *testing.T) {
	// LM Studio reports the GGUF architecture "gptoss"; the name still refines it.
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v0/models", jsonHandler(map[string]any{
		"object": "list",
		"data":   []map[string]any{{"id": "openai/gpt-oss-120b", "arch": "gptoss", "type": "llm"}},
	}))
	srv := httptest.NewServer(mux)
	defer srv.Close()

	api := detect(t, srv.URL+"/v1/chat/completions", "openai", "gpt-oss-120b")
	require.NotNil(t, api)
	require.Equal(t, "gpt-oss", api.ModelFamily)
	require.NotNil(t, api.Thinking)
	require.Equal(t, system.ThinkingModeTunable, api.Thinking.Mode)
	require.Nil(t, api.Bindings[system.IntentReasoningDisable])
	require.Equal(t, "chat_template_kwargs.reasoning_effort", api.Bindings[system.IntentReasoningEffort].Param)
}

func TestDetectUnknownArchitectureDoesNotBecomeFamily(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v0/models", jsonHandler(map[string]any{
		"object": "list",
		"data":   []map[string]any{{"id": "some-embedder", "arch": "bert", "type": "embeddings"}},
	}))
	srv := httptest.NewServer(mux)
	defer srv.Close()

	api := detect(t, srv.URL+"/v1/chat/completions", "openai", "some-embedder")
	require.NotNil(t, api)
	require.Equal(t, "lmstudio", api.Stack)
	require.Equal(t, "", api.ModelFamily)
}

func TestDetectServedIDRefinesGenericArchitecture(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v0/models", jsonHandler(map[string]any{
		"object": "list",
		"data":   []map[string]any{{"id": "deepseek-r1-distill", "arch": "deepseek2", "type": "llm"}},
	}))
	srv := httptest.NewServer(mux)
	defer srv.Close()

	api := detect(t, srv.URL+"/v1/chat/completions", "openai", "deepseek-r1-distill")
	require.NotNil(t, api)
	require.Equal(t, "deepseek-r1", api.ModelFamily)
	require.Equal(t, system.ThinkingModeAlwaysOn, api.Thinking.Mode)
}

func TestDetectHostedVendorSkipsTemplateKwargDefaults(t *testing.T) {
	// Groq serves qwen3 but is not an HF-template engine: chat_template_kwargs
	// would be rejected, so no family default may be advertised.
	d := newTestDetector()
	api := d.Detect(context.Background(), config.ModelConfig{
		ModelName: "qwen/qwen3-32b",
		ApiType:   "openai",
		ApiURL:    "https://api.groq.com/openai/v1/chat/completions",
	})
	require.NotNil(t, api)
	require.Equal(t, "groq", api.Stack)
	require.Equal(t, "qwen3", api.ModelFamily)
	require.Empty(t, api.Bindings)
	require.Nil(t, api.Thinking)
}

func TestDetectHostedVendorKeepsNativeFamilyDefaults(t *testing.T) {
	// claude defaults describe Anthropic's own API: valid there, not elsewhere.
	d := newTestDetector()
	api := d.Detect(context.Background(), config.ModelConfig{
		ModelName: "claude-sonnet-4-5",
		ApiType:   "openai",
		ApiURL:    "https://api.together.xyz/v1/chat/completions",
	})
	require.NotNil(t, api)
	require.Equal(t, "together", api.Stack)
	require.Empty(t, api.Bindings)
}

func TestDetectOllamaRewritesBudgetFamily(t *testing.T) {
	// seed-oss default is a numeric template-kwarg budget; on Ollama only the
	// reasoning_effort toggle exists, so the budget intent must be dropped and
	// both toggle directions must remain.
	mux := http.NewServeMux()
	mux.HandleFunc("/api/tags", jsonHandler(map[string]any{"models": []map[string]any{{"name": "seed-oss:36b"}}}))
	mux.HandleFunc("/api/show", jsonHandler(map[string]any{"template": "{{ .Prompt }}"}))
	srv := httptest.NewServer(mux)
	defer srv.Close()

	api := detect(t, srv.URL+"/v1/chat/completions", "openai", "seed-oss:36b")
	require.NotNil(t, api)
	require.Equal(t, "ollama", api.Stack)
	require.Nil(t, api.Bindings[system.IntentReasoningBudget])
	require.Equal(t, "reasoning_effort", api.Bindings[system.IntentReasoningDisable].Param)
	require.Equal(t, "none", api.Bindings[system.IntentReasoningDisable].Value)
	require.Equal(t, "reasoning_effort", api.Bindings[system.IntentReasoningEnable].Param)
	require.Equal(t, "medium", api.Bindings[system.IntentReasoningEnable].Value)
	require.Equal(t, system.ThinkingModeControllable, api.Thinking.Mode)
}

func TestDetectOpenRouterImportsAllReasoningBindings(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/models", jsonHandler(map[string]any{
		"data": []map[string]any{{"id": "x/y", "supported_parameters": []string{"reasoning"}}},
	}))
	srv := httptest.NewServer(mux)
	defer srv.Close()

	api := detect(t, srv.URL+"/api/v1/chat/completions", "openai", "x/y")
	require.NotNil(t, api)
	require.Equal(t, "reasoning.effort", api.Bindings[system.IntentReasoningDisable].Param)
	require.Equal(t, "reasoning.enabled", api.Bindings[system.IntentReasoningEnable].Param)
	require.Equal(t, true, api.Bindings[system.IntentReasoningEnable].Value)
	require.Equal(t, "reasoning.effort", api.Bindings[system.IntentReasoningEffort].Param)
	require.Equal(t, "reasoning.max_tokens", api.Bindings[system.IntentReasoningBudget].Param)
	require.Equal(t, "reasoning.exclude", api.Bindings[system.IntentReasoningFormat].Param)
	// the registry list is authoritative: the stack's default param list must not replace it
	require.Equal(t, []string{"reasoning"}, api.Parameters)
	// stack table still contributes request shapes
	require.Equal(t, "top_a", api.Bindings[system.IntentSamplingTopA].Param)
}

func TestTraceNeverContainsApiKeyOrURLCredentials(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/props", jsonHandler(map[string]any{"total_slots": 1, "chat_template": qwen3Template}))
	srv := httptest.NewServer(mux)
	defer srv.Close()

	u := strings.Replace(srv.URL, "http://", "http://user:URLSECRET@", 1)
	d := newTestDetector()
	_, trace := d.DetectWithTrace(context.Background(), config.ModelConfig{
		ModelName: "qwen3-32b", ApiType: "openai", ApiURL: u + "/v1/chat/completions?key=QUERYSECRET", ApiKey: "sk-HEADERSECRET",
	})
	joined := strings.Join(trace, "\n")
	require.NotContains(t, joined, "sk-HEADERSECRET")
	require.NotContains(t, joined, "URLSECRET")
	require.NotContains(t, joined, "QUERYSECRET")
}

func TestRedactURL(t *testing.T) {
	require.Equal(t, "https://h.example/v1", RedactURL("https://u:p@h.example/v1?k=v#f"))
	// never echo input that could not be parsed (it may be a pasted secret)
	require.Equal(t, "<unparseable url>", RedactURL("not a url"))
	require.Equal(t, "<unparseable url>", RedactURL("http://[::1"))
	require.Equal(t, "<unparseable url>", RedactURL(""))
	require.Equal(t, "<unparseable url>", RedactURL("/v1/chat/completions?key=SECRET"))
}

func TestCacheKeyChangesWithApiKeyStackAndFamily(t *testing.T) {
	base := config.ModelConfig{ApiURL: "u", ModelName: "m", ApiKey: "k1"}
	a := cacheKey(base)
	require.NotContains(t, a, "k1")
	other := base
	other.ApiKey = "k2"
	require.NotEqual(t, a, cacheKey(other))
	other = base
	other.ApiStack = "vllm"
	require.NotEqual(t, a, cacheKey(other), "R7: key changes when apiStack becomes set")
	other = base
	other.ModelFamily = "qwen3"
	require.NotEqual(t, a, cacheKey(other), "R7: key changes with modelFamily")
	require.Equal(t, a, cacheKey(base))
}

// --- stack tables end-to-end ---

func TestDetectVLLMStackTableForReasoningModel(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/version", jsonHandler(map[string]any{"version": "0.18.0"}))
	mux.HandleFunc("/v1/models", jsonHandler(map[string]any{
		"object": "list", "data": []map[string]any{{"id": "Qwen/Qwen3-32B", "owned_by": "vllm"}},
	}))
	srv := httptest.NewServer(mux)
	defer srv.Close()

	api := detect(t, srv.URL+"/v1/chat/completions", "openai", "Qwen/Qwen3-32B")
	require.NotNil(t, api)
	require.Equal(t, "vllm", api.Stack)
	// family layer owns the toggle; stack layer adds model-independent knobs
	require.Equal(t, "chat_template_kwargs.enable_thinking", api.Bindings[system.IntentReasoningDisable].Param)
	require.Equal(t, "thinking_token_budget", api.Bindings[system.IntentReasoningBudget].Param)
	require.Equal(t, "include_reasoning", api.Bindings[system.IntentReasoningFormat].Param)
	require.Equal(t, "top_k", api.Bindings[system.IntentSamplingTopK].Param)
	require.Equal(t, "structured_outputs.choice", api.Bindings[system.IntentResponseFormatChoice].Param)
	require.Equal(t, "array", api.Bindings[system.IntentResponseFormatChoice].ParamType)
	require.Equal(t, map[string]any{"type": "json_object"}, api.Bindings[system.IntentResponseFormatJSON].Value)
	require.Contains(t, api.Parameters, "logprobs")
	require.Equal(t, system.ThinkingModeControllable, api.Thinking.Mode)
}

func TestDetectVLLMStackTableForPlainModelHasNoReasoning(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/version", jsonHandler(map[string]any{"version": "0.18.0"}))
	srv := httptest.NewServer(mux)
	defer srv.Close()

	api := detect(t, srv.URL+"/v1/chat/completions", "openai", "meta-llama/Llama-3.3-70B-Instruct")
	require.NotNil(t, api)
	require.Equal(t, "vllm", api.Stack)
	require.Equal(t, "llama", api.ModelFamily)
	// a stack's reasoning knobs must not make a non-reasoning model look tunable
	require.Nil(t, api.Thinking)
	for intent := range api.Bindings {
		require.False(t, strings.HasPrefix(intent, "reasoning."), intent)
	}
	require.Equal(t, "repetition_penalty", api.Bindings[system.IntentSamplingRepetitionPenalty].Param)
	require.Equal(t, "structured_outputs.grammar", api.Bindings[system.IntentResponseFormatGrammar].Param)
}

func TestDetectLlamaCppStackTable(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/props", jsonHandler(map[string]any{"total_slots": 1, "chat_template": deepseekV31Template}))
	srv := httptest.NewServer(mux)
	defer srv.Close()

	api := detect(t, srv.URL+"/v1/chat/completions", "openai", "deepseek-v3.1")
	require.NotNil(t, api)
	require.Equal(t, "repeat_penalty", api.Bindings[system.IntentSamplingRepetitionPenalty].Param)
	require.Equal(t, "grammar", api.Bindings[system.IntentResponseFormatGrammar].Param)
	require.Equal(t, "cache_prompt", api.Bindings[system.IntentCachePrompt].Param)
	require.Equal(t, "reasoning_format", api.Bindings[system.IntentReasoningFormat].Param)
	require.ElementsMatch(t, []string{"none", "auto", "deepseek", "deepseek-legacy"}, api.Bindings[system.IntentReasoningFormat].EnumValues)
	require.Equal(t, "reasoning_budget", api.Bindings[system.IntentReasoningBudget].Param)
}

func TestDetectSGLangStackTable(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/get_model_info", jsonHandler(map[string]any{"model_path": "Qwen/Qwen3-8B", "is_generation": true}))
	srv := httptest.NewServer(mux)
	defer srv.Close()

	api := detect(t, srv.URL+"/v1/chat/completions", "openai", "alias")
	require.NotNil(t, api)
	require.Equal(t, "sglang", api.Stack)
	require.Equal(t, "separate_reasoning", api.Bindings[system.IntentReasoningFormat].Param)
	require.Equal(t, "ebnf", api.Bindings[system.IntentResponseFormatGrammar].Param)
	require.Equal(t, "min_p", api.Bindings[system.IntentSamplingMinP].Param)
	require.Contains(t, api.Parameters, "user")
}

func TestDetectOllamaGptOssTunesEffortOnly(t *testing.T) {
	// Ollama lists the thinking capability for gpt-oss, but booleans are
	// ignored for it: the family default (effort levels) is what applies,
	// rewritten onto /v1 reasoning_effort, with no disable.
	mux := http.NewServeMux()
	mux.HandleFunc("/api/tags", jsonHandler(map[string]any{"models": []map[string]any{{"name": "gpt-oss:120b"}}}))
	mux.HandleFunc("/api/show", jsonHandler(map[string]any{
		"template":     "{{ .Prompt }}",
		"capabilities": []string{"completion", "thinking"},
		"model_info":   map[string]any{"general.architecture": "gptoss"},
	}))
	srv := httptest.NewServer(mux)
	defer srv.Close()

	api := detect(t, srv.URL+"/v1/chat/completions", "openai", "gpt-oss:120b")
	require.NotNil(t, api)
	require.Equal(t, "ollama", api.Stack)
	require.Equal(t, "gpt-oss", api.ModelFamily)
	require.Equal(t, system.ThinkingModeTunable, api.Thinking.Mode)
	require.Nil(t, api.Bindings[system.IntentReasoningDisable])
	effort := api.Bindings[system.IntentReasoningEffort]
	require.NotNil(t, effort)
	require.Equal(t, system.BindingKindBodyParam, effort.Kind)
	require.Equal(t, "reasoning_effort", effort.Param)
	require.ElementsMatch(t, []string{"low", "medium", "high"}, effort.EnumValues)
}

func TestDetectLiteLLM(t *testing.T) {
	var sawAuth, sawGroup string
	mux := http.NewServeMux()
	mux.HandleFunc("/health/liveliness", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`"I'm alive!"`))
	})
	mux.HandleFunc("/model_group/info", func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		sawGroup = r.URL.Query().Get("model_group")
		jsonHandler(map[string]any{"data": []map[string]any{{
			"model_group":             "claude-sonnet-4-5",
			"supported_openai_params": []string{"temperature", "tools", "reasoning_effort", "response_format"},
			"supports_reasoning":      true,
		}}})(w, r)
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	d := newTestDetector()
	api := d.Detect(context.Background(), config.ModelConfig{
		ModelName: "claude-sonnet-4-5", ApiType: "openai", ApiURL: srv.URL + "/v1/chat/completions", ApiKey: "sk-litellm",
	})
	require.NotNil(t, api)
	require.Equal(t, "litellm", api.Stack)
	require.Equal(t, "Bearer sk-litellm", sawAuth)
	require.Equal(t, "claude-sonnet-4-5", sawGroup)
	require.Equal(t, []string{"reasoning_effort", "response_format", "temperature", "tools"}, api.Parameters)
	// claude family defaults describe Anthropic's own API, not a gateway:
	// litellm's provider-neutral reasoning knobs apply instead
	require.Equal(t, "reasoning_effort", api.Bindings[system.IntentReasoningDisable].Param)
	require.Equal(t, "none", api.Bindings[system.IntentReasoningDisable].Value)
	require.Equal(t, "thinking.budget_tokens", api.Bindings[system.IntentReasoningBudget].Param)
	require.Equal(t, system.ThinkingModeControllable, api.Thinking.Mode)
}

func TestDetectLiteLLMWithoutReasoningSupportSkipsReasoning(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/health/liveliness", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`"I'm alive!"`))
	})
	mux.HandleFunc("/model_group/info", jsonHandler(map[string]any{"data": []map[string]any{{
		"model_group": "llama-3.3-70b", "supported_openai_params": []string{"temperature"}, "supports_reasoning": false,
	}}}))
	srv := httptest.NewServer(mux)
	defer srv.Close()

	api := detect(t, srv.URL+"/chat/completions", "openai", "llama-3.3-70b")
	require.NotNil(t, api)
	require.Equal(t, "litellm", api.Stack)
	require.Nil(t, api.Thinking)
	require.Nil(t, api.Bindings[system.IntentReasoningEffort])
	require.Equal(t, "response_format", api.Bindings[system.IntentResponseFormatJSON].Param)
}

// --- second-review regressions ---

func TestDetectAlwaysOnFamilyOnLiteLLMGetsNoToggle(t *testing.T) {
	// deepseek-r1 reasons unconditionally; a gateway's generic toggle must
	// not be advertised for it alongside an always_on mode.
	mux := http.NewServeMux()
	mux.HandleFunc("/health/liveliness", func(w http.ResponseWriter, r *http.Request) { _, _ = w.Write([]byte(`"I'm alive!"`)) })
	mux.HandleFunc("/model_group/info", jsonHandler(map[string]any{"data": []map[string]any{{
		"model_group": "deepseek-r1", "supported_openai_params": []string{"temperature"}, "supports_reasoning": true,
	}}}))
	srv := httptest.NewServer(mux)
	defer srv.Close()

	api := detect(t, srv.URL+"/v1/chat/completions", "openai", "deepseek-r1")
	require.NotNil(t, api)
	require.Equal(t, "litellm", api.Stack)
	require.Equal(t, system.ThinkingModeAlwaysOn, api.Thinking.Mode)
	require.Nil(t, api.Bindings[system.IntentReasoningDisable])
	require.Nil(t, api.Bindings[system.IntentReasoningEnable])
	require.Nil(t, api.Bindings[system.IntentReasoningEffort])
	require.Nil(t, api.Bindings[system.IntentReasoningBudget])
}

func TestDetectOllamaAlwaysOnVariantIgnoresCapabilityToggle(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/tags", jsonHandler(map[string]any{"models": []map[string]any{{"name": "qwen3-thinking:32b"}}}))
	mux.HandleFunc("/api/show", jsonHandler(map[string]any{
		"template": "{{ .Prompt }}", "capabilities": []string{"thinking"}, "model_info": map[string]any{"general.architecture": "qwen3"},
	}))
	srv := httptest.NewServer(mux)
	defer srv.Close()

	api := detect(t, srv.URL+"/v1/chat/completions", "openai", "qwen3-thinking:32b")
	require.NotNil(t, api)
	require.Equal(t, system.ThinkingModeAlwaysOn, api.Thinking.Mode)
	require.Nil(t, api.Bindings[system.IntentReasoningDisable])
	require.Nil(t, api.Bindings[system.IntentReasoningEffort])
}

func TestDetectVeniceReasoningModel(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/models", jsonHandler(map[string]any{
		"data": []map[string]any{{
			"id": "qwen3-235b", "model_spec": map[string]any{"capabilities": map[string]any{"supportsReasoning": true, "supportsFunctionCalling": true}},
		}},
	}))
	srv := httptest.NewServer(mux)
	defer srv.Close()

	api := detect(t, srv.URL+"/api/v1/chat/completions", "openai", "qwen3-235b")
	require.NotNil(t, api)
	require.Equal(t, "venice", api.Stack)
	require.Equal(t, system.ThinkingModeControllable, api.Thinking.Mode)
	require.Equal(t, "venice_parameters.disable_thinking", api.Bindings[system.IntentReasoningDisable].Param)
	require.Equal(t, true, api.Bindings[system.IntentReasoningDisable].Value)
	require.Equal(t, "reasoning_effort", api.Bindings[system.IntentReasoningEffort].Param)
	require.Equal(t, "venice_parameters.strip_thinking_response", api.Bindings[system.IntentReasoningFormat].Param)
	// the family's template kwargs never apply on a gateway
	require.NotEqual(t, system.BindingKindTemplateKwarg, api.Bindings[system.IntentReasoningDisable].Kind)
}

func TestDetectVeniceNonReasoningModelHasNoReasoning(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/models", jsonHandler(map[string]any{
		"data": []map[string]any{{"id": "llama-3.3-70b", "model_spec": map[string]any{"capabilities": map[string]any{"supportsReasoning": false}}}},
	}))
	srv := httptest.NewServer(mux)
	defer srv.Close()

	api := detect(t, srv.URL+"/api/v1/chat/completions", "openai", "llama-3.3-70b")
	require.NotNil(t, api)
	require.Nil(t, api.Thinking)
	require.Nil(t, api.Bindings[system.IntentReasoningEffort])
}

func TestDetectLiteLLMDegradedIntrospection(t *testing.T) {
	// 401 on model_group/info: stack still identified, nothing imported.
	mux := http.NewServeMux()
	mux.HandleFunc("/health/liveliness", func(w http.ResponseWriter, r *http.Request) { _, _ = w.Write([]byte(`I'm alive!`)) }) // unquoted variant
	mux.HandleFunc("/model_group/info", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusUnauthorized) })
	srv := httptest.NewServer(mux)
	defer srv.Close()

	api := detect(t, srv.URL+"/v1/chat/completions", "openai", "qwen3-32b")
	require.NotNil(t, api)
	require.Equal(t, "litellm", api.Stack)
	require.Nil(t, api.Thinking)
	require.Contains(t, api.Parameters, "messages") // stack default list
	require.Nil(t, api.Bindings[system.IntentReasoningDisable])

	// bare-object (no data wrapper) response shape is accepted too
	mux2 := http.NewServeMux()
	mux2.HandleFunc("/health/liveliness", func(w http.ResponseWriter, r *http.Request) { _, _ = w.Write([]byte(`"I'm alive!"`)) })
	mux2.HandleFunc("/model_group/info", jsonHandler(map[string]any{
		"model_group": "qwen3-32b", "supported_openai_params": []string{"tools"}, "supports_reasoning": true,
	}))
	srv2 := httptest.NewServer(mux2)
	defer srv2.Close()

	api = detect(t, srv2.URL+"/v1/chat/completions", "openai", "qwen3-32b")
	require.NotNil(t, api)
	require.Equal(t, []string{"tools"}, api.Parameters)
	require.Equal(t, system.ThinkingModeControllable, api.Thinking.Mode)
	require.Equal(t, "reasoning_effort", api.Bindings[system.IntentReasoningEnable].Param)
}

func TestDetectRegistryListingOverLimitIsIgnoredNotFatal(t *testing.T) {
	prev := maxRegistryBody
	maxRegistryBody = 512
	t.Cleanup(func() { maxRegistryBody = prev })

	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/models", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":[{"id":"x/y","supported_parameters":["reasoning"],"pad":"` + strings.Repeat("p", 2048) + `"}]}`))
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	d := newTestDetector()
	api, trace := d.DetectWithTrace(context.Background(), config.ModelConfig{ModelName: "x/y", ApiType: "openai", ApiURL: srv.URL + "/api/v1/chat/completions"})
	require.Contains(t, strings.Join(trace, "\n"), "exceeds the 512-byte limit")
	// name heuristics still run; nothing crashes
	_ = api
}

func TestDetectOpenAIHostAdvertisesStandardParameters(t *testing.T) {
	d := newTestDetector()
	api := d.Detect(context.Background(), config.ModelConfig{ModelName: "gpt-4o-mini", ApiType: "openai", ApiURL: "https://api.openai.com/v1/chat/completions"})
	require.NotNil(t, api)
	require.Contains(t, api.Parameters, "response_format")
	require.Contains(t, api.Parameters, "seed")
}

func TestIgnoreHostVendorsFallsBackToShapeDetection(t *testing.T) {
	// A venice-shaped listing must identify the stack even when hostname
	// recognition is off (custom-domain scenario); and a listing that lacks
	// the configured model is reported as config drift rather than silence.
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/models", jsonHandler(map[string]any{
		"data": []map[string]any{
			{"id": "mistral-small-2603", "model_spec": map[string]any{"capabilities": map[string]any{"supportsReasoning": false}}},
			{"id": "qwen3-235b", "model_spec": map[string]any{"capabilities": map[string]any{"supportsReasoning": true}}},
		},
	}))
	srv := httptest.NewServer(mux)
	defer srv.Close()

	d := NewDetector(lib.NewTestLogger(), Options{IgnoreHostVendors: true})
	api, trace := d.DetectWithTrace(context.Background(), config.ModelConfig{ModelName: "qwen3-235b", ApiType: "openai", ApiURL: srv.URL + "/api/v1/chat/completions"})
	joined := strings.Join(trace, "\n")
	require.Contains(t, joined, "hostname recognition disabled")
	require.NotNil(t, api)
	require.Equal(t, "venice", api.Stack)
	require.Equal(t, "venice_parameters.disable_thinking", api.Bindings[system.IntentReasoningDisable].Param)

	api, trace = d.DetectWithTrace(context.Background(), config.ModelConfig{ModelName: "mistral-31-24b", ApiType: "openai", ApiURL: srv.URL + "/api/v1/chat/completions"})
	joined = strings.Join(trace, "\n")
	require.Contains(t, joined, "2 entries but none matches model \"mistral-31-24b\"")
	require.NotNil(t, api)
	require.Equal(t, "", api.Stack)
	require.Equal(t, "mistral", api.ModelFamily)
}

// --- rulings R1, R3, R4, R7 and the R5 addendum's probe side ---

// R3: everything the detector composes says where it came from.
func TestDetectSetsDetectedSource(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/props", jsonHandler(map[string]any{"total_slots": 1, "chat_template": qwen3Template}))
	srv := httptest.NewServer(mux)
	defer srv.Close()
	api := detect(t, srv.URL+"/v1/chat/completions", "openai", "qwen3-32b")
	require.Equal(t, system.ApiSpecSourceDetected, api.Source)
	closed := httptest.NewServer(http.NotFoundHandler())
	closed.Close()
	unreachable := detect(t, closed.URL+"/v1/chat/completions", "openai", "qwen3-32b")
	require.Equal(t, system.ApiSpecSourceDetected, unreachable.Source)
}

// R1: a declared apiStack is composed statically; the backend is never probed.
func TestDetectDeclaredStackSkipsProbing(t *testing.T) {
	var hits int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&hits, 1)
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()
	d := newTestDetector()
	cfg := config.ModelConfig{ModelName: "qwen3-32b", ApiType: "openai", ApiStack: " VLLM ", ApiURL: srv.URL + "/v1/chat/completions"}
	api := d.Detect(context.Background(), cfg)
	require.NotNil(t, api)
	require.Equal(t, "vllm", api.Stack)
	require.Equal(t, system.ApiSpecSourceDeclared, api.Source)
	traced, trace := d.DetectWithTrace(context.Background(), cfg)
	require.Equal(t, api, traced)
	require.Len(t, trace, 1)
	require.Contains(t, trace[0], "declared")
	require.Equal(t, int32(0), atomic.LoadInt32(&hits))
}

// R1 is evaluated first: a declared apiStack composes statically even when
// the config names neither an endpoint nor a model (nothing to detect, but
// something to declare).
func TestDetectDeclaredStackWinsOverEmptyEndpoint(t *testing.T) {
	d := newTestDetector()
	cfg := config.ModelConfig{ApiStack: "litellm"}
	api := d.Detect(context.Background(), cfg)
	require.NotNil(t, api)
	require.Equal(t, "litellm", api.Stack)
	require.Equal(t, system.ApiSpecSourceDeclared, api.Source)
	traced, trace := d.DetectWithTrace(context.Background(), cfg)
	require.Equal(t, api, traced)
	require.Len(t, trace, 1)
	require.Contains(t, trace[0], "declared")
	require.Nil(t, d.Detect(context.Background(), config.ModelConfig{}), "nothing declared, nothing to detect")
	_, trace = d.DetectWithTrace(context.Background(), config.ModelConfig{})
	require.Contains(t, trace[0], "nothing to detect")
}

// Probes carry the configured key: a redirect may only be followed to the
// same host over the same or a better scheme.
func TestRefuseCrossHostRedirect(t *testing.T) {
	hop := func(from, to string) (*http.Request, []*http.Request) {
		prev, err := http.NewRequest(http.MethodGet, from, nil)
		require.NoError(t, err)
		next, err := http.NewRequest(http.MethodGet, to, nil)
		require.NoError(t, err)
		return next, []*http.Request{prev}
	}
	followed := []struct{ from, to string }{
		{"https://h.example/v1/models", "https://h.example/v1/models/"},
		{"http://h.example/x", "https://h.example/x"},
		{"http://h.example:8080/x", "http://H.example:8080/y"},
	}
	for _, c := range followed {
		require.NoError(t, refuseCrossHostRedirect(hop(c.from, c.to)), "%s -> %s", c.from, c.to)
	}
	refused := []struct{ from, to string }{
		{"https://h.example/x", "https://other.example/x"},
		{"https://h.example/x", "https://api.h.example/x"},
		{"https://h.example/x", "http://h.example/x"},
		{"http://h.example:8080/x", "http://h.example:9090/x"},
		{"https://h.example/x", "https://h.example:8443/x"},
	}
	for _, c := range refused {
		require.ErrorIs(t, refuseCrossHostRedirect(hop(c.from, c.to)), http.ErrUseLastResponse, "%s -> %s", c.from, c.to)
	}
	// the standard 10-hop cap is kept
	next, via := hop("https://h.example/0", "https://h.example/11")
	for len(via) < 10 {
		via = append(via, via[0])
	}
	err := refuseCrossHostRedirect(next, via)
	require.Error(t, err)
	require.NotErrorIs(t, err, http.ErrUseLastResponse)
}

func TestProbesRefuseCrossHostRedirects(t *testing.T) {
	listing := jsonHandler(map[string]any{"data": []map[string]any{veniceEntry("qwen3-235b", true)}})

	t.Run("cross-host redirect is not followed", func(t *testing.T) {
		other := http.NewServeMux()
		other.HandleFunc("/v1/models", listing)
		elsewhere, elsewhereLog := loggingServer(t, other)
		mux := http.NewServeMux()
		mux.HandleFunc("/v1/models", func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, elsewhere.URL+"/v1/models", http.StatusFound)
		})
		srv := httptest.NewServer(mux)
		defer srv.Close()

		d := newTestDetector()
		api, trace := d.DetectWithTrace(context.Background(), config.ModelConfig{ModelName: "qwen3-235b", ApiType: "openai", ApiURL: srv.URL + "/v1/chat/completions", ApiKey: "sk-secret"})
		require.Empty(t, elsewhereLog.seen(), "the other host must not see the request (nor the key)")
		require.Contains(t, strings.Join(trace, "\n"), "/v1/models -> HTTP 302")
		require.NotNil(t, api)
		require.Equal(t, "", api.Stack, "the redirect target's listing was never read")
	})

	t.Run("same-host redirect is followed", func(t *testing.T) {
		mux := http.NewServeMux()
		mux.HandleFunc("/v1/models", func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, "/v1/models/", http.StatusMovedPermanently)
		})
		mux.HandleFunc("/v1/models/", listing)
		srv := httptest.NewServer(mux)
		defer srv.Close()

		api := detect(t, srv.URL+"/v1/chat/completions", "openai", "qwen3-235b")
		require.NotNil(t, api)
		require.Equal(t, "venice", api.Stack, "the trailing-slash 301 was followed")
		require.Equal(t, system.ThinkingModeControllable, api.Thinking.Mode)
	})
}

// R4: when nothing answers, a family with default bindings still gets none —
// the wire vocabulary is unknown; only the family (and always_on) is reported.
func TestDetectUnreachableKnownFamilyHasNoBindings(t *testing.T) {
	srv := httptest.NewServer(http.NotFoundHandler())
	defer srv.Close()
	api := detect(t, srv.URL+"/v1/chat/completions", "openai", "qwen3-32b")
	require.NotNil(t, api)
	require.Equal(t, "", api.Stack)
	require.Equal(t, "qwen3", api.ModelFamily)
	require.Nil(t, api.Thinking)
	require.Empty(t, api.Bindings)
	require.Empty(t, api.Parameters)
	require.Nil(t, detect(t, srv.URL+"/v1/chat/completions", "openai", "unknown-model-x"), "nothing known at all: no api block")
}

// R1: detection never overrides an explicit modelFamily.
func TestDetectExplicitFamilyIsNeverOverridden(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v0/models", jsonHandler(map[string]any{"object": "list", "data": []map[string]any{{"id": "alias", "arch": "llama", "type": "llm"}}}))
	srv := httptest.NewServer(mux)
	defer srv.Close()
	api := newTestDetector().Detect(context.Background(), config.ModelConfig{ModelName: "alias", ApiType: "openai", ApiURL: srv.URL + "/v1/chat/completions", ModelFamily: "qwen3"})
	require.Equal(t, "lmstudio", api.Stack)
	require.Equal(t, "qwen3", api.ModelFamily)
	require.Equal(t, "chat_template_kwargs.enable_thinking", api.Bindings[system.IntentReasoningDisable].Param)
}

// R7: the cached spec is canonical; every caller gets its own copy.
func TestDetectReturnsIndependentCopies(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/props", jsonHandler(map[string]any{"total_slots": 1, "chat_template": qwen3Template}))
	srv := httptest.NewServer(mux)
	defer srv.Close()
	d := newTestDetector()
	cfg := config.ModelConfig{ModelName: "qwen3-32b", ApiType: "openai", ApiURL: srv.URL + "/v1/chat/completions"}
	first := d.Detect(context.Background(), cfg)
	first.DeclaredAt = 12345
	first.Bindings[system.IntentReasoningDisable].Param = "mutated"
	first.Parameters[0] = "mutated"
	second := d.Detect(context.Background(), cfg)
	require.Zero(t, second.DeclaredAt)
	require.Equal(t, "chat_template_kwargs.enable_thinking", second.Bindings[system.IntentReasoningDisable].Param)
	require.NotEqual(t, "mutated", second.Parameters[0])
	d.mu.Lock()
	for _, e := range d.cache {
		require.Zero(t, e.api.DeclaredAt)
	}
	d.mu.Unlock()
}

// A pass cut short by the caller's deadline must not pin its partial
// result for the cache TTL.
func TestDetectDoesNotCacheTimedOutPass(t *testing.T) {
	var calls int32
	mux := http.NewServeMux()
	mux.HandleFunc("/props", func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		jsonHandler(map[string]any{"total_slots": 1, "chat_template": qwen3Template})(w, r)
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()
	d := newTestDetector()
	cfg := config.ModelConfig{ModelName: "qwen3-32b", ApiType: "openai", ApiURL: srv.URL + "/v1/chat/completions"}
	expired, cancel := context.WithCancel(context.Background())
	cancel()
	partial := d.Detect(expired, cfg)
	require.NotNil(t, partial)
	require.Equal(t, "", partial.Stack, "no probe can succeed on an expired context")
	full := d.Detect(context.Background(), cfg)
	require.Equal(t, "llamacpp", full.Stack, "the partial result was not cached")
	require.Equal(t, int32(1), atomic.LoadInt32(&calls))
}

// R5 addendum (probe side): supported_reasoning_efforts from
// /model_group/info narrows the litellm effort enum and drops the disable
// knob when "none" is not offered.
func TestDetectLiteLLMSupportedReasoningEfforts(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/health/liveliness", func(w http.ResponseWriter, r *http.Request) { _, _ = w.Write([]byte(`"I'm alive!"`)) })
	mux.HandleFunc("/model_group/info", jsonHandler(map[string]any{"data": []map[string]any{{
		"model_group": "claude-sonnet-4-5", "supported_openai_params": []string{"temperature", "reasoning_effort"},
		"supports_reasoning": true, "supported_reasoning_efforts": []string{"low", "medium", "high"},
	}}}))
	srv := httptest.NewServer(mux)
	defer srv.Close()
	api := detect(t, srv.URL+"/v1/chat/completions", "openai", "claude-sonnet-4-5")
	require.Equal(t, "litellm", api.Stack)
	require.Equal(t, []string{"low", "medium", "high"}, api.Bindings[system.IntentReasoningEffort].EnumValues)
	require.Nil(t, api.Bindings[system.IntentReasoningDisable])
	require.NotNil(t, api.Bindings[system.IntentReasoningEnable])
}
