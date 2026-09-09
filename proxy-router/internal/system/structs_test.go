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
