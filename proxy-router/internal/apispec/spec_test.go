package apispec

import (
	"strings"
	"testing"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/config"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/system"
	"github.com/stretchr/testify/require"
)

func build(stack, name, family string) *system.ModelApiSpec {
	apiType, ok := config.StackTransport[stack]
	if !ok {
		apiType = "openai"
	}
	return Build(config.ModelConfig{ModelName: name, ApiType: apiType, ApiStack: stack, ModelFamily: family, ApiURL: "http://h/v1/chat/completions"})
}

func TestBuildVLLMQwen3(t *testing.T) {
	api := build("vllm", "Qwen/Qwen3-235B-A22B", "")
	require.NotNil(t, api)
	require.Equal(t, "vllm", api.Stack)
	require.Equal(t, "qwen3", api.ModelFamily)
	require.Equal(t, system.ThinkingModeControllable, api.Thinking.Mode)
	disable := api.Bindings[system.IntentReasoningDisable]
	require.Equal(t, system.BindingKindTemplateKwarg, disable.Kind)
	require.Equal(t, "chat_template_kwargs.enable_thinking", disable.Param)
	require.Equal(t, false, disable.Value)
	require.Equal(t, true, api.Bindings[system.IntentReasoningEnable].Value)
	require.Equal(t, "thinking_token_budget", api.Bindings[system.IntentReasoningBudget].Param)
	require.Equal(t, "top_k", api.Bindings[system.IntentSamplingTopK].Param)
	require.Equal(t, map[string]any{"type": "json_object"}, api.Bindings[system.IntentResponseFormatJSON].Value)
	require.Contains(t, api.Parameters, "logprobs")
}

func TestBuildExplicitModelFamilyWins(t *testing.T) {
	api := build("vllm", "my-private-alias", "deepseek-v3.1")
	require.Equal(t, "deepseek-v3.1", api.ModelFamily)
	require.Equal(t, "chat_template_kwargs.thinking", api.Bindings[system.IntentReasoningDisable].Param)
}

func TestBuildPlainModelHasNoReasoning(t *testing.T) {
	api := build("vllm", "meta-llama/Llama-3.3-70B-Instruct", "")
	require.Equal(t, "llama", api.ModelFamily)
	require.Nil(t, api.Thinking)
	for intent := range api.Bindings {
		require.False(t, strings.HasPrefix(intent, "reasoning."), intent)
	}
	require.Equal(t, "repetition_penalty", api.Bindings[system.IntentSamplingRepetitionPenalty].Param)
}

func TestBuildAlwaysOnKeepsOnlyReasoningFormat(t *testing.T) {
	api := build("sglang", "deepseek-ai/DeepSeek-R1", "")
	require.Equal(t, system.ThinkingModeAlwaysOn, api.Thinking.Mode)
	require.Nil(t, api.Bindings[system.IntentReasoningDisable])
	require.Nil(t, api.Bindings[system.IntentReasoningEffort])
	require.Equal(t, "separate_reasoning", api.Bindings[system.IntentReasoningFormat].Param)
}

func TestBuildGptOssOnLlamacppReasoningBindings(t *testing.T) {
	api := build("llamacpp", "gpt-oss-120b", "")
	require.Equal(t, system.ThinkingModeControllable, api.Thinking.Mode)
	require.Equal(t, "chat_template_kwargs.reasoning_effort", api.Bindings[system.IntentReasoningEffort].Param)

	budget := api.Bindings[system.IntentReasoningBudget]
	require.NotNil(t, budget)
	require.Equal(t, "reasoning_budget", budget.Param)
	require.NotEmpty(t, budget.Hint)

	disable := api.Bindings[system.IntentReasoningDisable]
	require.NotNil(t, disable)
	require.Equal(t, "reasoning_budget", disable.Param)
	require.Equal(t, 0, disable.Value)
}

func TestBuildOllamaRewritesToReasoningEffort(t *testing.T) {
	api := build("ollama", "qwen3:32b", "")
	require.Equal(t, "reasoning_effort", api.Bindings[system.IntentReasoningDisable].Param)
	require.Equal(t, "none", api.Bindings[system.IntentReasoningDisable].Value)
	require.Equal(t, system.BindingKindBodyParam, api.Bindings[system.IntentReasoningEnable].Kind)
	require.Equal(t, "reasoning_effort", api.Bindings[system.IntentReasoningEnable].Param)
	require.Equal(t, "medium", api.Bindings[system.IntentReasoningEnable].Value)
	require.Equal(t, "options.num_ctx", api.Bindings[system.IntentContextNumCtx].Param)

	gptoss := build("ollama", "gpt-oss:120b", "")
	require.Equal(t, system.ThinkingModeTunable, gptoss.Thinking.Mode)
	require.Nil(t, gptoss.Bindings[system.IntentReasoningDisable])
	require.Equal(t, "reasoning_effort", gptoss.Bindings[system.IntentReasoningEffort].Param)

	seed := build("ollama", "seed-oss:36b", "")
	require.Nil(t, seed.Bindings[system.IntentReasoningBudget], "budget kwarg has no ollama equivalent")
	require.Equal(t, system.ThinkingModeControllable, seed.Thinking.Mode)
}

