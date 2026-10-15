package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/apispec"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/config"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/lib"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/system"
	"github.com/stretchr/testify/require"
)

func i64(v int64) *int64 { return &v }
func bp(v bool) *bool    { return &v }

func legacySpec() *system.ModelApiSpec {
	return &system.ModelApiSpec{
		Stack:       "vllm",
		ModelFamily: "qwen3",
		Source:      system.ApiSpecSourceDetected,
		Thinking:    &system.ThinkingSpec{Mode: system.ThinkingModeControllable},
		Bindings: map[string]*system.ParamBinding{
			system.IntentReasoningDisable: {Kind: system.BindingKindTemplateKwarg, Param: "chat_template_kwargs.enable_thinking", ParamType: "boolean", Value: false},
			system.IntentReasoningEffort:  {Kind: system.BindingKindTemplateKwarg, Param: "chat_template_kwargs.reasoning_effort", ParamType: "enum", EnumValues: []string{"low", "medium", "high"}},
		},
		Parameters: []string{"tools", "reasoning"},
	}
}

func fullSpec() *system.ModelApiSpec {
	api := legacySpec()
	api.Thinking.Default = "on"
	api.Bindings[system.IntentReasoningDisable].Verified = bp(true)
	api.Bindings[system.IntentReasoningDisable].VerifiedAt = 1790330400
	api.Bindings[system.IntentReasoningEffort].Verified = bp(false)
	api.Capabilities = &system.ModelCapabilities{
		Context:        i64(131072),
		ContextTrained: i64(262144),
		MaxOutput:      i64(8192),
		Modalities:     &system.Modalities{Input: []string{"text", "image"}, Output: []string{"text"}},
		Tools:          bp(true),
		ToolsParallel:  bp(false),
		Structured:     []string{"json_object", "json_schema"},
		Reasoning:      bp(true),
		WebSearch:      bp(false),
		Embeddings:     bp(false),
		PromptCache:    bp(true),
		Quantization:   "fp8",
	}
	api.Evidence = map[string]string{
		system.EvidenceKeyThinkingDefault:                        "vllm:/v1/models.max_model_len",
		system.EvidenceKeyCapability("context"):                  "vllm:/v1/models.max_model_len",
		system.EvidenceKeyCapability("tools"):                    "verify:tool_choice@1790330400",
		system.EvidenceKeyBinding(system.IntentReasoningDisable): "verify:no-reasoning@1790330400",
		system.EvidenceKeyBinding(system.IntentReasoningEffort):  "verify:rejected@1790330400",
	}
	return api
}

