package main

import (
	"testing"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/system"
	"github.com/stretchr/testify/require"
)

func TestFormatResultRendersBindingsAndTrace(t *testing.T) {
	api := &system.ModelApiSpec{
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
	out := formatResult(api, []string{"GET http://a:8000/props -> HTTP 200", "identified vllm by /version"})
	require.Contains(t, out, "vllm")
	require.Contains(t, out, "qwen3")
	require.Contains(t, out, "source:      detected")
	require.Contains(t, out, "controllable")
	require.Contains(t, out, "reasoning.disable")
	require.Contains(t, out, "chat_template_kwargs.enable_thinking")
	require.Contains(t, out, "low|medium|high")
	require.Contains(t, out, "tools")
	require.Contains(t, out, "identified vllm by /version")

	// a stack reached through a gateway prints the via line; a direct one does not
	require.NotContains(t, out, "via:")
	viaOut := formatResult(&system.ModelApiSpec{Stack: "venice", Via: "litellm", Source: system.ApiSpecSourceDetected}, nil)
	require.Contains(t, viaOut, "stack:       venice")
	require.Contains(t, viaOut, "via:         litellm")

	// nil result still renders something sensible
	out = formatResult(nil, []string{"nothing worked"})
	require.Contains(t, out, "nothing detected")
	require.Contains(t, out, "nothing worked")
}
