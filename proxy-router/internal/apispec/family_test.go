package apispec

import (
	"testing"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/system"
	"github.com/stretchr/testify/require"
)

func TestFamilyFromNameTable(t *testing.T) {
	cases := map[string]string{
		"Qwen/Qwen3-235B-A22B":          "qwen3",
		"qwen3:32b":                     "qwen3",
		"QwQ-32B":                       "qwq",
		"Qwen/Qwen2.5-72B-Instruct":     "qwen2.5",
		"deepseek-r1:70b":               "deepseek-r1",
		"deepseek-ai/DeepSeek-V3.1":     "deepseek-v3.1",
		"deepseek-chat":                 "deepseek",
		"zai-org/GLM-4.6":               "glm",
		"meta-llama/Llama-3.3-70B":      "llama",
		"mistral-large-latest":          "mistral",
		"mixtral-8x22b":                 "mixtral",
		"google/gemma-3-27b-it":         "gemma",
		"microsoft/phi-4":               "phi",
		"ibm-granite/granite-3.2-8b":    "granite",
		"tencent/Hunyuan-A13B":          "hunyuan",
		"nvidia/Llama-3.3-Nemotron-49B": "nemotron",
		"LGAI-EXAONE/EXAONE-4.0-32B":    "exaone",
		"moonshotai/Kimi-K2-Instruct":   "kimi",
		"MiniMax-M2":                    "minimax",
		"ByteDance-Seed/Seed-OSS-36B":   "seed-oss",
		"openai/gpt-oss-120b":           "gpt-oss",
		"gpt-4o":                        "gpt",
		"gpt-5":                         "gpt",
		"o3-mini":                       "o-series",
		"claude-opus-4-6":               "claude",
		"grok-4":                        "grok",
		"gemini-2.5-flash":              "gemini",
		"command-r-plus":                "command",
		"unknown-model-x":               "",
	}
	for name, want := range cases {
		require.Equal(t, want, FamilyFromName(name), "name %s", name)
	}
}

func TestBindingsForFamilyDefaults(t *testing.T) {
	_, b := bindingsForFamily("qwen3", "qwen3-32b")
	require.Equal(t, "chat_template_kwargs.enable_thinking", b[system.IntentReasoningDisable].Param)
	require.Equal(t, false, b[system.IntentReasoningDisable].Value)
	require.Equal(t, true, b[system.IntentReasoningEnable].Value)

	alwaysOn, b := bindingsForFamily("qwen3", "qwen3-235b-thinking-2507")
	require.True(t, alwaysOn)
	require.Empty(t, b)

	for _, f := range []string{"qwq", "deepseek-r1"} {
		alwaysOn, _ := bindingsForFamily(f, "m")
		require.True(t, alwaysOn, f)
	}

	_, b = bindingsForFamily("deepseek-v3.1", "m")
	require.Equal(t, "chat_template_kwargs.thinking", b[system.IntentReasoningDisable].Param)
	_, b = bindingsForFamily("granite", "m")
	require.Equal(t, "chat_template_kwargs.thinking", b[system.IntentReasoningDisable].Param)

	_, b = bindingsForFamily("gpt-oss", "gpt-oss-120b")
	require.Nil(t, b[system.IntentReasoningDisable])
	effort := b[system.IntentReasoningEffort]
	require.NotNil(t, effort)
	require.Equal(t, "chat_template_kwargs.reasoning_effort", effort.Param)
	require.Equal(t, "enum", effort.ParamType)
	require.ElementsMatch(t, []string{"low", "medium", "high"}, effort.EnumValues)

	_, b = bindingsForFamily("seed-oss", "m")
	require.Equal(t, "chat_template_kwargs.thinking_budget", b[system.IntentReasoningBudget].Param)
	require.Equal(t, "number", b[system.IntentReasoningBudget].ParamType)
	require.Equal(t, 0, b[system.IntentReasoningDisable].Value)

	_, b = bindingsForFamily("nemotron", "m")
	require.Equal(t, system.BindingKindSystemPrompt, b[system.IntentReasoningDisable].Kind)
	require.NotEmpty(t, b[system.IntentReasoningDisable].Hint)
	require.NotEmpty(t, b[system.IntentReasoningEnable].Hint)

	_, b = bindingsForFamily("claude", "claude-opus-4-6")
	require.Equal(t, map[string]any{"type": "disabled"}, b[system.IntentReasoningDisable].Value)
	require.Equal(t, map[string]any{"type": "adaptive"}, b[system.IntentReasoningEnable].Value)
	require.Equal(t, "output_config.effort", b[system.IntentReasoningEffort].Param)
	require.Nil(t, b[system.IntentReasoningBudget], "budget_tokens is gone on 4.6+")
	_, b = bindingsForFamily("claude", "claude-sonnet-4-5-20250929")
	require.Equal(t, map[string]any{"type": "enabled", "budget_tokens": 1024}, b[system.IntentReasoningEnable].Value)
	require.Equal(t, "thinking.budget_tokens", b[system.IntentReasoningBudget].Param)
	require.ElementsMatch(t, []string{"low", "medium", "high"}, b[system.IntentReasoningEffort].EnumValues)
	_, b = bindingsForFamily("claude", "claude-3-7-sonnet")
	require.Nil(t, b[system.IntentReasoningEffort])
	require.NotNil(t, b[system.IntentReasoningBudget])
	_, b = bindingsForFamily("claude", "claude-fable-5-1")
	require.Nil(t, b[system.IntentReasoningDisable], "always-on: cannot be disabled")
	require.Nil(t, b[system.IntentReasoningEnable])
	require.Contains(t, b[system.IntentReasoningEffort].EnumValues, "xhigh")
	_, b = bindingsForFamily("o-series", "m")
	require.Nil(t, b[system.IntentReasoningDisable])
	require.Equal(t, "reasoning_effort", b[system.IntentReasoningEffort].Param)
	_, b = bindingsForFamily("gemini", "m")
	require.Equal(t, "generationConfig.thinkingConfig.thinkingBudget", b[system.IntentReasoningDisable].Param)
	require.Equal(t, 0, b[system.IntentReasoningDisable].Value)

	for _, f := range []string{"llama", "mistral", ""} {
		alwaysOn, b := bindingsForFamily(f, "m")
		require.False(t, alwaysOn, f)
		require.Empty(t, b, f)
	}
}

