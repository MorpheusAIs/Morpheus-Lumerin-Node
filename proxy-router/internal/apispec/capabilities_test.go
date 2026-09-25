package apispec

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/config"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/system"
	"github.com/stretchr/testify/require"
)

func i64(v int64) *int64 { return &v }
func bp(v bool) *bool    { return &v }

func joined(lines []string) string { return strings.Join(lines, "\n") }

func vllmFacts() CapabilityFacts {
	var f CapabilityFacts
	f.SetContext(40960, "vllm:/v1/models.max_model_len")
	f.SetStructured([]string{"json_object", "json_schema", "grammar", "regex", "choice"}, "vllm:built-in")
	return f
}

func TestCapabilityFactsSettersValidate(t *testing.T) {
	var f CapabilityFacts
	require.True(t, f.Empty())
	require.False(t, f.SetContext(0, "s"), "0 is not a capability value")
	require.False(t, f.SetContext(-1, "s"))
	require.False(t, f.SetContext(1<<31, "s"))
	require.True(t, f.SetContext(1<<31-1, "s"))
	require.True(t, f.SetContextTrained(131072, "t"))
	require.False(t, f.SetMaxOutput(0, "m"))
	require.True(t, f.SetMaxOutput(4096, "m"))
	require.Equal(t, int64(1<<31-1), *f.Context)
	require.Equal(t, "s", f.Sources[CapFieldContext])
	require.Equal(t, "t", f.Sources[CapFieldContextTrained])
	require.Equal(t, "m", f.Sources[CapFieldMaxOutput])

	require.False(t, f.SetInputModalities([]string{"vision", "3d"}, "x"), "nothing in the vocabulary: field untouched")
	require.Nil(t, f.Modalities)
	require.True(t, f.SetInputModalities([]string{"text", "vision", "image", "text"}, "mod"))
	require.Equal(t, []string{"text", "image"}, f.Modalities.Input, "filtered, deduped, order kept")
	require.True(t, f.AddInputModality("audio", "add"))
	require.True(t, f.AddInputModality("audio", "add-again"), "re-adding is accepted and idempotent")
	require.Equal(t, []string{"text", "image", "audio"}, f.Modalities.Input)
	require.Equal(t, "add-again", f.Sources[CapFieldModalitiesInput])
	require.False(t, f.AddInputModality("embeddings", "add"), "embeddings is an output modality only")
	require.True(t, f.AddOutputModality("embeddings", "out"))
	require.False(t, f.AddOutputModality("pdf", "out"))
	require.True(t, f.SetOutputModalities([]string{"text"}, "out2"))
	require.Equal(t, []string{"text"}, f.Modalities.Output)

	require.False(t, f.SetStructured([]string{"xml"}, "st"))
	require.True(t, f.SetStructured([]string{"json_schema", "xml", "json_schema", "grammar"}, "st"))
	require.Equal(t, []string{"json_schema", "grammar"}, f.Structured)

	require.False(t, f.SetQuantization("", "q"))
	require.False(t, f.SetQuantization("fp8 (dynamic)", "q"), "not a plain label")
	require.True(t, f.SetQuantization("Q4_K_M", "q"))
	require.Equal(t, "Q4_K_M", f.Quantization)

	f.SetTools(false, "tools")
	f.SetToolsParallel(true, "par")
	f.SetReasoning(false, "r")
	f.SetWebSearch(true, "w")
	f.SetEmbeddings(false, "e")
	f.SetPromptCache(true, "p")
	require.False(t, *f.Tools)
	require.True(t, *f.ToolsParallel)
	require.False(t, *f.Reasoning)
	require.True(t, *f.WebSearch)
	require.False(t, *f.Embeddings)
	require.True(t, *f.PromptCache)
	require.False(t, f.Empty())
	require.Len(t, f.Sources, 13, "every wire field has a source")
}

func TestCapabilityFactsFillFromKeepsWhatIsSet(t *testing.T) {
	var hop1, hop2 CapabilityFacts
	hop1.SetContext(8192, "litellm:/model_group/info.max_input_tokens")
	hop1.SetTools(true, "litellm:/model_group/info.supports_function_calling")
	hop2.SetContext(40960, "vllm:/v1/models.max_model_len")
	hop2.SetInputModalities([]string{"text"}, "vllm:x")
	hop2.FillFrom(&hop1)
	require.Equal(t, int64(40960), *hop2.Context, "already set: kept")
	require.Equal(t, "vllm:/v1/models.max_model_len", hop2.Sources[CapFieldContext])
	require.True(t, *hop2.Tools, "unset: filled")
	require.Equal(t, "litellm:/model_group/info.supports_function_calling", hop2.Sources[CapFieldTools])
	*hop1.Tools = false
	require.True(t, *hop2.Tools, "copied, not aliased")
	hop2.FillFrom(nil)
}

func TestComposeRule1PrecedenceDeclaredBackendRegistry(t *testing.T) {
	facts := vllmFacts()
	facts.SetTools(true, "vllm:verify")
	facts.SetWebSearch(false, "vllm:none")
	declared := &config.DeclaredCapabilities{Context: i64(32768), Tools: bp(false), Structured: []string{"json_schema"}}
	registry := &RegistryFacts{Name: "models.dev", Date: "2026-09-25", ModelCapabilities: system.ModelCapabilities{Context: i64(131072), MaxOutput: i64(8192), Tools: bp(true), WebSearch: bp(true), Embeddings: bp(false), Modalities: &system.Modalities{Input: []string{"text", "image"}}}}
	api, lines := ComposeWithTrace(Evidence{Stack: "vllm", ModelName: "llama-3.3-70b", Caps: facts, DeclaredCaps: declared, Registry: registry})
	require.NotNil(t, api.Capabilities)
	c := api.Capabilities
	require.Equal(t, int64(32768), *c.Context, "declared beats backend and registry")
	require.False(t, *c.Tools, "declared false beats backend true")
	require.Equal(t, []string{"json_schema"}, c.Structured, "declared beats the backend's built-in list")
	require.False(t, *c.WebSearch, "backend beats registry")
	require.Equal(t, int64(8192), *c.MaxOutput, "registry fills what nothing else set")
	require.False(t, *c.Embeddings)
	require.Equal(t, []string{"text", "image"}, c.Modalities.Input)
	require.Nil(t, c.ContextTrained)
	require.Equal(t, map[string]string{
		"capabilities.context":          "declared",
		"capabilities.tools":            "declared",
		"capabilities.structured":       "declared",
		"capabilities.webSearch":        "vllm:none",
		"capabilities.maxOutput":        "registry:models.dev@2026-09-25",
		"capabilities.embeddings":       "registry:models.dev@2026-09-25",
		"capabilities.modalities.input": "registry:models.dev@2026-09-25",
	}, api.Evidence)
	tr := joined(lines)
	require.Contains(t, tr, "capabilities.context = 32768 (declared)")
	require.Contains(t, tr, "capabilities.tools = false (declared)")
	require.Contains(t, tr, "capabilities.structured = json_schema (declared)")
	require.Contains(t, tr, "capabilities.webSearch = false (vllm:none)")
	require.Contains(t, tr, "capabilities.maxOutput = 8192 (registry:models.dev@2026-09-25)")
	require.Contains(t, tr, "capabilities.modalities.input = text|image (registry:models.dev@2026-09-25)")
	require.Equal(t, 1, strings.Count(tr, "capabilities.context = "), "a field is set once, traced once")
	require.Nil(t, c.Reasoning, "llama with no reasoning bindings and no report: unknown")
	require.NotContains(t, tr, "context: served window", "a declared context is not a served one")
}