func TestFormatResult(t *testing.T) {
	trace := []string{"GET http://a:8000/props -> HTTP 200", "identified vllm by /version"}
	tests := []struct {
		name  string
		api   *system.ModelApiSpec
		trace []string
		want  string
	}{
		{
			name:  "legacy spec without capabilities, evidence, thinking default or verified flags",
			api:   legacySpec(),
			trace: trace,
			want: "  result:\n" +
				"    stack:       vllm\n" +
				"    modelFamily: qwen3\n" +
				"    source:      detected\n" +
				"    thinking:    controllable\n" +
				"    bindings:\n" +
				"      reasoning.disable: template_kwarg chat_template_kwargs.enable_thinking (boolean, value=false)\n" +
				"      reasoning.effort:  template_kwarg chat_template_kwargs.reasoning_effort (enum: low|medium|high)\n" +
				"    parameters:  tools, reasoning\n" +
				"  how:\n" +
				"    - GET http://a:8000/props -> HTTP 200\n" +
				"    - identified vllm by /version\n",
		},
		{
			name:  "fully populated spec",
			api:   fullSpec(),
			trace: trace,
			want: "  result:\n" +
				"    stack:       vllm\n" +
				"    modelFamily: qwen3\n" +
				"    source:      detected\n" +
				"    thinking:    controllable (default on)\n" +
				"    capabilities:\n" +
				"      context:           131072\n" +
				"      contextTrained:    262144\n" +
				"      maxOutput:         8192\n" +
				"      modalities.input:  text, image\n" +
				"      modalities.output: text\n" +
				"      tools:             true\n" +
				"      toolsParallel:     false\n" +
				"      structured:        json_object, json_schema\n" +
				"      reasoning:         true\n" +
				"      webSearch:         false\n" +
				"      embeddings:        false\n" +
				"      promptCache:       true\n" +
				"      quantization:      fp8\n" +
				"    bindings:\n" +
				"      reasoning.disable: template_kwarg chat_template_kwargs.enable_thinking (boolean, value=false) [verified 2026-09-25T10:00:00Z]\n" +
				"      reasoning.effort:  template_kwarg chat_template_kwargs.reasoning_effort (enum: low|medium|high) [unverified: verify:rejected@1790330400]\n" +
				"    parameters:  tools, reasoning\n" +
				"    evidence:\n" +
				"      bindings.reasoning.disable -> verify:no-reasoning@1790330400\n" +
				"      bindings.reasoning.effort -> verify:rejected@1790330400\n" +
				"      capabilities.context -> vllm:/v1/models.max_model_len\n" +
				"      capabilities.tools -> verify:tool_choice@1790330400\n" +
				"      thinking.default -> vllm:/v1/models.max_model_len\n" +
				"  how:\n" +
				"    - GET http://a:8000/props -> HTTP 200\n" +
				"    - identified vllm by /version\n",
		},
		{
			name: "via gateway, empty capabilities and modalities, verified without timestamp, unverified with timestamp but no evidence",
			api: &system.ModelApiSpec{
				Stack:        "venice",
				Via:          "litellm",
				Source:       system.ApiSpecSourceDetected,
				Capabilities: &system.ModelCapabilities{Modalities: &system.Modalities{}},
				Bindings: map[string]*system.ParamBinding{
					system.IntentReasoningDisable: {Kind: system.BindingKindBodyParam, Param: "reasoning_effort", Value: "none", Verified: bp(true)},
					system.IntentReasoningEffort:  {Kind: system.BindingKindBodyParam, Param: "reasoning_effort", Verified: bp(false), VerifiedAt: 1790330400},
				},
			},
			want: "  result:\n" +
				"    stack:       venice\n" +
				"    via:         litellm\n" +
				"    modelFamily: -\n" +
				"    source:      detected\n" +
				"    thinking:    unknown\n" +
				"    bindings:\n" +
				"      reasoning.disable: body_param reasoning_effort (value=none) [verified]\n" +
				"      reasoning.effort:  body_param reasoning_effort [unverified: reason not recorded]\n" +
				"  how:\n" +
				"    (no steps recorded)\n",
		},
		{
			name:  "nothing detected",
			api:   nil,
			trace: []string{"nothing worked"},
			want: "  result: nothing detected\n" +
				"  how:\n" +
				"    - nothing worked\n",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.want, formatResult(tc.api, tc.trace))
		})
	}
}

