package main

import (
	"encoding/json"
	"testing"

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