func TestComposeRule2ServedBeatsTrained(t *testing.T) {
	t.Run("both known", func(t *testing.T) {
		var f CapabilityFacts
		f.SetContext(8192, "llamacpp:/props.default_generation_settings.n_ctx")
		f.SetContextTrained(131072, "llamacpp:/v1/models.meta.n_ctx_train")
		api := Compose(Evidence{Stack: "llamacpp", ModelName: "llama-3.3-70b", Caps: f})
		require.Equal(t, int64(8192), *api.Capabilities.Context)
		require.Equal(t, int64(131072), *api.Capabilities.ContextTrained)
		require.Equal(t, "llamacpp:/props.default_generation_settings.n_ctx", api.Evidence["capabilities.context"])
		require.Equal(t, "llamacpp:/v1/models.meta.n_ctx_train", api.Evidence["capabilities.contextTrained"])
	})
	t.Run("only trained known", func(t *testing.T) {
		var f CapabilityFacts
		f.SetContextTrained(131072, "ollama:/api/show.model_info.llama.context_length")
		api, lines := ComposeWithTrace(Evidence{Stack: "ollama", ModelName: "llama3.3", Caps: f})
		require.Equal(t, int64(131072), *api.Capabilities.Context, "the trained window stands in")
		require.Nil(t, api.Capabilities.ContextTrained, "not repeated")
		require.Equal(t, "ollama:/api/show.model_info.llama.context_length", api.Evidence["capabilities.context"], "the evidence names the trained source")
		require.NotContains(t, api.Evidence, "capabilities.contextTrained")
		require.Contains(t, joined(lines), "capabilities.context = 131072 (ollama:/api/show.model_info.llama.context_length)")
	})
	t.Run("declared served with backend trained", func(t *testing.T) {
		var f CapabilityFacts
		f.SetContextTrained(131072, "ollama:/api/show.model_info.llama.context_length")
		api := Compose(Evidence{Stack: "ollama", ModelName: "llama3.3", Caps: f, DeclaredCaps: &config.DeclaredCapabilities{Context: i64(16384)}})
		require.Equal(t, int64(16384), *api.Capabilities.Context)
		require.Equal(t, int64(131072), *api.Capabilities.ContextTrained, "the trained window is still a fact")
		require.Equal(t, "declared", api.Evidence["capabilities.context"])
		require.Equal(t, "ollama:/api/show.model_info.llama.context_length", api.Evidence["capabilities.contextTrained"])
	})
}

func TestComposeRule3Reasoning(t *testing.T) {
	declaredCtx := &config.DeclaredCapabilities{Context: i64(8192)}
	t.Run("always_on", func(t *testing.T) {
		api := Compose(Evidence{Stack: "vllm", ModelName: "deepseek-r1", DeclaredCaps: declaredCtx})
		require.True(t, *api.Capabilities.Reasoning)
		require.Equal(t, SourceDerivedThinkingMode, api.Evidence["capabilities.reasoning"])
	})
	t.Run("surviving reasoning binding", func(t *testing.T) {
		api, lines := ComposeWithTrace(Evidence{Stack: "vllm", ModelName: "qwen3-32b", DeclaredCaps: declaredCtx})
		require.True(t, *api.Capabilities.Reasoning)
		require.Equal(t, SourceDerivedBindings, api.Evidence["capabilities.reasoning"])
		require.Contains(t, joined(lines), "capabilities.reasoning = true (derived:bindings.reasoning)")
	})
	t.Run("derived after the litellm filter", func(t *testing.T) {
		api := Compose(Evidence{Stack: "litellm", ModelName: "qwen3-32b", GatewayReasoning: true, LiteLLMSeen: true, LiteLLMSupportedParams: []string{"temperature"}, DeclaredCaps: declaredCtx})
		require.Nil(t, api.Bindings[system.IntentReasoningDisable])
		require.NotNil(t, api.Bindings[system.IntentReasoningBudget], "the one surviving reasoning knob")
		require.True(t, *api.Capabilities.Reasoning)
		require.Equal(t, SourceDerivedBindings, api.Evidence["capabilities.reasoning"])
	})
	t.Run("never false without a report", func(t *testing.T) {
		api := Compose(Evidence{Stack: "vllm", ModelName: "llama-3.3-70b", DeclaredCaps: declaredCtx})
		require.Nil(t, api.Capabilities.Reasoning)
	})
	t.Run("backend false stands even with family bindings", func(t *testing.T) {
		var f CapabilityFacts
		f.SetReasoning(false, "litellm:/model_group/info.supports_reasoning")
		api := Compose(Evidence{Stack: "litellm", ModelName: "qwen3-32b", Declared: true, Caps: f})
		require.NotNil(t, api.Bindings[system.IntentReasoningDisable], "documented knobs stay")
		require.False(t, *api.Capabilities.Reasoning)
		require.Equal(t, "litellm:/model_group/info.supports_reasoning", api.Evidence["capabilities.reasoning"])
	})
	t.Run("backend true", func(t *testing.T) {
		var f CapabilityFacts
		f.SetReasoning(true, "sglang:/model_info.reasoning_parser")
		api := Compose(Evidence{Stack: "sglang", ModelName: "llama-3.3-70b", Caps: f})
		require.True(t, *api.Capabilities.Reasoning)
		require.Equal(t, "sglang:/model_info.reasoning_parser", api.Evidence["capabilities.reasoning"])
	})
	t.Run("derived true beats registry false", func(t *testing.T) {
		reg := &RegistryFacts{Name: "litellm", Date: "2026-09-25", ModelCapabilities: system.ModelCapabilities{Reasoning: bp(false)}}
		api := Compose(Evidence{Stack: "vllm", ModelName: "qwen3-32b", Registry: reg})
		require.True(t, *api.Capabilities.Reasoning)
		require.Equal(t, SourceDerivedBindings, api.Evidence["capabilities.reasoning"])
		plain, lines := ComposeWithTrace(Evidence{Stack: "vllm", ModelName: "llama-3.3-70b", Registry: reg})
		require.Nil(t, plain.Capabilities, "nothing derived, and a registry false is not a backend report: unset")
		require.Contains(t, joined(lines), "capabilities.reasoning: registry false ignored — only a backend may report false (registry:litellm@2026-09-25)")
	})
	t.Run("registry true fills", func(t *testing.T) {
		reg := &RegistryFacts{Name: "models.dev", Date: "2026-09-25", ModelCapabilities: system.ModelCapabilities{Reasoning: bp(true)}}
		api := Compose(Evidence{Stack: "vllm", ModelName: "llama-3.3-70b", Registry: reg})
		require.True(t, *api.Capabilities.Reasoning)
		require.Equal(t, "registry:models.dev@2026-09-25", api.Evidence["capabilities.reasoning"])
	})
	t.Run("an empty declared block opens no gate", func(t *testing.T) {
		api := Compose(Evidence{Stack: "vllm", ModelName: "qwen3-32b", DeclaredCaps: &config.DeclaredCapabilities{}})
		require.NotNil(t, api.Bindings[system.IntentReasoningDisable], "family bindings as before")
		require.Nil(t, api.Capabilities, "{} is no declared fact: no block, no derived reasoning")
		require.Nil(t, api.Evidence)
	})
	t.Run("no capability evidence: no block at all", func(t *testing.T) {
		api := Compose(Evidence{Stack: "vllm", ModelName: "deepseek-r1"})
		require.Nil(t, api.Capabilities)
		require.Nil(t, api.Evidence)
	})
}

