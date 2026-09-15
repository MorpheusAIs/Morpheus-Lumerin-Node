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

// requestLog records every request a fake server saw.
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

// loggingServer serves mux and records every request, handled or not (a
// 404 on a probe is still a request the upstream saw).
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

// litellmServer fakes a LiteLLM proxy: liveliness, one model group and the
// /model/info deployment list.
func litellmServer(t *testing.T, group map[string]any, deployments []map[string]any) (*httptest.Server, *requestLog) {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("/health/liveliness", func(w http.ResponseWriter, r *http.Request) { _, _ = w.Write([]byte(`"I'm alive!"`)) })
	mux.HandleFunc("/model_group/info", jsonHandler(map[string]any{"data": []map[string]any{group}}))
	mux.HandleFunc("/model/info", jsonHandler(map[string]any{"data": deployments}))
	return loggingServer(t, mux)
}

// registryServer fakes a Venice- or OpenRouter-shaped /api/v1/models listing.
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

// ollamaServer fakes an Ollama at an arbitrary (unrecognised) host: /api/tags
// plus /api/show for the model, with the given capabilities.
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

// vllmServer fakes a vLLM: /version plus a /v1/models listing serving id.
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

// mapHostToVendor makes stackFromHost recognise srvURL's host as vendor for
// the duration of the test (httptest hosts are never real vendor hostnames).
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

// requireLiteLLMVocabulary asserts R5(d): the bindings are exactly what a
// declared litellm preset advertises for referenceModel (a name of the
// family detection found) — nothing from the upstream's own vocabulary
// (venice_parameters.*, OpenRouter's reasoning.*, chat_template_kwargs.*,
// Ollama's native knobs or hints).
func requireLiteLLMVocabulary(t *testing.T, api *system.ModelApiSpec, referenceModel string) {
	t.Helper()
	require.Equal(t, "litellm", api.Stack)
	require.Equal(t, system.ApiSpecSourceDetected, api.Source)
	declared := apispec.Build(config.ModelConfig{ModelName: referenceModel, ApiType: "openai", ApiStack: "litellm"})
	require.NotNil(t, declared)
	require.Equal(t, declared.Bindings, api.Bindings, "bindings must be the litellm stack table's, nothing imported from the upstream")
	for intent, b := range api.Bindings {
		require.False(t, strings.HasPrefix(b.Param, "venice_parameters."), intent)
		require.False(t, strings.HasPrefix(b.Param, "reasoning."), intent)
		require.False(t, strings.HasPrefix(b.Param, "options."), intent)
		require.NotEqual(t, system.BindingKindTemplateKwarg, b.Kind, intent)
		require.NotEqual(t, system.BindingKindNativeBodyParam, b.Kind, intent)
		require.NotContains(t, b.Hint, "/api/chat", intent)
	}
}

// R5 happy path: LiteLLM -> openai/<model> at a Venice api_base -> Venice
// says the model reasons. Family from the upstream id, reasoning from the
// hop, bindings in LiteLLM's vocabulary, key never sent upstream. LiteLLM's
// own map says supports_reasoning=false: the upstream wins (OQ4, OR).
func TestDetectLiteLLMTwoHopVeniceReasoning(t *testing.T) {
	venice, veniceLog := registryServer(t, veniceEntry("qwen3-235b", true), 0)
	mapHostToVendor(t, venice.URL, "venice")
	litellm, _ := litellmServer(t, noReasoningGroup, []map[string]any{
		deployment("other", "openai/other-model", venice.URL+"/api/v1", nil),
		deployment("MY-CHAT", "openai/qwen3-235b", venice.URL+"/api/v1", map[string]any{"api_key": "sk-UPSTREAMSECRET"}),
	})

	api, trace := detectVia(t, litellm.URL, "my-chat", DefaultOptions())
	require.NotNil(t, api)
	require.Equal(t, "litellm", api.Stack)
	require.Equal(t, system.ApiSpecSourceDetected, api.Source)
	require.Equal(t, "qwen3", api.ModelFamily)
	require.Equal(t, system.ThinkingModeControllable, api.Thinking.Mode)
	require.Equal(t, "reasoning_effort", api.Bindings[system.IntentReasoningDisable].Param)
	requireLiteLLMVocabulary(t, api, "qwen3-235b")
	require.Equal(t, []string{"reasoning_effort", "temperature", "tools"}, api.Parameters, "LiteLLM's own list, not Venice capabilities")
	veniceLog.requireAnonymous(t)
	require.Contains(t, trace, "upstream venice listing reports reasoning support")
	require.NotContains(t, trace, "sk-UPSTREAMSECRET")
	require.NotContains(t, trace, "sk-litellm")
}

