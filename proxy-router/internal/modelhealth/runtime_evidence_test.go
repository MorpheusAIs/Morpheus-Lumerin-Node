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
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/apiprobe"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/apispec"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/blockchainapi/structs"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/config"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/lib"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/system"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
)

// chatBackend is an httptest chat-completions endpoint whose answer the test
// swaps between sweeps. A 200 answers the arithmetic prompt correctly.
type chatBackend struct {
	mu        sync.Mutex
	status    int
	headers   map[string]string
	model     string
	reasoning string
	errorBody string
	server    *httptest.Server
	// requests records every request (path and decoded body) in order.
	requests []probeRequest
	// script, when set, answers a request before the default handling; ok
	// false falls through to it.
	script func(r probeRequest) (status int, body string, ok bool)
}

type probeRequest struct {
	path string
	body map[string]any
}

func (r probeRequest) has(key string) bool { _, ok := r.body[key]; return ok }

func (r probeRequest) firstContent() string {
	msgs, _ := r.body["messages"].([]any)
	if len(msgs) == 0 {
		return ""
	}
	msg, _ := msgs[0].(map[string]any)
	if s, ok := msg["content"].(string); ok {
		return s
	}
	parts, _ := msg["content"].([]any)
	for _, p := range parts {
		if part, ok := p.(map[string]any); ok && part["type"] == "text" {
			return part["text"].(string)
		}
	}
	return ""
}

func (r probeRequest) hasImagePart() bool {
	msgs, _ := r.body["messages"].([]any)
	if len(msgs) == 0 {
		return false
	}
	parts, _ := msgs[0].(map[string]any)["content"].([]any)
	for _, p := range parts {
		if part, ok := p.(map[string]any); ok && part["type"] == "image_url" {
			return true
		}
	}
	return false
}

func newChatBackend(t *testing.T) *chatBackend {
	t.Helper()
	b := &chatBackend{status: http.StatusOK}
	b.server = httptest.NewServer(http.HandlerFunc(b.serve))
	t.Cleanup(b.server.Close)
	return b
}

func (b *chatBackend) set(status int, model, reasoning, errorBody string, headers map[string]string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.status, b.model, b.reasoning, b.errorBody, b.headers = status, model, reasoning, errorBody, headers
}

func (b *chatBackend) serve(w http.ResponseWriter, r *http.Request) {
	b.mu.Lock()
	defer b.mu.Unlock()
	var body map[string]any
	_ = json.NewDecoder(r.Body).Decode(&body)
	req := probeRequest{path: r.URL.Path, body: body}
	b.requests = append(b.requests, req)
	for k, v := range b.headers {
		w.Header().Set(k, v)
	}
	w.Header().Set("Content-Type", "application/json")
	if b.script != nil {
		if status, resp, ok := b.script(req); ok {
			w.WriteHeader(status)
			_, _ = w.Write([]byte(resp))
			return
		}
	}
	if b.status != http.StatusOK {
		w.WriteHeader(b.status)
		_, _ = w.Write([]byte(b.errorBody))
		return
	}
	var x, y int
	_, _ = fmt.Sscanf(req.firstContent(), "What is %d+%d?", &x, &y)
	message := map[string]any{"role": "assistant", "content": fmt.Sprintf("%d", x+y)}
	if b.reasoning != "" {
		message["reasoning_content"] = b.reasoning
	}
	resp := map[string]any{
		"id": "chatcmpl-1", "object": "chat.completion",
		"choices": []any{map[string]any{"index": 0, "message": message, "finish_reason": "stop"}},
	}
	if b.model != "" {
		resp["model"] = b.model
	}
	_ = json.NewEncoder(w).Encode(resp)
}

func (b *chatBackend) adapter() aiengine.AIEngineStream {
	return aiengine.NewOpenAIEngine("configured-name", b.server.URL+chatPath, "", time.Minute, &lib.LoggerMock{}, nil)
}

const chatPath = "/v1/chat/completions"

func (b *chatBackend) recorded() []probeRequest {
	b.mu.Lock()
	defer b.mu.Unlock()
	return append([]probeRequest(nil), b.requests...)
}

func (b *chatBackend) setScript(script func(r probeRequest) (int, string, bool)) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.script = script
}

// runtimeDeps wires one bid-holding LLM model to the backend; det nil means
// detection is off.
func runtimeDeps(b *chatBackend, cfg config.ModelConfig, det ApiDetector) *mockDeps {
	return &mockDeps{
		bids: []*structs.Bid{bidFor(modelLLM)}, tags: map[common.Hash][]string{modelLLM: {"llm"}},
		modelIDs: []common.Hash{modelLLM}, adapter: b.adapter(), apiDetect: det,
		configs: map[common.Hash]config.ModelConfig{modelLLM: cfg},
	}
}

var vllmCfg = config.ModelConfig{ModelName: "my-alias", ApiType: "openai", ApiURL: "http://llm:8000/v1/chat/completions"}

func detectedVLLMNoFamily() *system.ModelApiSpec {
	return &system.ModelApiSpec{Stack: "vllm", Source: system.ApiSpecSourceDetected}
}

func apiJSON(t *testing.T, api *system.ModelApiSpec) string {
	t.Helper()
	c := api.Clone()
	c.DeclaredAt = 0
	raw, err := json.Marshal(c)
	require.NoError(t, err)
	return string(raw)
}

func TestRuntimeServedModelIDRefinesFamily(t *testing.T) {
	backend := newChatBackend(t)
	backend.set(http.StatusOK, "qwen3-8b", "", "", nil)
	det := &fakeDetector{spec: detectedVLLMNoFamily()}
	checker := newTestChecker(runtimeDeps(backend, vllmCfg, det))
	checker.checkAll(context.Background(), common.Address{})

	llm := reportByID(t, checker.GetReports(), modelLLM)
	require.Equal(t, system.ModelHealthStatusHealthy, llm.Status)
	require.NotNil(t, llm.Api)
	require.Equal(t, "qwen3", llm.Api.ModelFamily, "the served model id names the family the detector could not")
	require.Equal(t, "", det.spec.ModelFamily, "the detector's spec is never written back")
}

func TestRuntimeReasoningSeenSetsReasoningCapability(t *testing.T) {
	backend := newChatBackend(t)
	backend.set(http.StatusOK, "", "the sum is easy", "", nil)
	checker := newTestChecker(runtimeDeps(backend, vllmCfg, &fakeDetector{spec: detectedVLLMNoFamily()}))
	checker.checkAll(context.Background(), common.Address{})

	llm := reportByID(t, checker.GetReports(), modelLLM)
	require.NotNil(t, llm.Api)
	require.NotNil(t, llm.Api.Capabilities)
	require.NotNil(t, llm.Api.Capabilities.Reasoning)
	require.True(t, *llm.Api.Capabilities.Reasoning)
	require.Equal(t, "health-probe:reasoning_seen", llm.Api.Evidence["capabilities.reasoning"])
}