func TestBuildGatewayPresetsSkipFamilyTemplateKwargs(t *testing.T) {
	venice := build("venice", "qwen3-235b", "")
	require.Equal(t, "venice_parameters.disable_thinking", venice.Bindings[system.IntentReasoningDisable].Param)
	require.Equal(t, system.ThinkingModeControllable, venice.Thinking.Mode)
	for intent, b := range venice.Bindings {
		require.NotEqual(t, system.BindingKindTemplateKwarg, b.Kind, intent)
	}
	require.NotEmpty(t, venice.Parameters, "venice preset must advertise its documented parameter list")
	require.Contains(t, venice.Parameters, "messages")
	require.Contains(t, venice.Parameters, "reasoning_effort")
	require.NotContains(t, venice.Parameters, "logit_bias", "logit_bias is not documented by Venice")

	veniceLlama := build("venice", "llama-3.3-70b", "")
	require.Nil(t, veniceLlama.Thinking)
	require.Nil(t, veniceLlama.Bindings[system.IntentReasoningEffort])

	litellm := build("litellm", "deepseek-r1", "")
	require.Equal(t, system.ThinkingModeAlwaysOn, litellm.Thinking.Mode)
	require.Nil(t, litellm.Bindings[system.IntentReasoningDisable])

	openrouter := build("openrouter", "qwen/qwen3-235b-a22b", "")
	require.Equal(t, "reasoning.effort", openrouter.Bindings[system.IntentReasoningDisable].Param)
	require.Equal(t, "none", openrouter.Bindings[system.IntentReasoningDisable].Value)
}

func TestBuildAnthropicIsGenerationAware(t *testing.T) {
	api := build("anthropic", "claude-opus-4-6", "")
	require.Equal(t, "anthropic", api.Stack)
	require.Equal(t, map[string]any{"type": "adaptive"}, api.Bindings[system.IntentReasoningEnable].Value)
	require.Equal(t, "output_config.effort", api.Bindings[system.IntentReasoningEffort].Param)
	require.Nil(t, api.Bindings[system.IntentReasoningBudget])
	require.Equal(t, "output_config.format", api.Bindings[system.IntentResponseFormatSchema].Param)
	require.Contains(t, api.Parameters, "stop_sequences")

	onClaudeai := Build(config.ModelConfig{ModelName: "claude-sonnet-4-5-20250929", ApiType: "claudeai", ApiStack: "anthropic", ApiURL: "https://api.anthropic.com/v1/messages"})
	require.NotNil(t, onClaudeai)
	require.Equal(t, "anthropic", onClaudeai.Stack)
	require.Equal(t, "thinking.budget_tokens", onClaudeai.Bindings[system.IntentReasoningBudget].Param)

	fable := build("anthropic", "claude-fable-5-1", "")
	require.Equal(t, system.ThinkingModeTunable, fable.Thinking.Mode)
	require.Nil(t, fable.Bindings[system.IntentReasoningDisable])
}

func TestBuildGenericOpenAIIsConservative(t *testing.T) {
	api := build("openai", "qwen3-32b", "")
	require.Equal(t, "openai", api.Stack)
	require.Equal(t, "qwen3", api.ModelFamily)
	require.Empty(t, api.Bindings, "generic openai preset must not guess template kwargs")
	require.Nil(t, api.Thinking)
	require.Contains(t, api.Parameters, "response_format")

	r1 := build("openai", "deepseek-r1", "")
	require.Equal(t, system.ThinkingModeAlwaysOn, r1.Thinking.Mode)
}

func TestBuildNilWithoutApiStack(t *testing.T) {
	for _, apiType := range []string{"openai", "claudeai", "prodia-v2", "hyperbolic-sd"} {
		require.Nil(t, Build(config.ModelConfig{ModelName: "qwen3-32b", ApiType: apiType, ApiURL: "http://h/v1"}), apiType)
	}
	require.Nil(t, Build(config.ModelConfig{ModelName: "qwen3-32b", ApiType: "openai", ApiStack: "bogus", ApiURL: "http://h/v1"}))
	require.Nil(t, build("bogus", "sd-xl", ""))
}

func TestBuildEveryPresetBuilds(t *testing.T) {
	for stack := range config.StackTransport {
		api := build(stack, "qwen3-32b", "")
		require.NotNil(t, api, stack)
		require.Equal(t, stack, api.Stack, stack)
		require.NotEmpty(t, api.Parameters, stack)
	}
}

func TestBuildResultIsNotAliasedToTables(t *testing.T) {
	a := build("vllm", "qwen3-32b", "")
	a.Bindings[system.IntentSamplingTopK].Param = "mutated"
	b := build("vllm", "qwen3-32b", "")
	require.Equal(t, "top_k", b.Bindings[system.IntentSamplingTopK].Param)
}

