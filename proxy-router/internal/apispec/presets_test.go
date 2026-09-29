package apispec

import (
	"testing"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/config"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/system"
	"github.com/stretchr/testify/require"
)

func TestStackFor(t *testing.T) {
	for stack := range config.StackTransport {
		require.Equal(t, stack, StackFor(stack))
	}
	require.Equal(t, "vllm", StackFor(" VLLM "), "normalized like config.NormalizeApiStack")
	for _, legacy := range []string{"claudeai", "prodia-sd", "prodia-sdxl", "prodia-v2", "hyperbolic-sd"} {
		require.Equal(t, "", StackFor(legacy), "legacy adapter %q is an apiType, not an apiStack", legacy)
	}
	require.Equal(t, "", StackFor("bogus"))
	require.Equal(t, "", StackFor(""))
	require.Equal(t, "", StackFor(" "), "whitespace-only means unset")
}

func TestStackTablesMatchConfigPresets(t *testing.T) {
	for stack := range stackBindings {
		require.Contains(t, config.StackTransport, stack, "stackBindings key %q is not a config.StackTransport preset", stack)
	}
	for stack := range stackParameters {
		require.Contains(t, config.StackTransport, stack, "stackParameters key %q is not a config.StackTransport preset", stack)
	}
	for stack := range config.StackTransport {
		require.Contains(t, stackParameters, stack, "preset %q has no stackParameters entry", stack)
	}
}

func assertWellFormed(t *testing.T, where string, b map[string]*system.ParamBinding) {
	t.Helper()
	kinds := map[string]bool{system.BindingKindBodyParam: true, system.BindingKindTemplateKwarg: true, system.BindingKindSystemPrompt: true, system.BindingKindNativeBodyParam: true}
	types := map[string]bool{"boolean": true, "number": true, "enum": true, "string": true, "object": true, "array": true}
	for intent, pb := range b {
		require.NotNil(t, pb, "%s/%s", where, intent)
		require.True(t, kinds[pb.Kind], "%s/%s kind %q", where, intent, pb.Kind)
		if pb.Kind != system.BindingKindSystemPrompt {
			require.NotEmpty(t, pb.Param, "%s/%s", where, intent)
			require.True(t, types[pb.ParamType], "%s/%s type %q", where, intent, pb.ParamType)
		}
		if pb.Kind == system.BindingKindTemplateKwarg {
			require.Contains(t, pb.Param, "chat_template_kwargs.", "%s/%s", where, intent)
		}
		if len(pb.EnumValues) > 0 {
			require.Equal(t, "enum", pb.ParamType, "%s/%s", where, intent)
		}
		if pb.ParamType == "enum" {
			require.NotEmpty(t, pb.EnumValues, "%s/%s enum without values", where, intent)
		}
		if _, isMap := pb.Value.(map[string]any); isMap {
			require.Equal(t, "object", pb.ParamType, "%s/%s", where, intent)
		}
		if pb.Kind == system.BindingKindNativeBodyParam {
			require.NotEmpty(t, pb.Hint, "%s/%s native binding needs an endpoint hint", where, intent)
		}
	}
}

func TestAllTablesWellFormed(t *testing.T) {
	for stack, table := range stackBindings {
		assertWellFormed(t, "stack:"+stack, table)
	}
	for stack, params := range stackParameters {
		require.NotEmpty(t, params, stack)
		require.Contains(t, params, "messages", stack)
	}
	for stack := range stackBindings {
		require.NotEmpty(t, stackParameters[stack], "stack %q has bindings but no stackParameters entry", stack)
	}
	for _, f := range []string{"qwen3", "glm", "hunyuan", "deepseek-v3.1", "granite", "gpt-oss", "seed-oss", "nemotron", "o-series", "gemini"} {
		_, b := bindingsForFamily(f, "m")
		assertWellFormed(t, "family:"+f, b)
	}
	for _, name := range []string{"claude-opus-4-6", "claude-sonnet-4-5", "claude-3-7-sonnet", "claude-fable-5-1", "claude-opus-5", "claude"} {
		_, b := bindingsForFamily("claude", name)
		assertWellFormed(t, "family:claude:"+name, b)
	}
	for _, fm := range [][2]string{{"gemma", "gemma-4-31b"}, {"kimi", "kimi-k2.5"}, {"kimi", "kimi-k2.6"}, {"minimax", "minimax-m3"}, {"exaone", "exaone-4.0-32b"}} {
		_, b := bindingsForFamily(fm[0], fm[1])
		require.NotEmpty(t, b, "family:%s:%s", fm[0], fm[1])
		assertWellFormed(t, "family:"+fm[0]+":"+fm[1], b)
	}
	assertWellFormed(t, "ollama-toggle", ollamaThinkBindings())
	for _, mk := range []func() bindingSet{func() bindingSet { return kwargBoolBindings("enable_thinking") }, effortKwargBindings, budgetKwargBindings} {
		b := mk()
		rewriteBindingsForOllama(b)
		assertWellFormed(t, "ollama-rewrite", b)
	}
}