func TestRuntimeLiteLLMHeadersComposeMinimalSpec(t *testing.T) {
	backend := newChatBackend(t)
	backend.set(http.StatusOK, "qwen3-32b", "", "", map[string]string{"X-LiteLLM-Version": "1.70.0"})
	cfg := vllmCfg
	cfg.ModelName = "qwen3-32b"
	checker := newTestChecker(runtimeDeps(backend, cfg, nil))
	checker.checkAll(context.Background(), common.Address{})

	llm := reportByID(t, checker.GetReports(), modelLLM)
	require.Equal(t, system.ModelHealthStatusHealthy, llm.Status)
	require.NotNil(t, llm.Api, "a LiteLLM proxy in the headers names the stack without detection")
	require.Equal(t, "litellm", llm.Api.Stack)
	require.Equal(t, system.ApiSpecSourceDetected, llm.Api.Source)
	require.Equal(t, "health-probe:headers", llm.Api.Evidence["stack"])
	require.Equal(t, "qwen3", llm.Api.ModelFamily)
	require.Nil(t, llm.Api.Thinking, "no reasoning evidence: no reasoning knob")
	for intent := range llm.Api.Bindings {
		require.NotContains(t, intent, "reasoning", "no reasoning evidence: no reasoning binding")
	}
	require.NotZero(t, llm.Api.DeclaredAt)
}

func TestRuntimeLiteLLMHeadersIgnoredWhenStackDetected(t *testing.T) {
	backend := newChatBackend(t)
	backend.set(http.StatusOK, "", "", "", map[string]string{"X-LiteLLM-Version": "1.50.0"})
	checker := newTestChecker(runtimeDeps(backend, vllmCfg, &fakeDetector{spec: detectedVLLMNoFamily()}))
	checker.checkAll(context.Background(), common.Address{})

	llm := reportByID(t, checker.GetReports(), modelLLM)
	require.Equal(t, system.ModelHealthStatusHealthy, llm.Status)
	require.NotNil(t, llm.Api)
	require.Equal(t, "vllm", llm.Api.Stack, "detection named the stack; the proxy's headers do not rename it")
	require.Equal(t, "", llm.Api.Evidence["stack"])
}

func TestRuntimeRejectedParamDropsBindingWithStableEvidence(t *testing.T) {
	const rejected = `{"error":{"message":"Unrecognized request argument supplied: chat_template_kwargs","type":"invalid_request_error"}}`
	backend := newChatBackend(t)
	backend.set(http.StatusBadRequest, "", "", rejected, nil)
	deps := runtimeDeps(backend, vllmCfg, &fakeDetector{spec: detectedQwen3()})
	checker := NewChecker(Deps{Adapters: deps, Bids: deps, Models: deps, ModelConfigs: deps, ApiDetect: deps.apiDetect, ApiDetectTTL: time.Hour},
		time.Hour, time.Second, 0, 0, lib.NewTestLogger())
	clock := time.Unix(1_700_000_000, 0)
	checker.runtime.now = func() time.Time { return clock }
	checker.checkAll(context.Background(), common.Address{})

	first := reportByID(t, checker.GetReports(), modelLLM)
	require.NotNil(t, first.Api)
	require.Nil(t, first.Api.Bindings[system.IntentReasoningDisable], "the backend refused the binding's parameter")
	require.Nil(t, first.Api.Thinking, "no reasoning binding remains")
	const evidence = "verify:rejected:chat_template_kwargs@1700000000"
	require.Equal(t, evidence, first.Api.Evidence[system.EvidenceKeyBinding(system.IntentReasoningDisable)])

	seeded := first
	seededAPI := first.Api.Clone()
	seededAPI.DeclaredAt = 12345
	seeded.Api = seededAPI
	checker.setReport(seeded)

	// A later plain 200 naming the served id adds a newer fact; the rejection's
	// evidence keeps its own timestamp and declaredAt stands.
	backend.set(http.StatusOK, "qwen3-8b", "", "", nil)
	clock = clock.Add(5 * time.Minute)
	checker.checkAll(context.Background(), common.Address{})
	second := reportByID(t, checker.GetReports(), modelLLM)
	require.Equal(t, system.ModelHealthStatusHealthy, second.Status)
	require.Equal(t, evidence, second.Api.Evidence[system.EvidenceKeyBinding(system.IntentReasoningDisable)])
	require.Equal(t, int64(12345), second.Api.DeclaredAt)
	require.Equal(t, apiJSON(t, seededAPI), apiJSON(t, second.Api))
}

func TestRuntimeStoreVerifierOnlyEntryLeavesAtUnset(t *testing.T) {
	store := newRuntimeStore(time.Minute)
	clock := time.Unix(1_700_000_000, 0)
	store.now = func() time.Time { return clock }
	supported := true
	store.byModel["m"] = &runtimeEntry{
		rt:         apispec.Runtime{ServedModelID: "qwen3-8b"},
		seen:       map[runtimeFact]time.Time{factServedModelID: clock},
		verify:     apispec.Runtime{ToolsSupported: &supported},
		verifiedAt: clock.Add(90 * time.Second),
		verifyKey:  "k",
	}

	rt, ok := store.harvest("m", nil)
	require.True(t, ok)
	require.Equal(t, clock.Unix(), rt.At)
	require.Equal(t, "qwen3-8b", rt.ServedModelID)

	// The harvested fact expires; the verifier's fact keeps the entry, and At
	// is left for MergeRuntime to fill with now rather than the zero time.
	clock = clock.Add(2 * time.Minute)
	rt, ok = store.harvest("m", nil)
	require.True(t, ok, "verifier facts expire on their own clock")
	require.Equal(t, "", rt.ServedModelID)
	require.NotNil(t, rt.ToolsSupported)
	require.Zero(t, rt.At)

	// Past its own TTL the verifier's fact goes too, and the entry with it.
	clock = clock.Add(time.Minute)
	_, ok = store.harvest("m", nil)
	require.False(t, ok)
	require.Empty(t, store.byModel)
}