func TestComposeRule4MandatoryReasoning(t *testing.T) {
	registry := bindingSet{
		system.IntentReasoningDisable: {Kind: bodyParam, Param: "reasoning.effort", ParamType: "string", Value: "none"},
		system.IntentReasoningEnable:  {Kind: bodyParam, Param: "reasoning.enabled", ParamType: "boolean", Value: true},
		system.IntentReasoningEffort:  {Kind: bodyParam, Param: "reasoning.effort", ParamType: "enum", EnumValues: []string{"low", "high"}},
	}
	api, lines := ComposeWithTrace(Evidence{Stack: "openrouter", ModelName: "x/y", GatewayReasoning: true, RegistryBindings: registry, MandatoryReasoning: true})
	require.Equal(t, system.ThinkingModeAlwaysOn, api.Thinking.Mode)
	require.Nil(t, api.Bindings[system.IntentReasoningDisable])
	require.Nil(t, api.Bindings[system.IntentReasoningEnable])
	require.NotNil(t, api.Bindings[system.IntentReasoningEffort], "effort is still tunable")
	require.Nil(t, api.Bindings[system.IntentReasoningBudget], "always_on skips the stack table's reasoning knobs")
	require.Equal(t, "reasoning.exclude", api.Bindings[system.IntentReasoningFormat].Param)
	tr := joined(lines)
	require.Contains(t, tr, "thinking: the backend reports reasoning as mandatory -> always_on")
	require.Contains(t, tr, "bindings: reasoning.disable dropped — reasoning is mandatory")
	require.Contains(t, tr, "bindings: reasoning.enable dropped — reasoning is mandatory")

	fam := Compose(Evidence{Stack: "vllm", ModelName: "qwen3-32b", MandatoryReasoning: true})
	require.Equal(t, system.ThinkingModeAlwaysOn, fam.Thinking.Mode)
	require.Nil(t, fam.Bindings[system.IntentReasoningDisable], "family toggle dropped too")
}

func TestComposeRule5EffortsOnEveryStack(t *testing.T) {
	t.Run("venice", func(t *testing.T) {
		api, lines := ComposeWithTrace(Evidence{Stack: "venice", ModelName: "qwen3-235b", GatewayReasoning: true, ReasoningEfforts: []string{"low", "medium", "high"}, EffortsSource: "venice:/api/v1/models.model_spec.capabilities.reasoningEffortOptions"})
		require.Equal(t, []string{"low", "medium", "high"}, api.Bindings[system.IntentReasoningEffort].EnumValues)
		require.Equal(t, "venice_parameters.disable_thinking", api.Bindings[system.IntentReasoningDisable].Param, "a disable that is not reasoning_effort none is untouched")
		require.Equal(t, "venice:/api/v1/models.model_spec.capabilities.reasoningEffortOptions", api.Evidence["bindings.reasoning.effort"])
		require.Contains(t, joined(lines), "bindings: reasoning.effort levels from venice:/api/v1/models.model_spec.capabilities.reasoningEffortOptions -> low|medium|high")
	})
	t.Run("openrouter without none drops the none-based disable", func(t *testing.T) {
		api, lines := ComposeWithTrace(Evidence{Stack: "openrouter", ModelName: "x/y", GatewayReasoning: true, ReasoningEfforts: []string{"low", "high"}, EffortsSource: "openrouter:/api/v1/models.reasoning.supported_efforts"})
		require.Equal(t, []string{"low", "high"}, api.Bindings[system.IntentReasoningEffort].EnumValues)
		require.Nil(t, api.Bindings[system.IntentReasoningDisable])
		require.Equal(t, true, api.Bindings[system.IntentReasoningEnable].Value, "openrouter's enable is a boolean, not an effort: value untouched")
		require.Contains(t, joined(lines), "reasoning.disable dropped — openrouter:/api/v1/models.reasoning.supported_efforts does not list none")
	})
	t.Run("ollama levels with none keep the toggle", func(t *testing.T) {
		api := Compose(Evidence{Stack: "ollama", ModelName: "some-hybrid", OllamaThinking: true, ReasoningEfforts: []string{"none", "low", "medium", "high"}, EffortsSource: "ollama:/api/show.thinking.values"})
		require.Equal(t, []string{"none", "low", "medium", "high"}, api.Bindings[system.IntentReasoningEffort].EnumValues)
		require.Equal(t, "none", api.Bindings[system.IntentReasoningDisable].Value)
		require.Equal(t, "low", api.Bindings[system.IntentReasoningEnable].Value)
		require.Equal(t, system.ThinkingModeControllable, api.Thinking.Mode)
	})
	t.Run("ollama levels without none: effort only", func(t *testing.T) {
		api := Compose(Evidence{Stack: "ollama", ModelName: "gpt-oss:120b", OllamaThinking: true, ReasoningEfforts: []string{"low", "medium", "high"}, EffortsSource: "ollama:/api/show.thinking.values"})
		require.Equal(t, []string{"low", "medium", "high"}, api.Bindings[system.IntentReasoningEffort].EnumValues)
		require.Nil(t, api.Bindings[system.IntentReasoningDisable])
		require.Equal(t, system.ThinkingModeTunable, api.Thinking.Mode)
	})
	t.Run("a list applies to the stack that reported it", func(t *testing.T) {
		api, lines := ComposeWithTrace(Evidence{Stack: "litellm", ModelName: "claude-sonnet-4-5", GatewayReasoning: true, ReasoningEfforts: []string{"low"}})
		require.Equal(t, []string{"low"}, api.Bindings[system.IntentReasoningEffort].EnumValues)
		require.Nil(t, api.Evidence, "no source: no evidence entry")
		require.Contains(t, joined(lines), "reasoning.effort levels from litellm supported_reasoning_efforts -> low")

		api, lines = ComposeWithTrace(Evidence{Stack: "venice", Via: "litellm", ModelName: "qwen3-235b", GatewayReasoning: true, ReasoningEfforts: []string{"low"}})
		require.Equal(t, stackBindings["venice"][system.IntentReasoningEffort].EnumValues, api.Bindings[system.IntentReasoningEffort].EnumValues, "litellm's list does not describe venice's param")
		require.Contains(t, joined(lines), "reasoning.effort levels from litellm supported_reasoning_efforts skipped — they describe litellm's param, not venice's")

		api, lines = ComposeWithTrace(Evidence{Stack: "venice", Via: "litellm", ModelName: "qwen3-235b", GatewayReasoning: true, ReasoningEfforts: []string{"low"}, EffortsSource: "litellm:/model_group/info.supported_reasoning_efforts"})
		require.Equal(t, stackBindings["venice"][system.IntentReasoningEffort].EnumValues, api.Bindings[system.IntentReasoningEffort].EnumValues)
		require.Contains(t, joined(lines), "skipped — they describe litellm's param, not venice's")

		api, _ = ComposeWithTrace(Evidence{Stack: "venice", Via: "litellm", ModelName: "qwen3-235b", GatewayReasoning: true, ReasoningEfforts: []string{"low"}, EffortsSource: "venice:/api/v1/models.model_spec.capabilities.reasoningEffortOptions"})
		require.Equal(t, []string{"low"}, api.Bindings[system.IntentReasoningEffort].EnumValues, "read from the upstream's own listing")

		api, lines = ComposeWithTrace(Evidence{Stack: "venice", ModelName: "qwen3-235b", GatewayReasoning: true, ReasoningEfforts: []string{"low"}})
		require.Equal(t, stackBindings["venice"][system.IntentReasoningEffort].EnumValues, api.Bindings[system.IntentReasoningEffort].EnumValues, "a source-less list is litellm's")
		require.Contains(t, joined(lines), "skipped — they describe litellm's param, not venice's")
	})
}