func TestJSONReportCarriesNewFields(t *testing.T) {
	raw, err := json.Marshal(jsonReport{ModelID: "0x1", ModelName: "qwen3", ApiType: "openai", Api: fullSpec(), Trace: []string{"step"}})
	require.NoError(t, err)

	var got struct {
		Api struct {
			Thinking struct {
				Mode    string `json:"mode"`
				Default string `json:"default"`
			} `json:"thinking"`
			Capabilities map[string]json.RawMessage `json:"capabilities"`
			Evidence     map[string]string          `json:"evidence"`
			Bindings     map[string]struct {
				Verified   *bool `json:"verified"`
				VerifiedAt int64 `json:"verifiedAt"`
			} `json:"bindings"`
		} `json:"api"`
	}
	require.NoError(t, json.Unmarshal(raw, &got))
	require.Equal(t, "on", got.Api.Thinking.Default)
	require.Equal(t, "131072", string(got.Api.Capabilities["context"]))
	require.Equal(t, "false", string(got.Api.Capabilities["toolsParallel"]))
	require.Equal(t, `{"input":["text","image"],"output":["text"]}`, string(got.Api.Capabilities["modalities"]))
	require.Equal(t, `"fp8"`, string(got.Api.Capabilities["quantization"]))
	require.Equal(t, "vllm:/v1/models.max_model_len", got.Api.Evidence["capabilities.context"])
	require.NotNil(t, got.Api.Bindings[system.IntentReasoningDisable].Verified)
	require.True(t, *got.Api.Bindings[system.IntentReasoningDisable].Verified)
	require.Equal(t, int64(1790330400), got.Api.Bindings[system.IntentReasoningDisable].VerifiedAt)
	require.NotNil(t, got.Api.Bindings[system.IntentReasoningEffort].Verified)
	require.False(t, *got.Api.Bindings[system.IntentReasoningEffort].Verified)

	legacy, err := json.Marshal(jsonReport{Api: legacySpec()})
	require.NoError(t, err)
	require.NotContains(t, string(legacy), `"capabilities"`)
	require.NotContains(t, string(legacy), `"evidence"`)
	require.NotContains(t, string(legacy), `"verified"`)
}

// The -verify path against a fake llama.cpp-shaped OpenAI backend: /props
// carries the Qwen3 template so detection yields a template_kwarg
// reasoning.disable binding the verifier can apply, and the chat endpoint
// answers every probe with one of two behaviours.

const testQwen3Template = `{%- if messages %}{% generation %}{% endgeneration %}{%- endif %}
{%- if enable_thinking is defined and enable_thinking is false %}<think></think>{%- endif %}`

const (
	testAt  = int64(1790330400)
	testKey = "sk-test-secret-key"
)

// fakeOpts selects the backend's behaviour. rejectKwargs answers a request
// carrying chat_template_kwargs with OpenAI's unrecognised-argument error;
// reasoning puts reasoning_content in every answer, so the binding cannot
// have switched it off; bare serves the chat endpoint only, so detection
// has no engine or listing to fingerprint; chatStatus, when set, is the
// bare status every chat request gets instead of an answer.
type fakeOpts struct {
	rejectKwargs bool
	reasoning    bool
	bare         bool
	chatStatus   int
}

func fakeBackend(t *testing.T, o fakeOpts) *httptest.Server {
	t.Helper()
	writeJSON := func(w http.ResponseWriter, status int, v any) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_ = json.NewEncoder(w).Encode(v)
	}
	mux := http.NewServeMux()
	if !o.bare {
		mux.HandleFunc("/props", func(w http.ResponseWriter, r *http.Request) {
			writeJSON(w, http.StatusOK, map[string]any{
				"default_generation_settings": map[string]any{"id": 0, "n_ctx": 8192},
				"total_slots":                 1,
				"chat_template":               testQwen3Template,
				"build_info":                  "b11177-abcdef0",
			})
		})
		mux.HandleFunc("/v1/models", func(w http.ResponseWriter, r *http.Request) {
			writeJSON(w, http.StatusOK, map[string]any{
				"object": "list",
				"data":   []map[string]any{{"id": "qwen3-32b", "object": "model", "owned_by": "llamacpp"}},
			})
		})
	}
	mux.HandleFunc("/v1/chat/completions", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.NotFound(w, r)
			return
		}
		require.Equal(t, "Bearer "+testKey, r.Header.Get("Authorization"), "the probes carry the model's key")
		var body map[string]any
		require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
		if o.chatStatus != 0 {
			w.WriteHeader(o.chatStatus)
			return
		}
		if _, ok := body["chat_template_kwargs"]; ok && o.rejectKwargs {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": map[string]any{
				"message": "Unrecognized request argument supplied: chat_template_kwargs",
				"type":    "invalid_request_error",
			}})
			return
		}
		message := map[string]any{"role": "assistant", "content": "5"}
		if o.reasoning {
			message["reasoning_content"] = "2 plus 3 is 5"
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"id": "chatcmpl-probe", "object": "chat.completion", "model": body["model"],
			"choices": []map[string]any{{"index": 0, "finish_reason": "length", "message": message}},
		})
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv
}