func TestRuntimeContextFromErrorFillsOnlyUnset(t *testing.T) {
	const vllmError = `{"object":"error","message":"This model's maximum context length is 4096 tokens. However, you requested 5000 tokens (4000 in the messages, 1000 in the completion). Please reduce the length of the messages or completion.","type":"BadRequestError","code":400}`
	const vllmV1Error = `{"object":"error","message":"The input (5000 tokens) exceeds the model's maximum context length (8192)","type":"BadRequestError","code":400}`

	t.Run("fills an unset context", func(t *testing.T) {
		backend := newChatBackend(t)
		backend.set(http.StatusBadRequest, "", "", vllmV1Error, nil)
		checker := newTestChecker(runtimeDeps(backend, vllmCfg, &fakeDetector{spec: detectedVLLMNoFamily()}))
		checker.checkAll(context.Background(), common.Address{})

		llm := reportByID(t, checker.GetReports(), modelLLM)
		require.Equal(t, system.ModelHealthStatusUnhealthy, llm.Status, "the probe failed")
		require.Equal(t, http.StatusBadRequest, llm.HttpStatus)
		require.NotNil(t, llm.Api, "the api block is still published after a failed probe")
		require.NotNil(t, llm.Api.Capabilities)
		require.Equal(t, int64(8192), *llm.Api.Capabilities.Context)
		require.Equal(t, apiprobe.ContextSourceVLLM, llm.Api.Evidence["capabilities.context"])
	})

	t.Run("openai form", func(t *testing.T) {
		backend := newChatBackend(t)
		backend.set(http.StatusBadRequest, "", "", vllmError, nil)
		checker := newTestChecker(runtimeDeps(backend, vllmCfg, &fakeDetector{spec: detectedVLLMNoFamily()}))
		checker.checkAll(context.Background(), common.Address{})

		llm := reportByID(t, checker.GetReports(), modelLLM)
		require.Equal(t, int64(4096), *llm.Api.Capabilities.Context)
		require.Equal(t, "health-probe:context_length_exceeded", llm.Api.Evidence["capabilities.context"])
	})

	t.Run("leaves a set context alone", func(t *testing.T) {
		backend := newChatBackend(t)
		backend.set(http.StatusBadRequest, "", "", vllmV1Error, nil)
		spec := detectedVLLMNoFamily()
		served := int64(32768)
		spec.Capabilities = &system.ModelCapabilities{Context: &served}
		spec.Evidence = map[string]string{"capabilities.context": "vllm:/v1/models.max_model_len"}
		checker := newTestChecker(runtimeDeps(backend, vllmCfg, &fakeDetector{spec: spec}))
		checker.checkAll(context.Background(), common.Address{})

		llm := reportByID(t, checker.GetReports(), modelLLM)
		require.Equal(t, int64(32768), *llm.Api.Capabilities.Context)
		require.Equal(t, "vllm:/v1/models.max_model_len", llm.Api.Evidence["capabilities.context"])
	})
}

func TestRuntimePlainResponseLeavesApiByteIdentical(t *testing.T) {
	backend := newChatBackend(t)
	backend.set(http.StatusOK, "", "", "", nil)
	spec := detectedQwen3()
	checker := newTestChecker(runtimeDeps(backend, vllmCfg, &fakeDetector{spec: spec}))
	checker.checkAll(context.Background(), common.Address{})

	llm := reportByID(t, checker.GetReports(), modelLLM)
	require.Equal(t, system.ModelHealthStatusHealthy, llm.Status)
	require.Equal(t, apiJSON(t, spec), apiJSON(t, llm.Api))
	require.Nil(t, llm.Api.Evidence)
}

func TestRuntimeDeclaredAtStableWhenMergedSpecEqual(t *testing.T) {
	backend := newChatBackend(t)
	backend.set(http.StatusOK, "qwen3-8b", "thinking", "", nil)
	checker := newTestChecker(runtimeDeps(backend, vllmCfg, &fakeDetector{spec: detectedVLLMNoFamily()}))
	checker.checkAll(context.Background(), common.Address{})
	first := reportByID(t, checker.GetReports(), modelLLM)
	require.NotZero(t, first.Api.DeclaredAt)
	require.Equal(t, "qwen3", first.Api.ModelFamily)

	seeded := first
	seededAPI := first.Api.Clone()
	seededAPI.DeclaredAt = 12345
	seeded.Api = seededAPI
	checker.setReport(seeded)

	checker.checkAll(context.Background(), common.Address{})
	second := reportByID(t, checker.GetReports(), modelLLM)
	require.Equal(t, int64(12345), second.Api.DeclaredAt, "an equal merged spec keeps DeclaredAt")
	require.Equal(t, apiJSON(t, seededAPI), apiJSON(t, second.Api))

	backend.set(http.StatusOK, "qwen3-8b", "", "", nil)
	checker.checkAll(context.Background(), common.Address{})
	third := reportByID(t, checker.GetReports(), modelLLM)
	require.Equal(t, int64(12345), third.Api.DeclaredAt, "the last harvest still carries reasoning: nothing changed")
}

func TestRuntimePersistsAcrossFailedProbesUntilTTL(t *testing.T) {
	const vllmV1Error = `{"object":"error","message":"The input (5000 tokens) exceeds the model's maximum context length (8192)","type":"BadRequestError","code":400}`
	backend := newChatBackend(t)
	backend.set(http.StatusBadRequest, "", "", vllmV1Error, nil)
	deps := runtimeDeps(backend, vllmCfg, &fakeDetector{spec: detectedVLLMNoFamily()})
	checker := NewChecker(Deps{Adapters: deps, Bids: deps, Models: deps, ModelConfigs: deps, ApiDetect: deps.apiDetect, ApiDetectTTL: 10 * time.Minute},
		time.Hour, time.Second, 0, 0, lib.NewTestLogger())
	clock := time.Unix(1_700_000_000, 0)
	checker.runtime.now = func() time.Time { return clock }

	checker.checkAll(context.Background(), common.Address{})
	require.Equal(t, int64(8192), *reportByID(t, checker.GetReports(), modelLLM).Api.Capabilities.Context)

	// The backend now fails without a parseable body: the last harvest stands.
	backend.set(http.StatusInternalServerError, "", "", `{"error":"upstream down"}`, nil)
	clock = clock.Add(5 * time.Minute)
	checker.checkAll(context.Background(), common.Address{})
	llm := reportByID(t, checker.GetReports(), modelLLM)
	require.Equal(t, system.ModelHealthStatusUnhealthy, llm.Status)
	require.NotNil(t, llm.Api.Capabilities)
	require.Equal(t, int64(8192), *llm.Api.Capabilities.Context)

	// A model whose adapter cannot be built never probes; the store still serves it.
	deps.adapterErr = fmt.Errorf("no adapter")
	clock = clock.Add(4 * time.Minute)
	checker.checkAll(context.Background(), common.Address{})
	require.Equal(t, int64(8192), *reportByID(t, checker.GetReports(), modelLLM).Api.Capabilities.Context)

	// Past the TTL the entry expires and the detection result stands alone.
	clock = clock.Add(2 * time.Minute)
	checker.checkAll(context.Background(), common.Address{})
	llm = reportByID(t, checker.GetReports(), modelLLM)
	require.Nil(t, llm.Api.Capabilities, "runtime evidence older than the TTL is dropped")
	require.Equal(t, "", llm.Api.Evidence["capabilities.context"])
}