func TestClaudeVersionParsing(t *testing.T) {
	for name, want := range map[string][2]int{
		"claude-opus-4-6":            {4, 6},
		"claude-sonnet-4-5-20250929": {4, 5},
		"claude-3-7-sonnet-latest":   {3, 7},
		"claude-opus-5":              {5, 0},
		"anthropic/claude-sonnet-5":  {5, 0},
		"claude-fable-5-1":           {5, 1},
	} {
		major, minor, ok := claudeVersion(name)
		require.True(t, ok, name)
		require.Equal(t, want, [2]int{major, minor}, name)
	}
	_, _, ok := claudeVersion("claude")
	require.False(t, ok)
}

func TestBindingsFromTemplate(t *testing.T) {
	require.Equal(t, "chat_template_kwargs.enable_thinking", bindingsFromTemplate(qwen3Template)[system.IntentReasoningDisable].Param)
	require.Equal(t, "chat_template_kwargs.thinking", bindingsFromTemplate(deepseekV31Template)[system.IntentReasoningDisable].Param)
	require.Equal(t, "chat_template_kwargs.reasoning_effort", bindingsFromTemplate(`{% if reasoning_effort %}x{% endif %}`)[system.IntentReasoningEffort].Param)
	require.Equal(t, "chat_template_kwargs.thinking_budget", bindingsFromTemplate(`{% set b = thinking_budget %}`)[system.IntentReasoningBudget].Param)
	require.Nil(t, bindingsFromTemplate("You are a thinking assistant. {{ messages }}"))
	require.Nil(t, bindingsFromTemplate(""))
}

func TestRefinesFamily(t *testing.T) {
	require.True(t, refinesFamily("", "qwen3"))
	require.True(t, refinesFamily("deepseek", "deepseek-r1"))
	require.False(t, refinesFamily("deepseek-r1", "deepseek"))
	require.False(t, refinesFamily("qwen3", "qwen3"))
	require.False(t, refinesFamily("qwen3", ""))
	require.False(t, refinesFamily("", ""))
}