// registryServerMulti fakes a Venice- or OpenRouter-shaped /api/v1/models
// listing with more than one entry, unlike registryServer's single entry: it
// mirrors Venice's real listing shape (100+ models), where matchModelEntry's
// single-entry-list fallback cannot mask a matching bug.
func registryServerMulti(t *testing.T, entries ...map[string]any) (*httptest.Server, *requestLog) {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/models", jsonHandler(map[string]any{"data": entries}))
	return loggingServer(t, mux)
}

// Venice appends inline "key=value" parameters to the model id behind a
// LiteLLM deployment (observed live:
// "openai/deepseek-v4-pro:include_venice_system_prompt=false"). With a
// multi-entry listing (so the single-entry fallback cannot apply) the bare
// id must still match Venice's listing entry: family "deepseek" from the
// bare id, reasoning from the hop, bindings in LiteLLM's vocabulary, and the
// raw suffixed id traced exactly once.
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
	require.Equal(t, "litellm", api.Stack)
	require.Equal(t, system.ApiSpecSourceDetected, api.Source)
	require.Equal(t, "deepseek", api.ModelFamily, "family from the bare upstream id")
	require.Equal(t, system.ThinkingModeControllable, api.Thinking.Mode)
	require.Equal(t, "reasoning_effort", api.Bindings[system.IntentReasoningDisable].Param, "litellm's knob, not venice's venice_parameters.*")
	require.Equal(t, "reasoning_effort", api.Bindings[system.IntentReasoningEnable].Param)
	require.Equal(t, []string{"reasoning_effort", "temperature", "tools"}, api.Parameters, "LiteLLM's own list, not Venice capabilities")
	veniceLog.requireAnonymous(t)
	require.Contains(t, trace, "upstream venice listing reports reasoning support")
	require.Contains(t, trace, `upstream model "deepseek-v4-pro:include_venice_system_prompt=false"`, "the raw suffixed id is still traced once")
}

func TestDetectLiteLLMTwoHopOpenRouterKeepsLiteLLMVocabulary(t *testing.T) {
	openrouter, orLog := registryServer(t, map[string]any{"id": "deepseek/deepseek-chat-v3.1", "supported_parameters": []string{"temperature", "reasoning", "tools"}}, 0)
	mapHostToVendor(t, openrouter.URL, "openrouter")
	litellm, _ := litellmServer(t, noReasoningGroup, []map[string]any{deployment("my-chat", "openrouter/deepseek/deepseek-chat-v3.1", openrouter.URL+"/api/v1", nil)})

	api, _ := detectVia(t, litellm.URL, "my-chat", DefaultOptions())
	require.Equal(t, "litellm", api.Stack)
	require.Equal(t, "deepseek-v3.1", api.ModelFamily)
	require.Equal(t, system.ThinkingModeControllable, api.Thinking.Mode)
	require.Equal(t, "reasoning_effort", api.Bindings[system.IntentReasoningDisable].Param, "litellm's knob, not OpenRouter's reasoning.effort")
	require.Nil(t, api.Bindings[system.IntentSamplingTopA], "no OpenRouter stack-table import")
	require.NotContains(t, api.Parameters, "reasoning", "no OpenRouter parameter import")
	requireLiteLLMVocabulary(t, api, "deepseek/deepseek-chat-v3.1")
	orLog.requireAnonymous(t)
}

