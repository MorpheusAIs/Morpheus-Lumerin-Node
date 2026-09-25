package modelhealth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
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
	for k, v := range b.headers {
		w.Header().Set(k, v)
	}
	w.Header().Set("Content-Type", "application/json")
	if b.status != http.StatusOK {
		w.WriteHeader(b.status)
		_, _ = w.Write([]byte(b.errorBody))
		return
	}
	var req struct {
		Messages []struct {
			Content string `json:"content"`
		} `json:"messages"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)
	var x, y int
	if len(req.Messages) > 0 {
		_, _ = fmt.Sscanf(req.Messages[0].Content, "What is %d+%d?", &x, &y)
	}
	message := map[string]any{"role": "assistant", "content": fmt.Sprintf("%d", x+y)}
	if b.reasoning != "" {
		message["reasoning_content"] = b.reasoning
	}
	body := map[string]any{
		"id": "chatcmpl-1", "object": "chat.completion",
		"choices": []any{map[string]any{"index": 0, "message": message, "finish_reason": "stop"}},
	}
	if b.model != "" {
		body["model"] = b.model
	}
	_ = json.NewEncoder(w).Encode(body)
}

func (b *chatBackend) adapter() aiengine.AIEngineStream {
	return aiengine.NewOpenAIEngine("configured-name", b.server.URL, "", time.Minute, &lib.LoggerMock{}, nil)
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
		rt:   apispec.Runtime{ToolsSupported: &supported, ServedModelID: "qwen3-8b"},
		seen: map[runtimeFact]time.Time{factServedModelID: clock},
	}

	rt, ok := store.harvest("m", nil)
	require.True(t, ok)
	require.Equal(t, clock.Unix(), rt.At)
	require.Equal(t, "qwen3-8b", rt.ServedModelID)

	// The harvested fact expires; the verifier's fact keeps the entry, and At
	// is left for MergeRuntime to fill with now rather than the zero time.
	clock = clock.Add(2 * time.Minute)
	rt, ok = store.harvest("m", nil)
	require.True(t, ok, "verifier facts are not the harvest's to expire")
	require.Equal(t, "", rt.ServedModelID)
	require.NotNil(t, rt.ToolsSupported)
	require.Zero(t, rt.At)
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