func TestComposeRule6AnthropicFacts(t *testing.T) {
	src := "anthropic:/v1/models/{id}.capabilities"
	t.Run("adaptive", func(t *testing.T) {
		f := &AnthropicFacts{ThinkingSupported: true, ThinkingTypes: []string{"adaptive", "enabled"}, EffortLevels: []string{"low", "medium", "high", "xhigh", "max"}, Structured: true, Source: src}
		api, lines := ComposeWithTrace(Evidence{Stack: "anthropic", ModelName: "claude-3-haiku-20240307", Anthropic: f})
		require.Equal(t, map[string]any{"type": "disabled"}, api.Bindings[system.IntentReasoningDisable].Value)
		require.Equal(t, map[string]any{"type": "adaptive"}, api.Bindings[system.IntentReasoningEnable].Value, "facts, not the 3.x name regex")
		require.Nil(t, api.Bindings[system.IntentReasoningBudget])
		require.Equal(t, []string{"low", "medium", "high", "xhigh", "max"}, api.Bindings[system.IntentReasoningEffort].EnumValues)
		require.Equal(t, "output_config.effort", api.Bindings[system.IntentReasoningEffort].Param)
		require.Equal(t, system.ThinkingModeControllable, api.Thinking.Mode)
		require.Equal(t, "thinking.display", api.Bindings[system.IntentReasoningFormat].Param, "stack table reasoning knob merged: reasoning known")
		require.Equal(t, []string{"json_schema"}, api.Capabilities.Structured)
		require.Equal(t, src, api.Evidence["capabilities.structured"])
		require.Equal(t, src, api.Evidence["bindings.reasoning.disable"])
		require.Equal(t, src, api.Evidence["bindings.reasoning.enable"])
		require.Equal(t, src, api.Evidence["bindings.reasoning.effort"])
		require.NotContains(t, api.Evidence, "bindings.reasoning.format", "table knob: no evidence entry")
		require.True(t, *api.Capabilities.Reasoning)
		require.Contains(t, joined(lines), "bindings: from anthropic's reported capabilities (thinking.supported=true types=[adaptive enabled] effort=[low medium high xhigh max] structured_outputs=true)")
	})
	t.Run("enabled only", func(t *testing.T) {
		f := &AnthropicFacts{ThinkingSupported: true, ThinkingTypes: []string{"enabled"}, Source: src}
		api := Compose(Evidence{Stack: "anthropic", ModelName: "claude-opus-4-6", Anthropic: f})
		require.Equal(t, map[string]any{"type": "enabled", "budget_tokens": 1024}, api.Bindings[system.IntentReasoningEnable].Value, "facts, not the 4.6 regex (adaptive)")
		require.Equal(t, "thinking.budget_tokens", api.Bindings[system.IntentReasoningBudget].Param)
		require.Nil(t, api.Bindings[system.IntentReasoningEffort])
		require.Nil(t, api.Capabilities.Structured, "no structured_outputs")
		require.True(t, *api.Capabilities.Reasoning, "vendor facts are capability evidence: reasoning derived from the thinking bindings")
		require.Equal(t, SourceDerivedBindings, api.Evidence["capabilities.reasoning"])
	})
	t.Run("thinking unsupported, efforts only", func(t *testing.T) {
		f := &AnthropicFacts{EffortLevels: []string{"low", "medium", "high", "xhigh", "max"}, Source: src}
		api := Compose(Evidence{Stack: "anthropic", ModelName: "claude-fable-5-1", Anthropic: f})
		require.Nil(t, api.Bindings[system.IntentReasoningDisable])
		require.Nil(t, api.Bindings[system.IntentReasoningEnable])
		require.Equal(t, system.ThinkingModeTunable, api.Thinking.Mode)
	})
	t.Run("nothing supported: no reasoning knobs, no family fallback", func(t *testing.T) {
		f := &AnthropicFacts{Source: src}
		api, lines := ComposeWithTrace(Evidence{Stack: "anthropic", ModelName: "claude-opus-4-6", Anthropic: f})
		require.Nil(t, api.Thinking)
		for intent := range api.Bindings {
			require.False(t, strings.HasPrefix(intent, "reasoning."), intent)
		}
		require.Equal(t, "cache_control", api.Bindings[system.IntentCachePrompt].Param, "non-reasoning table knobs still merge")
		require.NotContains(t, joined(lines), "family default")
	})
	t.Run("junk effort levels are dropped", func(t *testing.T) {
		f := &AnthropicFacts{ThinkingSupported: true, ThinkingTypes: []string{"adaptive"}, EffortLevels: []string{"low", "hi gh\n", "low"}, Source: src}
		api, lines := ComposeWithTrace(Evidence{Stack: "anthropic", ModelName: "claude-opus-4-6", Anthropic: f})
		require.Equal(t, []string{"low"}, api.Bindings[system.IntentReasoningEffort].EnumValues)
		require.Contains(t, joined(lines), "anthropic effort levels: 2 of 3 backend-reported entries dropped")
	})
	t.Run("effort levels are capped like every reported list", func(t *testing.T) {
		levels := make([]string, 0, maxReasoningEfforts+4)
		for i := 0; i < maxReasoningEfforts+4; i++ {
			levels = append(levels, fmt.Sprintf("level%d", i))
		}
		f := &AnthropicFacts{EffortLevels: levels, Source: src}
		api := Compose(Evidence{Stack: "anthropic", ModelName: "claude-opus-4-6", Anthropic: f})
		require.Len(t, api.Bindings[system.IntentReasoningEffort].EnumValues, maxReasoningEfforts)
	})
	t.Run("structured fact does not write into the caller's facts", func(t *testing.T) {
		var caller CapabilityFacts
		caller.SetContext(200000, "anthropic:/v1/models/{id}.max_input_tokens")
		f := &AnthropicFacts{ThinkingSupported: true, ThinkingTypes: []string{"adaptive"}, Structured: true, Source: src}
		ev := Evidence{Stack: "anthropic", ModelName: "claude-opus-4-6", Anthropic: f, Caps: caller}
		api := Compose(ev)
		require.Equal(t, []string{"json_schema"}, api.Capabilities.Structured)
		require.Equal(t, src, api.Evidence["capabilities.structured"])
		require.Equal(t, map[string]string{CapFieldContext: "anthropic:/v1/models/{id}.max_input_tokens"}, caller.Sources)
		require.Nil(t, caller.Structured)
		again := Compose(ev)
		require.Equal(t, api, again, "composing the same evidence twice is stable")
	})
	t.Run("skipped off the anthropic stack", func(t *testing.T) {
		f := &AnthropicFacts{ThinkingSupported: true, ThinkingTypes: []string{"adaptive"}, Source: src}
		api, lines := ComposeWithTrace(Evidence{Stack: "litellm", ModelName: "claude-sonnet-4-5", GatewayReasoning: true, Anthropic: f})
		require.Equal(t, "reasoning_effort", api.Bindings[system.IntentReasoningDisable].Param, "litellm's own vocabulary")
		require.Contains(t, joined(lines), `anthropic model facts skipped — stack is "litellm"`)
	})
	t.Run("no facts: regex fallback unchanged", func(t *testing.T) {
		api := Compose(Evidence{Stack: "anthropic", ModelName: "claude-sonnet-4-5"})
		require.Equal(t, map[string]any{"type": "enabled", "budget_tokens": 1024}, api.Bindings[system.IntentReasoningEnable].Value)
		require.Nil(t, api.Evidence)
	})
}