func TestBuildOSeriesOnOpenAIPreset(t *testing.T) {
	api := build("openai", "o4-mini", "")
	require.NotNil(t, api)
	require.Equal(t, "o-series", api.ModelFamily)
	require.NotNil(t, api.Thinking)
	require.Equal(t, system.ThinkingModeTunable, api.Thinking.Mode)
	effort := api.Bindings[system.IntentReasoningEffort]
	require.NotNil(t, effort)
	require.Equal(t, "reasoning_effort", effort.Param)
	require.Nil(t, api.Bindings[system.IntentReasoningDisable])
}

func TestBuildFamilyWinsOverStackTableForSameIntent(t *testing.T) {
	api := build("ollama", "gpt-oss:120b", "")
	effort := api.Bindings[system.IntentReasoningEffort]
	require.NotNil(t, effort)
	require.Equal(t, []string{"low", "medium", "high"}, effort.EnumValues, "family's enum values must win over the ollama stack table's [low medium high max none]")
}

func TestBuildResultDeepCopiesNestedValues(t *testing.T) {
	a := build("vllm", "gpt-oss-120b", "")
	jsonFormat, ok := a.Bindings[system.IntentResponseFormatJSON].Value.(map[string]any)
	require.True(t, ok)
	jsonFormat["type"] = "mutated"
	a.Bindings[system.IntentReasoningEffort].EnumValues[0] = "mutated"

	b := build("vllm", "gpt-oss-120b", "")
	require.Equal(t, map[string]any{"type": "json_object"}, b.Bindings[system.IntentResponseFormatJSON].Value)
	require.Equal(t, []string{"low", "medium", "high"}, b.Bindings[system.IntentReasoningEffort].EnumValues)
}

func TestBuildThinkingNameOverridesFamilyDefault(t *testing.T) {
	api := build("vllm", "qwen3-235b-a22b-thinking-2507", "")
	require.NotNil(t, api.Thinking)
	require.Equal(t, system.ThinkingModeAlwaysOn, api.Thinking.Mode)
	require.Nil(t, api.Bindings[system.IntentReasoningDisable])
}

func TestBuildNameAwareFamilies(t *testing.T) {
	g4 := build("vllm", "google/gemma-4-31B-it", "")
	require.Equal(t, "gemma", g4.ModelFamily)
	require.Equal(t, system.ThinkingModeControllable, g4.Thinking.Mode)
	require.Equal(t, "chat_template_kwargs.enable_thinking", g4.Bindings[system.IntentReasoningDisable].Param)

	g3 := build("vllm", "google/gemma-3-27b-it", "")
	require.Equal(t, "gemma", g3.ModelFamily)
	require.Nil(t, g3.Thinking)
	require.Nil(t, g3.Bindings[system.IntentReasoningDisable])

	mg := build("vllm", "google/medgemma-4b-it", "")
	require.Equal(t, "gemma", mg.ModelFamily)
	require.Nil(t, mg.Thinking)
	require.Nil(t, mg.Bindings[system.IntentReasoningDisable])

	m2 := build("sglang", "MiniMaxAI/MiniMax-M2.1", "")
	require.Equal(t, "minimax", m2.ModelFamily)
	require.Equal(t, system.ThinkingModeAlwaysOn, m2.Thinking.Mode)

	m3 := build("vllm", "MiniMaxAI/MiniMax-M3", "")
	require.Equal(t, "minimax", m3.ModelFamily)
	require.Equal(t, system.ThinkingModeControllable, m3.Thinking.Mode)
	require.Equal(t, "chat_template_kwargs.thinking_mode", m3.Bindings[system.IntentReasoningDisable].Param)
	require.Equal(t, "disabled", m3.Bindings[system.IntentReasoningDisable].Value)
	require.Equal(t, "enabled", m3.Bindings[system.IntentReasoningEnable].Value)

	m3o := build("ollama", "minimax-m3", "")
	require.Equal(t, system.ThinkingModeControllable, m3o.Thinking.Mode)
	require.Equal(t, system.BindingKindBodyParam, m3o.Bindings[system.IntentReasoningDisable].Kind)
	require.Equal(t, "reasoning_effort", m3o.Bindings[system.IntentReasoningDisable].Param)
	require.Equal(t, "none", m3o.Bindings[system.IntentReasoningDisable].Value)
	require.Equal(t, "medium", m3o.Bindings[system.IntentReasoningEnable].Value)

	k26 := build("vllm", "moonshotai/Kimi-K2.6", "")
	require.Equal(t, "kimi", k26.ModelFamily)
	require.Equal(t, system.ThinkingModeControllable, k26.Thinking.Mode)
	require.Equal(t, "chat_template_kwargs.thinking", k26.Bindings[system.IntentReasoningDisable].Param)

	k3 := build("vllm", "moonshotai/Kimi-K3", "")
	require.Equal(t, "kimi", k3.ModelFamily)
	require.Equal(t, system.ThinkingModeAlwaysOn, k3.Thinking.Mode)

	ex := build("llamacpp", "LGAI-EXAONE/EXAONE-3.5-7.8B-Instruct", "exaone")
	require.Equal(t, "exaone", ex.ModelFamily)
	require.Nil(t, ex.Thinking)
}
