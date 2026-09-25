package apispec

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/config"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/system"
	"github.com/stretchr/testify/require"
)

func boolp(v bool) *bool { return &v }

func int64p(v int64) *int64 { return &v }

func vllmQwenSpec(t *testing.T) *system.ModelApiSpec {
	t.Helper()
	ev := Evidence{Stack: "vllm", ModelName: "qwen3-32b", ServedModelID: "Qwen/Qwen3-32B"}
	ev.Caps.SetContext(40960, "vllm:/v1/models.max_model_len")
	api := Compose(ev)
	require.NotNil(t, api)
	require.Equal(t, system.ThinkingModeControllable, api.Thinking.Mode)
	require.Equal(t, "chat_template_kwargs.enable_thinking", api.Bindings[system.IntentReasoningDisable].Param)
	return api
}

func mustJSON(t *testing.T, v any) string {
	t.Helper()
	raw, err := json.Marshal(v)
	require.NoError(t, err)
	return string(raw)
}

func traceHas(trace []string, substr string) bool {
	for _, line := range trace {
		if strings.Contains(line, substr) {
			return true
		}
	}
	return false
}

func TestMergeRuntimeNilAndEmpty(t *testing.T) {
	require.Nil(t, MergeRuntime(nil, Runtime{ContextLimit: 4096, ReasoningSeen: true}))

	for name, api := range map[string]*system.ModelApiSpec{
		"composed with capabilities": vllmQwenSpec(t),
		"built preset":               Build(config.ModelConfig{ModelName: "deepseek-r1", ApiType: "openai", ApiStack: "ollama", ApiURL: "http://h/v1/chat/completions"}),
		"family only":                Compose(Evidence{ModelName: "qwen3-32b"}),
	} {
		t.Run(name, func(t *testing.T) {
			require.NotNil(t, api)
			before := mustJSON(t, api)
			trace := MergeRuntime(api, Runtime{})
			require.Empty(t, trace)
			require.Equal(t, before, mustJSON(t, api))
			require.True(t, api.Evidence == nil || len(api.Evidence) > 0, "no empty evidence map is created")
		})
	}
}

func TestMergeRuntimeContext(t *testing.T) {
	t.Run("fills an unset context with the default source", func(t *testing.T) {
		api := Compose(Evidence{Stack: "vllm", ModelName: "qwen3-32b"})
		require.Nil(t, api.Capabilities)
		trace := MergeRuntime(api, Runtime{ContextLimit: 8192, At: 1700000000})
		require.Equal(t, int64(8192), *api.Capabilities.Context)
		require.Equal(t, "health-probe:context_length_exceeded", api.Evidence["capabilities.context"])
		require.Contains(t, trace, "capabilities.context = 8192 (health-probe:context_length_exceeded)")
	})
	t.Run("uses the given source", func(t *testing.T) {
		api := Compose(Evidence{Stack: "llamacpp", ModelName: "qwen3-32b"})
		MergeRuntime(api, Runtime{ContextLimit: 4096, ContextSource: "health-probe:llamacpp.exceed_context_size_error"})
		require.Equal(t, "health-probe:llamacpp.exceed_context_size_error", api.Evidence["capabilities.context"])
	})
	t.Run("never overwrites a set context", func(t *testing.T) {
		api := vllmQwenSpec(t)
		before := mustJSON(t, api)
		trace := MergeRuntime(api, Runtime{ContextLimit: 8192})
		require.Equal(t, before, mustJSON(t, api))
		require.True(t, traceHas(trace, "capabilities.context: runtime value ignored — already set (vllm:/v1/models.max_model_len)"), trace)
	})
	t.Run("never overwrites a set context that has no evidence entry", func(t *testing.T) {
		api := &system.ModelApiSpec{Capabilities: &system.ModelCapabilities{Context: int64p(4096)}}
		before := mustJSON(t, api)
		trace := MergeRuntime(api, Runtime{ContextLimit: 8192})
		require.Equal(t, before, mustJSON(t, api))
		require.Nil(t, api.Evidence)
		require.Contains(t, trace, "capabilities.context: runtime value ignored — already set (no evidence entry)")
	})
	t.Run("rejects an out-of-range value", func(t *testing.T) {
		api := Compose(Evidence{Stack: "vllm", ModelName: "qwen3-32b"})
		before := mustJSON(t, api)
		trace := MergeRuntime(api, Runtime{ContextLimit: -5})
		require.Equal(t, before, mustJSON(t, api))
		require.True(t, traceHas(trace, "capabilities.context: runtime value -5 out of"), trace)
		trace = MergeRuntime(api, Runtime{ContextLimit: 1 << 40})
		require.Equal(t, before, mustJSON(t, api))
		require.True(t, traceHas(trace, "out of"), trace)
	})
}