func TestIsKnownFamily(t *testing.T) {
	require.True(t, IsKnownFamily("qwen3"))
	require.True(t, IsKnownFamily("llama"))
	require.True(t, IsKnownFamily("o-series"))
	require.True(t, IsKnownFamily(" Gemma "), "case-insensitive and trimmed like Build")
	for _, f := range []string{"gemma", "kimi", "minimax", "exaone", "qwq", "hunyuan"} {
		require.True(t, IsKnownFamily(f), f)
	}
	for _, r := range familyRules {
		require.True(t, IsKnownFamily(r.family), "familyRules produces %q but IsKnownFamily rejects it", r.family)
	}
	require.False(t, IsKnownFamily("qwen-3"))
	require.False(t, IsKnownFamily("bogus"))
	require.False(t, IsKnownFamily(""))
}

func TestFamilyFromName(t *testing.T) {
	require.Equal(t, "qwen3", FamilyFromName("Qwen/Qwen3-235B-A22B"))
	require.Equal(t, "deepseek-r1", FamilyFromName("deepseek-r1:70b"))
	require.Equal(t, "deepseek-v3.1", FamilyFromName("deepseek-chat-v3.1"))
	require.Equal(t, "gpt-oss", FamilyFromName("openai/gpt-oss-120b"))
	require.Equal(t, "o-series", FamilyFromName("o4-mini"))
	require.Equal(t, "claude", FamilyFromName("claude-sonnet-4-5"))
	require.Equal(t, "nemotron", FamilyFromName("Llama-3.3-Nemotron-Super-49B"))
	require.Equal(t, "llama", FamilyFromName("llama-3.3-70b"))
	require.Equal(t, "gemma", FamilyFromName("google/gemma-4-31B-it"))
	require.Equal(t, "kimi", FamilyFromName("moonshotai/Kimi-K2.5"))
	require.Equal(t, "minimax", FamilyFromName("MiniMaxAI/MiniMax-M2"))
	require.Equal(t, "exaone", FamilyFromName("LGAI-EXAONE/EXAONE-4.0-32B"))
	require.Equal(t, "", FamilyFromName("some-embedder"))
}

func requireKwargBool(t *testing.T, b bindingSet, kwarg, name string) {
	t.Helper()
	param := "chat_template_kwargs." + kwarg
	require.NotNil(t, b[system.IntentReasoningDisable], name)
	require.NotNil(t, b[system.IntentReasoningEnable], name)
	require.Equal(t, system.BindingKindTemplateKwarg, b[system.IntentReasoningDisable].Kind, name)
	require.Equal(t, param, b[system.IntentReasoningDisable].Param, name)
	require.Equal(t, false, b[system.IntentReasoningDisable].Value, name)
	require.Equal(t, param, b[system.IntentReasoningEnable].Param, name)
	require.Equal(t, true, b[system.IntentReasoningEnable].Value, name)
}

func requireNoReasoning(t *testing.T, family, name string) {
	t.Helper()
	alwaysOn, b := bindingsForFamily(family, name)
	require.False(t, alwaysOn, name)
	require.Nil(t, b, name)
}

func requireAlwaysOn(t *testing.T, family, name string) {
	t.Helper()
	alwaysOn, b := bindingsForFamily(family, name)
	require.True(t, alwaysOn, name)
	require.Nil(t, b, name)
}

