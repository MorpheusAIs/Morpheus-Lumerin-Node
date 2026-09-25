package apidetect

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/apispec"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/config"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/lib"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/system"
	"github.com/stretchr/testify/require"
)

type requestLog struct {
	mu      sync.Mutex
	paths   []string
	auth    []string
	headers []http.Header
}

func (l *requestLog) record(r *http.Request) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.paths = append(l.paths, r.URL.Path)
	l.auth = append(l.auth, r.Header.Get("Authorization"))
	l.headers = append(l.headers, r.Header.Clone())
}

func (l *requestLog) sawPath(p string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	for _, seen := range l.paths {
		if seen == p {
			return true
		}
	}
	return false
}

func (l *requestLog) seen() []string {
	l.mu.Lock()
	defer l.mu.Unlock()
	return append([]string(nil), l.paths...)
}

func (l *requestLog) requireAnonymous(t *testing.T) {
	t.Helper()
	l.mu.Lock()
	defer l.mu.Unlock()
	require.NotEmpty(t, l.paths, "the upstream was expected to be probed")
	for i, a := range l.auth {
		require.Equal(t, "", a, "request %d (%s) carried a credential", i, l.paths[i])
		for _, h := range []string{"Cookie", "X-Api-Key", "Api-Key"} {
			require.Equal(t, "", l.headers[i].Get(h), "request %d (%s) carried a %s header", i, l.paths[i], h)
		}
	}
}

func loggingServer(t *testing.T, mux *http.ServeMux) (*httptest.Server, *requestLog) {
	t.Helper()
	log := &requestLog{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.record(r)
		mux.ServeHTTP(w, r)
	}))
	t.Cleanup(srv.Close)
	return srv, log
}

func litellmServer(t *testing.T, group map[string]any, deployments []map[string]any) (*httptest.Server, *requestLog) {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("/health/liveliness", func(w http.ResponseWriter, r *http.Request) { _, _ = w.Write([]byte(`"I'm alive!"`)) })
	mux.HandleFunc("/model_group/info", jsonHandler(map[string]any{"data": []map[string]any{group}}))
	mux.HandleFunc("/model/info", jsonHandler(map[string]any{"data": deployments}))
	return loggingServer(t, mux)
}

func registryServer(t *testing.T, entry map[string]any, status int) (*httptest.Server, *requestLog) {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/models", func(w http.ResponseWriter, r *http.Request) {
		if status != 0 {
			w.WriteHeader(status)
			return
		}
		jsonHandler(map[string]any{"data": []map[string]any{entry}})(w, r)
	})
	return loggingServer(t, mux)
}

func ollamaServer(t *testing.T, arch string, capabilities []string) (*httptest.Server, *requestLog) {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("/api/tags", jsonHandler(map[string]any{"models": []map[string]any{{"name": "qwen3:32b"}}}))
	mux.HandleFunc("/api/show", jsonHandler(map[string]any{
		"template":     qwen3Template,
		"capabilities": capabilities,
		"model_info":   map[string]any{"general.architecture": arch},
	}))
	return loggingServer(t, mux)
}

func vllmServer(t *testing.T, id string) (*httptest.Server, *requestLog) {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("/version", jsonHandler(map[string]any{"version": "0.18.0"}))
	mux.HandleFunc("/v1/models", jsonHandler(map[string]any{"object": "list", "data": []map[string]any{{"id": id, "owned_by": "vllm"}}}))
	return loggingServer(t, mux)
}

func veniceEntry(id string, reasons bool) map[string]any {
	return map[string]any{"id": id, "model_spec": map[string]any{"capabilities": map[string]any{"supportsReasoning": reasons}}}
}

func mapHostToVendor(t *testing.T, srvURL, vendor string) {
	t.Helper()
	prev := stackFromHost
	want, err := url.Parse(srvURL)
	require.NoError(t, err)
	stackFromHost = func(u string) string {
		if p, err := url.Parse(u); err == nil && p.Host == want.Host {
			return vendor
		}
		return prev(u)
	}
	t.Cleanup(func() { stackFromHost = prev })
}

func deployment(name, model, apiBase string, extra map[string]any) map[string]any {
	params := map[string]any{"model": model}
	if apiBase != "" {
		params["api_base"] = apiBase
	}
	for k, v := range extra {
		params[k] = v
	}
	return map[string]any{"model_name": name, "litellm_params": params, "model_info": map[string]any{"id": "dep-" + name}}
}

var noReasoningGroup = map[string]any{"model_group": "my-chat", "supported_openai_params": []string{"temperature", "tools", "reasoning_effort"}, "supports_reasoning": false}

func detectVia(t *testing.T, litellmURL, modelName string, opts Options) (*system.ModelApiSpec, string) {
	t.Helper()
	d := NewDetector(lib.NewTestLogger(), opts)
	api, trace := d.DetectWithTrace(context.Background(), config.ModelConfig{ModelName: modelName, ApiType: "openai", ApiURL: litellmURL + "/v1/chat/completions", ApiKey: "sk-litellm"})
	return api, strings.Join(trace, "\n")
}

func requireLiteLLMVocabulary(t *testing.T, api *system.ModelApiSpec, referenceModel string, supported []string) {
	t.Helper()
	require.Equal(t, "litellm", api.Stack)
	require.Equal(t, "", api.Via)
	require.Equal(t, system.ApiSpecSourceDetected, api.Source)
	declared := apispec.Build(config.ModelConfig{ModelName: referenceModel, ApiType: "openai", ApiStack: "litellm"})
	require.NotNil(t, declared)
	listed := map[string]bool{}
	for _, p := range supported {
		listed[p] = true
	}
	standard := map[string]bool{}
	for _, p := range declared.Parameters {
		standard[p] = true
	}
	expected := map[string]*system.ParamBinding{}
	for intent, b := range declared.Bindings {
		if root, _, _ := strings.Cut(b.Param, "."); standard[root] && !listed[root] {
			continue
		}
		expected[intent] = b
	}
	require.Equal(t, expected, api.Bindings, "bindings must be the litellm stack table's (narrowed to what litellm forwards), nothing imported from the upstream")
	for intent, b := range api.Bindings {
		require.False(t, strings.HasPrefix(b.Param, "venice_parameters."), intent)
		require.False(t, strings.HasPrefix(b.Param, "reasoning."), intent)
		require.False(t, strings.HasPrefix(b.Param, "options."), intent)
		require.NotEqual(t, system.BindingKindTemplateKwarg, b.Kind, intent)
		require.NotEqual(t, system.BindingKindNativeBodyParam, b.Kind, intent)
		require.NotContains(t, b.Hint, "/api/chat", intent)
	}
}

func TestDetectLiteLLMTwoHopVeniceReasoning(t *testing.T) {
	venice, veniceLog := registryServer(t, veniceEntry("qwen3-235b", true), 0)
	mapHostToVendor(t, venice.URL, "venice")
	litellm, _ := litellmServer(t, noReasoningGroup, []map[string]any{
		deployment("other", "openai/other-model", venice.URL+"/api/v1", nil),
		deployment("MY-CHAT", "openai/qwen3-235b", venice.URL+"/api/v1", map[string]any{"api_key": "sk-UPSTREAMSECRET"}),
	})

	api, trace := detectVia(t, litellm.URL, "my-chat", DefaultOptions())
	require.NotNil(t, api)
	require.Equal(t, "venice", api.Stack)
	require.Equal(t, "litellm", api.Via)
	require.Equal(t, system.ApiSpecSourceDetected, api.Source)
	require.Equal(t, "qwen3", api.ModelFamily)
	require.Equal(t, system.ThinkingModeControllable, api.Thinking.Mode)
	require.Equal(t, "venice_parameters.disable_thinking", api.Bindings[system.IntentReasoningDisable].Param, "venice's knob: forwarded by litellm as a provider kwarg")
	require.Equal(t, "reasoning_effort", api.Bindings[system.IntentReasoningEffort].Param, "listed in supported_openai_params")
	require.Nil(t, api.Bindings[system.IntentReasoningEnable], "venice has no enable knob; litellm's is not borrowed")
	require.Nil(t, api.Bindings[system.IntentReasoningBudget], "litellm's thinking.budget_tokens is not venice's")
	require.Equal(t, "top_k", api.Bindings[system.IntentSamplingTopK].Param)
	require.Nil(t, api.Bindings[system.IntentResponseFormatJSON], "response_format is standard and not listed for the group")
	require.Nil(t, api.Bindings[system.IntentStreamIncludeUsage])
	require.Nil(t, api.Bindings[system.IntentToolsParallel])
	for intent, b := range api.Bindings {
		require.NotEqual(t, system.BindingKindTemplateKwarg, b.Kind, intent)
	}
	require.Equal(t, []string{"reasoning_effort", "temperature", "tools"}, api.Parameters, "venice's documented list ∩ litellm's supported list; never Venice's capability names")
	veniceLog.requireAnonymous(t)
	require.Contains(t, trace, "upstream venice listing reports reasoning support")
	require.Contains(t, trace, "upstream venice identified: it becomes the stack (via litellm)")
	require.NotContains(t, trace, "sk-UPSTREAMSECRET")
	require.NotContains(t, trace, "sk-litellm")
}