func TestMergeRuntimeTools(t *testing.T) {
	t.Run("sets tools with the default verify source", func(t *testing.T) {
		api := Compose(Evidence{Stack: "vllm", ModelName: "qwen3-32b"})
		trace := MergeRuntime(api, Runtime{ToolsSupported: boolp(false), At: 42})
		require.False(t, *api.Capabilities.Tools)
		require.Equal(t, "verify:tools@42", api.Evidence["capabilities.tools"])
		require.Contains(t, trace, "capabilities.tools = false (verify:tools@42)")
	})
	t.Run("overrides a backend-reported value", func(t *testing.T) {
		ev := Evidence{Stack: "vllm", ModelName: "qwen3-32b"}
		ev.Caps.SetTools(true, "vllm:/v1/models.tool_parser")
		api := Compose(ev)
		require.True(t, *api.Capabilities.Tools)
		MergeRuntime(api, Runtime{ToolsSupported: boolp(false), ToolsSource: "verify:tool_choice_auto@7"})
		require.False(t, *api.Capabilities.Tools)
		require.Equal(t, "verify:tool_choice_auto@7", api.Evidence["capabilities.tools"])
	})
	t.Run("never overrides a declared value", func(t *testing.T) {
		api := Compose(Evidence{Stack: "vllm", ModelName: "qwen3-32b", DeclaredCaps: &config.DeclaredCapabilities{Tools: boolp(true)}})
		require.Equal(t, SourceDeclared, api.Evidence["capabilities.tools"])
		before := mustJSON(t, api)
		trace := MergeRuntime(api, Runtime{ToolsSupported: boolp(false)})
		require.Equal(t, before, mustJSON(t, api))
		require.True(t, traceHas(trace, "capabilities.tools: runtime value ignored — declared (declared)"), trace)
	})
}

func TestMergeRuntimeReasoningSeen(t *testing.T) {
	t.Run("with disable applied: reasoning true and always_on", func(t *testing.T) {
		api := vllmQwenSpec(t)
		trace := MergeRuntime(api, Runtime{ReasoningSeen: true, ReasoningDisabledInRequest: true, At: 99})
		require.True(t, *api.Capabilities.Reasoning)
		require.Equal(t, "verify:reasoning_seen_with_disable@99", api.Evidence["capabilities.reasoning"])
		require.Equal(t, system.ThinkingModeAlwaysOn, api.Thinking.Mode)
		require.Contains(t, trace, "thinking.mode: controllable -> always_on (verify:reasoning_seen_with_disable@99)")
		require.NotNil(t, api.Bindings[system.IntentReasoningDisable], "the binding stays; the verifier's verdict marks it")
	})
	t.Run("with disable applied and no thinking block", func(t *testing.T) {
		api := Compose(Evidence{Stack: "vllm", ModelName: "some-model"})
		require.Nil(t, api.Thinking)
		trace := MergeRuntime(api, Runtime{ReasoningSeen: true, ReasoningDisabledInRequest: true, At: 5})
		require.Equal(t, system.ThinkingModeAlwaysOn, api.Thinking.Mode)
		require.Contains(t, trace, "thinking.mode = always_on (verify:reasoning_seen_with_disable@5)")
	})
	t.Run("already always_on is left alone", func(t *testing.T) {
		api := Compose(Evidence{Stack: "vllm", ModelName: "deepseek-r1"})
		require.Equal(t, system.ThinkingModeAlwaysOn, api.Thinking.Mode)
		trace := MergeRuntime(api, Runtime{ReasoningSeen: true, ReasoningDisabledInRequest: true, At: 5})
		require.Equal(t, system.ThinkingModeAlwaysOn, api.Thinking.Mode)
		require.False(t, traceHas(trace, "thinking.mode"), trace)
		require.True(t, *api.Capabilities.Reasoning)
	})
	t.Run("without a disable attempt the mode is untouched", func(t *testing.T) {
		api := vllmQwenSpec(t)
		trace := MergeRuntime(api, Runtime{ReasoningSeen: true, At: 5})
		require.Equal(t, system.ThinkingModeControllable, api.Thinking.Mode)
		require.True(t, *api.Capabilities.Reasoning)
		require.Equal(t, "health-probe:reasoning_seen", api.Evidence["capabilities.reasoning"])
		require.False(t, traceHas(trace, "thinking.mode"), trace)
	})
	t.Run("disable attempted without reasoning proves nothing", func(t *testing.T) {
		api := vllmQwenSpec(t)
		before := mustJSON(t, api)
		require.Empty(t, MergeRuntime(api, Runtime{ReasoningDisabledInRequest: true}))
		require.Equal(t, before, mustJSON(t, api))
	})
}

