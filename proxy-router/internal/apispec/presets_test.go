package apispec

import (
	"testing"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/config"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/system"
	"github.com/stretchr/testify/require"
)

// TransportFor answers only for apiStack presets: legacy adapter names are
// apiType values, so they are not accepted here (the validation path relies
// on that to reject `apiStack: claudeai`).
func TestTransportFor(t *testing.T) {
	for _, preset := range []string{"vllm", "sglang", "llamacpp", "ollama", "venice", "openrouter", "litellm", "openai"} {
		adapter, ok := TransportFor(preset)
		require.True(t, ok, preset)
		require.Equal(t, "openai", adapter, preset)
	}
	adapter, ok := TransportFor("anthropic")
	require.True(t, ok)
	require.Equal(t, "claudeai", adapter)
	for _, legacy := range []string{"claudeai", "prodia-sd", "prodia-sdxl", "prodia-v2", "hyperbolic-sd"} {
		_, ok = TransportFor(legacy)
		require.False(t, ok, "legacy adapter %q is an apiType, not an apiStack", legacy)
	}
	_, ok = TransportFor("bogus")
	require.False(t, ok)
	_, ok = TransportFor("")
	require.False(t, ok)
}

func TestStackFor(t *testing.T) {
	require.Equal(t, "vllm", StackFor("vllm"))
	require.Equal(t, "anthropic", StackFor("anthropic"))
	require.Equal(t, "openai", StackFor("openai"))
	require.Equal(t, "", StackFor("claudeai"), "no legacy alias: claudeai is a transport, not a stack")
	require.Equal(t, "", StackFor("prodia-v2"))
	require.Equal(t, "", StackFor("bogus"))
	require.Equal(t, "", StackFor(""))
}

// The stack tables here and config.StackTransport must describe the same
// nine presets, or a preset could validate at load but build no spec (or
// the reverse).
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

// assertWellFormed checks a binding set: known kind, a param (except
// system_prompt), known type, enum values iff enum type, object values only
// with object type, endpoint hint on native params.
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
	// Every stack with a bindings table must also have a documented parameter
	// list (composition rule 5: `parameters` is the preset's documented
	// list) — a stack that binds params it doesn't also advertise is a gap.
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
	for _, fm := range [][2]string{{"gemma", "gemma-4-31b"}, {"kimi", "kimi-k2.5"}, {"kimi", "kimi-k2.6"}, {"exaone", "exaone-4.0-32b"}} {
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

// IsKnownFamily accepts exactly the labels FamilyFromName can produce (the
// familyRules table plus o-series), so an explicit modelFamily never warns
// when the inferred value would not — including families without bindings
// such as llama.
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

// requireKwargBool asserts a family binding set is exactly the boolean
// chat-template kwarg toggle for kwarg (disable=false, enable=true).
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

// requireNoReasoning asserts a member gets neither always-on nor bindings.
func requireNoReasoning(t *testing.T, family, name string) {
	t.Helper()
	alwaysOn, b := bindingsForFamily(family, name)
	require.False(t, alwaysOn, name)
	require.Nil(t, b, name)
}

// requireAlwaysOn asserts a member reasons unconditionally, with no bindings.
func requireAlwaysOn(t *testing.T, family, name string) {
	t.Helper()
	alwaysOn, b := bindingsForFamily(family, name)
	require.True(t, alwaysOn, name)
	require.Nil(t, b, name)
}

// Gemma: only Gemma 4 has a (default-off) thinking mode toggled by
// enable_thinking; Gemma 3 and older have none.
func TestBindingsForFamilyGemma(t *testing.T) {
	for _, name := range []string{"gemma-4-31b", "google/gemma-4-31B-it", "gemma-4-26b-a4b-it", "gemma4-e4b", "gemma_4_12b"} {
		alwaysOn, b := bindingsForFamily("gemma", name)
		require.False(t, alwaysOn, name)
		requireKwargBool(t, b, "enable_thinking", name)
	}
	requireNoReasoning(t, "gemma", "gemma-3-27b")
	requireNoReasoning(t, "gemma", "google/gemma-3-27b-it")
	requireNoReasoning(t, "gemma", "gemma-2-9b")
	requireNoReasoning(t, "gemma", "gemma")
}

// Kimi: K2 Thinking and K2.7-Code always reason, K2.5 / K2.6 are hybrid via
// the `thinking` chat-template kwarg (on by default), K2 Instruct has no
// thinking mode.
func TestBindingsForFamilyKimi(t *testing.T) {
	requireNoReasoning(t, "kimi", "Kimi-K2-Instruct")
	requireNoReasoning(t, "kimi", "moonshotai/Kimi-K2-Instruct-0905")
	requireNoReasoning(t, "kimi", "kimi")
	requireAlwaysOn(t, "kimi", "Kimi-K2-Thinking")
	requireAlwaysOn(t, "kimi", "moonshotai/Kimi-K2.7-Code")
	for _, name := range []string{"Kimi-K2.5", "moonshotai/Kimi-K2.6"} {
		alwaysOn, b := bindingsForFamily("kimi", name)
		require.False(t, alwaysOn, name)
		requireKwargBool(t, b, "thinking", name)
	}
}

// MiniMax: M2 and every M2.x are interleaved-thinking models with no off
// switch; Text-01 has no thinking mode; M1 is left without bindings because
// no official source states whether its thinking can be disabled.
func TestBindingsForFamilyMinimax(t *testing.T) {
	for _, name := range []string{"MiniMax-M2", "MiniMaxAI/MiniMax-M2.1", "MiniMax-M2.5", "MiniMax-M2.7-highspeed"} {
		requireAlwaysOn(t, "minimax", name)
	}
	requireNoReasoning(t, "minimax", "MiniMax-Text-01")
	requireNoReasoning(t, "minimax", "MiniMaxAI/MiniMax-M1-80k")
	requireNoReasoning(t, "minimax", "minimax")
}

// EXAONE: 4.x is hybrid via enable_thinking (default off), Deep always
// reasons, 3.5 and older have no reasoning mode.
func TestBindingsForFamilyExaone(t *testing.T) {
	for _, name := range []string{"EXAONE-4.0-32B", "LGAI-EXAONE/EXAONE-4.0.1-32B", "exaone-4.0-1.2b"} {
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