func TestBindingsForFamilyGemma(t *testing.T) {
	for _, name := range []string{"gemma-4-31b", "google/gemma-4-31B-it", "gemma-4-26b-a4b-it", "gemma4-e4b", "gemma_4_12b", "gemma4:31b", "gemma-4"} {
		alwaysOn, b := bindingsForFamily("gemma", name)
		require.False(t, alwaysOn, name)
		requireKwargBool(t, b, "enable_thinking", name)
	}
	requireNoReasoning(t, "gemma", "gemma-3-27b")
	requireNoReasoning(t, "gemma", "google/gemma-3-27b-it")
	requireNoReasoning(t, "gemma", "gemma-3-4b-it")
	requireNoReasoning(t, "gemma", "google/gemma-3n-E4B-it")
	requireNoReasoning(t, "gemma", "gemma-2-9b")
	requireNoReasoning(t, "gemma", "google/medgemma-4b-it")
	requireNoReasoning(t, "gemma", "medgemma-4b-pt")
	requireNoReasoning(t, "gemma", "gemma")
}

func TestBindingsForFamilyKimi(t *testing.T) {
	requireNoReasoning(t, "kimi", "Kimi-K2-Instruct")
	requireNoReasoning(t, "kimi", "moonshotai/Kimi-K2-Instruct-0905")
	requireNoReasoning(t, "kimi", "kimi")
	requireAlwaysOn(t, "kimi", "Kimi-K2-Thinking")
	requireAlwaysOn(t, "kimi", "moonshotai/Kimi-K2.7-Code")
	requireAlwaysOn(t, "kimi", "moonshotai/Kimi-K3")
	requireAlwaysOn(t, "kimi", "kimi_k3")
	for _, name := range []string{"Kimi-K2.5", "moonshotai/Kimi-K2.6"} {
		alwaysOn, b := bindingsForFamily("kimi", name)
		require.False(t, alwaysOn, name)
		requireKwargBool(t, b, "thinking", name)
	}
}

func TestBindingsForFamilyMinimax(t *testing.T) {
	for _, name := range []string{"MiniMax-M2", "MiniMaxAI/MiniMax-M2.1", "MiniMax-M2.5", "MiniMax-M2.7-highspeed"} {
		requireAlwaysOn(t, "minimax", name)
	}
	for _, name := range []string{"MiniMax-M3", "MiniMaxAI/MiniMax-M3"} {
		alwaysOn, b := bindingsForFamily("minimax", name)
		require.False(t, alwaysOn, name)
		for _, intent := range []string{system.IntentReasoningDisable, system.IntentReasoningEnable} {
			require.NotNil(t, b[intent], name)
			require.Equal(t, system.BindingKindTemplateKwarg, b[intent].Kind, name)
			require.Equal(t, "chat_template_kwargs.thinking_mode", b[intent].Param, name)
			require.Equal(t, "string", b[intent].ParamType, name)
		}
		require.Equal(t, "disabled", b[system.IntentReasoningDisable].Value, name)
		require.Equal(t, "enabled", b[system.IntentReasoningEnable].Value, name)
	}
	requireNoReasoning(t, "minimax", "MiniMax-Text-01")
	requireNoReasoning(t, "minimax", "MiniMaxAI/MiniMax-M1-80k")
	requireNoReasoning(t, "minimax", "minimax")
}

func TestBindingsForFamilyExaone(t *testing.T) {
	for _, name := range []string{"EXAONE-4.0-32B", "LGAI-EXAONE/EXAONE-4.0.1-32B", "exaone-4.0-1.2b", "exaone4:32b", "exaone4", "exaone_4"} {
		alwaysOn, b := bindingsForFamily("exaone", name)
		require.False(t, alwaysOn, name)
		requireKwargBool(t, b, "enable_thinking", name)
	}
	requireAlwaysOn(t, "exaone", "EXAONE-Deep-32B")
	requireAlwaysOn(t, "exaone", "LGAI-EXAONE/EXAONE-Deep-7.8B")
	requireNoReasoning(t, "exaone", "EXAONE-3.5-7.8B")
	requireNoReasoning(t, "exaone", "LGAI-EXAONE/EXAONE-3.5-32B-Instruct")
	requireNoReasoning(t, "exaone", "exaone")
}

func withStackTable(t *testing.T, stack string, b bindingSet, params []string) {
	t.Helper()
	prevB, hadB := stackBindings[stack]
	prevP, hadP := stackParameters[stack]
	stackBindings[stack] = b
	stackParameters[stack] = params
	t.Cleanup(func() {
		if hadB {
			stackBindings[stack] = prevB
		} else {
			delete(stackBindings, stack)
		}
		if hadP {
			stackParameters[stack] = prevP
		} else {
			delete(stackParameters, stack)
		}
	})
}