func writeConfig(t *testing.T, models ...map[string]any) string {
	t.Helper()
	raw, err := json.Marshal(map[string]any{"models": models})
	require.NoError(t, err)
	path := filepath.Join(t.TempDir(), "models-config.json")
	require.NoError(t, os.WriteFile(path, raw, 0o644))
	return path
}

func pinClock(t *testing.T) {
	t.Helper()
	prev := now
	now = func() time.Time { return time.Unix(testAt, 0) }
	t.Cleanup(func() { now = prev })
}

func runCLI(t *testing.T, args ...string) (code int, stdout, stderr string) {
	t.Helper()
	var out, errOut bytes.Buffer
	code = run(args, &out, &errOut)
	return code, out.String(), errOut.String()
}

func TestVerifyRequiresExactlyOneModel(t *testing.T) {
	srv := fakeBackend(t, fakeOpts{})
	cfg := writeConfig(t,
		map[string]any{"modelId": "0x01", "modelName": "qwen3-32b", "apiType": "openai", "apiUrl": srv.URL + "/v1/chat/completions"},
		map[string]any{"modelId": "0x02", "modelName": "qwen3-8b", "apiType": "openai", "apiUrl": srv.URL + "/v1/chat/completions"},
	)

	code, _, stderr := runCLI(t, "-config", cfg, "-verify")
	require.Equal(t, 2, code)
	require.Contains(t, stderr, "-verify needs -model")

	code, _, stderr = runCLI(t, "-config", cfg, "-verify", "-model", "qwen3")
	require.Equal(t, 2, code)
	require.Contains(t, stderr, "matches 2")
	require.Contains(t, stderr, "qwen3-32b (0x01)")
}

func TestVerifyVerifiedBindingText(t *testing.T) {
	pinClock(t)
	srv := fakeBackend(t, fakeOpts{})
	cfg := writeConfig(t, map[string]any{"modelId": "0x01", "modelName": "qwen3-32b", "apiType": "openai", "apiUrl": srv.URL + "/v1/chat/completions", "apiKey": testKey})

	code, stdout, stderr := runCLI(t, "-config", cfg, "-verify", "-model", "qwen3-32b")
	require.Equal(t, 0, code, stderr)
	require.NotContains(t, stdout+stderr, testKey)
	require.Empty(t, stderr, "a self-hosted stack gets no cost warning")

	require.Contains(t, stdout, "    stack:       llamacpp\n")
	require.Contains(t, stdout, "      reasoning.disable: template_kwarg chat_template_kwargs.enable_thinking (boolean, value=false) [verified 2026-09-25T10:00:00Z]\n")
	require.Contains(t, stdout, "      tools:             true\n")
	require.Contains(t, stdout, "      bindings.reasoning.disable -> verify:reasoning.disable@1790330400\n")
	require.Contains(t, stdout, "      capabilities.structured -> llamacpp:/props\n", "the stack already lists json_schema; the probe adds nothing")
	require.Contains(t, stdout, "      capabilities.tools -> verify:tools@1790330400\n")
	require.Contains(t, stdout, "      capabilities.modalities.input -> verify:image@1790330400\n")

	want := "  verify:\n" +
		"    requests:\n" +
		"      - POST /v1/chat/completions -> HTTP 200\n" +
		"      - POST /v1/chat/completions -> HTTP 200\n" +
		"      - POST /v1/chat/completions -> HTTP 200\n" +
		"      - POST /v1/chat/completions -> HTTP 200\n" +
		"    verdicts:\n" +
		"      - reasoning.disable: accepted (HTTP 200): no reasoning content with the binding applied\n" +
		"      - tools: accepted (HTTP 200): accepted\n" +
		"      - json_schema: accepted (HTTP 200): accepted\n" +
		"      - image: accepted (HTTP 200): accepted\n" +
		"    merge:\n" +
		"      - bindings.reasoning.disable verified = true (verify:reasoning.disable@1790330400): no reasoning content with the binding applied\n"
	require.Contains(t, stdout, want)
	require.Contains(t, stdout, "      - capabilities.tools = true (verify:tools@1790330400)\n")
	require.Contains(t, stdout, "      - capabilities.modalities.input = image (verify:image@1790330400)\n")
	require.Contains(t, stdout, "      - capabilities.structured: runtime value [json_schema] adds nothing to [json_object json_schema grammar]\n")
	require.Contains(t, stdout, "      - capabilities.modalities.input: text prepended — a chat model accepts text (verify:image@1790330400)\n")
}