// Multi-entry listing, so matchModelEntry's single-entry fallback cannot
// mask a matching bug.
func registryServerMulti(t *testing.T, entries ...map[string]any) (*httptest.Server, *requestLog) {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/models", jsonHandler(map[string]any{"data": entries}))
	return loggingServer(t, mux)
}

func TestDetectLiteLLMTwoHopVeniceInlineParamsMatchByBareID(t *testing.T) {
	venice, veniceLog := registryServerMulti(t,
		veniceEntry("deepseek-v4-pro", true),
		veniceEntry("some-other-model", false),
	)
	mapHostToVendor(t, venice.URL, "venice")
	litellm, _ := litellmServer(t, noReasoningGroup, []map[string]any{
		deployment("my-chat", "openai/deepseek-v4-pro:include_venice_system_prompt=false", venice.URL+"/api/v1", nil),
	})

	api, trace := detectVia(t, litellm.URL, "my-chat", DefaultOptions())
	require.NotNil(t, api)
	require.Equal(t, "venice", api.Stack)
	require.Equal(t, "litellm", api.Via)
	require.Equal(t, system.ApiSpecSourceDetected, api.Source)
	require.Equal(t, "deepseek", api.ModelFamily, "family from the bare upstream id")
	require.Equal(t, system.ThinkingModeControllable, api.Thinking.Mode)
	require.Equal(t, "venice_parameters.disable_thinking", api.Bindings[system.IntentReasoningDisable].Param, "venice's knob, not litellm's reasoning_effort: none")
	require.Nil(t, api.Bindings[system.IntentReasoningEnable])
	require.Equal(t, []string{"reasoning_effort", "temperature", "tools"}, api.Parameters, "venice's documented list ∩ litellm's supported list, not Venice capabilities")
	veniceLog.requireAnonymous(t)
	require.Contains(t, trace, "upstream venice listing reports reasoning support")
	require.Contains(t, trace, `upstream model "deepseek-v4-pro:include_venice_system_prompt=false"`, "the raw suffixed id is still traced once")
}

func TestDetectLiteLLMTwoHopOpenRouterBecomesStack(t *testing.T) {
	openrouter, orLog := registryServer(t, map[string]any{"id": "deepseek/deepseek-chat-v3.1", "supported_parameters": []string{"temperature", "reasoning", "tools"}}, 0)
	mapHostToVendor(t, openrouter.URL, "openrouter")
	litellm, _ := litellmServer(t, noReasoningGroup, []map[string]any{deployment("my-chat", "openrouter/deepseek/deepseek-chat-v3.1", openrouter.URL+"/api/v1", nil)})

	api, trace := detectVia(t, litellm.URL, "my-chat", DefaultOptions())
	require.Equal(t, "openrouter", api.Stack)
	require.Equal(t, "litellm", api.Via)
	require.Equal(t, "deepseek-v3.1", api.ModelFamily)
	require.Equal(t, system.ThinkingModeControllable, api.Thinking.Mode)
	require.Equal(t, "reasoning.effort", api.Bindings[system.IntentReasoningDisable].Param, "OpenRouter's knob, forwarded verbatim")
	require.Equal(t, "reasoning.enabled", api.Bindings[system.IntentReasoningEnable].Param)
	require.Equal(t, "top_a", api.Bindings[system.IntentSamplingTopA].Param, "openrouter stack table, non-standard root")
	require.Nil(t, api.Bindings[system.IntentResponseFormatJSON], "response_format is standard and not listed for the group")
	for intent, b := range api.Bindings {
		require.NotEqual(t, system.BindingKindTemplateKwarg, b.Kind, intent)
	}
	require.Equal(t, []string{"reasoning_effort", "temperature", "tools"}, api.Parameters, "openrouter's documented list ∩ litellm's supported list; the listing's supported_parameters never cross")
	orLog.requireAnonymous(t)
	require.Contains(t, trace, "upstream openrouter identified: it becomes the stack (via litellm)")
}

func TestDetectLiteLLMTwoHopUpstream401FamilyFromIdOnly(t *testing.T) {
	venice, veniceLog := registryServer(t, nil, http.StatusUnauthorized)
	mapHostToVendor(t, venice.URL, "venice")
	litellm, _ := litellmServer(t, noReasoningGroup, []map[string]any{deployment("my-chat", "openai/qwen3-235b", venice.URL+"/api/v1", nil)})

	api, trace := detectVia(t, litellm.URL, "my-chat", DefaultOptions())
	require.Equal(t, "litellm", api.Stack, "the listing could not be read: the upstream is not identified")
	require.Equal(t, "", api.Via)
	require.Equal(t, "qwen3", api.ModelFamily)
	require.Nil(t, api.Thinking)
	require.Nil(t, api.Bindings[system.IntentReasoningDisable])
	require.Contains(t, trace, "HTTP 401")
	veniceLog.requireAnonymous(t)
}

func TestDetectLiteLLMTwoHopUpstreamIdRefinesFamily(t *testing.T) {
	litellm, _ := litellmServer(t, noReasoningGroup, []map[string]any{deployment("my-chat", "anthropic/claude-opus-4-6", "", nil)})
	api, _ := detectVia(t, litellm.URL, "my-chat", DefaultOptions())
	require.Equal(t, "anthropic", api.Stack)
	require.Equal(t, "litellm", api.Via)
	require.Equal(t, "claude", api.ModelFamily)
	require.Equal(t, "thinking", api.Bindings[system.IntentReasoningEnable].Param, "claude's native knob, identified via the litellm prefix alone")
	require.Equal(t, map[string]any{"type": "adaptive"}, api.Bindings[system.IntentReasoningEnable].Value)

	litellm2, _ := litellmServer(t, noReasoningGroup, []map[string]any{deployment("my-chat", "bedrock/converse/deepseek-r1-distill", "", nil)})
	r1, _ := detectVia(t, litellm2.URL, "my-chat", DefaultOptions())
	require.Equal(t, "litellm", r1.Stack, "bedrock/ is a documented cloud prefix: never becomes the stack")
	require.Equal(t, "", r1.Via)
	require.Equal(t, "deepseek-r1", r1.ModelFamily)
	require.Equal(t, system.ThinkingModeAlwaysOn, r1.Thinking.Mode)
}

func TestDetectLiteLLMTwoHopDocumentedCloudPrefixIsNotProbed(t *testing.T) {
	upstream, upLog := registryServer(t, veniceEntry("deepseek-r1-distill", true), 0)
	litellm, _ := litellmServer(t, noReasoningGroup, []map[string]any{deployment("my-chat", "bedrock/converse/deepseek-r1-distill", upstream.URL+"/api/v1", nil)})

	api, trace := detectVia(t, litellm.URL, "my-chat", DefaultOptions())
	require.Equal(t, "litellm", api.Stack)
	require.Equal(t, "deepseek-r1", api.ModelFamily, "family from the bare upstream id")
	require.Equal(t, system.ThinkingModeAlwaysOn, api.Thinking.Mode)
	require.Empty(t, upLog.seen(), "a hosted vendor / cloud upstream is never read")
	require.Contains(t, trace, `provider "bedrock"`)
	require.Contains(t, trace, "upstream not probed")

	litellm2, _ := litellmServer(t, noReasoningGroup, []map[string]any{deployment("my-chat", "anthropic/claude-opus-4-6", upstream.URL, nil)})
	api, trace = detectVia(t, litellm2.URL, "my-chat", DefaultOptions())
	require.Equal(t, "anthropic", api.Stack)
	require.Equal(t, "litellm", api.Via)
	require.Equal(t, "claude", api.ModelFamily)
	require.Empty(t, upLog.seen(), "anthropic is recognised by its litellm prefix alone; never read")
	require.Contains(t, trace, "upstream anthropic recognised")
}