func TestMergeStackBindingsFillsMissingIntentsOnly(t *testing.T) {
	withStackTable(t, "teststack", bindingSet{
		system.IntentSamplingTopK:       {Kind: system.BindingKindBodyParam, Param: "top_k", ParamType: "number"},
		system.IntentReasoningDisable:   {Kind: system.BindingKindBodyParam, Param: "stack_level_disable", ParamType: "boolean", Value: false},
		system.IntentResponseFormatJSON: {Kind: system.BindingKindBodyParam, Param: "response_format", ParamType: "object", Value: map[string]any{"type": "json_object"}},
	}, []string{"temperature", "top_p"})

	api := &system.ModelApiSpec{Stack: "teststack"}
	established := bindingSet{
		system.IntentReasoningDisable: {Kind: system.BindingKindTemplateKwarg, Param: "chat_template_kwargs.enable_thinking", ParamType: "boolean", Value: false},
	}
	mergeStackTables(api, "teststack", established, true, false)

	require.Equal(t, "chat_template_kwargs.enable_thinking", api.Bindings[system.IntentReasoningDisable].Param)
	require.Equal(t, "top_k", api.Bindings[system.IntentSamplingTopK].Param)
	require.Equal(t, "response_format", api.Bindings[system.IntentResponseFormatJSON].Param)
	require.Equal(t, []string{"temperature", "top_p"}, api.Parameters)

	api.Bindings[system.IntentSamplingTopK].Param = "mutated"
	require.Equal(t, "top_k", stackBindings["teststack"][system.IntentSamplingTopK].Param)
}

func TestMergeStackTablesKeepsRegistryParameters(t *testing.T) {
	withStackTable(t, "teststack", nil, []string{"temperature"})

	api := &system.ModelApiSpec{Stack: "teststack", Parameters: []string{"tools", "reasoning"}}
	mergeStackTables(api, "teststack", nil, false, false)
	require.Equal(t, []string{"tools", "reasoning"}, api.Parameters)
}

func TestMergeStackTablesNoopForUnknownStack(t *testing.T) {
	api := &system.ModelApiSpec{Stack: "", ModelFamily: "llama"}
	mergeStackTables(api, "", nil, false, false)
	require.Empty(t, api.Bindings)
	require.Empty(t, api.Parameters)
}

func TestMergeStackTablesGatesReasoningOnEvidence(t *testing.T) {
	withStackTable(t, "teststack", bindingSet{
		system.IntentReasoningBudget: {Kind: system.BindingKindBodyParam, Param: "budget", ParamType: "number"},
		system.IntentSamplingTopK:    {Kind: system.BindingKindBodyParam, Param: "top_k", ParamType: "number"},
	}, nil)

	api := &system.ModelApiSpec{Stack: "teststack"}
	mergeStackTables(api, "teststack", nil, false, false)
	require.Nil(t, api.Bindings[system.IntentReasoningBudget])
	require.NotNil(t, api.Bindings[system.IntentSamplingTopK])

	api = &system.ModelApiSpec{Stack: "teststack"}
	mergeStackTables(api, "teststack", nil, true, false)
	require.NotNil(t, api.Bindings[system.IntentReasoningBudget])
}

func TestMergeStackTablesAlwaysOnKeepsOnlyFormat(t *testing.T) {
	withStackTable(t, "teststack", bindingSet{
		system.IntentReasoningDisable: {Kind: system.BindingKindBodyParam, Param: "off", ParamType: "boolean", Value: true},
		system.IntentReasoningEffort:  {Kind: system.BindingKindBodyParam, Param: "effort", ParamType: "enum", EnumValues: []string{"low"}},
		system.IntentReasoningFormat:  {Kind: system.BindingKindBodyParam, Param: "separate", ParamType: "boolean"},
	}, nil)

	api := &system.ModelApiSpec{Stack: "teststack"}
	mergeStackTables(api, "teststack", nil, true, true)
	require.Nil(t, api.Bindings[system.IntentReasoningDisable])
	require.Nil(t, api.Bindings[system.IntentReasoningEffort])
	require.NotNil(t, api.Bindings[system.IntentReasoningFormat])
}

func TestRewriteBindingsForOllamaEffortKeepsLevels(t *testing.T) {
	b := effortKwargBindings()
	rewriteBindingsForOllama(b)
	effort := b[system.IntentReasoningEffort]
	require.Equal(t, system.BindingKindBodyParam, effort.Kind)
	require.Equal(t, "reasoning_effort", effort.Param)
	require.ElementsMatch(t, []string{"low", "medium", "high"}, effort.EnumValues)
}