func TestVerifyRejectedBindingJSON(t *testing.T) {
	pinClock(t)
	srv := fakeBackend(t, fakeOpts{rejectKwargs: true})
	cfg := writeConfig(t, map[string]any{"modelId": "0x01", "modelName": "qwen3-32b", "apiType": "openai", "apiUrl": srv.URL + "/v1/chat/completions", "apiKey": testKey})

	code, stdout, stderr := runCLI(t, "-config", cfg, "-verify", "-model", "0x01", "-json")
	require.Equal(t, 0, code, stderr)
	require.NotContains(t, stdout+stderr, testKey)

	var reports []struct {
		Api            *system.ModelApiSpec `json:"api"`
		Runtime        *apispec.Runtime     `json:"runtime"`
		VerifyTrace    []string             `json:"verifyTrace"`
		VerifyRequests []probeRequest       `json:"verifyRequests"`
		VerifyVerdicts []probeVerdict       `json:"verifyVerdicts"`
	}
	require.NoError(t, json.Unmarshal([]byte(stdout), &reports))
	require.Len(t, reports, 1)
	r := reports[0]

	require.NotNil(t, r.Runtime)
	require.Equal(t, []string{"chat_template_kwargs"}, r.Runtime.RejectedParams)
	require.Empty(t, r.Runtime.Verified)
	require.NotNil(t, r.Runtime.ToolsSupported)
	require.True(t, *r.Runtime.ToolsSupported)
	require.Equal(t, testAt, r.Runtime.At)

	require.NotNil(t, r.Api)
	require.Equal(t, "llamacpp", r.Api.Stack)
	require.Nil(t, r.Api.Bindings[system.IntentReasoningDisable], "the rejected binding is dropped")
	require.Nil(t, r.Api.Bindings[system.IntentReasoningEnable], "the sibling on the rejected parameter goes with it")
	require.Equal(t, "verify:rejected:chat_template_kwargs@1790330400", r.Api.Evidence[system.EvidenceKeyBinding(system.IntentReasoningDisable)])
	require.Equal(t, "verify:tools@1790330400", r.Api.Evidence[system.EvidenceKeyCapability("tools")])

	require.Equal(t, []probeRequest{
		{Method: "POST", Path: "/v1/chat/completions", Status: 400},
		{Method: "POST", Path: "/v1/chat/completions", Status: 200},
		{Method: "POST", Path: "/v1/chat/completions", Status: 200},
		{Method: "POST", Path: "/v1/chat/completions", Status: 200},
	}, r.VerifyRequests)
	require.Len(t, r.VerifyVerdicts, 4)
	require.Equal(t, probeVerdict{Probe: "reasoning.disable", Outcome: "rejected", Status: 400, Reason: "backend rejected chat_template_kwargs", RejectedParams: []string{"chat_template_kwargs"}}, r.VerifyVerdicts[0])

	joined := strings.Join(r.VerifyTrace, "\n")
	require.Contains(t, joined, "bindings: reasoning.disable dropped — backend rejected parameter chat_template_kwargs (verify:rejected:chat_template_kwargs@1790330400)")
	require.Contains(t, joined, "bindings: reasoning.enable dropped")
	require.Contains(t, joined, "thinking.mode: controllable -> tunable after the drop", "reasoning.budget from the stack table survives the drop")
}