func TestMergeRuntimeVerified(t *testing.T) {
	api := vllmQwenSpec(t)
	trace := MergeRuntime(api, Runtime{At: 1000, Verified: map[string]VerifiedBinding{
		system.IntentReasoningDisable: {Verified: true, Reason: "2xx without reasoning content"},
		system.IntentSamplingTopK:     {Verified: false, At: 900},
		"no.such.intent":              {Verified: true},
	}})
	disable := api.Bindings[system.IntentReasoningDisable]
	require.True(t, *disable.Verified)
	require.Equal(t, int64(1000), disable.VerifiedAt)
	require.Equal(t, "verify:reasoning.disable@1000", api.Evidence["bindings.reasoning.disable"])
	topK := api.Bindings[system.IntentSamplingTopK]
	require.False(t, *topK.Verified)
	require.Equal(t, int64(900), topK.VerifiedAt)
	require.Equal(t, "verify:sampling.top_k@900", api.Evidence["bindings.sampling.top_k"])
	require.Contains(t, trace, "bindings.reasoning.disable verified = true (verify:reasoning.disable@1000): 2xx without reasoning content")
	require.Contains(t, trace, "bindings.sampling.top_k verified = false (verify:sampling.top_k@900)")
	require.Contains(t, trace, "verify: no.such.intent has no binding; verdict ignored")
	for intent, b := range api.Bindings {
		if intent != system.IntentReasoningDisable && intent != system.IntentSamplingTopK {
			require.Nil(t, b.Verified, intent)
			require.Zero(t, b.VerifiedAt, intent)
		}
	}

	raw := mustJSON(t, api)
	require.Contains(t, raw, `"verified":true,"verifiedAt":1000`)
	require.Contains(t, raw, `"verified":false,"verifiedAt":900`)
}

