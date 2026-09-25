package system

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestModelApiSpecValueFalseAndZeroSurviveJSON(t *testing.T) {
	report := ModelHealthReport{
		ModelID: "0x01",
		Status:  ModelHealthStatusHealthy,
		Api: &ModelApiSpec{
			Stack:       "vllm",
			ModelFamily: "qwen3",
			Thinking:    &ThinkingSpec{Mode: ThinkingModeControllable},
			Bindings: map[string]*ParamBinding{
				IntentReasoningDisable: {Kind: BindingKindTemplateKwarg, Param: "chat_template_kwargs.enable_thinking", ParamType: "boolean", Value: false},
				IntentReasoningBudget:  {Kind: BindingKindBodyParam, Param: "thinking_token_budget", ParamType: "number", Value: 0},
			},
			Parameters: []string{"messages", "temperature"},
			DeclaredAt: 1,
		},
	}
	raw, err := json.Marshal(report)
	require.NoError(t, err)
	s := string(raw)
	require.Contains(t, s, `"api":{`)
	require.Contains(t, s, `"stack":"vllm"`)
	require.Contains(t, s, `"modelFamily":"qwen3"`)
	require.Contains(t, s, `"thinking":{"mode":"controllable"}`)
	require.Contains(t, s, `"value":false`)
	require.Contains(t, s, `"value":0`)
	require.Contains(t, s, `"declaredAt":1`)

	var back ModelHealthReport
	require.NoError(t, json.Unmarshal(raw, &back))
	require.Equal(t, false, back.Api.Bindings[IntentReasoningDisable].Value)
}

func TestModelHealthReportOmitsApiWhenNil(t *testing.T) {
	raw, err := json.Marshal(ModelHealthReport{ModelID: "0x01", Status: ModelHealthStatusNoBid})
	require.NoError(t, err)
	require.NotContains(t, string(raw), `"api"`)
}

func TestModelApiSpecSourceOnWire(t *testing.T) {
	raw, err := json.Marshal(ModelApiSpec{Stack: "vllm", Source: ApiSpecSourceDetected})
	require.NoError(t, err)
	require.Contains(t, string(raw), `"source":"detected"`)
	raw, err = json.Marshal(ModelApiSpec{Stack: "vllm"})
	require.NoError(t, err)
	require.NotContains(t, string(raw), `"source"`)
	require.Equal(t, "declared", ApiSpecSourceDeclared)
}

func TestModelApiSpecCloneIsDeep(t *testing.T) {
	var nilSpec *ModelApiSpec
	require.Nil(t, nilSpec.Clone())
	orig := &ModelApiSpec{
		Stack: "vllm", Source: ApiSpecSourceDetected, Thinking: &ThinkingSpec{Mode: ThinkingModeTunable},
		Bindings: map[string]*ParamBinding{
			IntentReasoningEffort:    {Kind: BindingKindTemplateKwarg, Param: "chat_template_kwargs.reasoning_effort", ParamType: "enum", EnumValues: []string{"low", "high"}},
			IntentResponseFormatJSON: {Kind: BindingKindBodyParam, Param: "response_format", ParamType: "object", Value: map[string]any{"type": "json_object"}},
		},
		Parameters: []string{"messages"}, DeclaredAt: 7,
	}
	c := orig.Clone()
	require.Equal(t, orig, c)
	c.DeclaredAt = 99
	c.Thinking.Mode = ThinkingModeAlwaysOn
	c.Bindings[IntentReasoningEffort].EnumValues[0] = "mutated"
	c.Bindings[IntentResponseFormatJSON].Value.(map[string]any)["type"] = "mutated"
	c.Parameters[0] = "mutated"
	delete(c.Bindings, IntentResponseFormatJSON)
	require.Equal(t, int64(7), orig.DeclaredAt)
	require.Equal(t, ThinkingModeTunable, orig.Thinking.Mode)
	require.Equal(t, []string{"low", "high"}, orig.Bindings[IntentReasoningEffort].EnumValues)
	require.Equal(t, map[string]any{"type": "json_object"}, orig.Bindings[IntentResponseFormatJSON].Value)
	require.Equal(t, []string{"messages"}, orig.Parameters)
}

func int64p(v int64) *int64 { return &v }
func boolp(v bool) *bool    { return &v }