func TestDetectLiteLLMTwoHopUpstreamLiteLLMIsNotFollowed(t *testing.T) {
	t.Run("openai prefix at an unrecognised host", func(t *testing.T) {
		upstream, upLog := litellmServer(t, noReasoningGroup, []map[string]any{deployment("qwen3-32b", "openai/qwen3-32b", "", nil)})
		front, _ := litellmServer(t, noReasoningGroup, []map[string]any{deployment("my-chat", "openai/qwen3-32b", upstream.URL+"/v1", nil)})
		api, trace := detectVia(t, front.URL, "my-chat", DefaultOptions())
		require.Equal(t, "litellm", api.Stack)
		require.Equal(t, "", api.Via, "a LiteLLM upstream is not an identified upstream stack")
		require.Equal(t, "qwen3", api.ModelFamily)
		require.Nil(t, api.Thinking, "nothing but the bare id is learned from a LiteLLM upstream")
		require.True(t, upLog.sawPath("/health/liveliness"), "the fingerprint identifies the upstream LiteLLM")
		for _, p := range []string{"/model_group/info", "/model/info"} {
			require.False(t, upLog.sawPath(p), "the upstream's LiteLLM endpoints must never be read: %s", p)
		}
		upLog.requireAnonymous(t)
		require.Contains(t, trace, "unrecognised host")
		require.Contains(t, trace, "itself a LiteLLM proxy; not followed")
	})

	t.Run("hosted_vllm prefix pointing at a LiteLLM", func(t *testing.T) {
		upstream, upLog := litellmServer(t, noReasoningGroup, []map[string]any{deployment("qwen3-32b", "openai/qwen3-32b", "", nil)})
		front, _ := litellmServer(t, noReasoningGroup, []map[string]any{deployment("my-chat", "hosted_vllm/qwen3-32b", upstream.URL, nil)})
		api, _ := detectVia(t, front.URL, "my-chat", DefaultOptions())
		require.Equal(t, "litellm", api.Stack)
		require.Equal(t, "", api.Via)
		require.Equal(t, "qwen3", api.ModelFamily)
		for _, p := range []string{"/model_group/info", "/model/info"} {
			require.False(t, upLog.sawPath(p), "the upstream's LiteLLM endpoints must never be read: %s", p)
		}
		upLog.requireAnonymous(t)
	})
}

func TestDetectLiteLLMTwoHopVLLMServedID(t *testing.T) {
	vllm, vllmLog := vllmServer(t, "deepseek-ai/DeepSeek-R1")
	litellm, _ := litellmServer(t, noReasoningGroup, []map[string]any{deployment("my-chat", "hosted_vllm/local-model", vllm.URL, nil)})

	api, trace := detectVia(t, litellm.URL, "my-chat", DefaultOptions())
	require.Equal(t, "vllm", api.Stack)
	require.Equal(t, "litellm", api.Via)
	require.Equal(t, "deepseek-r1", api.ModelFamily)
	require.Equal(t, system.ThinkingModeAlwaysOn, api.Thinking.Mode)
	require.Nil(t, api.Bindings[system.IntentReasoningBudget], "an always-on family takes no budget knob")
	require.Equal(t, "include_reasoning", api.Bindings[system.IntentReasoningFormat].Param)
	require.Equal(t, "top_k", api.Bindings[system.IntentSamplingTopK].Param, "vllm's stack table, non-standard root")
	require.Nil(t, api.Bindings[system.IntentResponseFormatJSON], "standard root not listed for the group")
	require.Equal(t, []string{"reasoning_effort", "temperature", "tools"}, api.Parameters, "vllm's documented list ∩ litellm's supported list")
	require.Contains(t, trace, `upstream vllm serves "deepseek-ai/DeepSeek-R1"`)
	require.Contains(t, trace, "upstream vllm identified: it becomes the stack (via litellm)")
	vllmLog.requireAnonymous(t)
}

func TestDetectLiteLLMTwoHopCustomProviderWinsOverPrefix(t *testing.T) {
	vllm, vllmLog := vllmServer(t, "Qwen/Qwen3-32B")
	litellm, _ := litellmServer(t, noReasoningGroup, []map[string]any{deployment("my-chat", "foo/bar", vllm.URL, map[string]any{"custom_llm_provider": "hosted_vllm"})})

	api, trace := detectVia(t, litellm.URL, "my-chat", DefaultOptions())
	require.Equal(t, "qwen3", api.ModelFamily)
	require.Equal(t, "vllm", api.Stack)
	require.Equal(t, "litellm", api.Via)
	require.Contains(t, trace, `provider "hosted_vllm", upstream model "foo/bar"`, "the prefix is not the provider, so the whole model string is the upstream id")
	vllmLog.requireAnonymous(t)
}

func TestDetectLiteLLMTwoHopUnknownPrefixIsNotProbed(t *testing.T) {
	venice, veniceLog := registryServer(t, veniceEntry("qwen3-32b", true), 0)
	mapHostToVendor(t, venice.URL, "venice")
	litellm, _ := litellmServer(t, noReasoningGroup, []map[string]any{deployment("my-chat", "myprovider/qwen3-32b", venice.URL+"/api/v1", nil)})

	api, trace := detectVia(t, litellm.URL, "my-chat", DefaultOptions())
	require.Equal(t, "qwen3", api.ModelFamily)
	require.Nil(t, api.Thinking)
	require.Empty(t, veniceLog.seen())
	require.Contains(t, trace, `provider "myprovider" is not in the documented prefix table`)
}

func TestDetectLiteLLMTwoHopNoMatchingDeployment(t *testing.T) {
	litellm, _ := litellmServer(t, noReasoningGroup, []map[string]any{deployment("someone-else", "openai/deepseek-r1", "", nil)})
	api, trace := detectVia(t, litellm.URL, "qwen3-32b", DefaultOptions())
	require.Equal(t, "litellm", api.Stack)
	require.Equal(t, "qwen3", api.ModelFamily, "family from the configured name only")
	require.Nil(t, api.Thinking)
	require.Contains(t, trace, `none is named "qwen3-32b"`)
}

func TestDetectLiteLLMTwoHopFirstCaseInsensitiveMatchWins(t *testing.T) {
	litellm, _ := litellmServer(t, noReasoningGroup, []map[string]any{
		deployment("My-Chat", "openai/llama-3.3-70b", "", nil),
		deployment("my-chat", "openai/qwen3-32b", "", nil),
	})
	api, _ := detectVia(t, litellm.URL, "MY-CHAT", DefaultOptions())
	require.Equal(t, "llama", api.ModelFamily)
}

func TestDetectLiteLLMTwoHopNoApiBaseIsNotProbed(t *testing.T) {
	litellm, _ := litellmServer(t, noReasoningGroup, []map[string]any{deployment("my-chat", "openai/gpt-4o", "", nil)})
	api, trace := detectVia(t, litellm.URL, "my-chat", DefaultOptions())
	require.Equal(t, "litellm", api.Stack)
	require.Equal(t, "gpt", api.ModelFamily)
	require.Contains(t, trace, "has no api_base; upstream not probed")
}

func TestDetectLiteLLMTwoHopTraceRedaction(t *testing.T) {
	venice, _ := registryServer(t, veniceEntry("qwen3-235b", true), 0)
	mapHostToVendor(t, venice.URL, "venice")
	withCreds := strings.Replace(venice.URL, "http://", "http://user:URLSECRET@", 1) + "/api/v1?key=QUERYSECRET"
	litellm, _ := litellmServer(t, noReasoningGroup, []map[string]any{deployment("my-chat", "openai/qwen3-235b", withCreds, map[string]any{"api_key": "sk-UPSTREAMSECRET"})})

	_, trace := detectVia(t, litellm.URL, "my-chat", DefaultOptions())
	require.NotContains(t, trace, "URLSECRET")
	require.NotContains(t, trace, "QUERYSECRET")
	require.NotContains(t, trace, "sk-UPSTREAMSECRET")
	require.NotContains(t, trace, "sk-litellm")
	require.Contains(t, trace, RedactURL(withCreds))
}

