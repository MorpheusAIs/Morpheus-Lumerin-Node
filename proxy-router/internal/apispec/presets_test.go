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
	assertWellFormed(t, "ollama-toggle", ollamaThinkBindings())
	for _, mk := range []func() bindingSet{func() bindingSet { return kwargBoolBindings("enable_thinking") }, effortKwargBindings, budgetKwargBindings} {
		b := mk()
		rewriteBindingsForOllama(b)
		assertWellFormed(t, "ollama-rewrite", b)
	}
}

func TestIsKnownFamily(t *testing.T) {
	require.True(t, IsKnownFamily("qwen3"))
	require.False(t, IsKnownFamily("qwen-3"))
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
	require.Equal(t, "", FamilyFromName("some-embedder"))
}