func TestModelCapabilitiesExplicitFalseAndZeroSurviveJSON(t *testing.T) {
	var caps ModelCapabilities
	require.NoError(t, json.Unmarshal([]byte(`{"tools": false, "context": 0}`), &caps))
	require.NotNil(t, caps.Tools)
	require.False(t, *caps.Tools)
	require.NotNil(t, caps.Context, "0 is invalid as a capability (setters reject it) but the struct must carry it")
	require.Equal(t, int64(0), *caps.Context)
	raw, err := json.Marshal(caps)
	require.NoError(t, err)
	require.JSONEq(t, `{"tools": false, "context": 0}`, string(raw))

	api := ModelApiSpec{
		Stack: "vllm",
		Capabilities: &ModelCapabilities{
			Context: int64p(40960), ContextTrained: int64p(131072), MaxOutput: int64p(4096),
			Modalities: &Modalities{Input: []string{"text", "image"}, Output: []string{"text"}},
			Tools:      boolp(true), ToolsParallel: boolp(false), Structured: []string{"json_object", "json_schema"},
			Reasoning: boolp(true), WebSearch: boolp(false), Embeddings: boolp(false), PromptCache: boolp(true), Quantization: "fp8",
		},
		Evidence: map[string]string{EvidenceKeyCapability("context"): "vllm:/v1/models.max_model_len", EvidenceKeyThinkingDefault: "ollama:/api/show.thinking.default"},
		Thinking: &ThinkingSpec{Mode: ThinkingModeControllable, Default: "on"},
		Bindings: map[string]*ParamBinding{
			IntentWebSearchEnable:  {Kind: BindingKindToolEntry, Param: "tools", ParamType: "object", Value: map[string]any{"type": "web_search_20250305", "name": "web_search"}, Verified: boolp(false), VerifiedAt: 5},
			IntentReasoningDisable: {Kind: BindingKindBodyParam, Param: "reasoning_effort", Value: "none"},
		},
	}
	raw, err = json.Marshal(api)
	require.NoError(t, err)
	s := string(raw)
	require.Contains(t, s, `"capabilities":{"context":40960,"contextTrained":131072,"maxOutput":4096,"modalities":{"input":["text","image"],"output":["text"]},"tools":true,"toolsParallel":false,"structured":["json_object","json_schema"],"reasoning":true,"webSearch":false,"embeddings":false,"promptCache":true,"quantization":"fp8"}`)
	require.Contains(t, s, `"evidence":{"capabilities.context":"vllm:/v1/models.max_model_len","thinking.default":"ollama:/api/show.thinking.default"}`)
	require.Contains(t, s, `"thinking":{"mode":"controllable","default":"on"}`)
	require.Contains(t, s, `"kind":"tool_entry","param":"tools"`)
	require.Contains(t, s, `"verified":false,"verifiedAt":5`)
	require.NotContains(t, s, `"verified":true`)

	var back ModelApiSpec
	require.NoError(t, json.Unmarshal(raw, &back))
	require.Equal(t, api.Capabilities, back.Capabilities)
	require.Equal(t, api.Evidence, back.Evidence)
	require.Equal(t, api.Thinking, back.Thinking)
	require.Equal(t, false, *back.Bindings[IntentWebSearchEnable].Verified)
	require.Nil(t, back.Bindings[IntentReasoningDisable].Verified, "never checked stays absent")
	require.Equal(t, int64(0), back.Bindings[IntentReasoningDisable].VerifiedAt)
}

func TestModelCapabilitiesNilVersusEmptyOnWire(t *testing.T) {
	raw, err := json.Marshal(ModelApiSpec{Stack: "vllm"})
	require.NoError(t, err)
	require.Equal(t, `{"stack":"vllm"}`, string(raw), "no capabilities, evidence or thinking.default without evidence")

	raw, err = json.Marshal(ModelApiSpec{Stack: "vllm", Capabilities: &ModelCapabilities{}, Evidence: map[string]string{}})
	require.NoError(t, err)
	require.Equal(t, `{"stack":"vllm","capabilities":{}}`, string(raw), "an empty struct is serialised; an empty map is not")

	raw, err = json.Marshal(ModelCapabilities{Structured: []string{}, Modalities: &Modalities{Input: []string{}}})
	require.NoError(t, err)
	require.Equal(t, `{"modalities":{}}`, string(raw), "empty lists are omitted like nil ones")

	raw, err = json.Marshal(ThinkingSpec{Mode: ThinkingModeAlwaysOn})
	require.NoError(t, err)
	require.Equal(t, `{"mode":"always_on"}`, string(raw))
}

