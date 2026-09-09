package apispec

import (
	"strings"
	"testing"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/config"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/system"
	"github.com/stretchr/testify/require"
)

func build(apiType, name, family string) *system.ModelApiSpec {
	return Build(config.ModelConfig{ModelName: name, ApiType: apiType, ModelFamily: family, ApiURL: "http://h/v1/chat/completions"})
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
	// stack layer adds model-independent knobs
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

func TestBuildGptOssIsTunable(t *testing.T) {
	api := build("llamacpp", "gpt-oss-120b", "")
	require.Equal(t, system.ThinkingModeTunable, api.Thinking.Mode)
	require.Nil(t, api.Bindings[system.IntentReasoningDisable])
	require.Equal(t, "chat_template_kwargs.reasoning_effort", api.Bindings[system.IntentReasoningEffort].Param)
}

func TestBuildOllamaRewritesToReasoningEffort(t *testing.T) {
	api := build("ollama", "qwen3:32b", "")
	require.Equal(t, "reasoning_effort", api.Bindings[system.IntentReasoningDisable].Param)
	require.Equal(t, "none", api.Bindings[system.IntentReasoningDisable].Value)
	require.Equal(t, system.BindingKindNativeBodyParam, api.Bindings[system.IntentReasoningEnable].Kind)
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

	legacy := build("claudeai", "claude-sonnet-4-5-20250929", "")
	require.Equal(t, "anthropic", legacy.Stack)
	require.Equal(t, "thinking.budget_tokens", legacy.Bindings[system.IntentReasoningBudget].Param)

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

func TestBuildImageAdaptersHaveNoSpec(t *testing.T) {
	require.Nil(t, build("prodia-v2", "sd-xl", ""))
	require.Nil(t, build("bogus", "", ""))
}

func TestBuildResultIsNotAliasedToTables(t *testing.T) {
	a := build("vllm", "qwen3-32b", "")
	a.Bindings[system.IntentSamplingTopK].Param = "mutated"
	b := build("vllm", "qwen3-32b", "")
	require.Equal(t, "top_k", b.Bindings[system.IntentSamplingTopK].Param)
}