func TestRuntimeFactsAccumulateAcrossSweeps(t *testing.T) {
	const vllmV1Error = `{"object":"error","message":"The input (5000 tokens) exceeds the model's maximum context length (8192)","type":"BadRequestError","code":400}`
	backend := newChatBackend(t)
	backend.set(http.StatusBadRequest, "", "", vllmV1Error, nil)
	checker := newTestChecker(runtimeDeps(backend, vllmCfg, &fakeDetector{spec: detectedVLLMNoFamily()}))
	checker.checkAll(context.Background(), common.Address{})

	backend.set(http.StatusOK, "qwen3-8b", "thinking", "", nil)
	checker.checkAll(context.Background(), common.Address{})
	llm := reportByID(t, checker.GetReports(), modelLLM)
	require.Equal(t, system.ModelHealthStatusHealthy, llm.Status)
	require.Equal(t, "qwen3", llm.Api.ModelFamily)
	require.True(t, *llm.Api.Capabilities.Reasoning)
	require.Equal(t, int64(8192), *llm.Api.Capabilities.Context, "a later 200 does not refute the window a 400 named")
}

func TestRuntimePruneForgetsRemovedModel(t *testing.T) {
	backend := newChatBackend(t)
	backend.set(http.StatusOK, "qwen3-8b", "", "", nil)
	deps := runtimeDeps(backend, vllmCfg, &fakeDetector{spec: detectedVLLMNoFamily()})
	checker := newTestChecker(deps)
	checker.checkAll(context.Background(), common.Address{})
	_, held := checker.runtime.byModel[modelLLM.Hex()]
	require.True(t, held)

	deps.modelIDs = nil
	deps.bids = nil
	checker.checkAll(context.Background(), common.Address{})
	_, held = checker.runtime.byModel[modelLLM.Hex()]
	require.False(t, held, "a model that left the sweep leaves the store")
}

func TestHarvestRuntime(t *testing.T) {
	now := time.Unix(1_700_000_000, 0)
	cases := []struct {
		name string
		meta *aiengine.ResponseMeta
		want map[string]any
		ok   bool
	}{
		{name: "nil meta", meta: nil},
		{name: "no response", meta: &aiengine.ResponseMeta{}},
		{name: "plain 200", meta: &aiengine.ResponseMeta{StatusCode: 200}},
		{name: "served id", meta: &aiengine.ResponseMeta{StatusCode: 200, Model: "qwen3-8b"}, want: map[string]any{"served": "qwen3-8b"}, ok: true},
		{name: "reasoning", meta: &aiengine.ResponseMeta{StatusCode: 200, ReasoningSeen: true}, want: map[string]any{"reasoning": true}, ok: true},
		{name: "litellm header", meta: &aiengine.ResponseMeta{StatusCode: 200, Headers: map[string]string{"x-litellm-model-id": "abc"}}, want: map[string]any{"hints": []string{"litellm"}}, ok: true},
		{name: "other headers are no hint", meta: &aiengine.ResponseMeta{StatusCode: 200, Headers: map[string]string{"server": "uvicorn", "openai-processing-ms": "1"}}},
		{name: "error body on 200 is ignored", meta: &aiengine.ResponseMeta{StatusCode: 200, ErrorBody: []byte(`maximum context length is 4096 tokens`)}},
		{name: "context from 400", meta: &aiengine.ResponseMeta{StatusCode: 400, ErrorBody: []byte(`{"error":{"message":"This model's maximum context length is 4096 tokens."}}`)},
			want: map[string]any{"context": int64(4096), "contextSource": apiprobe.ContextSourceOpenAI}, ok: true},
		{name: "rejected params from 400", meta: &aiengine.ResponseMeta{StatusCode: 400, ErrorBody: []byte(`{"error":{"message":"Unrecognized request argument supplied: chat_template_kwargs"}}`)},
			want: map[string]any{"rejected": []string{"chat_template_kwargs"}}, ok: true},
		{name: "500 body is not a request rejection", meta: &aiengine.ResponseMeta{StatusCode: 500, ErrorBody: []byte(`maximum context length is 4096 tokens`)}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rt, facts := harvestRuntime(tc.meta, now)
			require.Equal(t, tc.ok, len(facts) > 0)
			if len(facts) == 0 {
				return
			}
			require.Equal(t, now.Unix(), rt.At)
			require.Equal(t, tc.want["served"], orZero(tc.want["served"], rt.ServedModelID))
			require.Equal(t, tc.want["reasoning"], orZero(tc.want["reasoning"], rt.ReasoningSeen))
			require.Equal(t, tc.want["hints"], orZero(tc.want["hints"], rt.StackHints))
			require.Equal(t, tc.want["context"], orZero(tc.want["context"], rt.ContextLimit))
			require.Equal(t, tc.want["contextSource"], orZero(tc.want["contextSource"], rt.ContextSource))
			require.Equal(t, tc.want["rejected"], orZero(tc.want["rejected"], rt.RejectedParams))
		})
	}
}

// orZero returns got when the case names an expectation, and nil when the
// case expects the field untouched and got is its zero value.
func orZero(want any, got any) any {
	if want != nil {
		return got
	}
	switch v := got.(type) {
	case string:
		if v == "" {
			return nil
		}
	case bool:
		if !v {
			return nil
		}
	case int64:
		if v == 0 {
			return nil
		}
	case []string:
		if v == nil {
			return nil
		}
	}
	return got
}

// --- verifier and capability probes (design contract §5) ---

const (
	kwargRejectedBody   = `{"error":{"message":"[{'type': 'extra_forbidden', 'loc': ('body', 'chat_template_kwargs'), 'msg': 'Extra inputs are not permitted', 'input': {'enable_thinking': False}}]","type":"BadRequestError","param":null,"code":400}}`
	toolChoiceRefusal   = `{"error":{"message":"\"auto\" tool choice requires --enable-auto-tool-choice and --tool-call-parser to be set","type":"BadRequestError","param":null,"code":400}}`
	plainBadRequest     = `{"error":{"message":"not supported","type":"BadRequestError","code":400}}`
	reasoningCompletion = `{"id":"1","object":"chat.completion","choices":[{"index":0,"message":{"role":"assistant","content":"5","reasoning_content":"thinking"},"finish_reason":"stop"}]}`
	verifyClock         = int64(1_700_000_000)
	verifyDelay         = 250 * time.Millisecond
)