func TestMergeRuntimeRejectedParams(t *testing.T) {
	t.Run("drops by root param and downgrades the mode", func(t *testing.T) {
		api := vllmQwenSpec(t)
		require.NotNil(t, api.Bindings[system.IntentReasoningBudget])
		trace := MergeRuntime(api, Runtime{RejectedParams: []string{"chat_template_kwargs"}, At: 77})
		require.Nil(t, api.Bindings[system.IntentReasoningDisable])
		require.Nil(t, api.Bindings[system.IntentReasoningEnable])
		require.Equal(t, "verify:rejected:chat_template_kwargs@77", api.Evidence["bindings.reasoning.disable"])
		require.Equal(t, "verify:rejected:chat_template_kwargs@77", api.Evidence["bindings.reasoning.enable"])
		require.NotNil(t, api.Bindings[system.IntentSamplingTopK], "unrelated bindings stay")
		require.Contains(t, trace, "bindings: reasoning.disable dropped — backend rejected parameter chat_template_kwargs (verify:rejected:chat_template_kwargs@77)")
		require.Equal(t, system.ThinkingModeTunable, api.Thinking.Mode)
		require.Contains(t, trace, "thinking.mode: controllable -> tunable after the drop")
	})
	t.Run("drops by the exact dotted param and leaves siblings", func(t *testing.T) {
		api := vllmQwenSpec(t)
		require.Equal(t, "thinking_token_budget", api.Bindings[system.IntentReasoningBudget].Param)
		require.Equal(t, "top_k", api.Bindings[system.IntentSamplingTopK].Param)
		trace := MergeRuntime(api, Runtime{RejectedParams: []string{"chat_template_kwargs.enable_thinking", "top_k"}, At: 78})
		require.Nil(t, api.Bindings[system.IntentReasoningDisable])
		require.Nil(t, api.Bindings[system.IntentReasoningEnable])
		require.Nil(t, api.Bindings[system.IntentSamplingTopK])
		require.NotNil(t, api.Bindings[system.IntentReasoningBudget])
		require.NotNil(t, api.Bindings[system.IntentSamplingMinP])
		require.Equal(t, "verify:rejected:chat_template_kwargs.enable_thinking@78", api.Evidence["bindings.reasoning.disable"])
		require.Equal(t, "verify:rejected:top_k@78", api.Evidence["bindings.sampling.top_k"])
		require.Contains(t, trace, "bindings: reasoning.enable dropped — backend rejected parameter chat_template_kwargs.enable_thinking (verify:rejected:chat_template_kwargs.enable_thinking@78)")
		require.Equal(t, system.ThinkingModeTunable, api.Thinking.Mode)
	})
	t.Run("a rejected path covers its descendants only", func(t *testing.T) {
		api := &system.ModelApiSpec{Bindings: map[string]*system.ParamBinding{
			"a": {Kind: system.BindingKindBodyParam, Param: "extra_body.thinking"},
			"b": {Kind: system.BindingKindBodyParam, Param: "extra_body.thinking.type"},
			"c": {Kind: system.BindingKindBodyParam, Param: "extra_body.top_k"},
			"d": {Kind: system.BindingKindBodyParam, Param: "extra_body_thinking"},
		}}
		MergeRuntime(api, Runtime{RejectedParams: []string{"extra_body.thinking"}, At: 1})
		require.Nil(t, api.Bindings["a"])
		require.Nil(t, api.Bindings["b"])
		require.NotNil(t, api.Bindings["c"])
		require.NotNil(t, api.Bindings["d"])
		require.Equal(t, "verify:rejected:extra_body.thinking@1", api.Evidence["bindings.b"])

		MergeRuntime(api, Runtime{RejectedParams: []string{"extra_body"}, At: 2})
		require.Nil(t, api.Bindings["c"])
		require.NotNil(t, api.Bindings["d"])
		require.Equal(t, "verify:rejected:extra_body@2", api.Evidence["bindings.c"])
	})
	t.Run("removes the thinking block when no reasoning binding remains", func(t *testing.T) {
		api := Compose(Evidence{Stack: "ollama", ModelName: "qwen3:32b", OllamaThinking: true})
		require.Equal(t, "reasoning_effort", api.Bindings[system.IntentReasoningDisable].Param)
		trace := MergeRuntime(api, Runtime{RejectedParams: []string{"reasoning_effort"}, At: 1})
		require.Nil(t, api.Bindings[system.IntentReasoningDisable])
		require.Nil(t, api.Bindings[system.IntentReasoningEffort])
		require.Nil(t, api.Thinking)
		require.Contains(t, trace, "thinking: controllable dropped — no reasoning binding remains")
	})
	t.Run("derived reasoning goes with the last reasoning binding", func(t *testing.T) {
		ev := Evidence{Stack: "ollama", ModelName: "qwen3:32b", OllamaThinking: true}
		ev.Caps.SetContext(8192, "ollama:/api/show.num_ctx")
		api := Compose(ev)
		require.Equal(t, SourceDerivedBindings, api.Evidence["capabilities.reasoning"])
		trace := MergeRuntime(api, Runtime{RejectedParams: []string{"reasoning_effort"}, At: 1})
		require.Nil(t, api.Thinking)
		require.Nil(t, api.Capabilities.Reasoning)
		require.NotContains(t, api.Evidence, "capabilities.reasoning")
		require.Equal(t, int64(8192), *api.Capabilities.Context)
		require.Contains(t, trace, "capabilities.reasoning dropped — derived:bindings.reasoning names no remaining binding")
		for key := range api.Evidence {
			if strings.HasPrefix(key, "capabilities.") {
				require.True(t, capFieldPresent(api.Capabilities, strings.TrimPrefix(key, "capabilities.")), key)
			}
		}
	})
	t.Run("derived reasoning alone leaves no empty capabilities block", func(t *testing.T) {
		api := Compose(Evidence{Stack: "ollama", ModelName: "qwen3:32b", OllamaThinking: true, Registry: &RegistryFacts{Name: "models.dev", Date: "2026-09-25"}})
		require.Equal(t, SourceDerivedBindings, api.Evidence["capabilities.reasoning"])
		MergeRuntime(api, Runtime{RejectedParams: []string{"reasoning_effort"}, At: 1})
		require.Nil(t, api.Capabilities)
		require.NotContains(t, api.Evidence, "capabilities.reasoning")
	})
	t.Run("backend-reported reasoning survives the drop", func(t *testing.T) {
		ev := Evidence{Stack: "ollama", ModelName: "qwen3:32b", OllamaThinking: true}
		ev.Caps.SetReasoning(true, "ollama:/api/show.capabilities")
		api := Compose(ev)
		MergeRuntime(api, Runtime{RejectedParams: []string{"reasoning_effort"}, At: 1})
		require.Nil(t, api.Thinking)
		require.True(t, *api.Capabilities.Reasoning)
		require.Equal(t, "ollama:/api/show.capabilities", api.Evidence["capabilities.reasoning"])
	})
	t.Run("reasoning seen in the same round trip re-sets reasoning after the drop", func(t *testing.T) {
		ev := Evidence{Stack: "ollama", ModelName: "qwen3:32b", OllamaThinking: true}
		ev.Caps.SetContext(8192, "ollama:/api/show.num_ctx")
		api := Compose(ev)
		MergeRuntime(api, Runtime{RejectedParams: []string{"reasoning_effort"}, ReasoningSeen: true, At: 1})
		require.True(t, *api.Capabilities.Reasoning)
		require.Equal(t, "health-probe:reasoning_seen", api.Evidence["capabilities.reasoning"])
	})
	t.Run("always_on survives a drop", func(t *testing.T) {
		api := Compose(Evidence{Stack: "ollama", ModelName: "deepseek-r1"})
		require.Equal(t, system.ThinkingModeAlwaysOn, api.Thinking.Mode)
		require.Equal(t, "options.top_k", api.Bindings[system.IntentSamplingTopK].Param)
		MergeRuntime(api, Runtime{RejectedParams: []string{"options"}, At: 3})
		require.Nil(t, api.Bindings[system.IntentSamplingTopK])
		require.Nil(t, api.Bindings[system.IntentContextNumCtx])
		require.NotNil(t, api.Bindings[system.IntentResponseFormatJSON])
		require.Equal(t, "verify:rejected:options@3", api.Evidence["bindings.context.num_ctx"])
		require.Equal(t, system.ThinkingModeAlwaysOn, api.Thinking.Mode)
	})
	t.Run("the tool_choice sentinel drops nothing", func(t *testing.T) {
		api := vllmQwenSpec(t)
		before := mustJSON(t, api)
		require.Empty(t, MergeRuntime(api, Runtime{RejectedParams: []string{"tool_choice", "unknown_param"}}))
		require.Equal(t, before, mustJSON(t, api))
	})
	t.Run("no bindings is a no-op", func(t *testing.T) {
		api := Compose(Evidence{ModelName: "qwen3-32b"})
		require.Nil(t, api.Bindings)
		before := mustJSON(t, api)
		require.Empty(t, MergeRuntime(api, Runtime{RejectedParams: []string{"top_k"}}))
		require.Equal(t, before, mustJSON(t, api))
	})
}