func TestComposeRule7ThinkingDefault(t *testing.T) {
	t.Run("reported with source", func(t *testing.T) {
		api, lines := ComposeWithTrace(Evidence{Stack: "ollama", ModelName: "qwen3:32b", OllamaThinking: true, ThinkingDefault: "on", ThinkingDefaultSource: "ollama:/api/show.thinking.default"})
		require.Equal(t, "on", api.Thinking.Default)
		require.Equal(t, "ollama:/api/show.thinking.default", api.Evidence["thinking.default"])
		require.Contains(t, joined(lines), "thinking.default = on (ollama:/api/show.thinking.default)")
		raw, err := json.Marshal(api.Thinking)
		require.NoError(t, err)
		require.Equal(t, `{"mode":"controllable","default":"on"}`, string(raw))
	})
	t.Run("level name", func(t *testing.T) {
		api := Compose(Evidence{Stack: "openrouter", ModelName: "x/y", GatewayReasoning: true, ThinkingDefault: "medium", ThinkingDefaultSource: "openrouter:/api/v1/models.reasoning.default_effort"})
		require.Equal(t, "medium", api.Thinking.Default)
	})
	t.Run("no source: not published", func(t *testing.T) {
		api, lines := ComposeWithTrace(Evidence{Stack: "ollama", ModelName: "qwen3:32b", OllamaThinking: true, ThinkingDefault: "off"})
		require.Equal(t, "", api.Thinking.Default)
		require.Nil(t, api.Evidence)
		require.Contains(t, joined(lines), "thinking.default = off ignored — no source")
	})
	t.Run("junk value ignored", func(t *testing.T) {
		api, lines := ComposeWithTrace(Evidence{Stack: "ollama", ModelName: "qwen3:32b", OllamaThinking: true, ThinkingDefault: "on\n", ThinkingDefaultSource: "s"})
		require.Equal(t, "", api.Thinking.Default)
		require.Nil(t, api.Evidence)
		require.Contains(t, joined(lines), "thinking.default: reported value is not on|off|<level>; ignored")
		require.NotContains(t, joined(lines), "on\n")
	})
	t.Run("no thinking mode: ignored", func(t *testing.T) {
		api, lines := ComposeWithTrace(Evidence{Stack: "vllm", ModelName: "llama-3.3-70b", ThinkingDefault: "off", ThinkingDefaultSource: "s"})
		require.Nil(t, api.Thinking)
		require.Nil(t, api.Evidence)
		require.Contains(t, joined(lines), "thinking.default = off ignored — thinking mode undetermined")
	})
}

func withVeniceWebSearchBinding(t *testing.T) {
	t.Helper()
	require.Equal(t, "venice_parameters.enable_web_search", stackBindings["venice"][system.IntentWebSearchEnable].Param, "the gate is exercised against the documented venice row")
}

// The openrouter web_search.enable value is [{"id": "web"}]: a composed spec
// owns its copy at every level, so a consumer editing it (or Clone of it)
// never edits the package table (contract amendment c).
func TestComposeWebSearchBindingValueDoesNotAliasTheTable(t *testing.T) {
	table := stackBindings["openrouter"][system.IntentWebSearchEnable]
	require.Equal(t, []any{map[string]any{"id": "web"}}, table.Value, "the documented openrouter row")

	var f CapabilityFacts
	f.SetWebSearch(true, "openrouter:/models.plugins")
	api := Compose(Evidence{Stack: "openrouter", ModelName: "vendor/model", Caps: f})
	got := api.Bindings[system.IntentWebSearchEnable]
	require.Equal(t, table.Value, got.Value)
	got.Value.([]any)[0].(map[string]any)["id"] = "mutated"
	got.Clone().Value.([]any)[0].(map[string]any)["id"] = "mutated again"
	require.Equal(t, []any{map[string]any{"id": "web"}}, table.Value)

	again := Compose(Evidence{Stack: "openrouter", ModelName: "vendor/model", Caps: f})
	require.Equal(t, []any{map[string]any{"id": "web"}}, again.Bindings[system.IntentWebSearchEnable].Value)
}