func TestDetectLiteLLMTwoHopTraceRedactionUnrecognisedHost(t *testing.T) {
	dead, _ := loggingServer(t, http.NewServeMux())
	withCreds := strings.Replace(dead.URL, "http://", "http://user:URLSECRET@", 1) + "/v1?key=QUERYSECRET"
	litellm, _ := litellmServer(t, noReasoningGroup, []map[string]any{deployment("my-chat", "openai/qwen3-32b", withCreds, map[string]any{"api_key": "sk-UPSTREAMSECRET"})})

	_, trace := detectVia(t, litellm.URL, "my-chat", DefaultOptions())
	require.NotContains(t, trace, "URLSECRET")
	require.NotContains(t, trace, "QUERYSECRET")
	require.NotContains(t, trace, "sk-UPSTREAMSECRET")
	require.NotContains(t, trace, "sk-litellm")
	require.Contains(t, trace, RedactURL(withCreds))
	require.Contains(t, trace, "fingerprinting engines at", "the unrecognised-host branch line itself must be exercised")
}

func TestDetectLiteLLMTwoHopDisabled(t *testing.T) {
	venice, veniceLog := registryServer(t, veniceEntry("qwen3-235b", true), 0)
	mapHostToVendor(t, venice.URL, "venice")
	litellm, llLog := litellmServer(t, noReasoningGroup, []map[string]any{deployment("my-chat", "openai/qwen3-235b", venice.URL+"/api/v1", nil)})

	api, _ := detectVia(t, litellm.URL, "my-chat", Options{})
	require.Equal(t, "litellm", api.Stack)
	require.Equal(t, "", api.ModelFamily, "without the hop nothing names the family")
	require.False(t, llLog.sawPath("/model/info"))
	require.Empty(t, veniceLog.seen())
	require.True(t, DefaultOptions().TwoHop, "production wiring follows the hop")
}

func TestDetectLiteLLMTwoHopOllamaCapability(t *testing.T) {
	for _, prefix := range []string{"ollama", "ollama_chat"} {
		t.Run(prefix, func(t *testing.T) {
			ollama, ollamaLog := ollamaServer(t, "qwen3", []string{"completion", "tools", "thinking"})
			litellm, _ := litellmServer(t, noReasoningGroup, []map[string]any{deployment("my-chat", prefix+"/qwen3:32b", ollama.URL, nil)})

			api, trace := detectVia(t, litellm.URL, "my-chat", DefaultOptions())
			require.NotNil(t, api)
			require.Equal(t, "qwen3", api.ModelFamily)
			require.Equal(t, system.ThinkingModeControllable, api.Thinking.Mode)
			requireLiteLLMVocabulary(t, api, "qwen3:32b", noReasoningGroup["supported_openai_params"].([]string))
			require.Nil(t, api.Bindings[system.IntentContextNumCtx], "no ollama stack-table import")
			require.Equal(t, "", api.Bindings[system.IntentReasoningDisable].Hint, "litellm's knob, not ollama's think toggle")
			require.Equal(t, []string{"reasoning_effort", "temperature", "tools"}, api.Parameters, "LiteLLM's own list, not ollama's")
			require.True(t, ollamaLog.sawPath("/api/show"))
			ollamaLog.requireAnonymous(t)
			require.Contains(t, trace, `upstream ollama reports architecture "qwen3"`)
			require.Contains(t, trace, "upstream ollama reports the thinking capability")
			require.Contains(t, trace, "upstream ollama stays evidence-only")
			require.NotContains(t, trace, "chat template", "the upstream's template is never evidence")
		})
	}

	t.Run("without the thinking capability", func(t *testing.T) {
		ollama, ollamaLog := ollamaServer(t, "llama", []string{"completion", "tools"})
		litellm, _ := litellmServer(t, noReasoningGroup, []map[string]any{deployment("my-chat", "ollama_chat/some-alias", ollama.URL, nil)})

		api, _ := detectVia(t, litellm.URL, "my-chat", DefaultOptions())
		require.Equal(t, "litellm", api.Stack)
		require.Equal(t, "", api.Via)
		require.Equal(t, "llama", api.ModelFamily, "family from the upstream architecture")
		require.Nil(t, api.Thinking)
		require.Nil(t, api.Bindings[system.IntentReasoningEffort])
		ollamaLog.requireAnonymous(t)
	})
}

func TestDetectLiteLLMTwoHopUnknownHostIsFingerprinted(t *testing.T) {
	t.Run("ollama behind openai/", func(t *testing.T) {
		ollama, ollamaLog := ollamaServer(t, "qwen3", []string{"thinking"})
		litellm, _ := litellmServer(t, noReasoningGroup, []map[string]any{deployment("my-chat", "openai/some-alias", ollama.URL+"/v1", nil)})

		api, trace := detectVia(t, litellm.URL, "my-chat", DefaultOptions())
		require.Equal(t, "qwen3", api.ModelFamily, "family from the upstream architecture")
		require.Equal(t, system.ThinkingModeControllable, api.Thinking.Mode)
		requireLiteLLMVocabulary(t, api, "qwen3", noReasoningGroup["supported_openai_params"].([]string))
		require.Nil(t, api.Bindings[system.IntentContextNumCtx])
		require.Equal(t, []string{"reasoning_effort", "temperature", "tools"}, api.Parameters)
		ollamaLog.requireAnonymous(t)
		require.Contains(t, trace, "unrecognised host; fingerprinting engines")
		require.Contains(t, trace, "identified ollama by /api/tags shape")
		require.Contains(t, trace, `upstream ollama reports architecture "qwen3"`)
		require.NotContains(t, trace, "chat template")
	})

	t.Run("vllm behind openai/", func(t *testing.T) {
		vllm, vllmLog := vllmServer(t, "deepseek-ai/DeepSeek-R1")
		litellm, _ := litellmServer(t, noReasoningGroup, []map[string]any{deployment("my-chat", "openai/local-model", vllm.URL+"/v1", nil)})

		api, trace := detectVia(t, litellm.URL, "my-chat", DefaultOptions())
		require.Equal(t, "vllm", api.Stack, "fingerprinted engine becomes the stack")
		require.Equal(t, "litellm", api.Via)
		require.Equal(t, "deepseek-r1", api.ModelFamily, "family from the upstream's served id")
		require.Equal(t, system.ThinkingModeAlwaysOn, api.Thinking.Mode)
		require.Equal(t, "top_k", api.Bindings[system.IntentSamplingTopK].Param, "vllm stack table")
		require.Equal(t, []string{"reasoning_effort", "temperature", "tools"}, api.Parameters)
		vllmLog.requireAnonymous(t)
		require.Contains(t, trace, `upstream vllm serves "deepseek-ai/DeepSeek-R1"`)
	})

	t.Run("nothing answers", func(t *testing.T) {
		dead, deadLog := loggingServer(t, http.NewServeMux())
		litellm, _ := litellmServer(t, noReasoningGroup, []map[string]any{deployment("my-chat", "openai/qwen3-32b", dead.URL+"/v1", nil)})

		api, trace := detectVia(t, litellm.URL, "my-chat", DefaultOptions())
		require.Equal(t, "litellm", api.Stack)
		require.Equal(t, "", api.Via)
		require.Equal(t, "qwen3", api.ModelFamily, "family from the bare upstream id")
		require.Nil(t, api.Thinking)
		deadLog.requireAnonymous(t)
		require.Contains(t, trace, "no engine fingerprint matched at the upstream")
	})
}