func TestMergeRuntimeCapabilityFacts(t *testing.T) {
	t.Run("facts land with their own sources", func(t *testing.T) {
		api := Compose(Evidence{Stack: "vllm", ModelName: "qwen3-32b"})
		var facts CapabilityFacts
		facts.SetStructured([]string{"json_schema"}, "verify:json_schema@3")
		facts.AddInputModality("image", "verify:image@3")
		facts.SetMaxOutput(0, "verify:bogus@3")
		trace := MergeRuntime(api, Runtime{CapabilityFacts: facts})
		require.Equal(t, []string{"json_schema"}, api.Capabilities.Structured)
		require.Equal(t, []string{"text", "image"}, api.Capabilities.Modalities.Input, "a proven image input implies text (amendment b)")
		require.Nil(t, api.Capabilities.MaxOutput)
		require.Equal(t, "verify:json_schema@3", api.Evidence["capabilities.structured"])
		require.Equal(t, "verify:image@3", api.Evidence["capabilities.modalities.input"])
		require.Contains(t, trace, "capabilities.structured = json_schema (verify:json_schema@3)")
		require.Contains(t, trace, "capabilities.modalities.input: text prepended — a chat model accepts text (verify:image@3)")
	})
	t.Run("a proven image input on a spec without modalities publishes text and image", func(t *testing.T) {
		api := Compose(Evidence{Stack: "vllm", ModelName: "qwen3-32b"})
		require.Nil(t, api.Capabilities)
		var facts CapabilityFacts
		facts.AddInputModality("image", verifySource("image", 3))
		MergeRuntime(api, Runtime{CapabilityFacts: facts})
		require.Equal(t, []string{"text", "image"}, api.Capabilities.Modalities.Input)
		require.Nil(t, api.Capabilities.Modalities.Output, "the output list stays as reported")
		require.Equal(t, verifySource("image", 3), api.Evidence["capabilities.modalities.input"])
	})
	t.Run("list facts extend a composed list instead of replacing it", func(t *testing.T) {
		ev := Evidence{Stack: "vllm", ModelName: "qwen3-32b"}
		ev.Caps.SetStructured([]string{"json_object", "json_schema", "grammar", "regex", "choice"}, "vllm:/v1/models.structured")
		ev.Caps.SetInputModalities([]string{"text", "image", "audio"}, "vllm:/v1/models.modalities")
		api := Compose(ev)
		before := mustJSON(t, api)
		var facts CapabilityFacts
		facts.SetStructured([]string{"json_schema"}, "verify:json_schema@3")
		facts.AddInputModality("image", "verify:image@3")
		trace := MergeRuntime(api, Runtime{CapabilityFacts: facts})
		require.Equal(t, before, mustJSON(t, api), "a proven entry that is already listed changes nothing")
		require.Contains(t, trace, "capabilities.structured: runtime value [json_schema] adds nothing to [json_object json_schema grammar regex choice]")
		require.Contains(t, trace, "capabilities.modalities.input: runtime value [image] adds nothing to [text image audio]")

		facts = CapabilityFacts{}
		facts.SetStructured([]string{"json_schema", "bogus"}, "verify:json_schema@4")
		facts.AddInputModality("video", "verify:video@4")
		facts.SetOutputModalities([]string{"audio"}, "verify:audio_out@4")
		trace = MergeRuntime(api, Runtime{CapabilityFacts: facts})
		require.Equal(t, []string{"json_object", "json_schema", "grammar", "regex", "choice"}, api.Capabilities.Structured)
		require.Equal(t, "vllm:/v1/models.structured", api.Evidence["capabilities.structured"])
		require.Equal(t, []string{"text", "image", "audio", "video"}, api.Capabilities.Modalities.Input)
		require.Equal(t, "verify:video@4", api.Evidence["capabilities.modalities.input"])
		require.Equal(t, []string{"audio"}, api.Capabilities.Modalities.Output)
		require.Equal(t, "verify:audio_out@4", api.Evidence["capabilities.modalities.output"])
		require.Contains(t, trace, "capabilities.modalities.input = text|image|audio|video (verify:video@4)")
	})
	t.Run("a single-entry composed list grows by the proven entry", func(t *testing.T) {
		ev := Evidence{Stack: "vllm", ModelName: "qwen3-32b"}
		ev.Caps.SetStructured([]string{"json_object"}, "vllm:/v1/models.structured")
		api := Compose(ev)
		var facts CapabilityFacts
		facts.SetStructured([]string{"json_schema"}, "verify:json_schema@5")
		MergeRuntime(api, Runtime{CapabilityFacts: facts})
		require.Equal(t, []string{"json_object", "json_schema"}, api.Capabilities.Structured)
		require.Equal(t, "verify:json_schema@5", api.Evidence["capabilities.structured"])
	})
	t.Run("facts override backend but not declared", func(t *testing.T) {
		ev := Evidence{Stack: "vllm", ModelName: "qwen3-32b", DeclaredCaps: &config.DeclaredCapabilities{Structured: []string{"json_object"}}}
		ev.Caps.SetContext(4096, "vllm:/v1/models.max_model_len")
		api := Compose(ev)
		var facts CapabilityFacts
		facts.SetStructured([]string{"json_schema"}, "verify:json_schema@3")
		facts.SetContext(32768, "verify:tokenize.max_model_len@3")
		trace := MergeRuntime(api, Runtime{CapabilityFacts: facts})
		require.Equal(t, []string{"json_object"}, api.Capabilities.Structured)
		require.Equal(t, int64(32768), *api.Capabilities.Context)
		require.Equal(t, "verify:tokenize.max_model_len@3", api.Evidence["capabilities.context"])
		require.True(t, traceHas(trace, "capabilities.structured: runtime value ignored — declared"), trace)
	})
	t.Run("an out-of-vocabulary fact leaves the spec unchanged", func(t *testing.T) {
		api := Compose(Evidence{Stack: "vllm", ModelName: "qwen3-32b"})
		before := mustJSON(t, api)
		facts := CapabilityFacts{ModelCapabilities: system.ModelCapabilities{Context: int64p(-1)}, Sources: map[string]string{CapFieldContext: "verify:x@1"}}
		MergeRuntime(api, Runtime{CapabilityFacts: facts})
		require.Equal(t, before, mustJSON(t, api))
	})
}