type verifyFlags struct{ verify, caps, hosted bool }

// verifyHarness is a checker with the probe flags set, a recording sleeper
// in place of the timer and a controllable clock for the runtime store.
type verifyHarness struct {
	checker *Checker
	deps    *mockDeps
	backend *chatBackend
	sleeps  []time.Duration
	clock   time.Time
}

func newVerifyHarness(t *testing.T, spec *system.ModelApiSpec, flags verifyFlags) *verifyHarness {
	t.Helper()
	h := &verifyHarness{backend: newChatBackend(t), clock: time.Unix(verifyClock, 0)}
	h.backend.set(http.StatusOK, "", "", "", nil)
	cfg := vllmCfg
	cfg.ApiKey = "sk-test"
	// The render request is derived from apiUrl, so it must name the test
	// backend, as the adapter does.
	cfg.ApiURL = h.backend.server.URL + chatPath
	h.deps = runtimeDeps(h.backend, cfg, &fakeDetector{spec: spec})
	h.checker = NewChecker(Deps{
		Adapters: h.deps, Bids: h.deps, Models: h.deps, ModelConfigs: h.deps, ApiDetect: h.deps.apiDetect,
		ApiDetectTTL: time.Hour, VerifyBindings: flags.verify, ProbeCapabilities: flags.caps, ProbeHosted: flags.hosted,
	}, time.Hour, time.Second, verifyDelay, 0, lib.NewTestLogger())
	h.checker.runtime.now = func() time.Time { return h.clock }
	h.checker.sleepFn = func(ctx context.Context, d time.Duration) bool {
		h.sleeps = append(h.sleeps, d)
		return true
	}
	return h
}

func (h *verifyHarness) sweep(t *testing.T) system.ModelHealthReport {
	t.Helper()
	h.checker.checkAll(context.Background(), common.Address{})
	return reportByID(t, h.checker.GetReports(), modelLLM)
}

// kinds names each recorded request by the probe that shaped it; a render
// carries the toggle at its on or off value, or no kwarg at all.
func kinds(reqs []probeRequest) []string {
	var out []string
	for _, r := range reqs {
		switch {
		case strings.HasSuffix(r.path, "/render") && r.has("chat_template_kwargs"):
			if kwargValueIs(r, true) {
				out = append(out, "render:on")
			} else {
				out = append(out, "render:off")
			}
		case strings.HasSuffix(r.path, "/render"):
			out = append(out, "render")
		case r.has("tool_choice"):
			out = append(out, "tools")
		case r.has("response_format"):
			out = append(out, "json_schema")
		case r.hasImagePart():
			out = append(out, "image")
		case r.has("max_tokens"):
			// The health probe sets no max_tokens; every probe request does,
			// and the one with no other marker carries the binding.
			out = append(out, "reasoning.disable")
		default:
			out = append(out, "health")
		}
	}
	return out
}

// kwargValueIs reports whether the request's chat_template_kwargs carries
// the value under any key.
func kwargValueIs(r probeRequest, value any) bool {
	kwargs, _ := r.body["chat_template_kwargs"].(map[string]any)
	for _, v := range kwargs {
		if v == value {
			return true
		}
	}
	return false
}

func detectedQwen3Both() *system.ModelApiSpec {
	spec := detectedQwen3()
	spec.Bindings[system.IntentReasoningEnable] = &system.ParamBinding{Kind: system.BindingKindTemplateKwarg, Param: "chat_template_kwargs.enable_thinking", ParamType: "boolean", Value: true}
	return spec
}

func TestVerifyFlagsOffSendNothing(t *testing.T) {
	h := newVerifyHarness(t, detectedQwen3(), verifyFlags{})
	llm := h.sweep(t)
	require.Equal(t, system.ModelHealthStatusHealthy, llm.Status)
	require.Equal(t, []string{"health"}, kinds(h.backend.recorded()))
	require.Empty(t, h.sleeps)
	require.Nil(t, llm.Api.Bindings[system.IntentReasoningDisable].Verified)
}

func TestVerifyBindingsSendsPacedProbesOncePerTTL(t *testing.T) {
	h := newVerifyHarness(t, detectedQwen3(), verifyFlags{verify: true})
	llm := h.sweep(t)
	require.Equal(t, system.ModelHealthStatusHealthy, llm.Status)
	reqs := h.backend.recorded()
	require.Equal(t, []string{"health", "reasoning.disable", "tools"}, kinds(reqs))
	require.Equal(t, []time.Duration{verifyDelay, verifyDelay}, h.sleeps, "every probe request is paced")

	disable := reqs[1]
	require.Equal(t, map[string]any{"enable_thinking": false}, disable.body["chat_template_kwargs"])
	require.Equal(t, float64(1), disable.body["max_tokens"])
	require.Contains(t, disable.firstContent(), "What is")
	tools := reqs[2]
	require.Equal(t, "auto", tools.body["tool_choice"])
	require.Len(t, tools.body["tools"].([]any), 1)
	require.Equal(t, float64(1), tools.body["max_tokens"])

	b := llm.Api.Bindings[system.IntentReasoningDisable]
	require.NotNil(t, b.Verified)
	require.True(t, *b.Verified)
	require.Equal(t, verifyClock, b.VerifiedAt)
	require.Equal(t, "verify:reasoning.disable@1700000000", llm.Api.Evidence[system.EvidenceKeyBinding(system.IntentReasoningDisable)])
	require.NotNil(t, llm.Api.Capabilities)
	require.True(t, *llm.Api.Capabilities.Tools)
	require.Equal(t, "verify:tools@1700000000", llm.Api.Evidence["capabilities.tools"])
	require.Equal(t, system.ThinkingModeControllable, llm.Api.Thinking.Mode)
	first := llm.Api.DeclaredAt

	// Within the TTL a sweep sends the health probe alone and republishes
	// the verdicts unchanged.
	h.clock = h.clock.Add(30 * time.Minute)
	llm = h.sweep(t)
	require.Equal(t, []string{"health", "reasoning.disable", "tools", "health"}, kinds(h.backend.recorded()))
	require.Len(t, h.sleeps, 2)
	require.True(t, *llm.Api.Bindings[system.IntentReasoningDisable].Verified)
	require.Equal(t, verifyClock, llm.Api.Bindings[system.IntentReasoningDisable].VerifiedAt)
	require.True(t, *llm.Api.Capabilities.Tools)
	require.Equal(t, first, llm.Api.DeclaredAt, "nothing changed")

	// Past the TTL the verifier runs again.
	h.clock = h.clock.Add(31 * time.Minute)
	llm = h.sweep(t)
	require.Equal(t, []string{"health", "reasoning.disable", "tools", "health", "health", "reasoning.disable", "tools"}, kinds(h.backend.recorded()))
	require.Equal(t, h.clock.Unix(), llm.Api.Bindings[system.IntentReasoningDisable].VerifiedAt)

	// With the flag off again the facts expire with the TTL.
	h.checker.deps.VerifyBindings = false
	h.clock = h.clock.Add(61 * time.Minute)
	llm = h.sweep(t)
	require.Len(t, h.backend.recorded(), 8)
	require.Nil(t, llm.Api.Bindings[system.IntentReasoningDisable].Verified)
	require.Nil(t, llm.Api.Capabilities)
	require.Empty(t, llm.Api.Evidence)
}