func TestComposeRule8WebSearchGatesTableBindings(t *testing.T) {
	withVeniceWebSearchBinding(t)
	t.Run("reported true", func(t *testing.T) {
		var f CapabilityFacts
		f.SetWebSearch(true, "venice:/api/v1/models.model_spec.capabilities.supportsWebSearch")
		api := Compose(Evidence{Stack: "venice", ModelName: "llama-3.3-70b", Caps: f})
		require.Equal(t, "venice_parameters.enable_web_search", api.Bindings[system.IntentWebSearchEnable].Param)
		require.Equal(t, "venice:/api/v1/models.model_spec.capabilities.supportsWebSearch", api.Evidence["bindings.web_search.enable"])
		require.True(t, *api.Capabilities.WebSearch)
	})
	t.Run("declared true", func(t *testing.T) {
		api := Compose(Evidence{Stack: "venice", ModelName: "llama-3.3-70b", DeclaredCaps: &config.DeclaredCapabilities{WebSearch: bp(true)}})
		require.NotNil(t, api.Bindings[system.IntentWebSearchEnable])
		require.Equal(t, "declared", api.Evidence["bindings.web_search.enable"])
	})
	t.Run("reported false", func(t *testing.T) {
		var f CapabilityFacts
		f.SetWebSearch(false, "venice:x")
		api, lines := ComposeWithTrace(Evidence{Stack: "venice", ModelName: "llama-3.3-70b", Caps: f})
		require.Nil(t, api.Bindings[system.IntentWebSearchEnable])
		require.False(t, *api.Capabilities.WebSearch)
		require.Contains(t, joined(lines), "bindings: web_search.enable dropped — capabilities.webSearch is not reported true")
	})
	t.Run("unknown", func(t *testing.T) {
		api := Compose(Evidence{Stack: "venice", ModelName: "llama-3.3-70b"})
		require.Nil(t, api.Bindings[system.IntentWebSearchEnable])
		require.NotNil(t, api.Bindings[system.IntentSamplingTopK], "other table knobs are not gated")
	})
	t.Run("registry true from a lower layer still gates", func(t *testing.T) {
		reg := &RegistryFacts{Name: "litellm", Date: "2026-09-25", ModelCapabilities: system.ModelCapabilities{WebSearch: bp(true)}}
		api := Compose(Evidence{Stack: "venice", ModelName: "llama-3.3-70b", Registry: reg})
		require.Nil(t, api.Bindings[system.IntentWebSearchEnable], "the registry fills after the tables merge: a snapshot never enables a knob")
		require.True(t, *api.Capabilities.WebSearch)
	})
	t.Run("listing-provided web_search binding is not gated", func(t *testing.T) {
		registry := bindingSet{system.IntentWebSearchEnable: {Kind: bodyParam, Param: "plugins", ParamType: "array", Value: []any{map[string]any{"id": "web"}}}}
		api := Compose(Evidence{Stack: "openrouter", ModelName: "x/y", RegistryBindings: registry})
		require.Equal(t, "plugins", api.Bindings[system.IntentWebSearchEnable].Param)
	})
}

func TestComposeRule9EvidenceOnlyFields(t *testing.T) {
	var f CapabilityFacts
	f.SetTools(true, "llamacpp:/props.chat_template_caps.supports_tool_calls")
	f.SetToolsParallel(true, "llamacpp:/props.chat_template_caps.supports_parallel_tool_calls")
	f.SetInputModalities([]string{"text", "image", "audio"}, "llamacpp:/props.modalities")
	f.SetStructured([]string{"json_object", "json_schema", "grammar"}, "llamacpp:built-in")
	f.SetEmbeddings(false, "llamacpp:/v1/models.capabilities")
	f.SetPromptCache(true, "llamacpp:cache_prompt")
	f.SetMaxOutput(2048, "litellm:/model_group/info.max_output_tokens")
	f.SetQuantization("Q4_K_M", "llamacpp:/v1/models.meta.ftype")
	api, lines := ComposeWithTrace(Evidence{Stack: "llamacpp", ModelName: "llama-3.3-70b", Caps: f})
	c := api.Capabilities
	require.True(t, *c.Tools)
	require.True(t, *c.ToolsParallel)
	require.Equal(t, []string{"text", "image", "audio"}, c.Modalities.Input)
	require.Nil(t, c.Modalities.Output)
	require.Equal(t, []string{"json_object", "json_schema", "grammar"}, c.Structured)
	require.False(t, *c.Embeddings)
	require.True(t, *c.PromptCache)
	require.Equal(t, int64(2048), *c.MaxOutput)
	require.Equal(t, "Q4_K_M", c.Quantization)
	require.Len(t, api.Evidence, 8)
	tr := joined(lines)
	require.Contains(t, tr, "capabilities.quantization = Q4_K_M (llamacpp:/v1/models.meta.ftype)")
	require.Contains(t, tr, "capabilities.toolsParallel = true (llamacpp:/props.chat_template_caps.supports_parallel_tool_calls)")

	raw, err := json.Marshal(api.Capabilities)
	require.NoError(t, err)
	require.Equal(t, `{"maxOutput":2048,"modalities":{"input":["text","image","audio"]},"tools":true,"toolsParallel":true,"structured":["json_object","json_schema","grammar"],"embeddings":false,"promptCache":true,"quantization":"Q4_K_M"}`, string(raw))

	plain := Compose(Evidence{Stack: "llamacpp", ModelName: "llama-3.3-70b"})
	require.Nil(t, plain.Capabilities, "the stack tables never imply a capability")
}