func TestDetectLiteLLMTwoHopTimeoutKeepsFirstHopEvidence(t *testing.T) {
	prev := hopTimeout
	hopTimeout = 150 * time.Millisecond
	t.Cleanup(func() { hopTimeout = prev })

	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/models", func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-r.Context().Done():
		case <-time.After(5 * time.Second):
		}
	})
	slow, slowLog := loggingServer(t, mux)
	mapHostToVendor(t, slow.URL, "venice")
	litellm, _ := litellmServer(t, noReasoningGroup, []map[string]any{deployment("my-chat", "openai/qwen3-235b", slow.URL+"/api/v1", nil)})

	d := NewDetector(lib.NewTestLogger(), DefaultOptions())
	cfg := config.ModelConfig{ModelName: "my-chat", ApiType: "openai", ApiURL: litellm.URL + "/v1/chat/completions", ApiKey: "sk-litellm"}
	api, lines := d.DetectWithTrace(context.Background(), cfg)
	trace := strings.Join(lines, "\n")
	require.NotNil(t, api)
	require.Equal(t, "litellm", api.Stack)
	require.Equal(t, "", api.Via, "a timed-out hop identifies nothing")
	require.Equal(t, "qwen3", api.ModelFamily)
	require.Nil(t, api.Thinking, "no upstream evidence arrived in time")
	require.Equal(t, []string{"reasoning_effort", "temperature", "tools"}, api.Parameters, "hop-1 evidence stands")
	require.Contains(t, trace, "second hop timed out after 150ms; hop-1 evidence stands")
	slowLog.requireAnonymous(t)

	cached := d.Detect(context.Background(), cfg)
	require.Equal(t, "litellm", cached.Stack)
	d.mu.Lock()
	require.Len(t, d.cache, 1, "a hop timeout is not a detection timeout")
	d.mu.Unlock()
}

func TestDetectLiteLLMTwoHopKeepsSupportedReasoningEfforts(t *testing.T) {
	group := map[string]any{
		"model_group": "my-chat", "supported_openai_params": []string{"temperature", "reasoning_effort"},
		"supports_reasoning": true, "supported_reasoning_efforts": []string{"low", "medium", "high"},
	}

	t.Run("stack litellm", func(t *testing.T) {
		dead, _ := loggingServer(t, http.NewServeMux())
		litellm, llLog := litellmServer(t, group, []map[string]any{deployment("my-chat", "openai/qwen3-235b", dead.URL+"/v1", nil)})

		d := NewDetector(lib.NewTestLogger(), DefaultOptions())
		api := d.Detect(context.Background(), config.ModelConfig{ModelName: "my-chat", ApiType: "openai", ApiURL: litellm.URL + "/v1/chat/completions", ApiKey: "sk-litellm"})
		require.NotNil(t, api)
		require.Equal(t, "litellm", api.Stack)
		require.Equal(t, "", api.Via)
		require.Equal(t, "qwen3", api.ModelFamily, "the hop ran: family from the upstream id")
		require.Equal(t, []string{"low", "medium", "high"}, api.Bindings[system.IntentReasoningEffort].EnumValues)
		require.Nil(t, api.Bindings[system.IntentReasoningDisable], "none is not offered")
		require.Equal(t, "low", api.Bindings[system.IntentReasoningEnable].Value, "first supported effort other than none")
		require.Equal(t, system.ThinkingModeControllable, api.Thinking.Mode)
		require.True(t, llLog.sawPath("/model/info"))
	})

	t.Run("identified upstream keeps its own enum", func(t *testing.T) {
		venice, _ := registryServer(t, veniceEntry("qwen3-235b", true), 0)
		mapHostToVendor(t, venice.URL, "venice")
		litellm, _ := litellmServer(t, group, []map[string]any{deployment("my-chat", "openai/qwen3-235b", venice.URL+"/api/v1", nil)})

		api, _ := detectVia(t, litellm.URL, "my-chat", DefaultOptions())
		require.Equal(t, "venice", api.Stack)
		require.Equal(t, "litellm", api.Via)
		require.Equal(t, []string{"none", "minimal", "low", "medium", "high", "xhigh", "max"}, api.Bindings[system.IntentReasoningEffort].EnumValues, "venice's documented enum")
		require.Equal(t, "venice_parameters.disable_thinking", api.Bindings[system.IntentReasoningDisable].Param)
		require.Equal(t, system.ThinkingModeControllable, api.Thinking.Mode)
	})
}

func TestHopAllowed(t *testing.T) {
	cases := []struct {
		name    string
		apiBase string
		allowed bool
	}{
		{"non-http scheme (file)", "file:///x", false},
		{"non-http scheme (ftp)", "ftp://h", false},
		{"ipv4 link-local (cloud metadata)", "http://169.254.169.254/v1", false},
		{"ipv4 unspecified", "http://0.0.0.0:8000", false},
		{"ipv6 link-local unicast", "http://[fe80::1]:11434", false},
		{"ipv6 link-local unicast with zone", "http://[fe80::1%25eth0]:11434", false},
		{"ipv6 link-local multicast", "http://[ff02::1]", false},
		{"v4-mapped ipv6 link-local (cloud metadata)", "http://[::ffff:169.254.169.254]/v1", false},
		{"loopback", "http://127.0.0.1:11434", true},
		{"private rfc1918", "http://10.0.0.5:8000", true},
		{"hostname: localhost (DNS is checked at dial time)", "http://localhost:1", true},
		{"hostname resolving to link-local (DNS is checked at dial time)", "http://169.254.169.254.nip.io/", true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			require.Equal(t, c.allowed, hopAllowed(c.apiBase), c.apiBase)
		})
	}
}

func TestHopDialControl(t *testing.T) {
	refused := []string{"169.254.169.254:80", "[fe80::1%eth0]:80", "0.0.0.0:80", "[ff02::1]:80", "[::ffff:169.254.169.254]:80"}
	for _, addr := range refused {
		t.Run("refused "+addr, func(t *testing.T) {
			err := hopDialControl("tcp", addr, nil)
			require.ErrorIs(t, err, errHopAddrRefused, addr)
		})
	}
	allowed := []string{"127.0.0.1:80", "10.0.0.5:80", "[::1]:80"}
	for _, addr := range allowed {
		t.Run("allowed "+addr, func(t *testing.T) {
			require.NoError(t, hopDialControl("tcp", addr, nil), addr)
		})
	}
	require.Error(t, hopDialControl("tcp", "not-an-address", nil))
	require.Error(t, hopDialControl("unix", "/var/run/x.sock", nil))
}

func TestHopClientIsGatedAndKeepsClientPolicies(t *testing.T) {
	d := NewDetector(lib.NewTestLogger(), Options{TwoHop: true, ProbeTimeout: 123 * time.Millisecond})
	require.NotSame(t, d.client, d.hopClient)
	require.Equal(t, d.client.Timeout, d.hopClient.Timeout)
	require.NotNil(t, d.hopClient.CheckRedirect)

	tr, ok := d.hopClient.Transport.(*http.Transport)
	require.True(t, ok, "the hop client needs its own transport to carry the gated dialer")
	require.NotNil(t, tr.DialContext)
	_, err := tr.DialContext(context.Background(), "tcp", "169.254.169.254:1")
	require.ErrorIs(t, err, errHopAddrRefused)
}

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestDetectLiteLLMTwoHopUsesHopClient(t *testing.T) {
	venice, veniceLog := registryServer(t, veniceEntry("qwen3-235b", true), 0)
	mapHostToVendor(t, venice.URL, "venice")
	litellm, _ := litellmServer(t, noReasoningGroup, []map[string]any{deployment("my-chat", "openai/qwen3-235b", venice.URL+"/api/v1", nil)})

	d := NewDetector(lib.NewTestLogger(), DefaultOptions())
	var mu sync.Mutex
	var viaHopClient []string
	inner := d.hopClient.Transport
	d.hopClient.Transport = roundTripperFunc(func(r *http.Request) (*http.Response, error) {
		mu.Lock()
		viaHopClient = append(viaHopClient, r.URL.Host)
		mu.Unlock()
		return inner.RoundTrip(r)
	})

	api, _ := d.DetectWithTrace(context.Background(), config.ModelConfig{ModelName: "my-chat", ApiType: "openai", ApiURL: litellm.URL + "/v1/chat/completions", ApiKey: "sk-litellm"})
	require.NotNil(t, api)
	require.Equal(t, "qwen3", api.ModelFamily)
	require.Equal(t, "venice", api.Stack)
	require.Equal(t, "litellm", api.Via)
	require.Equal(t, system.ThinkingModeControllable, api.Thinking.Mode, "the hop ran and its evidence arrived")
	veniceLog.requireAnonymous(t)

	litellmHost := strings.TrimPrefix(litellm.URL, "http://")
	veniceHost := strings.TrimPrefix(venice.URL, "http://")
	mu.Lock()
	defer mu.Unlock()
	require.NotEmpty(t, viaHopClient, "the upstream must be reached through the hop client")
	for _, host := range viaHopClient {
		require.Equal(t, veniceHost, host, "only the upstream is dialled through the hop client")
		require.NotEqual(t, litellmHost, host)
	}
}