// Upstream rejects the anonymous request: the hop is skipped silently and
// the bare upstream id still decides the family.
func TestDetectLiteLLMTwoHopUpstream401FamilyFromIdOnly(t *testing.T) {
	venice, veniceLog := registryServer(t, nil, http.StatusUnauthorized)
	mapHostToVendor(t, venice.URL, "venice")
	litellm, _ := litellmServer(t, noReasoningGroup, []map[string]any{deployment("my-chat", "openai/qwen3-235b", venice.URL+"/api/v1", nil)})

	api, trace := detectVia(t, litellm.URL, "my-chat", DefaultOptions())
	require.Equal(t, "litellm", api.Stack)
	require.Equal(t, "qwen3", api.ModelFamily)
	require.Nil(t, api.Thinking)
	require.Nil(t, api.Bindings[system.IntentReasoningDisable])
	require.Contains(t, trace, "HTTP 401")
	veniceLog.requireAnonymous(t)
}

func TestDetectLiteLLMTwoHopUpstreamIdRefinesFamily(t *testing.T) {
	litellm, _ := litellmServer(t, noReasoningGroup, []map[string]any{deployment("my-chat", "anthropic/claude-opus-4-6", "", nil)})
	api, _ := detectVia(t, litellm.URL, "my-chat", DefaultOptions())
	require.Equal(t, "claude", api.ModelFamily)
	require.Nil(t, api.Bindings[system.IntentReasoningEnable], "claude's native knobs never leak through litellm")

	litellm2, _ := litellmServer(t, noReasoningGroup, []map[string]any{deployment("my-chat", "bedrock/converse/deepseek-r1-distill", "", nil)})
	r1, _ := detectVia(t, litellm2.URL, "my-chat", DefaultOptions())
	require.Equal(t, "deepseek-r1", r1.ModelFamily)
	require.Equal(t, system.ThinkingModeAlwaysOn, r1.Thinking.Mode)
}

// A documented cloud / hosted-vendor prefix (bedrock/, azure/, groq/, ...)
// is recognised only so the bare model id can be taken: its upstream is
// never probed, even when the deployment carries an api_base.
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

	// anthropic/ likewise: a hosted vendor, no second hop.
	litellm2, _ := litellmServer(t, noReasoningGroup, []map[string]any{deployment("my-chat", "anthropic/claude-opus-4-6", upstream.URL, nil)})
	api, trace = detectVia(t, litellm2.URL, "my-chat", DefaultOptions())
	require.Equal(t, "claude", api.ModelFamily)
	require.Empty(t, upLog.seen())
	require.Contains(t, trace, `upstream "anthropic" is a hosted vendor; no second hop`)
}