func TestComposeRule10Registry(t *testing.T) {
	reg := &RegistryFacts{Name: "models.dev+litellm", Date: "2026-09-25", ModelCapabilities: system.ModelCapabilities{
		Context: i64(131072), MaxOutput: i64(16384), Tools: bp(true), Reasoning: bp(false), WebSearch: bp(false), PromptCache: bp(true),
		Modalities: &system.Modalities{Input: []string{"text", "image", "pdf"}, Output: []string{"text"}}, Structured: []string{"json_schema"}, Quantization: "bf16",
	}}
	t.Run("fills unset fields only", func(t *testing.T) {
		var f CapabilityFacts
		f.SetContext(8192, "ollama:/api/ps.context_length")
		f.SetTools(false, "ollama:/api/show.capabilities")
		api, lines := ComposeWithTrace(Evidence{Stack: "ollama", ModelName: "llama3.3", Caps: f, Registry: reg})
		c := api.Capabilities
		require.Equal(t, int64(8192), *c.Context)
		require.False(t, *c.Tools)
		require.Equal(t, int64(16384), *c.MaxOutput)
		require.Nil(t, c.Reasoning, "a registry false is not a backend report")
		require.NotContains(t, api.Evidence, "capabilities.reasoning")
		require.True(t, *c.PromptCache)
		require.Equal(t, []string{"text", "image", "pdf"}, c.Modalities.Input)
		require.Equal(t, "bf16", c.Quantization)
		require.Equal(t, "registry:models.dev+litellm@2026-09-25", api.Evidence["capabilities.maxOutput"])
		require.Equal(t, "ollama:/api/ps.context_length", api.Evidence["capabilities.context"])
		require.Equal(t, "registry:models.dev+litellm@2026-09-25", reg.Source())
		require.Contains(t, joined(lines), "context: served window 8192 is far below the registry's 131072")
	})
	t.Run("no gap line at or above a quarter", func(t *testing.T) {
		var f CapabilityFacts
		f.SetContext(32768, "ollama:/api/ps.context_length")
		_, lines := ComposeWithTrace(Evidence{Stack: "ollama", ModelName: "llama3.3", Caps: f, Registry: reg})
		require.NotContains(t, joined(lines), "far below")
	})
	t.Run("no gap line for a declared or registry context", func(t *testing.T) {
		_, lines := ComposeWithTrace(Evidence{Stack: "ollama", ModelName: "llama3.3", DeclaredCaps: &config.DeclaredCapabilities{Context: i64(2048)}, Registry: reg})
		require.NotContains(t, joined(lines), "far below")
		api, lines := ComposeWithTrace(Evidence{Stack: "ollama", ModelName: "llama3.3", Registry: reg})
		require.Equal(t, int64(131072), *api.Capabilities.Context)
		require.NotContains(t, joined(lines), "far below")
	})
	t.Run("out-of-bounds and unknown registry values are ignored", func(t *testing.T) {
		bad := &RegistryFacts{Name: "x", Date: "2026-01-01", ModelCapabilities: system.ModelCapabilities{Context: i64(0), MaxOutput: i64(1 << 40), Structured: []string{"xml"}, Modalities: &system.Modalities{Input: []string{"vision", "image"}}, Quantization: "bf 16"}}
		api, lines := ComposeWithTrace(Evidence{Stack: "openai", ModelName: "gpt-4o", Registry: bad})
		require.Nil(t, api.Capabilities.Context)
		require.Nil(t, api.Capabilities.MaxOutput)
		require.Nil(t, api.Capabilities.Structured)
		require.Equal(t, []string{"text", "image"}, api.Capabilities.Modalities.Input)
		require.Equal(t, "", api.Capabilities.Quantization)
		tr := joined(lines)
		require.Contains(t, tr, "capabilities.context: registry value 0 out of (0, 2147483648); ignored")
		require.Contains(t, tr, "capabilities.maxOutput: registry value 1099511627776 out of (0, 2147483648); ignored")
		require.Contains(t, tr, "capabilities.structured: registry value [xml] has no known entry; ignored")
		require.Contains(t, tr, "capabilities.modalities.input: registry value [vision image] narrowed to the known entries [image]")
		require.Contains(t, tr, "capabilities.quantization: registry value is not a plain label; ignored")
	})
	t.Run("nil registry", func(t *testing.T) {
		var r *RegistryFacts
		require.Equal(t, "", r.Source())
		require.Equal(t, "registry:models.dev@2026-09-25", RegistrySource("models.dev", "2026-09-25"))
	})
}

func TestComposeRule12ClearDetectedCapabilities(t *testing.T) {
	var f CapabilityFacts
	f.SetContext(8192, "ollama:/api/ps.context_length")
	f.SetInputModalities([]string{"text", "image"}, "ollama:/api/show.capabilities")
	f.SetQuantization("Q4_K_M", "ollama:/api/show.details.quantization_level")
	reg := &RegistryFacts{Name: "models.dev", Date: "2026-09-25", ModelCapabilities: system.ModelCapabilities{MaxOutput: i64(4096), Modalities: &system.Modalities{Output: []string{"text"}}}}
	api := Compose(Evidence{Stack: "ollama", ModelName: "qwen3:32b", OllamaThinking: true, Caps: f, Registry: reg, DeclaredCaps: &config.DeclaredCapabilities{Tools: bp(true)}, ThinkingDefault: "on", ThinkingDefaultSource: "ollama:/api/show.thinking.default", ReasoningEfforts: []string{"none", "low"}, EffortsSource: "ollama:/api/show.thinking.values"})
	require.Equal(t, "on", api.Thinking.Default)
	require.Contains(t, api.Evidence, "bindings.reasoning.effort")

	// What apidetect does on a transport mismatch, minus the stack drop itself.
	api.Bindings = nil
	ClearDetectedCapabilities(api, nil)
	c := api.Capabilities
	require.NotNil(t, c)
	require.Nil(t, c.Context, "from the dropped stack's endpoint")
	require.Equal(t, "", c.Quantization)
	require.Nil(t, c.Modalities.Input)
	require.Equal(t, []string{"text"}, c.Modalities.Output, "registry fact kept")
	require.Equal(t, int64(4096), *c.MaxOutput, "registry fact kept")
	require.True(t, *c.Tools, "declared fact kept")
	require.Nil(t, c.Reasoning, "derived from the dropped bindings")
	require.Equal(t, "", api.Thinking.Default)
	require.Equal(t, map[string]string{
		"capabilities.tools":             "declared",
		"capabilities.maxOutput":         "registry:models.dev@2026-09-25",
		"capabilities.modalities.output": "registry:models.dev@2026-09-25",
	}, api.Evidence)

	t.Run("everything detected: block and map go", func(t *testing.T) {
		var f CapabilityFacts
		f.SetContext(8192, "vllm:/v1/models.max_model_len")
		f.SetInputModalities([]string{"text"}, "vllm:x")
		api := Compose(Evidence{Stack: "vllm", ModelName: "qwen3-32b", Caps: f})
		ClearDetectedCapabilities(api, nil)
		require.Nil(t, api.Capabilities)
		require.Nil(t, api.Evidence)
	})
	t.Run("custom keep", func(t *testing.T) {
		var f CapabilityFacts
		f.SetContext(8192, "vllm:/v1/models.max_model_len")
		f.SetTools(true, "verify:tools@1790000000")
		api := Compose(Evidence{Stack: "vllm", ModelName: "llama-3.3-70b", Caps: f})
		ClearDetectedCapabilities(api, func(source string) bool { return strings.HasPrefix(source, "verify:") })
		require.Nil(t, api.Capabilities.Context)
		require.True(t, *api.Capabilities.Tools)
		require.Equal(t, map[string]string{"capabilities.tools": "verify:tools@1790000000"}, api.Evidence)
	})
	t.Run("nil and evidence-less specs", func(t *testing.T) {
		ClearDetectedCapabilities(nil, nil)
		api := &system.ModelApiSpec{Stack: "vllm", Capabilities: &system.ModelCapabilities{Tools: bp(true)}}
		ClearDetectedCapabilities(api, nil)
		require.True(t, *api.Capabilities.Tools, "no evidence: nothing to judge, nothing touched")
	})
	require.True(t, IsDeclaredOrRegistrySource("declared"))
	require.True(t, IsDeclaredOrRegistrySource("registry:models.dev@2026-09-25"))
	require.False(t, IsDeclaredOrRegistrySource("vllm:/v1/models.max_model_len"))
	require.False(t, IsDeclaredOrRegistrySource("verify:tools@1"))
}

