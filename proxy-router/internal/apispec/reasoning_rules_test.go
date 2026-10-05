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

func TestContainsVersionNeedsANonDigitBoundary(t *testing.T) {
	for _, tc := range []struct {
		name, sub string
		want      bool
	}{
		{"Kimi-K2.5", "k2-5", true},
		{"kimi-k2-6-thinking", "k2-6", true},
		{"kimi-k2-50b", "k2-5", false},
		{"kimi-k2-50b-k2-5", "k2-5", true},
		{"EXAONE-4.0.1-32B", "exaone-4", true},
		{"exaone-40b", "exaone-4", false},
	} {
		require.Equal(t, tc.want, containsVersion(tc.name, tc.sub), tc.name)
	}
	_, b := bindingsForFamily("kimi", "kimi-k2-50b")
	require.Nil(t, b, "k2-50b is not K2.5")
	_, b = bindingsForFamily("exaone", "exaone-40b")
	require.Nil(t, b, "exaone-40b is not EXAONE 4")
}

func TestClaudeVersionIgnoresSeparatorSpelling(t *testing.T) {
	for _, name := range []string{"claude_sonnet_4_5", "claude-sonnet-4-5@20250929", "Claude Sonnet 4.5"} {
		_, b := claudeBindings(name)
		require.NotNil(t, b[system.IntentReasoningBudget], "%s parses as 4.5", name)
		require.Equal(t, []string{"low", "medium", "high"}, b[system.IntentReasoningEffort].EnumValues, name)
	}
	_, b := claudeBindings("claude-opus-4-1@20250805")
	require.Nil(t, b[system.IntentReasoningEffort], "4.1 takes no effort")
	require.NotNil(t, b[system.IntentReasoningBudget])
}

func TestComposeThinkingLabelKeepsServedHybridToggles(t *testing.T) {
	hybrid, lines := ComposeWithTrace(Evidence{Stack: "vllm", ModelName: "team-qwen3-thinking", ServedModelID: "Qwen/Qwen3-32B"})
	require.Equal(t, "qwen3", hybrid.ModelFamily)
	require.Equal(t, system.ThinkingModeControllable, hybrid.Thinking.Mode, "the served weights are a hybrid; the operator's label does not remove its switch")
	require.Equal(t, "chat_template_kwargs.enable_thinking", hybrid.Bindings[system.IntentReasoningDisable].Param)
	require.Contains(t, strings.Join(lines, "\n"), "names no checkpoint")

	checkpoint := Compose(Evidence{Stack: "vllm", ModelName: "team-qwen3", ServedModelID: "Qwen/Qwen3-235B-A22B-Thinking-2507"})
	require.Nil(t, checkpoint.Bindings[system.IntentReasoningDisable], "the served id names the checkpoint")
	require.Nil(t, checkpoint.Bindings[system.IntentReasoningEnable])
	require.Equal(t, system.ThinkingModeTunable, checkpoint.Thinking.Mode)
}

func TestWithoutTemplateTogglesKeepsAPIParamsAndLimits(t *testing.T) {
	kept := withoutTemplateToggles(budgetKwargBindings())
	require.Nil(t, kept[system.IntentReasoningDisable], "thinking_budget=0 is a template switch")
	require.NotNil(t, kept[system.IntentReasoningBudget], "a budget still limits a checkpoint")

	_, claude := claudeBindings("claude-sonnet-4-5")
	require.Len(t, withoutTemplateToggles(claude), len(claude), "API params are validated by the vendor, so they stay")

	_, nemotron := bindingsForFamily("nemotron", "llama-3.3-nemotron-super-49b")
	require.Empty(t, withoutTemplateToggles(nemotron), "system prompt switches go too")
}

func TestComposeThinkingCheckpointKeepsFamilyBudget(t *testing.T) {
	api := Compose(Evidence{Stack: "vllm", ModelName: "seed-oss-36b-thinking"})
	require.Equal(t, "seed-oss", api.ModelFamily)
	require.Nil(t, api.Bindings[system.IntentReasoningDisable], "thinking_budget=0 is the template switch a checkpoint may ignore")
	require.Equal(t, "chat_template_kwargs.thinking_budget", api.Bindings[system.IntentReasoningBudget].Param, "the family's budget still applies")
}

func TestComposeListedEffortLevelsAreReasoningEvidence(t *testing.T) {
	api := Compose(Evidence{Stack: "venice", ModelName: "some-new-model", RegistryEfforts: []string{"low", "high"}})
	require.NotNil(t, api.Thinking, "a listing that names effort levels says the model reasons")
	require.Equal(t, []string{"low", "high"}, api.Bindings[system.IntentReasoningEffort].EnumValues)
}

func TestComposeOllamaThinkingCheckpointGetsNoEffortNone(t *testing.T) {
	checkpoint := Compose(Evidence{Stack: "ollama", ModelName: "qwen3-thinking:32b", Architecture: "qwen3", OllamaThinking: true})
	effort := checkpoint.Bindings[system.IntentReasoningEffort]
	require.NotNil(t, effort)
	require.NotContains(t, effort.EnumValues, "none", "none is ollama's think=false, the toggle a checkpoint is not given")
	require.Contains(t, stackBindings["ollama"][system.IntentReasoningEffort].EnumValues, "none", "the shared stack table is untouched")

	hybrid := Compose(Evidence{Stack: "ollama", ModelName: "qwen3:8b", Architecture: "qwen3", OllamaThinking: true})
	require.NotNil(t, hybrid.Bindings[system.IntentReasoningDisable])
	require.Contains(t, hybrid.Bindings[system.IntentReasoningEffort].EnumValues, "none")
}

func TestNamesIgnoreVeniceInlineParams(t *testing.T) {
	require.False(t, thinkingNamed("llama-3.3-70b:strip_thinking_response=true"), "a parameter is not a thinking token")
	require.False(t, thinkingNamed("qwen3-32b:disable_thinking=true"))
	require.True(t, thinkingNamed("qwen3-thinking:32b"), "an ollama tag is part of the name")
	require.Equal(t, "kimi-k2-5", normalizeModelName("Kimi-K2.5:enable_web_search=on&include_venice_system_prompt=false"))

	plain := Compose(Evidence{Stack: "venice", ModelName: "llama-3.3-70b:strip_thinking_response=true"})
	require.Nil(t, plain.Thinking, "no reasoning evidence from a parameter")
	require.Nil(t, plain.Bindings[system.IntentReasoningDisable])
}

func TestComposeNoEffortDropsTheEffortKnob(t *testing.T) {
	api, lines := ComposeWithTrace(Evidence{Stack: "venice", ModelName: "qwen3-235b", GatewayReasoning: true, NoEffort: true})
	require.Nil(t, api.Bindings[system.IntentReasoningEffort])
	require.Equal(t, "venice_parameters.disable_thinking", api.Bindings[system.IntentReasoningDisable].Param)
	require.Equal(t, system.ThinkingModeControllable, api.Thinking.Mode)
	require.Contains(t, strings.Join(lines, "\n"), "takes no effort level")
}
