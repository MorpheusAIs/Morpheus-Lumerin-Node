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

// TestBuildGptOssOnLlamacppReasoningBindings covers a thinking-capable family
// (gpt-oss contributes only reasoning.effort) combined with llama.cpp's
// stack-level reasoning_budget param. Per the llama-server README,
// reasoning_budget: 0 ends thinking immediately, so llama.cpp can disable
// thinking even though gpt-oss's own kwarg has no explicit off switch — that
// makes the mode controllable, not merely tunable.
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
	// mirrors litellm: reasoning_effort (any level but none) also works on
	// ollama's OpenAI-compatible /v1 surface, so gateways can enable thinking
	// without the native-only `think` boolean.
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
	// composition rule 5: parameters must be the preset's documented list,
	// never nil, even when the preset's bindings table is sparse.
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

// TestBuildOSeriesOnOpenAIPreset covers controller ruling 2: the o-series
// family's native vendor is the generic openai preset (standard
// reasoning_effort body param), so o4-mini on the openai preset must still
// get a tunable reasoning spec even though openai is a gateway stack.
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

// TestBuildFamilyWinsOverStackTableForSameIntent covers precedence: when
// both the family layer and the stack table bind the same intent, the
// family's binding (the model-specific one) must win, not the stack's more
// generic one.
func TestBuildFamilyWinsOverStackTableForSameIntent(t *testing.T) {
	api := build("ollama", "gpt-oss:120b", "")
	effort := api.Bindings[system.IntentReasoningEffort]
	require.NotNil(t, effort)
	require.Equal(t, []string{"low", "medium", "high"}, effort.EnumValues, "family's enum values must win over the ollama stack table's [low medium high max none]")
}

// TestBuildResultDeepCopiesNestedValues covers the deep-copy invariant:
// mutating a returned spec's nested map/slice values must not corrupt the
// shared static tables that a subsequent Build reads from.
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

// TestBuildThinkingNameOverridesFamilyDefault covers the "-thinking" model
// name override: a model name containing "thinking" reasons unconditionally
// regardless of what its family would otherwise default to.
func TestBuildThinkingNameOverridesFamilyDefault(t *testing.T) {
	api := build("vllm", "qwen3-235b-a22b-thinking-2507", "")
	require.NotNil(t, api.Thinking)
	require.Equal(t, system.ThinkingModeAlwaysOn, api.Thinking.Mode)
	require.Nil(t, api.Bindings[system.IntentReasoningDisable])
}