func TestVerifyReasoningContentStillPresent(t *testing.T) {
	h := newVerifyHarness(t, detectedQwen3(), verifyFlags{verify: true})
	h.backend.setScript(func(r probeRequest) (int, string, bool) {
		if r.has("chat_template_kwargs") {
			return http.StatusOK, reasoningCompletion, true
		}
		return 0, "", false
	})
	llm := h.sweep(t)
	require.Equal(t, []string{"health", "reasoning.disable", "tools"}, kinds(h.backend.recorded()))
	b := llm.Api.Bindings[system.IntentReasoningDisable]
	require.NotNil(t, b.Verified)
	require.False(t, *b.Verified)
	require.Equal(t, "verify:reasoning_content_still_present@1700000000", llm.Api.Evidence[system.EvidenceKeyBinding(system.IntentReasoningDisable)])
	require.Equal(t, system.ThinkingModeAlwaysOn, llm.Api.Thinking.Mode, "reasoning with the disable applied: the model reasons unconditionally")
	require.True(t, *llm.Api.Capabilities.Reasoning)
	require.Equal(t, "verify:reasoning_seen_with_disable@1700000000", llm.Api.Evidence["capabilities.reasoning"])
	first := llm.Api.DeclaredAt

	h.clock = h.clock.Add(10 * time.Minute)
	llm = h.sweep(t)
	require.Equal(t, first, llm.Api.DeclaredAt, "the refuted verdict republishes with its own timestamp")
	require.Equal(t, "verify:reasoning_seen_with_disable@1700000000", llm.Api.Evidence["capabilities.reasoning"])
}

func TestVerifyRejectedParamDropsBinding(t *testing.T) {
	h := newVerifyHarness(t, detectedQwen3(), verifyFlags{verify: true})
	h.backend.setScript(func(r probeRequest) (int, string, bool) {
		if r.has("chat_template_kwargs") {
			return http.StatusBadRequest, kwargRejectedBody, true
		}
		return 0, "", false
	})
	llm := h.sweep(t)
	require.Equal(t, []string{"health", "reasoning.disable", "tools"}, kinds(h.backend.recorded()))
	require.Nil(t, llm.Api.Bindings[system.IntentReasoningDisable], "the backend refused the parameter")
	require.Nil(t, llm.Api.Thinking)
	require.Equal(t, "verify:rejected:chat_template_kwargs@1700000000", llm.Api.Evidence[system.EvidenceKeyBinding(system.IntentReasoningDisable)])
	require.True(t, *llm.Api.Capabilities.Tools)

	// Within the TTL the drop is republished without a request; past it the
	// rejection expires with the other verdicts, the binding is back in the
	// merged spec and is verified again (a backend upgrade may accept it).
	h.clock = h.clock.Add(30 * time.Minute)
	llm = h.sweep(t)
	require.Equal(t, []string{"health", "reasoning.disable", "tools", "health"}, kinds(h.backend.recorded()))
	require.Nil(t, llm.Api.Bindings[system.IntentReasoningDisable])
	require.Equal(t, "verify:rejected:chat_template_kwargs@1700000000", llm.Api.Evidence[system.EvidenceKeyBinding(system.IntentReasoningDisable)])

	h.clock = h.clock.Add(31 * time.Minute)
	llm = h.sweep(t)
	require.Equal(t, []string{"health", "reasoning.disable", "tools", "health", "health", "reasoning.disable", "tools"}, kinds(h.backend.recorded()))
	require.Nil(t, llm.Api.Bindings[system.IntentReasoningDisable])
	require.Equal(t, fmt.Sprintf("verify:rejected:chat_template_kwargs@%d", h.clock.Unix()), llm.Api.Evidence[system.EvidenceKeyBinding(system.IntentReasoningDisable)])
}

func TestVerifyToolsVerdicts(t *testing.T) {
	t.Run("the vLLM refusal publishes tools false", func(t *testing.T) {
		h := newVerifyHarness(t, detectedVLLMNoFamily(), verifyFlags{verify: true})
		h.backend.setScript(func(r probeRequest) (int, string, bool) {
			if r.has("tool_choice") {
				return http.StatusBadRequest, toolChoiceRefusal, true
			}
			return 0, "", false
		})
		llm := h.sweep(t)
		require.Equal(t, []string{"health", "tools"}, kinds(h.backend.recorded()), "no reasoning binding: only the tool probe")
		require.False(t, *llm.Api.Capabilities.Tools)
		require.Equal(t, "verify:tool_choice_auto@1700000000", llm.Api.Evidence["capabilities.tools"], "the evidence names the server-side refusal, not a model fact")
	})

	t.Run("another 4xx publishes nothing and is not asked again before the TTL", func(t *testing.T) {
		// The health response carries no model, no reasoning and no gateway
		// header, so the run leaves no harvested fact and no verdict: the
		// once-per-TTL gate must survive on its own.
		h := newVerifyHarness(t, detectedVLLMNoFamily(), verifyFlags{verify: true})
		h.backend.setScript(func(r probeRequest) (int, string, bool) {
			if r.has("tool_choice") {
				return http.StatusBadRequest, plainBadRequest, true
			}
			return 0, "", false
		})
		llm := h.sweep(t)
		require.Equal(t, []string{"health", "tools"}, kinds(h.backend.recorded()))
		require.Nil(t, llm.Api.Capabilities)
		require.Empty(t, llm.Api.Evidence)
		first := llm.Api.DeclaredAt

		h.clock = h.clock.Add(time.Minute)
		llm = h.sweep(t)
		require.Equal(t, []string{"health", "tools", "health"}, kinds(h.backend.recorded()), "within the TTL the health probe goes alone")
		require.Nil(t, llm.Api.Capabilities)
		require.Empty(t, llm.Api.Evidence)
		require.Equal(t, first, llm.Api.DeclaredAt)

		h.clock = h.clock.Add(61 * time.Minute)
		h.sweep(t)
		require.Equal(t, []string{"health", "tools", "health", "health", "tools"}, kinds(h.backend.recorded()), "past the TTL the probe is sent again")
	})

	t.Run("known tools are not probed", func(t *testing.T) {
		spec := detectedVLLMNoFamily()
		supported := true
		spec.Capabilities = &system.ModelCapabilities{Tools: &supported}
		spec.Evidence = map[string]string{"capabilities.tools": "llamacpp:/props.chat_template_caps.supports_tool_calls"}
		h := newVerifyHarness(t, spec, verifyFlags{verify: true})
		h.sweep(t)
		require.Equal(t, []string{"health"}, kinds(h.backend.recorded()))
	})
}