// A 2xx that still carries reasoning_content refutes the binding: the
// binding line names the reason the checker publishes for the same verdict
// (verify:<reason slug>@<at>, modelhealth.Checker.publishedApi), not the
// intent's own name MergeRuntime wrote.
func TestVerifyRefutedBindingNamesReason(t *testing.T) {
	pinClock(t)
	srv := fakeBackend(t, fakeOpts{reasoning: true})
	cfg := writeConfig(t, map[string]any{"modelId": "0x01", "modelName": "qwen3-32b", "apiType": "openai", "apiUrl": srv.URL + "/v1/chat/completions", "apiKey": testKey})

	code, stdout, stderr := runCLI(t, "-config", cfg, "-verify", "-model", "qwen3-32b")
	require.Equal(t, 0, code, stderr)
	require.NotContains(t, stdout+stderr, testKey)

	require.Contains(t, stdout, "    thinking:    always_on\n")
	require.Contains(t, stdout, "      reasoning.disable: template_kwarg chat_template_kwargs.enable_thinking (boolean, value=false) [unverified: verify:reasoning_content_still_present@1790330400]\n")
	require.Contains(t, stdout, "      bindings.reasoning.disable -> verify:reasoning_content_still_present@1790330400\n")
	result, _, found := strings.Cut(stdout, "  verify:\n")
	require.True(t, found)
	require.NotContains(t, result, "verify:reasoning.disable@", "the api block carries the reason, not the intent's name")
	require.Contains(t, stdout, "      - reasoning.disable: refuted (HTTP 200): reasoning content still present\n")
	require.Contains(t, stdout, "      - bindings.reasoning.disable verified = false (verify:reasoning.disable@1790330400): reasoning content still present\n", "the merge trace keeps MergeRuntime's wording")
	require.Contains(t, stdout, "      - thinking.mode: controllable -> always_on (verify:reasoning_seen_with_disable@1790330400)\n")
}

// Detection finds nothing behind a bare chat endpoint: the cost warning
// reaches stderr before the first probe, the probes still run, and their
// verdicts make the api block (with the wire contract's source value)
// rather than leaving the report empty.
func TestVerifyUndetectedBackendPublishesProbeFacts(t *testing.T) {
	pinClock(t)
	srv := fakeBackend(t, fakeOpts{bare: true})
	cfg := writeConfig(t, map[string]any{"modelId": "0x01", "modelName": "mystery-model", "apiType": "openai", "apiUrl": srv.URL + "/v1/chat/completions", "apiKey": testKey})

	code, stdout, stderr := runCLI(t, "-config", cfg, "-verify", "-model", "mystery-model", "-json")
	require.Equal(t, 0, code, stderr)
	require.NotContains(t, stdout+stderr, testKey)
	require.True(t, strings.HasPrefix(stderr, "warning: stack not identified; -verify sends up to 6 inference requests"), stderr)

	var reports []struct {
		Api            *system.ModelApiSpec `json:"api"`
		Trace          []string             `json:"trace"`
		VerifyRequests []probeRequest       `json:"verifyRequests"`
		VerifyVerdicts []probeVerdict       `json:"verifyVerdicts"`
	}
	require.NoError(t, json.Unmarshal([]byte(stdout), &reports))
	require.Len(t, reports, 1)
	r := reports[0]

	require.Contains(t, strings.Join(r.Trace, "\n"), "no engine fingerprint matched")
	require.NotNil(t, r.Api)
	require.Equal(t, system.ApiSpecSourceDetected, r.Api.Source)
	require.Empty(t, r.Api.Stack)
	require.NotNil(t, r.Api.Capabilities)
	require.NotNil(t, r.Api.Capabilities.Tools)
	require.True(t, *r.Api.Capabilities.Tools)
	require.Equal(t, "verify:tools@1790330400", r.Api.Evidence[system.EvidenceKeyCapability("tools")])
	require.Equal(t, []string{"json_schema"}, r.Api.Capabilities.Structured)
	require.Equal(t, []probeRequest{
		{Method: "POST", Path: "/v1/chat/completions", Status: 200},
		{Method: "POST", Path: "/v1/chat/completions", Status: 200},
		{Method: "POST", Path: "/v1/chat/completions", Status: 200},
	}, r.VerifyRequests, "no reasoning.disable binding, so three probes")
	require.Len(t, r.VerifyVerdicts, 3)
}