func TestDetectLiteLLMTwoHopHopTargetGate(t *testing.T) {
	refused := []struct{ name, apiBase string }{
		{"non-http scheme (file)", "file:///x"},
		{"non-http scheme (ftp)", "ftp://h"},
		{"ipv4 link-local (cloud metadata)", "http://169.254.169.254/v1"},
		{"ipv4 unspecified", "http://0.0.0.0:8000"},
		{"ipv6 link-local unicast", "http://[fe80::1]:11434"},
		{"ipv6 link-local multicast", "http://[ff02::1]"},
	}
	for _, c := range refused {
		t.Run(c.name, func(t *testing.T) {
			litellm, _ := litellmServer(t, noReasoningGroup, []map[string]any{deployment("my-chat", "openai/qwen3-32b", c.apiBase, nil)})
			api, trace := detectVia(t, litellm.URL, "my-chat", DefaultOptions())
			require.Equal(t, "litellm", api.Stack)
			require.Equal(t, "", api.Via)
			require.Equal(t, "qwen3", api.ModelFamily, "bare-id family evidence still applies")
			require.Nil(t, api.Thinking, "the hop never ran")
			require.Contains(t, trace, "refused for the second hop (scheme/address)")
			require.Contains(t, trace, RedactURL(c.apiBase))
			require.NotContains(t, trace, "litellm openai-compatible upstream at an unrecognised host", "no probe may start")
		})
	}

	t.Run("allowed: loopback (httptest server)", func(t *testing.T) {
		vllm, vllmLog := vllmServer(t, "deepseek-ai/DeepSeek-R1")
		litellm, _ := litellmServer(t, noReasoningGroup, []map[string]any{deployment("my-chat", "hosted_vllm/local-model", vllm.URL, nil)})
		api, trace := detectVia(t, litellm.URL, "my-chat", DefaultOptions())
		require.Equal(t, "deepseek-r1", api.ModelFamily, "the hop reached the loopback upstream")
		require.Equal(t, "vllm", api.Stack)
		require.Equal(t, "litellm", api.Via)
		require.NotContains(t, trace, "refused for the second hop")
		vllmLog.requireAnonymous(t)
	})
}

func TestDetectLiteLLMTwoHopCutShortByDetectionDeadline(t *testing.T) {
	prev := hopTimeout
	hopTimeout = 10 * time.Second // so the detection deadline, not the hop timeout, cuts the hop
	t.Cleanup(func() { hopTimeout = prev })

	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/models", func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-r.Context().Done():
		case <-time.After(5 * time.Second):
		}
	})
	slow, slowLog := loggingServer(t, mux)
	mapHostToVendor(t, slow.URL, "venice")
	litellm, _ := litellmServer(t, noReasoningGroup, []map[string]any{deployment("my-chat", "openai/qwen3-235b", slow.URL+"/api/v1", nil)})

	ctx, cancel := context.WithTimeout(context.Background(), 400*time.Millisecond)
	defer cancel()
	d := NewDetector(lib.NewTestLogger(), DefaultOptions())
	api, lines := d.DetectWithTrace(ctx, config.ModelConfig{ModelName: "my-chat", ApiType: "openai", ApiURL: litellm.URL + "/v1/chat/completions", ApiKey: "sk-litellm"})
	trace := strings.Join(lines, "\n")
	require.NotNil(t, api)
	require.Equal(t, "litellm", api.Stack)
	require.Equal(t, "", api.Via)
	require.Equal(t, "qwen3", api.ModelFamily, "family from the bare upstream id")
	require.Nil(t, api.Thinking, "no upstream evidence arrived in time")
	require.Equal(t, []string{"reasoning_effort", "temperature", "tools"}, api.Parameters, "hop-1 evidence stands")
	require.Contains(t, trace, "second hop cut short by the detection deadline; hop-1 evidence stands")
	require.NotContains(t, trace, "second hop timed out after")
	slowLog.requireAnonymous(t)
}

func TestDetectLiteLLMTwoHopModelInfoOverBodyCapSkipsHop(t *testing.T) {
	upstream, upLog := registryServer(t, veniceEntry("deepseek-r1", true), 0)
	mapHostToVendor(t, upstream.URL, "venice")

	mux := http.NewServeMux()
	mux.HandleFunc("/health/liveliness", func(w http.ResponseWriter, r *http.Request) { _, _ = w.Write([]byte(`"I'm alive!"`)) })
	mux.HandleFunc("/model_group/info", jsonHandler(map[string]any{"data": []map[string]any{noReasoningGroup}}))
	mux.HandleFunc("/model/info", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		dep, err := json.Marshal(deployment("my-chat", "openai/deepseek-r1", upstream.URL+"/api/v1", nil))
		require.NoError(t, err)
		_, _ = w.Write([]byte(`{"data":[` + string(dep) + `],"pad":"` + strings.Repeat("p", maxProbeBody+1024) + `"}`))
	})
	litellm, _ := loggingServer(t, mux)

	api, trace := detectVia(t, litellm.URL, "my-chat", DefaultOptions())
	require.NotNil(t, api)
	require.Equal(t, "litellm", api.Stack)
	require.Equal(t, "", api.ModelFamily, "the bare upstream id was never read; the alias names no family")
	require.Nil(t, api.Thinking)
	require.Equal(t, []string{"reasoning_effort", "temperature", "tools"}, api.Parameters, "hop-1 evidence stands")
	require.Contains(t, trace, fmt.Sprintf("/model/info -> HTTP 200, but the body exceeds the %d-byte limit; ignored", maxProbeBody))
	require.NotContains(t, trace, "litellm /model/info: deployment")
	require.Empty(t, upLog.seen(), "the hop never ran")
}

var liveGroup = map[string]any{
	"model_group":             "venice-deepseek-v4-pro",
	"supported_openai_params": []string{"temperature", "tools", "response_format", "stream_options", "max_tokens", "user"},
	"supports_reasoning":      false,
}

func withReasoningEffort(group map[string]any) map[string]any {
	out := make(map[string]any, len(group))
	for k, v := range group {
		out[k] = v
	}
	out["supported_openai_params"] = append([]string{"reasoning_effort"}, group["supported_openai_params"].([]string)...)
	return out
}

func requireNoReasoningEffortBinding(t *testing.T, api *system.ModelApiSpec) {
	t.Helper()
	for intent, b := range api.Bindings {
		root, _, _ := strings.Cut(b.Param, ".")
		require.NotEqual(t, "reasoning_effort", root, "%s must not be rooted at reasoning_effort", intent)
	}
}

func requireSubset(t *testing.T, list, allowed []string) {
	t.Helper()
	set := map[string]bool{}
	for _, a := range allowed {
		set[a] = true
	}
	for _, p := range list {
		require.True(t, set[p], "%q is not in %v", p, allowed)
	}
}