func TestMergeRuntimeStackHintAndServedID(t *testing.T) {
	t.Run("hint names the stack only when detection left it empty", func(t *testing.T) {
		api := Compose(Evidence{ModelName: "qwen3-32b"})
		require.Equal(t, "", api.Stack)
		trace := MergeRuntime(api, Runtime{StackHints: []string{"bogus", "litellm"}})
		require.Equal(t, "litellm", api.Stack)
		require.Equal(t, "health-probe:headers", api.Evidence["stack"])
		require.Contains(t, trace, "stack = litellm (health-probe:headers)")
	})
	t.Run("hint is ignored when the stack is known", func(t *testing.T) {
		api := vllmQwenSpec(t)
		trace := MergeRuntime(api, Runtime{StackHints: []string{"litellm"}})
		require.Equal(t, "vllm", api.Stack)
		require.True(t, traceHas(trace, "stack hints [litellm] ignored"), trace)
	})
	t.Run("served id refines the family", func(t *testing.T) {
		api := Compose(Evidence{Stack: "openai", ModelName: "my-alias"})
		require.Equal(t, "", api.ModelFamily)
		trace := MergeRuntime(api, Runtime{ServedModelID: "Qwen/Qwen3-32B"})
		require.Equal(t, "qwen3", api.ModelFamily)
		require.Contains(t, trace, `family "qwen3": from runtime-served model id`)
	})
}

func TestMergeRuntimeOnCloneLeavesCacheIntact(t *testing.T) {
	cached := vllmQwenSpec(t)
	before := mustJSON(t, cached)
	live := cached.Clone()
	MergeRuntime(live, Runtime{
		ContextLimit: 1, ToolsSupported: boolp(true), ReasoningSeen: true, ReasoningDisabledInRequest: true,
		RejectedParams: []string{"top_k"}, Verified: map[string]VerifiedBinding{system.IntentReasoningDisable: {Verified: false}},
	})
	require.Equal(t, before, mustJSON(t, cached))
	require.NotEqual(t, before, mustJSON(t, live))
}