// R5(c): at most one hop — an upstream that is itself LiteLLM is never
// followed. At an unrecognised openai/ host the engine fingerprint (OQ3)
// identifies it by its liveliness endpoint and stops there: its model group
// and deployments are never read, no credential is ever sent.
func TestDetectLiteLLMTwoHopUpstreamLiteLLMIsNotFollowed(t *testing.T) {
	t.Run("openai prefix at an unrecognised host", func(t *testing.T) {
		upstream, upLog := litellmServer(t, noReasoningGroup, []map[string]any{deployment("qwen3-32b", "openai/qwen3-32b", "", nil)})
		front, _ := litellmServer(t, noReasoningGroup, []map[string]any{deployment("my-chat", "openai/qwen3-32b", upstream.URL+"/v1", nil)})
		api, trace := detectVia(t, front.URL, "my-chat", DefaultOptions())
		require.Equal(t, "litellm", api.Stack)
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
	require.Equal(t, "litellm", api.Stack)
	require.Equal(t, "deepseek-r1", api.ModelFamily)
	require.Equal(t, system.ThinkingModeAlwaysOn, api.Thinking.Mode)
	require.Nil(t, api.Bindings[system.IntentReasoningBudget], "vllm's thinking_token_budget never leaks through litellm")
	require.Nil(t, api.Bindings[system.IntentSamplingTopK], "no vllm stack-table import")
	require.Equal(t, []string{"reasoning_effort", "temperature", "tools"}, api.Parameters, "LiteLLM's own list, not vllm's")
	require.Contains(t, trace, `upstream vllm serves "deepseek-ai/DeepSeek-R1"`)
	vllmLog.requireAnonymous(t)
}

func TestDetectLiteLLMTwoHopCustomProviderWinsOverPrefix(t *testing.T) {
	vllm, vllmLog := vllmServer(t, "Qwen/Qwen3-32B")
	litellm, _ := litellmServer(t, noReasoningGroup, []map[string]any{deployment("my-chat", "foo/bar", vllm.URL, map[string]any{"custom_llm_provider": "hosted_vllm"})})

	api, trace := detectVia(t, litellm.URL, "my-chat", DefaultOptions())
	require.Equal(t, "qwen3", api.ModelFamily)
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

// openai/ without api_base is OpenAI itself: nothing to probe, the bare id
// still names the family.
func TestDetectLiteLLMTwoHopNoApiBaseIsNotProbed(t *testing.T) {
	litellm, _ := litellmServer(t, noReasoningGroup, []map[string]any{deployment("my-chat", "openai/gpt-4o", "", nil)})
	api, trace := detectVia(t, litellm.URL, "my-chat", DefaultOptions())
	require.Equal(t, "litellm", api.Stack)
	require.Equal(t, "gpt", api.ModelFamily)
	require.Contains(t, trace, "has no api_base; upstream not probed")
}

// R5(e): api_base is redacted and an upstream key is never printed.
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

// Same credentialed api_base form as above, but the host is not mapped to
// any vendor: the second hop takes the "unrecognised host" fingerprint
// branch, whose "fingerprinting engines at …" line prints parsed base
// candidates rather than going through RedactURL directly — this pins that
// it still carries no userinfo either.
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

// OQ2: an ollama/ or ollama_chat/ upstream is read once, anonymously:
// /api/show gives the architecture (family) and the thinking capability
// says the model reasons — as gateway evidence, never as Ollama's own
// capability path, so the bindings stay litellm's.
func TestDetectLiteLLMTwoHopOllamaCapability(t *testing.T) {
	for _, prefix := range []string{"ollama", "ollama_chat"} {
		t.Run(prefix, func(t *testing.T) {
			ollama, ollamaLog := ollamaServer(t, "qwen3", []string{"completion", "tools", "thinking"})
			litellm, _ := litellmServer(t, noReasoningGroup, []map[string]any{deployment("my-chat", prefix+"/qwen3:32b", ollama.URL, nil)})

			api, trace := detectVia(t, litellm.URL, "my-chat", DefaultOptions())
			require.NotNil(t, api)
			require.Equal(t, "qwen3", api.ModelFamily)
			require.Equal(t, system.ThinkingModeControllable, api.Thinking.Mode)
			requireLiteLLMVocabulary(t, api, "qwen3:32b")
			require.Nil(t, api.Bindings[system.IntentContextNumCtx], "no ollama stack-table import")
			require.Equal(t, "", api.Bindings[system.IntentReasoningDisable].Hint, "litellm's knob, not ollama's think toggle")
			require.Equal(t, []string{"reasoning_effort", "temperature", "tools"}, api.Parameters, "LiteLLM's own list, not ollama's")
			require.True(t, ollamaLog.sawPath("/api/show"))
			ollamaLog.requireAnonymous(t)
			require.Contains(t, trace, `upstream ollama reports architecture "qwen3"`)
			require.Contains(t, trace, "upstream ollama reports the thinking capability")
			require.NotContains(t, trace, "chat template", "the upstream's template is never evidence")
		})
	}

	t.Run("without the thinking capability", func(t *testing.T) {
		ollama, ollamaLog := ollamaServer(t, "llama", []string{"completion", "tools"})
		litellm, _ := litellmServer(t, noReasoningGroup, []map[string]any{deployment("my-chat", "ollama_chat/some-alias", ollama.URL, nil)})

		api, _ := detectVia(t, litellm.URL, "my-chat", DefaultOptions())
		require.Equal(t, "litellm", api.Stack)
		require.Equal(t, "llama", api.ModelFamily, "family from the upstream architecture")
		require.Nil(t, api.Thinking)
		require.Nil(t, api.Bindings[system.IntentReasoningEffort])
		ollamaLog.requireAnonymous(t)
	})
}

// OQ3: openai/ at a host that is no recognised vendor is fingerprinted like
// a self-hosted engine, anonymously; only family / reasoning facts cross the
// hop and the spec's stack stays litellm.
func TestDetectLiteLLMTwoHopUnknownHostIsFingerprinted(t *testing.T) {
	t.Run("ollama behind openai/", func(t *testing.T) {
		ollama, ollamaLog := ollamaServer(t, "qwen3", []string{"thinking"})
		litellm, _ := litellmServer(t, noReasoningGroup, []map[string]any{deployment("my-chat", "openai/some-alias", ollama.URL+"/v1", nil)})

		api, trace := detectVia(t, litellm.URL, "my-chat", DefaultOptions())
		require.Equal(t, "qwen3", api.ModelFamily, "family from the upstream architecture")
		require.Equal(t, system.ThinkingModeControllable, api.Thinking.Mode)
		requireLiteLLMVocabulary(t, api, "qwen3")
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
		require.Equal(t, "litellm", api.Stack)
		require.Equal(t, "deepseek-r1", api.ModelFamily, "family from the upstream's served id")
		require.Equal(t, system.ThinkingModeAlwaysOn, api.Thinking.Mode)
		require.Nil(t, api.Bindings[system.IntentSamplingTopK], "no vllm stack-table import")
		require.Equal(t, []string{"reasoning_effort", "temperature", "tools"}, api.Parameters)
		vllmLog.requireAnonymous(t)
		require.Contains(t, trace, `upstream vllm serves "deepseek-ai/DeepSeek-R1"`)
	})

	t.Run("nothing answers", func(t *testing.T) {
		dead, deadLog := loggingServer(t, http.NewServeMux())
		litellm, _ := litellmServer(t, noReasoningGroup, []map[string]any{deployment("my-chat", "openai/qwen3-32b", dead.URL+"/v1", nil)})

		api, trace := detectVia(t, litellm.URL, "my-chat", DefaultOptions())
		require.Equal(t, "litellm", api.Stack)
		require.Equal(t, "qwen3", api.ModelFamily, "family from the bare upstream id")
		require.Nil(t, api.Thinking)
		deadLog.requireAnonymous(t)
		require.Contains(t, trace, "no engine fingerprint matched at the upstream")
	})
}

// Hop budget: a slow upstream cannot cost hop-1 its evidence — the hop is
// cut off after its own budget, the trace says so and the litellm result
// (stack, parameters, family from the bare id) stands.
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
	require.Equal(t, "qwen3", api.ModelFamily)
	require.Nil(t, api.Thinking, "no upstream evidence arrived in time")
	require.Equal(t, []string{"reasoning_effort", "temperature", "tools"}, api.Parameters, "hop-1 evidence stands")
	require.Contains(t, trace, "second hop timed out after 150ms; hop-1 evidence stands")
	slowLog.requireAnonymous(t)

	// The detection itself was not cut short: the result is cacheable.
	cached := d.Detect(context.Background(), cfg)
	require.Equal(t, "litellm", cached.Stack)
	d.mu.Lock()
	require.Len(t, d.cache, 1, "a hop timeout is not a detection timeout")
	d.mu.Unlock()
}

// R5 addendum end-to-end with the hop in place: supported_reasoning_efforts
// still narrows the litellm effort enum (OQ1 value rule included) after
// probeLiteLLM falls through to the second hop.
func TestDetectLiteLLMTwoHopKeepsSupportedReasoningEfforts(t *testing.T) {
	venice, _ := registryServer(t, veniceEntry("qwen3-235b", true), 0)
	mapHostToVendor(t, venice.URL, "venice")
	group := map[string]any{
		"model_group": "my-chat", "supported_openai_params": []string{"temperature", "reasoning_effort"},
		"supports_reasoning": true, "supported_reasoning_efforts": []string{"low", "medium", "high"},
	}
	litellm, llLog := litellmServer(t, group, []map[string]any{deployment("my-chat", "openai/qwen3-235b", venice.URL+"/api/v1", nil)})

	d := NewDetector(lib.NewTestLogger(), DefaultOptions())
	api := d.Detect(context.Background(), config.ModelConfig{ModelName: "my-chat", ApiType: "openai", ApiURL: litellm.URL + "/v1/chat/completions", ApiKey: "sk-litellm"})
	require.NotNil(t, api)
	require.Equal(t, "litellm", api.Stack)
	require.Equal(t, "qwen3", api.ModelFamily, "the hop ran: family from the upstream id")
	require.Equal(t, []string{"low", "medium", "high"}, api.Bindings[system.IntentReasoningEffort].EnumValues)
	require.Nil(t, api.Bindings[system.IntentReasoningDisable], "none is not offered")
	require.Equal(t, "low", api.Bindings[system.IntentReasoningEnable].Value, "first supported effort other than none")
	require.Equal(t, system.ThinkingModeControllable, api.Thinking.Mode)
	require.True(t, llLog.sawPath("/model/info"))
}

// Controller ruling: hopAllowed gates the second hop's target by scheme and
// address before any request is sent. Non-http(s) schemes and the classic
// SSRF literal-IP ranges (unspecified, link-local unicast/multicast,
// multicast — cloud metadata, mDNS, ...) are refused, zoned IPv6 literals
// and v4-mapped IPv6 forms included. Loopback and private (RFC1918 and
// equivalent) addresses stay allowed: self-hosted vLLM/Ollama upstreams
// legitimately live there. A hostname passes the URL gate — what it resolves
// to is checked at dial time (TestHopDialControl).
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

// The dial-time gate: the second hop's dialer runs hopDialControl on the
// resolved address, so a hostname that resolves to a refused range (DNS
// rebinding, 169.254.169.254.nip.io and the like) is stopped before the
// connection is made. Exercised on address strings only — no DNS, no
// network.
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
	// fail closed on anything that is not an ip:port
	require.Error(t, hopDialControl("tcp", "not-an-address", nil))
	require.Error(t, hopDialControl("unix", "/var/run/x.sock", nil))
}

// The hop client is wired with the gate and the detector client's policies:
// its transport dials through hopDialControl (a literal refused address is
// rejected before any connection is attempted), and it keeps the per-probe
// timeout and the same-host redirect policy.
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

// roundTripperFunc adapts a function to http.RoundTripper.
type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

// Every request of the second hop goes through the hop client (and so
// through its gated dialer); the first hop's requests to the LiteLLM host
// never do.
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

// End-to-end: the gate is actually wired into the second hop. A refused
// api_base never reaches baseCandidates/fingerprintEngines — the refusal is
// traced and the bare upstream id still names the family (hop-1 evidence is
// unaffected) — while a loopback or private-address api_base still runs the
// hop as before.
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
		require.NotContains(t, trace, "refused for the second hop")
		vllmLog.requireAnonymous(t)
	})
}

// The other side of the hop budget: when the whole detection deadline
// expires while the hop is in flight, the trace attributes the cut to the
// detection deadline (not to the hop's own timeout) and hop-1 evidence
// still stands.
func TestDetectLiteLLMTwoHopCutShortByDetectionDeadline(t *testing.T) {
	prev := hopTimeout
	hopTimeout = 10 * time.Second // never the binding limit here
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
	require.Equal(t, "qwen3", api.ModelFamily, "family from the bare upstream id")
	require.Nil(t, api.Thinking, "no upstream evidence arrived in time")
	require.Equal(t, []string{"reasoning_effort", "temperature", "tools"}, api.Parameters, "hop-1 evidence stands")
	require.Contains(t, trace, "second hop cut short by the detection deadline; hop-1 evidence stands")
	require.NotContains(t, trace, "second hop timed out after")
	slowLog.requireAnonymous(t)
}

// /model/info is read under the probe body cap like every service
// endpoint: a listing larger than maxProbeBody is ignored (traced), so no
// deployment is found and the hop is skipped — the upstream is never
// dialled and the family comes from the configured name alone.
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