// Detection finds nothing and every probe gets a 5xx: no fact was proven,
// so the report says nothing was detected instead of printing an api block
// of dashes with no source.
func TestVerifyNothingProvenPrintsNothingDetected(t *testing.T) {
	pinClock(t)
	srv := fakeBackend(t, fakeOpts{bare: true, chatStatus: http.StatusBadGateway})
	cfg := writeConfig(t, map[string]any{"modelId": "0x01", "modelName": "mystery-model", "apiType": "openai", "apiUrl": srv.URL + "/v1/chat/completions", "apiKey": testKey})

	code, stdout, stderr := runCLI(t, "-config", cfg, "-verify", "-model", "mystery-model")
	require.Equal(t, 0, code, stderr)
	require.True(t, strings.HasPrefix(stderr, "warning: stack not identified"), stderr)
	require.Contains(t, stdout, "  result: nothing detected\n")
	require.NotContains(t, stdout, "stack:       -")
	require.Contains(t, stdout, "      - POST /v1/chat/completions -> HTTP 502\n")
	require.Contains(t, stdout, "      - tools: unknown (HTTP 502): HTTP 502\n")
	require.Contains(t, stdout, "    merge:\n      (nothing changed)\n")
}

// An account out of credit answers 402 to the first probe: the run stops
// there, as the checker's does, and the report says how many were not sent.
func TestVerifyStopsAtAccountLimit(t *testing.T) {
	pinClock(t)
	srv := fakeBackend(t, fakeOpts{bare: true, chatStatus: http.StatusPaymentRequired})
	cfg := writeConfig(t, map[string]any{"modelId": "0x01", "modelName": "mystery-model", "apiType": "openai", "apiUrl": srv.URL + "/v1/chat/completions", "apiKey": testKey})

	code, stdout, stderr := runCLI(t, "-config", cfg, "-verify", "-model", "mystery-model")
	require.Equal(t, 0, code, stderr)
	require.Contains(t, stdout, "    requests:\n      - POST /v1/chat/completions -> HTTP 402\n    verdicts:\n", "one request, no more")
	require.Contains(t, stdout, "      - tools: unknown (HTTP 402): account or rate limit\n")
	require.Contains(t, stdout, "      (2 more probes not sent: the backend answered with an account or rate-limit status)\n")

	code, stdout, stderr = runCLI(t, "-config", cfg, "-verify", "-model", "mystery-model", "-json")
	require.Equal(t, 0, code, stderr)
	var reports []struct {
		VerifyVerdicts []probeVerdict `json:"verifyVerdicts"`
		VerifyNotSent  int            `json:"verifyNotSent"`
	}
	require.NoError(t, json.Unmarshal([]byte(stdout), &reports))
	require.Len(t, reports, 1)
	require.Len(t, reports[0].VerifyVerdicts, 1)
	require.Equal(t, 2, reports[0].VerifyNotSent)
}