func TestDetectLiteLLMTwoHopIdentifiedUpstreamBecomesStack(t *testing.T) {
	t.Run("venice, reasoning_effort not forwarded", func(t *testing.T) {
		venice, veniceLog := registryServerMulti(t, veniceEntry("deepseek-v4-pro", true), veniceEntry("llama-3.3-70b", false))
		mapHostToVendor(t, venice.URL, "venice")
		litellm, _ := litellmServer(t, liveGroup, []map[string]any{
			deployment("venice-deepseek-v4-pro", "openai/deepseek-v4-pro:include_venice_system_prompt=false", venice.URL+"/api/v1", map[string]any{"api_key": "sk-UPSTREAMSECRET"}),
		})

		api, trace := detectVia(t, litellm.URL, "venice-deepseek-v4-pro", DefaultOptions())
		require.NotNil(t, api)
		require.Equal(t, "venice", api.Stack)
		require.Equal(t, "litellm", api.Via)
		require.Equal(t, system.ApiSpecSourceDetected, api.Source)
		require.Equal(t, "deepseek", api.ModelFamily)
		require.Equal(t, system.ThinkingModeControllable, api.Thinking.Mode)
		disable := api.Bindings[system.IntentReasoningDisable]
		require.NotNil(t, disable)
		require.Equal(t, "venice_parameters.disable_thinking", disable.Param)
		require.Equal(t, true, disable.Value)
		require.Equal(t, "venice_parameters.strip_thinking_response", api.Bindings[system.IntentReasoningFormat].Param)
		require.Nil(t, api.Bindings[system.IntentReasoningEffort], "reasoning_effort is standard and not in supported_openai_params")
		requireNoReasoningEffortBinding(t, api)
		require.Equal(t, "top_k", api.Bindings[system.IntentSamplingTopK].Param, "non-standard root: forwarded")
		require.Equal(t, "response_format", api.Bindings[system.IntentResponseFormatJSON].Param, "standard root in the supported list")
		require.Equal(t, "stream_options.include_usage", api.Bindings[system.IntentStreamIncludeUsage].Param)
		require.Nil(t, api.Bindings[system.IntentToolsParallel], "parallel_tool_calls is standard and not supported")
		require.Nil(t, api.Bindings[system.IntentReasoningBudget], "litellm's thinking.budget_tokens is not venice's knob")
		require.Equal(t, []string{"max_tokens", "response_format", "stream_options", "temperature", "tools", "user"}, api.Parameters)
		requireSubset(t, api.Parameters, liveGroup["supported_openai_params"].([]string))
		require.NotContains(t, api.Parameters, "reasoning", "venice's capability list never crosses the hop")

		veniceLog.requireAnonymous(t)
		require.NotContains(t, trace, "sk-UPSTREAMSECRET")
		require.NotContains(t, trace, "sk-litellm")
		require.Contains(t, trace, "upstream venice listing reports reasoning support")
		require.Contains(t, trace, "upstream venice identified: it becomes the stack (via litellm)")
		require.Contains(t, trace, "bindings: reasoning.effort dropped — litellm does not forward reasoning_effort for this model")
		require.Contains(t, trace, "bindings: tools.parallel dropped — litellm does not forward parallel_tool_calls for this model")
	})

	t.Run("venice, reasoning_effort forwarded", func(t *testing.T) {
		venice, _ := registryServerMulti(t, veniceEntry("deepseek-v4-pro", true), veniceEntry("llama-3.3-70b", false))
		mapHostToVendor(t, venice.URL, "venice")
		litellm, _ := litellmServer(t, withReasoningEffort(liveGroup), []map[string]any{
			deployment("venice-deepseek-v4-pro", "openai/deepseek-v4-pro:include_venice_system_prompt=false", venice.URL+"/api/v1", nil),
		})

		api, _ := detectVia(t, litellm.URL, "venice-deepseek-v4-pro", DefaultOptions())
		require.Equal(t, "venice", api.Stack)
		require.Equal(t, "litellm", api.Via)
		effort := api.Bindings[system.IntentReasoningEffort]
		require.NotNil(t, effort, "venice's reasoning.effort binding survives when LiteLLM forwards reasoning_effort")
		require.Equal(t, "reasoning_effort", effort.Param)
		require.Equal(t, []string{"none", "minimal", "low", "medium", "high", "xhigh", "max"}, effort.EnumValues, "venice's documented enum")
		require.Equal(t, "venice_parameters.disable_thinking", api.Bindings[system.IntentReasoningDisable].Param)
		require.Contains(t, api.Parameters, "reasoning_effort")
		require.Equal(t, system.ThinkingModeControllable, api.Thinking.Mode)
	})

	t.Run("hosted_vllm upstream becomes vllm", func(t *testing.T) {
		vllm, vllmLog := vllmServer(t, "Qwen/Qwen3-32B")
		litellm, _ := litellmServer(t, liveGroup, []map[string]any{deployment("venice-deepseek-v4-pro", "hosted_vllm/local-model", vllm.URL, nil)})

		api, trace := detectVia(t, litellm.URL, "venice-deepseek-v4-pro", DefaultOptions())
		require.Equal(t, "vllm", api.Stack)
		require.Equal(t, "litellm", api.Via)
		require.Equal(t, "qwen3", api.ModelFamily, "family from the upstream's served id")
		disable := api.Bindings[system.IntentReasoningDisable]
		require.NotNil(t, disable)
		require.Equal(t, system.BindingKindTemplateKwarg, disable.Kind, "vllm honours the family's template kwargs")
		require.Equal(t, "chat_template_kwargs.enable_thinking", disable.Param)
		require.Equal(t, system.ThinkingModeControllable, api.Thinking.Mode)
		require.Equal(t, "thinking_token_budget", api.Bindings[system.IntentReasoningBudget].Param)
		require.Equal(t, "top_k", api.Bindings[system.IntentSamplingTopK].Param)
		require.Equal(t, "stream_options.include_usage", api.Bindings[system.IntentStreamIncludeUsage].Param, "stream_options is in the supported list")
		require.Nil(t, api.Bindings[system.IntentToolsParallel])
		requireNoReasoningEffortBinding(t, api)
		requireSubset(t, api.Parameters, liveGroup["supported_openai_params"].([]string))
		vllmLog.requireAnonymous(t)
		require.Contains(t, trace, `upstream vllm serves "Qwen/Qwen3-32B"`)

		narrow := map[string]any{"model_group": "venice-deepseek-v4-pro", "supported_openai_params": []string{"temperature"}, "supports_reasoning": false}
		litellm2, _ := litellmServer(t, narrow, []map[string]any{deployment("venice-deepseek-v4-pro", "hosted_vllm/local-model", vllm.URL, nil)})
		api, _ = detectVia(t, litellm2.URL, "venice-deepseek-v4-pro", DefaultOptions())
		require.Equal(t, "vllm", api.Stack)
		require.Nil(t, api.Bindings[system.IntentStreamIncludeUsage])
		require.Equal(t, "chat_template_kwargs.enable_thinking", api.Bindings[system.IntentReasoningDisable].Param)
		require.Equal(t, []string{"temperature"}, api.Parameters)
	})

	t.Run("openai upstream by hostname, no request needed", func(t *testing.T) {
		upstream, upLog := loggingServer(t, http.NewServeMux())
		mapHostToVendor(t, upstream.URL, "openai")
		litellm, _ := litellmServer(t, withReasoningEffort(liveGroup), []map[string]any{deployment("venice-deepseek-v4-pro", "openai/o4-mini", upstream.URL+"/v1", nil)})

		api, trace := detectVia(t, litellm.URL, "venice-deepseek-v4-pro", DefaultOptions())
		require.Equal(t, "openai", api.Stack)
		require.Equal(t, "litellm", api.Via)
		require.Equal(t, "o-series", api.ModelFamily)
		require.Equal(t, system.ThinkingModeTunable, api.Thinking.Mode)
		require.Equal(t, "reasoning_effort", api.Bindings[system.IntentReasoningEffort].Param, "o-series' native knob on the openai preset, forwarded")
		requireSubset(t, api.Parameters, withReasoningEffort(liveGroup)["supported_openai_params"].([]string))
		require.Empty(t, upLog.seen(), "openai is identified by hostname alone: nothing is read from it")
		require.Contains(t, trace, "upstream openai recognised by hostname")
	})

	t.Run("ollama upstream stays evidence-only", func(t *testing.T) {
		for _, prefix := range []string{"ollama", "ollama_chat"} {
			ollama, ollamaLog := ollamaServer(t, "qwen3", []string{"thinking"})
			litellm, _ := litellmServer(t, withReasoningEffort(liveGroup), []map[string]any{deployment("venice-deepseek-v4-pro", prefix+"/qwen3:32b", ollama.URL, nil)})

			api, trace := detectVia(t, litellm.URL, "venice-deepseek-v4-pro", DefaultOptions())
			require.Equal(t, "litellm", api.Stack, prefix)
			require.Equal(t, "", api.Via, prefix)
			require.Equal(t, "qwen3", api.ModelFamily)
			require.Equal(t, system.ThinkingModeControllable, api.Thinking.Mode)
			require.Equal(t, "reasoning_effort", api.Bindings[system.IntentReasoningDisable].Param, "litellm's knob")
			require.Nil(t, api.Bindings[system.IntentContextNumCtx], "no ollama stack-table import")
			ollamaLog.requireAnonymous(t)
			require.Contains(t, trace, "upstream ollama stays evidence-only")
		}
	})

	t.Run("unidentified upstream keeps stack litellm with the filter", func(t *testing.T) {
		dead, deadLog := loggingServer(t, http.NewServeMux())
		litellm, _ := litellmServer(t, liveGroup, []map[string]any{deployment("venice-deepseek-v4-pro", "openai/qwen3-32b", dead.URL+"/v1", nil)})

		api, trace := detectVia(t, litellm.URL, "venice-deepseek-v4-pro", DefaultOptions())
		require.Equal(t, "litellm", api.Stack)
		require.Equal(t, "", api.Via)
		require.Equal(t, "qwen3", api.ModelFamily)
		requireNoReasoningEffortBinding(t, api)
		require.Equal(t, "response_format", api.Bindings[system.IntentResponseFormatJSON].Param)
		require.Nil(t, api.Bindings[system.IntentToolsParallel])
		require.Equal(t, []string{"max_tokens", "response_format", "stream_options", "temperature", "tools", "user"}, api.Parameters)
		deadLog.requireAnonymous(t)
		require.Contains(t, trace, "no engine fingerprint matched at the upstream")

		reasons := map[string]any{"model_group": "venice-deepseek-v4-pro", "supported_openai_params": []string{"temperature"}, "supports_reasoning": true}
		litellm2, _ := litellmServer(t, reasons, []map[string]any{deployment("venice-deepseek-v4-pro", "openai/qwen3-32b", dead.URL+"/v1", nil)})
		api, _ = detectVia(t, litellm2.URL, "venice-deepseek-v4-pro", DefaultOptions())
		require.Equal(t, "litellm", api.Stack)
		requireNoReasoningEffortBinding(t, api)
		require.Nil(t, api.Bindings[system.IntentReasoningDisable])
		require.Equal(t, "thinking.budget_tokens", api.Bindings[system.IntentReasoningBudget].Param)
		require.Equal(t, system.ThinkingModeTunable, api.Thinking.Mode, "derived after the filter")
	})

	t.Run("supported list unavailable: only reasoning_effort is dropped", func(t *testing.T) {
		venice, _ := registryServerMulti(t, veniceEntry("deepseek-v4-pro", true), veniceEntry("llama-3.3-70b", false))
		mapHostToVendor(t, venice.URL, "venice")
		mux := http.NewServeMux()
		mux.HandleFunc("/health/liveliness", func(w http.ResponseWriter, r *http.Request) { _, _ = w.Write([]byte(`"I'm alive!"`)) })
		mux.HandleFunc("/model_group/info", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusUnauthorized) })
		mux.HandleFunc("/model/info", jsonHandler(map[string]any{"data": []map[string]any{
			deployment("venice-deepseek-v4-pro", "openai/deepseek-v4-pro", venice.URL+"/api/v1", nil),
		}}))
		litellm, _ := loggingServer(t, mux)

		api, trace := detectVia(t, litellm.URL, "venice-deepseek-v4-pro", DefaultOptions())
		require.Equal(t, "venice", api.Stack)
		require.Equal(t, "litellm", api.Via)
		require.Equal(t, "venice_parameters.disable_thinking", api.Bindings[system.IntentReasoningDisable].Param)
		require.Nil(t, api.Bindings[system.IntentReasoningEffort])
		requireNoReasoningEffortBinding(t, api)
		require.Equal(t, "response_format", api.Bindings[system.IntentResponseFormatJSON].Param, "kept: nothing says it is not forwarded")
		require.Equal(t, "parallel_tool_calls", api.Bindings[system.IntentToolsParallel].Param)
		require.NotContains(t, api.Parameters, "reasoning_effort")
		require.Contains(t, api.Parameters, "messages", "venice's documented list minus reasoning_effort")
		require.Equal(t, system.ThinkingModeControllable, api.Thinking.Mode)
		require.Contains(t, trace, "litellm supported_openai_params unavailable")
	})

	t.Run("anthropic upstream becomes the stack (prefix or an anthropic-mapped host)", func(t *testing.T) {
		upstream, upLog := loggingServer(t, http.NewServeMux())
		for _, dep := range []map[string]any{
			deployment("venice-deepseek-v4-pro", "anthropic/claude-opus-4-6", upstream.URL, nil),
			deployment("venice-deepseek-v4-pro", "openai/claude-opus-4-6", upstream.URL+"/v1", nil),
		} {
			if strings.HasPrefix(dep["litellm_params"].(map[string]any)["model"].(string), "openai/") {
				mapHostToVendor(t, upstream.URL, "anthropic")
			}
			litellm, _ := litellmServer(t, withReasoningEffort(liveGroup), []map[string]any{dep})
			api, trace := detectVia(t, litellm.URL, "venice-deepseek-v4-pro", DefaultOptions())
			require.Equal(t, "anthropic", api.Stack)
			require.Equal(t, "litellm", api.Via)
			require.Equal(t, "claude", api.ModelFamily)
			require.Equal(t, "thinking", api.Bindings[system.IntentReasoningDisable].Param, "claude family default, non-standard root: forwarded")
			require.Equal(t, map[string]any{"type": "disabled"}, api.Bindings[system.IntentReasoningDisable].Value)
			require.Nil(t, api.Bindings[system.IntentResponseFormatGrammar])
			for _, b := range api.Bindings {
				require.NotEqual(t, system.BindingKindTemplateKwarg, b.Kind, "anthropic's own vocabulary, never a chat-template kwarg")
			}
			require.Equal(t, []string{"max_tokens", "reasoning_effort", "response_format", "stream_options", "temperature", "tools", "user"}, api.Parameters, "litellm's list verbatim (sorted), not intersected with the anthropic table")
			require.NotContains(t, api.Parameters, "stop_sequences", "anthropic's own params are not re-added to litellm's list")
			require.Empty(t, upLog.seen(), "anthropic is never read")
			require.Contains(t, trace, "upstream anthropic recognised")
			require.Contains(t, trace, "upstream anthropic identified: it becomes the stack (via litellm)")
		}
	})

	t.Run("anthropic upstream needs no api_base", func(t *testing.T) {
		litellm, _ := litellmServer(t, liveGroup, []map[string]any{deployment("venice-deepseek-v4-pro", "anthropic/claude-sonnet-4-5", "", nil)})
		api, trace := detectVia(t, litellm.URL, "venice-deepseek-v4-pro", DefaultOptions())
		require.Equal(t, "anthropic", api.Stack)
		require.Equal(t, "litellm", api.Via)
		require.Equal(t, "claude", api.ModelFamily)
		require.Equal(t, map[string]any{"type": "disabled"}, api.Bindings[system.IntentReasoningDisable].Value)
		require.Equal(t, []string{"max_tokens", "response_format", "stream_options", "temperature", "tools", "user"}, api.Parameters)
		require.NotContains(t, trace, "has no api_base")
	})

	t.Run("an identified stack survives a transport mismatch (via is litellm's transport, not the upstream's)", func(t *testing.T) {
		venice, _ := registryServerMulti(t, veniceEntry("deepseek-v4-pro", true), veniceEntry("llama-3.3-70b", false))
		mapHostToVendor(t, venice.URL, "venice")
		litellm, _ := litellmServer(t, liveGroup, []map[string]any{deployment("venice-deepseek-v4-pro", "openai/deepseek-v4-pro", venice.URL+"/api/v1", nil)})

		d := NewDetector(lib.NewTestLogger(), DefaultOptions())
		api, lines := d.DetectWithTrace(context.Background(), config.ModelConfig{ModelName: "venice-deepseek-v4-pro", ApiType: "claudeai", ApiURL: litellm.URL + "/v1/messages", ApiKey: "sk-litellm"})
		require.NotNil(t, api)
		require.Equal(t, "venice", api.Stack)
		require.Equal(t, "litellm", api.Via)
		require.Equal(t, "deepseek", api.ModelFamily)
		require.NotEmpty(t, api.Bindings)
		require.NotContains(t, strings.Join(lines, "\n"), "stack dropped")
	})

	t.Run("declared preset is untouched", func(t *testing.T) {
		d := NewDetector(lib.NewTestLogger(), DefaultOptions())
		api := d.Detect(context.Background(), config.ModelConfig{ModelName: "qwen3-32b", ApiType: "openai", ApiStack: "litellm", ApiURL: "http://h/v1/chat/completions"})
		require.Equal(t, "litellm", api.Stack)
		require.Equal(t, "", api.Via)
		require.Equal(t, system.ApiSpecSourceDeclared, api.Source)
		require.Equal(t, "reasoning_effort", api.Bindings[system.IntentReasoningEffort].Param)
		require.Equal(t, "parallel_tool_calls", api.Bindings[system.IntentToolsParallel].Param)
		require.Equal(t, system.ThinkingModeControllable, api.Thinking.Mode)
	})
}