func TestBuildAppliesDeclaredCapabilities(t *testing.T) {
	caps := &config.DeclaredCapabilities{Context: i64(32768), MaxOutput: i64(4096), Modalities: &config.DeclaredModalities{Input: []string{"text", "image"}}, Tools: bp(true), Structured: []string{"json_schema"}, WebSearch: bp(false), Embeddings: bp(false)}
	t.Run("with apiStack", func(t *testing.T) {
		api := Build(config.ModelConfig{ModelName: "qwen3-235b-a22b", ApiType: "openai", ApiStack: "vllm", ModelFamily: "qwen3", ApiURL: "http://h/v1/chat/completions", Capabilities: caps})
		require.Equal(t, system.ApiSpecSourceDeclared, api.Source)
		c := api.Capabilities
		require.Equal(t, int64(32768), *c.Context)
		require.Equal(t, int64(4096), *c.MaxOutput)
		require.Equal(t, []string{"text", "image"}, c.Modalities.Input)
		require.True(t, *c.Tools)
		require.Equal(t, []string{"json_schema"}, c.Structured)
		require.False(t, *c.WebSearch)
		require.False(t, *c.Embeddings)
		require.True(t, *c.Reasoning, "qwen3 on vllm has reasoning bindings")
		for _, key := range []string{"context", "maxOutput", "modalities.input", "tools", "structured", "webSearch", "embeddings"} {
			require.Equal(t, "declared", api.Evidence["capabilities."+key], key)
		}
		require.Equal(t, SourceDerivedBindings, api.Evidence["capabilities.reasoning"])
		raw, err := json.Marshal(api.Capabilities)
		require.NoError(t, err)
		require.Equal(t, `{"context":32768,"maxOutput":4096,"modalities":{"input":["text","image"]},"tools":true,"structured":["json_schema"],"reasoning":true,"webSearch":false,"embeddings":false}`, string(raw))
	})
	t.Run("without apiStack: a capabilities-only block", func(t *testing.T) {
		api := Build(config.ModelConfig{ModelName: "my-private-alias", ApiType: "openai", ApiURL: "http://h/v1", Capabilities: &config.DeclaredCapabilities{Context: i64(8192)}})
		require.NotNil(t, api)
		require.Equal(t, "", api.Stack)
		require.Equal(t, system.ApiSpecSourceDeclared, api.Source)
		require.Nil(t, api.Bindings)
		require.Nil(t, api.Parameters)
		require.Equal(t, int64(8192), *api.Capabilities.Context)
		require.Equal(t, map[string]string{"capabilities.context": "declared"}, api.Evidence)
	})
	t.Run("nothing declared: nil as before", func(t *testing.T) {
		require.Nil(t, Build(config.ModelConfig{ModelName: "my-private-alias", ApiType: "openai", ApiURL: "http://h/v1"}))
	})
	t.Run("result does not alias the config", func(t *testing.T) {
		api := Build(config.ModelConfig{ModelName: "qwen3-32b", ApiType: "openai", ApiStack: "vllm", ApiURL: "http://h/v1", Capabilities: caps})
		*api.Capabilities.Context = 1
		api.Capabilities.Modalities.Input[0] = "mutated"
		api.Capabilities.Structured[0] = "mutated"
		require.Equal(t, int64(32768), *caps.Context)
		require.Equal(t, []string{"text", "image"}, caps.Modalities.Input)
		require.Equal(t, []string{"json_schema"}, caps.Structured)
	})
}

func TestComposeResultDoesNotAliasEvidence(t *testing.T) {
	var f CapabilityFacts
	f.SetContext(8192, "vllm:/v1/models.max_model_len")
	f.SetInputModalities([]string{"text"}, "vllm:x")
	f.SetStructured([]string{"json_schema"}, "vllm:y")
	sources := map[string]string{}
	for k, v := range f.Sources {
		sources[k] = v
	}
	api := Compose(Evidence{Stack: "vllm", ModelName: "llama-3.3-70b", Caps: f})
	*api.Capabilities.Context = 1
	api.Capabilities.Modalities.Input[0] = "mutated"
	api.Capabilities.Structured[0] = "mutated"
	require.Equal(t, int64(8192), *f.Context)
	require.Equal(t, []string{"text"}, f.Modalities.Input)
	require.Equal(t, []string{"json_schema"}, f.Structured)
	require.Equal(t, sources, f.Sources, "the caller's sources map is not written to")
}

func TestComposeUndeterminedStackStillCarriesFacts(t *testing.T) {
	var f CapabilityFacts
	f.SetContext(8192, "litellm:/model_group/info.max_input_tokens")
	api := Compose(Evidence{ModelName: "unknown-model-x", Caps: f})
	require.NotNil(t, api, "a capability fact is something known")
	require.Equal(t, "", api.Stack)
	require.Equal(t, int64(8192), *api.Capabilities.Context)
	require.Nil(t, Compose(Evidence{ModelName: "unknown-model-x"}))
}

// Amendment b: a chat model accepts text, so an input list a producer built
// from its extra-modality flags gains "text" in front under the producer's
// evidence; an output list is left as reported.
func TestComposeImpliesTextInput(t *testing.T) {
	t.Run("lmstudio vision", func(t *testing.T) {
		var f CapabilityFacts
		require.True(t, f.SetInputModalities([]string{"image"}, "lmstudio:/api/v1/models.capabilities.vision"))
		api, lines := ComposeWithTrace(Evidence{Stack: "lmstudio", ModelName: "gemma-4-26b", Caps: f})
		require.Equal(t, []string{"text", "image"}, api.Capabilities.Modalities.Input)
		require.Equal(t, "lmstudio:/api/v1/models.capabilities.vision", api.Evidence["capabilities.modalities.input"])
		require.Contains(t, lines, "capabilities.modalities.input: text prepended — a chat model accepts text (lmstudio:/api/v1/models.capabilities.vision)")
	})
	t.Run("text already listed", func(t *testing.T) {
		var f CapabilityFacts
		require.True(t, f.SetInputModalities([]string{"image", "text"}, "x:/m"))
		api, lines := ComposeWithTrace(Evidence{Stack: "vllm", ModelName: "m", Caps: f})
		require.Equal(t, []string{"image", "text"}, api.Capabilities.Modalities.Input, "order kept, nothing added")
		for _, l := range lines {
			require.NotContains(t, l, "text prepended")
		}
	})
	t.Run("declared input", func(t *testing.T) {
		declared := &config.DeclaredCapabilities{Modalities: &config.DeclaredModalities{Input: []string{"pdf"}, Output: []string{"image"}}}
		api, _ := ComposeWithTrace(Evidence{Stack: "openai", ModelName: "m", Declared: true, DeclaredCaps: declared})
		require.Equal(t, []string{"text", "pdf"}, api.Capabilities.Modalities.Input)
		require.Equal(t, []string{"image"}, api.Capabilities.Modalities.Output, "output stays as reported")
		require.Equal(t, SourceDeclared, api.Evidence["capabilities.modalities.input"])
	})
	t.Run("output alone adds no input", func(t *testing.T) {
		var f CapabilityFacts
		require.True(t, f.SetOutputModalities([]string{"image"}, "x:/m"))
		api, _ := ComposeWithTrace(Evidence{Stack: "vllm", ModelName: "m", Caps: f})
		require.Nil(t, api.Capabilities.Modalities.Input)
		require.Equal(t, []string{"image"}, api.Capabilities.Modalities.Output)
	})
}