// renderScript renders a template that thinks by default: the off value
// changes the prompt when differ is set, the on value never does.
func renderScript(differ bool) func(r probeRequest) (int, string, bool) {
	return func(r probeRequest) (int, string, bool) {
		if !strings.HasSuffix(r.path, "/render") {
			return 0, "", false
		}
		if differ && kwargValueIs(r, false) {
			return http.StatusOK, `{"request_id":"r2","token_ids":[1,2,3,9]}`, true
		}
		return http.StatusOK, `{"request_id":"r1","token_ids":[1,2,3]}`, true
	}
}

func TestProbeCapabilities(t *testing.T) {
	t.Run("accepted probes publish structured, image input and the template bindings", func(t *testing.T) {
		h := newVerifyHarness(t, detectedQwen3Both(), verifyFlags{caps: true})
		h.backend.setScript(renderScript(true))
		llm := h.sweep(t)
		reqs := h.backend.recorded()
		require.Equal(t, []string{"health", "json_schema", "image", "render:off", "render:on"}, kinds(reqs))
		require.Len(t, h.sleeps, 4, "both render requests are paced")
		require.Equal(t, chatPath+"/render", reqs[3].path)
		require.Equal(t, "my-alias", reqs[3].body["model"])
		require.Equal(t, float64(1), reqs[1].body["max_tokens"])
		require.Equal(t, float64(1), reqs[2].body["max_tokens"])

		require.Equal(t, []string{"json_schema"}, llm.Api.Capabilities.Structured)
		require.Equal(t, "verify:json_schema@1700000000", llm.Api.Evidence["capabilities.structured"])
		require.Equal(t, []string{"text", "image"}, llm.Api.Capabilities.Modalities.Input)
		require.Equal(t, "verify:image@1700000000", llm.Api.Evidence["capabilities.modalities.input"])
		for _, intent := range []string{system.IntentReasoningDisable, system.IntentReasoningEnable} {
			b := llm.Api.Bindings[intent]
			require.NotNil(t, b.Verified, intent)
			require.True(t, *b.Verified, intent)
			require.Equal(t, verifyClock, b.VerifiedAt)
			require.Equal(t, "verify:"+intent+"@1700000000", llm.Api.Evidence[system.EvidenceKeyBinding(intent)])
		}
		require.Nil(t, llm.Api.Capabilities.Tools, "the tool probe belongs to MODEL_API_VERIFY_BINDINGS")
	})

	t.Run("refused probes and an unchanged render publish the refutation only", func(t *testing.T) {
		h := newVerifyHarness(t, detectedQwen3Both(), verifyFlags{caps: true})
		render := renderScript(false)
		h.backend.setScript(func(r probeRequest) (int, string, bool) {
			if r.has("response_format") || r.hasImagePart() {
				return http.StatusBadRequest, plainBadRequest, true
			}
			return render(r)
		})
		llm := h.sweep(t)
		require.Equal(t, []string{"health", "json_schema", "image", "render:off", "render:on"}, kinds(h.backend.recorded()))
		require.Nil(t, llm.Api.Capabilities, "a refused list-valued probe publishes nothing")
		for _, intent := range []string{system.IntentReasoningDisable, system.IntentReasoningEnable} {
			require.False(t, *llm.Api.Bindings[intent].Verified, intent)
			require.Equal(t, "verify:rendered_prompt_unchanged@1700000000", llm.Api.Evidence[system.EvidenceKeyBinding(intent)])
		}
		require.Equal(t, system.ThinkingModeControllable, llm.Api.Thinking.Mode)
	})

	t.Run("known structured and image input are not probed; a non-vllm stack is not rendered", func(t *testing.T) {
		spec := detectedQwen3Both()
		spec.Stack = "sglang"
		spec.Capabilities = &system.ModelCapabilities{Structured: []string{"json_object", "json_schema"}, Modalities: &system.Modalities{Input: []string{"text", "image"}}}
		spec.Evidence = map[string]string{"capabilities.structured": "sglang:/model_info", "capabilities.modalities.input": "sglang:/model_info.has_image_understanding"}
		h := newVerifyHarness(t, spec, verifyFlags{caps: true})
		h.sweep(t)
		require.Equal(t, []string{"health"}, kinds(h.backend.recorded()))
	})

	t.Run("the render is skipped once the completion with the kwarg verified the binding", func(t *testing.T) {
		h := newVerifyHarness(t, detectedQwen3Both(), verifyFlags{verify: true, caps: true})
		h.backend.setScript(renderScript(true))
		llm := h.sweep(t)
		require.Equal(t, []string{"health", "reasoning.disable", "tools", "json_schema", "image"}, kinds(h.backend.recorded()))
		require.True(t, *llm.Api.Bindings[system.IntentReasoningDisable].Verified)
		require.Nil(t, llm.Api.Bindings[system.IntentReasoningEnable].Verified, "the completion probe says nothing about the enable binding")
	})

	t.Run("a refuted completion is not overridden by a differing render", func(t *testing.T) {
		h := newVerifyHarness(t, detectedQwen3Both(), verifyFlags{verify: true, caps: true})
		render := renderScript(true)
		h.backend.setScript(func(r probeRequest) (int, string, bool) {
			if r.path == chatPath && r.has("chat_template_kwargs") {
				return http.StatusOK, reasoningCompletion, true
			}
			return render(r)
		})
		llm := h.sweep(t)
		require.Equal(t, []string{"health", "reasoning.disable", "tools", "json_schema", "image"}, kinds(h.backend.recorded()), "the completion decided the binding; no render")
		b := llm.Api.Bindings[system.IntentReasoningDisable]
		require.NotNil(t, b.Verified)
		require.False(t, *b.Verified)
		require.Equal(t, "verify:reasoning_content_still_present@1700000000", llm.Api.Evidence[system.EvidenceKeyBinding(system.IntentReasoningDisable)])
		require.Equal(t, system.ThinkingModeAlwaysOn, llm.Api.Thinking.Mode)
		require.Equal(t, "verify:reasoning_seen_with_disable@1700000000", llm.Api.Evidence["capabilities.reasoning"])
		require.Nil(t, llm.Api.Bindings[system.IntentReasoningEnable].Verified)
	})
}

