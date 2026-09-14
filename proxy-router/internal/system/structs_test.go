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