// NotSent counts the probes an account answer left unsent, the vLLM render
// check only when it would have run: while the completion left
// reasoning.disable undecided.
func TestVerifyNotSentCountsTheRenderCheck(t *testing.T) {
	const okAnswer = `{"id":"1","object":"chat.completion","model":"qwen3-8b","choices":[{"index":0,"message":{"role":"assistant","content":"5"},"finish_reason":"stop"}]}`
	for _, tc := range []struct {
		name    string
		disable int // the reasoning.disable completion's status; the tool probe answers 402
		verdict string
		notSent int
	}{
		{"the first probe answers 402", http.StatusPaymentRequired, "unknown", 4},
		{"the completion decided the binding", http.StatusOK, "accepted", 2},
		{"the completion left it undecided", http.StatusBadRequest, "unknown", 3},
	} {
		t.Run(tc.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				var body map[string]any
				require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
				w.Header().Set("Content-Type", "application/json")
				switch {
				case r.URL.Path != "/v1/chat/completions":
					w.WriteHeader(http.StatusNotFound)
				case body["chat_template_kwargs"] != nil && tc.disable != http.StatusOK:
					w.WriteHeader(tc.disable)
					_, _ = w.Write([]byte(`{"error":{"message":"no"}}`))
				case body["tools"] != nil:
					w.WriteHeader(http.StatusPaymentRequired)
					_, _ = w.Write([]byte(`{"error":{"message":"insufficient balance"}}`))
				default:
					_, _ = w.Write([]byte(okAnswer))
				}
			}))
			t.Cleanup(srv.Close)
			cfg := config.ModelConfig{ModelName: "qwen3-8b", ApiType: "openai", ApiURL: srv.URL + "/v1/chat/completions", ApiKey: testKey}
			api := &system.ModelApiSpec{Stack: "vllm", ModelFamily: "qwen3", Source: system.ApiSpecSourceDetected, Thinking: &system.ThinkingSpec{Mode: system.ThinkingModeControllable},
				Bindings: map[string]*system.ParamBinding{
					system.IntentReasoningDisable: {Kind: system.BindingKindTemplateKwarg, Param: "chat_template_kwargs.enable_thinking", ParamType: "boolean", Value: false},
				}}
			res, err := verify(context.Background(), cfg, api, 5*time.Second, lib.NewTestLogger())
			require.NoError(t, err)
			require.Equal(t, tc.notSent, res.NotSent)
			require.Equal(t, tc.verdict, res.Verdicts[0].Outcome)
			require.Equal(t, http.StatusPaymentRequired, res.Verdicts[len(res.Verdicts)-1].Status, "the run ends at the account answer")
			require.Len(t, res.Requests, len(res.Verdicts))
		})
	}
}

func TestVerifyWarnsBeforeProbingHostedStack(t *testing.T) {
	require.Contains(t, hostedWarning(&system.ModelApiSpec{Stack: "openai"}), "hosted vendor")
	require.Contains(t, hostedWarning(&system.ModelApiSpec{Stack: "vllm", Via: "litellm"}), "through litellm")
	require.Contains(t, hostedWarning(nil), "not identified")
	require.Empty(t, hostedWarning(&system.ModelApiSpec{Stack: "vllm"}))
}

func TestProbeRequestString(t *testing.T) {
	require.Equal(t, "POST /v1/chat/completions -> HTTP 400", probeRequest{Method: "POST", Path: "/v1/chat/completions", Status: 400}.String())
	require.Equal(t, "POST /v1/chat/completions -> no response", probeRequest{Method: "POST", Path: "/v1/chat/completions", Failed: true}.String())
}

// -h keeps flag.ExitOnError's exit 0; a flag the command does not know is
// still a parse error.
func TestHelpExitsZero(t *testing.T) {
	code, _, stderr := runCLI(t, "-h")
	require.Equal(t, 0, code)
	require.Contains(t, stderr, "-verify")

	code, _, stderr = runCLI(t, "-no-such-flag")
	require.Equal(t, 2, code)
	require.Contains(t, stderr, "flag provided but not defined")
}