func TestEvidenceKeyHelpers(t *testing.T) {
	require.Equal(t, "capabilities.context", EvidenceKeyCapability("context"))
	require.Equal(t, "capabilities.modalities.input", EvidenceKeyCapability("modalities.input"))
	require.Equal(t, "bindings.reasoning.disable", EvidenceKeyBinding(IntentReasoningDisable))
	require.Equal(t, "bindings.web_search.enable", EvidenceKeyBinding(IntentWebSearchEnable))
	require.Equal(t, "thinking.default", EvidenceKeyThinkingDefault)
	require.Equal(t, "tool_entry", BindingKindToolEntry)
	require.Equal(t, "web_search.enable", IntentWebSearchEnable)
	require.Equal(t, "web_search.citations", IntentWebSearchCitations)
}

func TestModelApiSpecCloneDeepCopiesCapabilitiesAndEvidence(t *testing.T) {
	orig := &ModelApiSpec{
		Stack:    "vllm",
		Thinking: &ThinkingSpec{Mode: ThinkingModeControllable, Default: "on"},
		Capabilities: &ModelCapabilities{
			Context: int64p(40960), ContextTrained: int64p(131072), MaxOutput: int64p(4096),
			Modalities: &Modalities{Input: []string{"text", "image"}, Output: []string{"text"}},
			Tools:      boolp(true), ToolsParallel: boolp(false), Structured: []string{"json_schema"},
			Reasoning: boolp(true), WebSearch: boolp(false), Embeddings: boolp(false), PromptCache: boolp(true), Quantization: "fp8",
		},
		Evidence: map[string]string{"capabilities.context": "vllm:/v1/models.max_model_len"},
		Bindings: map[string]*ParamBinding{IntentReasoningDisable: {Kind: BindingKindBodyParam, Param: "x", Verified: boolp(true), VerifiedAt: 3}},
	}
	c := orig.Clone()
	require.Equal(t, orig, c)

	*c.Capabilities.Context = 1
	*c.Capabilities.ContextTrained = 1
	*c.Capabilities.MaxOutput = 1
	*c.Capabilities.Tools = false
	*c.Capabilities.ToolsParallel = true
	*c.Capabilities.Reasoning = false
	*c.Capabilities.WebSearch = true
	*c.Capabilities.Embeddings = true
	*c.Capabilities.PromptCache = false
	c.Capabilities.Modalities.Input[0] = "mutated"
	c.Capabilities.Modalities.Output[0] = "mutated"
	c.Capabilities.Structured[0] = "mutated"
	c.Capabilities.Quantization = "mutated"
	c.Evidence["capabilities.context"] = "mutated"
	c.Thinking.Default = "off"
	*c.Bindings[IntentReasoningDisable].Verified = false

	require.Equal(t, int64(40960), *orig.Capabilities.Context)
	require.Equal(t, int64(131072), *orig.Capabilities.ContextTrained)
	require.Equal(t, int64(4096), *orig.Capabilities.MaxOutput)
	require.True(t, *orig.Capabilities.Tools)
	require.False(t, *orig.Capabilities.ToolsParallel)
	require.True(t, *orig.Capabilities.Reasoning)
	require.False(t, *orig.Capabilities.WebSearch)
	require.False(t, *orig.Capabilities.Embeddings)
	require.True(t, *orig.Capabilities.PromptCache)
	require.Equal(t, []string{"text", "image"}, orig.Capabilities.Modalities.Input)
	require.Equal(t, []string{"text"}, orig.Capabilities.Modalities.Output)
	require.Equal(t, []string{"json_schema"}, orig.Capabilities.Structured)
	require.Equal(t, "fp8", orig.Capabilities.Quantization)
	require.Equal(t, "vllm:/v1/models.max_model_len", orig.Evidence["capabilities.context"])
	require.Equal(t, "on", orig.Thinking.Default)
	require.True(t, *orig.Bindings[IntentReasoningDisable].Verified)

	var nilCaps *ModelCapabilities
	require.Nil(t, nilCaps.Clone())
	empty := (&ModelCapabilities{}).Clone()
	require.Equal(t, &ModelCapabilities{}, empty)
	require.Nil(t, (&ModelApiSpec{Stack: "vllm"}).Clone().Evidence, "nil stays nil, not an empty map")
}
