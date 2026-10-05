package apispec

import (
	"strings"
	"testing"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/system"
	"github.com/stretchr/testify/require"
)

func TestFamilyFromNameIgnoresSeparatorSpelling(t *testing.T) {
	cases := map[string]string{
		"kimi-k2-7-code":            "kimi",
		"Kimi K2.7 Code":            "kimi",
		"deepseek_v3_1":             "deepseek-v3.1",
		"deepseek-ai/DeepSeek-V3.1": "deepseek-v3.1",
		"DeepSeek-V3.1-Terminus":    "deepseek-v3.1",
		"deepseek-chat-v3.1":        "deepseek-v3.1",
		"deepseek_r1_distill":       "deepseek-r1",
		"EXAONE_4.0_32B":            "exaone",
		"Phi_4_reasoning":           "phi",
		"seed_oss_36b":              "seed-oss",
		"gpt_oss_120b":              "gpt-oss",
	}
	for name, want := range cases {
		require.Equal(t, want, FamilyFromName(name), name)
	}
}

func TestFamilyMembersIgnoreSeparatorSpelling(t *testing.T) {
	for _, name := range []string{"minimax-m27", "MiniMax M2.7", "minimax_m2_5"} {
		requireAlwaysOn(t, "minimax", name)
	}
	for _, name := range []string{"exaone_deep_32b", "EXAONE Deep 32B"} {
		requireAlwaysOn(t, "exaone", name)
	}
	for _, name := range []string{"exaone_4_0_32b", "EXAONE 4.0 32B"} {
		_, b := bindingsForFamily("exaone", name)
		requireKwargBool(t, b, "enable_thinking", name)
	}
	for _, name := range []string{"Gemma 4 31B", "gemma_4_12b"} {
		_, b := bindingsForFamily("gemma", name)
		requireKwargBool(t, b, "enable_thinking", name)
	}
}

func TestThinkingNamed(t *testing.T) {
	for _, name := range []string{
		"Qwen/Qwen3-235B-A22B-Thinking-2507", "Qwen 3 235B A22B Thinking 2507", "kimi-k2-thinking",
		"glm-4.7-thinking:web", "qwen3-thinking:32b", "thinking",
	} {
		require.True(t, thinkingNamed(name), name)
	}
	for _, name := range []string{
		"glm-5.1-non-thinking", "glm-5.1-non-thinking:web", "GLM 5.1 Non Thinking", "model-no-thinking",
		"nonthinking-model", "deep-thinker", "qwen3-32b", "",
	} {
		require.False(t, thinkingNamed(name), name)
	}
}

func TestComposeVeniceListingNoneLiftsFamilyAlwaysOn(t *testing.T) {
	api, lines := ComposeWithTrace(Evidence{Stack: "venice", ModelName: "minimax-m27", GatewayReasoning: true, RegistryEfforts: []string{"none", "low", "medium", "high"}})
	require.Equal(t, "minimax", api.ModelFamily)
	require.Equal(t, system.ThinkingModeControllable, api.Thinking.Mode)
	require.Equal(t, "venice_parameters.disable_thinking", api.Bindings[system.IntentReasoningDisable].Param)
	effort := api.Bindings[system.IntentReasoningEffort]
	require.Equal(t, "reasoning.effort", effort.Param)
	require.Equal(t, []string{"none", "low", "medium", "high"}, effort.EnumValues, "the model's own levels from the listing")
	require.Contains(t, strings.Join(lines, "\n"), "offers reasoning effort none")
}

func TestComposeVeniceListingWithoutNoneKeepsAlwaysOn(t *testing.T) {
	for _, name := range []string{"kimi-k2-7-code", "minimax-m25"} {
		api := Compose(Evidence{Stack: "venice", ModelName: name, GatewayReasoning: true, RegistryEfforts: []string{"low", "medium", "high"}})
		require.Equal(t, system.ThinkingModeAlwaysOn, api.Thinking.Mode, name)
		require.Nil(t, api.Bindings[system.IntentReasoningDisable], name)
		effort := api.Bindings[system.IntentReasoningEffort]
		require.NotNil(t, effort, "%s: the listing names effort levels, so it can still think less", name)
		require.Equal(t, []string{"low", "medium", "high"}, effort.EnumValues, name)
		require.Equal(t, "venice_parameters.strip_thinking_response", api.Bindings[system.IntentReasoningFormat].Param, name)
	}

	r1 := Compose(Evidence{Stack: "venice", ModelName: "deepseek-r1", GatewayReasoning: true})
	require.Equal(t, system.ThinkingModeAlwaysOn, r1.Thinking.Mode)
	require.Nil(t, r1.Bindings[system.IntentReasoningEffort], "no listed levels: an always-on model takes no effort knob")
	require.Nil(t, r1.Bindings[system.IntentReasoningDisable])
}

func TestComposeVeniceThinkingCheckpointIsControllable(t *testing.T) {
	api := Compose(Evidence{Stack: "venice", ModelName: "qwen3-235b-a22b-thinking-2507", GatewayReasoning: true, RegistryEfforts: []string{"low", "medium", "high"}})
	require.Equal(t, system.ThinkingModeControllable, api.Thinking.Mode, "the name no longer forces always_on; venice's own switch applies")
	require.Equal(t, "venice_parameters.disable_thinking", api.Bindings[system.IntentReasoningDisable].Param)
	require.Equal(t, []string{"low", "medium", "high"}, api.Bindings[system.IntentReasoningEffort].EnumValues)
}

func TestComposeThinkingNameIsReasoningEvidenceOnGateways(t *testing.T) {
	api := Compose(Evidence{Stack: "venice", ModelName: "glm-4.7-thinking"})
	require.NotNil(t, api.Thinking, "a detected gateway gets its reasoning knobs from a thinking name alone")
	require.Equal(t, "venice_parameters.disable_thinking", api.Bindings[system.IntentReasoningDisable].Param)

	plain := Compose(Evidence{Stack: "venice", ModelName: "glm-5.1-non-thinking"})
	require.Nil(t, plain.Thinking, "a non-thinking name is no reasoning evidence for a detected gateway")
	require.Nil(t, plain.Bindings[system.IntentReasoningDisable])
}

func TestComposeVeniceEffortSurvivesLiteLLMFilter(t *testing.T) {
	api := Compose(Evidence{
		Stack: "venice", Via: "litellm", ModelName: "my-chat", ServedModelID: "minimax-m27", GatewayReasoning: true,
		RegistryEfforts: []string{"none", "low", "medium", "high"}, LiteLLMSeen: true, LiteLLMSupportedParams: []string{"temperature", "tools"},
	})
	require.Equal(t, system.ThinkingModeControllable, api.Thinking.Mode)
	effort := api.Bindings[system.IntentReasoningEffort]
	require.NotNil(t, effort, "reasoning.effort is not a standard root, so litellm forwards it")
	require.Equal(t, "reasoning.effort", effort.Param)
	require.Equal(t, []string{"none", "low", "medium", "high"}, effort.EnumValues)
	require.Equal(t, "venice_parameters.disable_thinking", api.Bindings[system.IntentReasoningDisable].Param)
}