func TestVerifyHostedGuard(t *testing.T) {
	hosted := func(stack, via string) *system.ModelApiSpec {
		return &system.ModelApiSpec{Stack: stack, Via: via, Source: system.ApiSpecSourceDetected,
			Bindings: map[string]*system.ParamBinding{system.IntentReasoningDisable: {Kind: system.BindingKindBodyParam, Param: "reasoning.enabled", ParamType: "boolean", Value: false}}}
	}
	gated := []*system.ModelApiSpec{hosted("openrouter", ""), hosted("venice", ""), hosted("litellm", ""), hosted("anthropic", ""), hosted("openai", ""), hosted("groq", ""), hosted("hf-router", ""), hosted("", ""), hosted("vllm", "litellm")}
	for _, spec := range gated {
		name := spec.Stack + "/" + spec.Via
		t.Run(name+" is skipped", func(t *testing.T) {
			h := newVerifyHarness(t, spec, verifyFlags{verify: true, caps: true})
			h.sweep(t)
			require.Equal(t, []string{"health"}, kinds(h.backend.recorded()))
			require.Empty(t, h.sleeps)
		})
	}
	for _, spec := range gated {
		name := spec.Stack + "/" + spec.Via
		t.Run(name+" is probed with the switch", func(t *testing.T) {
			h := newVerifyHarness(t, spec, verifyFlags{verify: true, hosted: true})
			llm := h.sweep(t)
			require.Equal(t, []string{"health", "reasoning.disable", "tools"}, kinds(h.backend.recorded()))
			require.True(t, *llm.Api.Bindings[system.IntentReasoningDisable].Verified)
		})
	}

	t.Run("MODEL_API_PROBE_HOSTED lets a hosted vendor be probed", func(t *testing.T) {
		h := newVerifyHarness(t, hosted("openrouter", ""), verifyFlags{verify: true, hosted: true})
		llm := h.sweep(t)
		reqs := h.backend.recorded()
		require.Equal(t, []string{"health", "reasoning.disable", "tools"}, kinds(reqs))
		require.Equal(t, map[string]any{"enabled": false}, reqs[1].body["reasoning"], "the body_param binding lands at its dotted path")
		require.True(t, *llm.Api.Bindings[system.IntentReasoningDisable].Verified)
	})

}

func TestVerifyOnlyAfterHealthyProbe(t *testing.T) {
	h := newVerifyHarness(t, detectedQwen3(), verifyFlags{verify: true, caps: true})
	h.backend.set(http.StatusInternalServerError, "", "", `{"error":"down"}`, nil)
	llm := h.sweep(t)
	require.Equal(t, system.ModelHealthStatusUnhealthy, llm.Status)
	require.Equal(t, []string{"health"}, kinds(h.backend.recorded()))
	require.Empty(t, h.sleeps)
}

func TestVerifySpecChangeReverifiesBeforeTTL(t *testing.T) {
	h := newVerifyHarness(t, detectedQwen3(), verifyFlags{verify: true})
	h.sweep(t)
	require.Equal(t, []string{"health", "reasoning.disable", "tools"}, kinds(h.backend.recorded()))

	det := h.deps.apiDetect.(*fakeDetector)
	det.spec = detectedQwen3()
	det.spec.Bindings[system.IntentReasoningDisable].Param = "chat_template_kwargs.thinking"
	h.clock = h.clock.Add(time.Minute)
	llm := h.sweep(t)
	reqs := h.backend.recorded()
	require.Equal(t, []string{"health", "reasoning.disable", "tools", "health", "reasoning.disable"}, kinds(reqs),
		"the new binding is verified; tool support is a backend fact the first run settled")
	require.Equal(t, map[string]any{"thinking": false}, reqs[4].body["chat_template_kwargs"])
	require.True(t, *llm.Api.Bindings[system.IntentReasoningDisable].Verified)
	require.Equal(t, h.clock.Unix(), llm.Api.Bindings[system.IntentReasoningDisable].VerifiedAt)
	require.True(t, *llm.Api.Capabilities.Tools, "carried over from the first run")
	require.Equal(t, "verify:tools@1700000000", llm.Api.Evidence["capabilities.tools"])
}

func TestVerifyCancelledRunRecordsNothing(t *testing.T) {
	h := newVerifyHarness(t, detectedQwen3(), verifyFlags{verify: true})
	ctx, cancel := context.WithCancel(context.Background())
	calls := 0
	h.checker.sleepFn = func(ctx context.Context, d time.Duration) bool {
		calls++
		if calls == 2 {
			cancel()
			return false
		}
		return true
	}
	h.checker.checkAll(ctx, common.Address{})
	require.Equal(t, []string{"health", "reasoning.disable"}, kinds(h.backend.recorded()), "the second probe was cancelled before its request")
	llm := reportByID(t, h.checker.GetReports(), modelLLM)
	require.Nil(t, llm.Api.Bindings[system.IntentReasoningDisable].Verified, "an interrupted run publishes nothing")

	// The next sweep runs the whole verifier again.
	h.checker.sleepFn = func(context.Context, time.Duration) bool { return true }
	h.clock = h.clock.Add(time.Minute)
	llm = h.sweep(t)
	require.Equal(t, []string{"health", "reasoning.disable", "health", "reasoning.disable", "tools"}, kinds(h.backend.recorded()))
	require.True(t, *llm.Api.Bindings[system.IntentReasoningDisable].Verified)
}

func TestVerifyKeyIgnoresRuntimeVerdicts(t *testing.T) {
	a := detectedQwen3()
	b := detectedQwen3()
	verified := true
	b.Bindings[system.IntentReasoningDisable].Verified = &verified
	b.Bindings[system.IntentReasoningDisable].VerifiedAt = 1
	b.Capabilities = &system.ModelCapabilities{Tools: &verified}
	require.Equal(t, verifyKey(a), verifyKey(b))

	c := detectedQwen3()
	c.Bindings[system.IntentReasoningDisable].Value = true
	require.NotEqual(t, verifyKey(a), verifyKey(c))
	d := detectedQwen3()
	d.Stack = "sglang"
	require.NotEqual(t, verifyKey(a), verifyKey(d))
}
