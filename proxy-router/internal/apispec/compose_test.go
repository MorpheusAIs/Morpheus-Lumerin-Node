package apispec

import (
	"fmt"
	"strings"
	"testing"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/config"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/system"
	"github.com/stretchr/testify/require"
)

const qwen3Template = `{%- if messages %}{% generation %}{% endgeneration %}{%- endif %}
{%- if enable_thinking is defined and enable_thinking is false %}<think></think>{%- endif %}`

const deepseekV31Template = `{% if not thinking is defined %}{% set thinking = false %}{% endif %}
{% if thinking %}<think>{% endif %}`

// R2: Build is exactly Compose over the declared evidence, for every preset.
func TestBuildIsComposeOverDeclaredEvidence(t *testing.T) {
	names := []string{"Qwen/Qwen3-235B-A22B", "deepseek-r1", "gpt-oss:120b", "claude-opus-4-6", "o4-mini", "meta-llama/Llama-3.3-70B-Instruct", "seed-oss:36b", "MiniMaxAI/MiniMax-M3", "some-embedder"}
	for stack, apiType := range config.StackTransport {
		for _, name := range names {
			for _, family := range []string{"", "deepseek-v3.1", " Llama "} {
				cfg := config.ModelConfig{ModelName: name, ApiType: apiType, ApiStack: stack, ModelFamily: family, ApiURL: "http://h/v1/chat/completions"}
				want := Compose(Evidence{Stack: stack, ModelName: name, ModelFamily: family, Declared: true})
				require.Equal(t, want, Build(cfg), "%s/%s/%q", stack, name, family)
			}
		}
	}
	require.Nil(t, Build(config.ModelConfig{ModelName: "qwen3-32b", ApiType: "openai", ApiStack: "bogus", ApiURL: "http://h/v1"}))
}

func TestComposeRegistryBindingsWinOverEverything(t *testing.T) {
	registry := bindingSet{system.IntentReasoningDisable: {Kind: system.BindingKindBodyParam, Param: "registry.off", ParamType: "boolean", Value: true}}
	api := Compose(Evidence{Stack: "vllm", ModelName: "qwen3-32b", ChatTemplate: qwen3Template, OllamaThinking: true, RegistryBindings: registry})
	require.Equal(t, "registry.off", api.Bindings[system.IntentReasoningDisable].Param)
	require.Equal(t, system.ThinkingModeControllable, api.Thinking.Mode)
	// the stack table still fills the other intents; the input map is not aliased
	require.Equal(t, "top_k", api.Bindings[system.IntentSamplingTopK].Param)
	api.Bindings[system.IntentReasoningDisable].Param = "mutated"
	require.Equal(t, "registry.off", registry[system.IntentReasoningDisable].Param)
}

func TestComposeOllamaCapabilityBeatsTemplate(t *testing.T) {
	api := Compose(Evidence{Stack: "ollama", ModelName: "qwen3:32b", ChatTemplate: qwen3Template, OllamaThinking: true})
	disable := api.Bindings[system.IntentReasoningDisable]
	require.Equal(t, system.BindingKindBodyParam, disable.Kind)
	require.Equal(t, "reasoning_effort", disable.Param)
	require.Equal(t, "none", disable.Value)
}

func TestComposeTemplateBeatsFamilyDefault(t *testing.T) {
	// the template says `thinking`; the qwen3 family default would say enable_thinking
	api := Compose(Evidence{Stack: "llamacpp", ModelName: "qwen3-32b", ChatTemplate: deepseekV31Template})
	require.Equal(t, "chat_template_kwargs.thinking", api.Bindings[system.IntentReasoningDisable].Param)
}

func TestComposeFamilyDefaultWhenNoEvidence(t *testing.T) {
	api := Compose(Evidence{Stack: "vllm", ModelName: "qwen3-32b"})
	require.Equal(t, "chat_template_kwargs.enable_thinking", api.Bindings[system.IntentReasoningDisable].Param)
}

// R4: an undetermined stack never carries bindings — at most the family and
// an always_on mode; nothing known at all composes to nil.
func TestComposeUndeterminedStackHasNoBindings(t *testing.T) {
	api := Compose(Evidence{ModelName: "qwen3-32b"})
	require.NotNil(t, api)
	require.Equal(t, "", api.Stack)
	require.Equal(t, "qwen3", api.ModelFamily)
	require.Nil(t, api.Thinking)
	require.Empty(t, api.Bindings)
	require.Empty(t, api.Parameters)

	r1 := Compose(Evidence{ModelName: "deepseek-r1-distill"})
	require.Equal(t, system.ThinkingModeAlwaysOn, r1.Thinking.Mode)
	require.Empty(t, r1.Bindings)

	require.Nil(t, Compose(Evidence{ModelName: "unknown-model-x"}))
	require.Nil(t, Compose(Evidence{}))
}

// R1: an explicit modelFamily is never overridden by backend evidence.
func TestComposeExplicitFamilyIsNeverOverridden(t *testing.T) {
	api := Compose(Evidence{Stack: "vllm", ModelName: "alias", ModelFamily: "llama", Architecture: "qwen3", ServedModelID: "deepseek-ai/DeepSeek-R1"})
	require.Equal(t, "llama", api.ModelFamily)
	require.Nil(t, api.Thinking)
}

func TestComposeFamilyEvidenceChain(t *testing.T) {
	// architecture beats served id beats name, unless the later one refines
	api := Compose(Evidence{Stack: "lmstudio", ModelName: "deepseek-r1-distill", Architecture: "deepseek2"})
	require.Equal(t, "deepseek-r1", api.ModelFamily)
	api = Compose(Evidence{Stack: "sglang", ModelName: "whatever-local-alias", ServedModelID: "deepseek-ai/DeepSeek-R1"})
	require.Equal(t, "deepseek-r1", api.ModelFamily)
	api = Compose(Evidence{Stack: "lmstudio", ModelName: "some-embedder", Architecture: "bert"})
	require.Equal(t, "", api.ModelFamily)
	// the served id, not the alias, drives the -thinking refinement
	api = Compose(Evidence{Stack: "vllm", ModelName: "alias", ServedModelID: "Qwen/Qwen3-235B-A22B-Thinking-2507"})
	require.Equal(t, system.ThinkingModeAlwaysOn, api.Thinking.Mode)
}

// A declared gateway preset merges its reasoning knobs on family knowledge
// alone (today's Build); a detected gateway needs probe evidence.
func TestComposeDetectedGatewayNeedsReasoningEvidence(t *testing.T) {
	detected := Compose(Evidence{Stack: "litellm", ModelName: "qwen3-32b"})
	require.Nil(t, detected.Thinking)
	require.Nil(t, detected.Bindings[system.IntentReasoningDisable])
	require.Equal(t, "response_format", detected.Bindings[system.IntentResponseFormatJSON].Param)

	declared := Compose(Evidence{Stack: "litellm", ModelName: "qwen3-32b", Declared: true})
	require.Equal(t, system.ThinkingModeControllable, declared.Thinking.Mode)
	require.Equal(t, "reasoning_effort", declared.Bindings[system.IntentReasoningDisable].Param)

	evidenced := Compose(Evidence{Stack: "litellm", ModelName: "qwen3-32b", GatewayReasoning: true})
	require.Equal(t, system.ThinkingModeControllable, evidenced.Thinking.Mode)

	// always-on families need no evidence on either path
	r1 := Compose(Evidence{Stack: "litellm", ModelName: "deepseek-r1"})
	require.Equal(t, system.ThinkingModeAlwaysOn, r1.Thinking.Mode)
	require.Nil(t, r1.Bindings[system.IntentReasoningEffort])
}

func TestComposeHostedVendorGatesFamilyDefaults(t *testing.T) {
	groq := Compose(Evidence{Stack: "groq", ModelName: "qwen/qwen3-32b"})
	require.Equal(t, "groq", groq.Stack)
	require.Equal(t, "qwen3", groq.ModelFamily)
	require.Empty(t, groq.Bindings)
	require.Nil(t, groq.Thinking)
	require.Empty(t, groq.Parameters)

	together := Compose(Evidence{Stack: "together", ModelName: "claude-sonnet-4-5"})
	require.Empty(t, together.Bindings)

	// the family's native vendor keeps its defaults
	anthropic := Compose(Evidence{Stack: "anthropic", ModelName: "claude-sonnet-4-5"})
	require.Equal(t, "thinking", anthropic.Bindings[system.IntentReasoningDisable].Param)
}

// Detected-only engines (tgi, lmstudio, koboldcpp) have no preset table and
// nothing documents that they honour chat-template kwargs: the family's
// template-kwarg defaults are gated for them like for a gateway. The family
// and an always_on mode are still reported.
func TestComposeDetectedOnlyEnginesGateFamilyDefaults(t *testing.T) {
	for _, stack := range []string{"tgi", "lmstudio", "koboldcpp"} {
		api, lines := ComposeWithTrace(Evidence{Stack: stack, ModelName: "qwen3-32b"})
		require.NotNil(t, api, stack)
		require.Equal(t, stack, api.Stack)
		require.Equal(t, "qwen3", api.ModelFamily, stack)
		require.Empty(t, api.Bindings, stack)
		require.Nil(t, api.Thinking, stack)
		require.Empty(t, api.Parameters, stack)
		require.Contains(t, strings.Join(lines, "\n"), `family default for "qwen3" skipped — "`+stack+`" does not accept it`)

		r1 := Compose(Evidence{Stack: stack, ModelName: "deepseek-r1-distill"})
		require.Equal(t, system.ThinkingModeAlwaysOn, r1.Thinking.Mode, stack)
		require.Empty(t, r1.Bindings, stack)
	}
	// the preset engines keep the family default
	vllm := Compose(Evidence{Stack: "vllm", ModelName: "qwen3-32b"})
	require.Equal(t, "chat_template_kwargs.enable_thinking", vllm.Bindings[system.IntentReasoningDisable].Param)
}

func TestComposeImportedParametersReplaceStackList(t *testing.T) {
	api := Compose(Evidence{Stack: "openrouter", ModelName: "x/y", GatewayReasoning: true, Parameters: []string{"tools", "reasoning"}})
	require.Equal(t, []string{"reasoning", "tools"}, api.Parameters)
	require.Equal(t, "reasoning.effort", api.Bindings[system.IntentReasoningDisable].Param)
	require.Equal(t, "top_a", api.Bindings[system.IntentSamplingTopA].Param)
}

// R5 addendum: LiteLLM's supported_reasoning_efforts narrow the litellm
// reasoning.effort enum; without "none" the reasoning.disable knob goes.
func TestComposeLiteLLMSupportedReasoningEfforts(t *testing.T) {
	api := Compose(Evidence{Stack: "litellm", ModelName: "claude-sonnet-4-5", GatewayReasoning: true, ReasoningEfforts: []string{"low", "medium", "high"}})
	require.Equal(t, []string{"low", "medium", "high"}, api.Bindings[system.IntentReasoningEffort].EnumValues)
	require.Nil(t, api.Bindings[system.IntentReasoningDisable])
	require.NotNil(t, api.Bindings[system.IntentReasoningEnable])

	withNone := Compose(Evidence{Stack: "litellm", ModelName: "claude-sonnet-4-5", GatewayReasoning: true, ReasoningEfforts: []string{"none", "low", "high"}})
	require.Equal(t, []string{"none", "low", "high"}, withNone.Bindings[system.IntentReasoningEffort].EnumValues)
	require.Equal(t, "none", withNone.Bindings[system.IntentReasoningDisable].Value)

	// supports_reasoning=false: efforts alone add nothing
	off := Compose(Evidence{Stack: "litellm", ModelName: "llama-3.3-70b", ReasoningEfforts: []string{"low"}})
	require.Nil(t, off.Thinking)
	require.Nil(t, off.Bindings[system.IntentReasoningEffort])

	// the shared table is untouched
	require.Equal(t, []string{"none", "minimal", "low", "medium", "high", "xhigh", "max"}, stackBindings["litellm"][system.IntentReasoningEffort].EnumValues)
}

// OQ1 ruling: the litellm reasoning.enable knob follows
// supported_reasoning_efforts too — its value becomes the first listed effort
// other than none (LiteLLM's order), and when the list has no such effort the
// knob is dropped. The thinking mode is derived after the fixup.
func TestComposeLiteLLMReasoningEnableFollowsSupportedEfforts(t *testing.T) {
	api := Compose(Evidence{Stack: "litellm", ModelName: "claude-sonnet-4-5", GatewayReasoning: true, ReasoningEfforts: []string{"low", "medium", "high"}})
	require.Equal(t, "low", api.Bindings[system.IntentReasoningEnable].Value)
	require.Equal(t, "reasoning_effort", api.Bindings[system.IntentReasoningEnable].Param)
	require.Equal(t, system.ThinkingModeControllable, api.Thinking.Mode)

	// none is skipped; the next entry in LiteLLM's order is taken
	withNone := Compose(Evidence{Stack: "litellm", ModelName: "claude-sonnet-4-5", GatewayReasoning: true, ReasoningEfforts: []string{"none", "high", "max"}})
	require.Equal(t, "high", withNone.Bindings[system.IntentReasoningEnable].Value)
	require.Equal(t, "none", withNone.Bindings[system.IntentReasoningDisable].Value)
	require.Equal(t, system.ThinkingModeControllable, withNone.Thinking.Mode)

	// nothing but none: thinking cannot be switched on through LiteLLM
	onlyNone := Compose(Evidence{Stack: "litellm", ModelName: "claude-sonnet-4-5", GatewayReasoning: true, ReasoningEfforts: []string{"none"}})
	require.Nil(t, onlyNone.Bindings[system.IntentReasoningEnable])
	require.Equal(t, "none", onlyNone.Bindings[system.IntentReasoningDisable].Value)
	require.Equal(t, []string{"none"}, onlyNone.Bindings[system.IntentReasoningEffort].EnumValues)
	require.Equal(t, system.ThinkingModeControllable, onlyNone.Thinking.Mode)

	// without the efforts list the table value stands, and the table itself
	// is never touched
	plain := Compose(Evidence{Stack: "litellm", ModelName: "claude-sonnet-4-5", GatewayReasoning: true})
	require.Equal(t, "medium", plain.Bindings[system.IntentReasoningEnable].Value)
	require.Equal(t, "medium", stackBindings["litellm"][system.IntentReasoningEnable].Value)

	_, lines := ComposeWithTrace(Evidence{Stack: "litellm", ModelName: "claude-sonnet-4-5", GatewayReasoning: true, ReasoningEfforts: []string{"none"}})
	require.Contains(t, strings.Join(lines, "\n"), "reasoning.enable dropped")
}

func TestComposeWithTraceExplainsDecisions(t *testing.T) {
	api, lines := ComposeWithTrace(Evidence{Stack: "llamacpp", ModelName: "qwen3-32b", ChatTemplate: qwen3Template})
	require.NotNil(t, api)
	joined := strings.Join(lines, "\n")
	require.Contains(t, joined, `family "qwen3": from configured model name`)
	require.Contains(t, joined, "chat template")
	require.Contains(t, joined, "enable_thinking")
	require.Contains(t, joined, "llamacpp stack table")

	_, lines = ComposeWithTrace(Evidence{Stack: "groq", ModelName: "qwen/qwen3-32b"})
	require.Contains(t, strings.Join(lines, "\n"), `family default for "qwen3" skipped`)

	_, lines = ComposeWithTrace(Evidence{ModelName: "qwen3-32b"})
	require.Contains(t, strings.Join(lines, "\n"), "stack undetermined")
}

// R3: the composer stamps where the block came from.
func TestComposeSetsSource(t *testing.T) {
	require.Equal(t, system.ApiSpecSourceDeclared, Build(config.ModelConfig{ModelName: "qwen3-32b", ApiType: "openai", ApiStack: "vllm", ApiURL: "http://h/v1"}).Source)
	require.Equal(t, system.ApiSpecSourceDetected, Compose(Evidence{Stack: "vllm", ModelName: "qwen3-32b"}).Source)
	require.Equal(t, system.ApiSpecSourceDetected, Compose(Evidence{ModelName: "qwen3-32b"}).Source, "a family-only detected spec still says detected")
}

// Rider (post-Task-1 review, closes an R4 gap): chat-template evidence must
// not bypass the "undetermined stack" guard the family-default step already
// enforces — an unknown wire vocabulary means no template-kwarg bindings
// either, not just no family default.
func TestComposeUndeterminedStackSkipsTemplateEvidence(t *testing.T) {
	api := Compose(Evidence{Stack: "", ModelName: "Qwen3-8B", ChatTemplate: qwen3Template})
	require.NotNil(t, api)
	require.Equal(t, "", api.Stack)
	require.Equal(t, "qwen3", api.ModelFamily, "family may still be reported")
	require.Nil(t, api.Thinking, "qwen3 is not an always-on family")
	require.Empty(t, api.Bindings)

	_, lines := ComposeWithTrace(Evidence{Stack: "", ModelName: "Qwen3-8B", ChatTemplate: qwen3Template})
	require.Contains(t, strings.Join(lines, "\n"), "chat template evidence skipped — stack undetermined")
}

func TestDescribeBinding(t *testing.T) {
	require.Equal(t, "unknown", DescribeBinding(nil))
	s := DescribeBinding(&system.ParamBinding{Kind: system.BindingKindBodyParam, Param: "thinking", ParamType: "object", Value: map[string]any{"type": "disabled"}, Hint: "h"})
	require.Equal(t, `body_param thinking (object, value={"type":"disabled"}) — h`, s)
	s = DescribeBinding(&system.ParamBinding{Kind: system.BindingKindTemplateKwarg, Param: "chat_template_kwargs.reasoning_effort", ParamType: "enum", EnumValues: []string{"low", "high"}})
	require.Equal(t, "template_kwarg chat_template_kwargs.reasoning_effort (enum: low|high)", s)
}

// Backend-reported lists (LiteLLM supported_openai_params and
// supported_reasoning_efforts, registry supported_parameters, Venice
// capabilities) are third-party strings that reach the public wire through
// the spec. They are sanitized once, here: only names matching
// ^[A-Za-z0-9_.-]{1,64}$ survive, duplicates go, order is preserved, and the
// lists are capped (64 parameters, 16 efforts); a drop is traced once with
// the counts.
func TestComposeSanitizesBackendReportedLists(t *testing.T) {
	t.Run("10 000 junk parameters plus a few valid ones", func(t *testing.T) {
		in := make([]string, 0, 10_008)
		for i := 0; i < 10_000; i++ {
			in = append(in, fmt.Sprintf("junk %d\n", i)) // space and newline: never a name
		}
		in = append(in, "", "temperature", strings.Repeat("a", 65), "tools", "tools", "top_p\x00", "reasoning_effort", "ünïcode")
		snapshot := append([]string(nil), in...)

		api, lines := ComposeWithTrace(Evidence{Stack: "litellm", ModelName: "qwen3-32b", Parameters: in})
		require.NotNil(t, api)
		require.Equal(t, []string{"reasoning_effort", "temperature", "tools"}, api.Parameters, "exactly the valid, deduped names")
		require.Contains(t, strings.Join(lines, "\n"), "parameters: 10005 of 10008 backend-reported entries dropped (invalid name, duplicate or beyond the 64 cap); 3 kept")
		require.Equal(t, snapshot, in, "the caller's list is never modified")
	})

	t.Run("cap of 64 parameters, first ones kept", func(t *testing.T) {
		in := make([]string, 0, 100)
		for i := 0; i < 100; i++ {
			in = append(in, fmt.Sprintf("p%03d", i))
		}
		api, lines := ComposeWithTrace(Evidence{Stack: "openrouter", ModelName: "x/y", Parameters: in})
		require.Len(t, api.Parameters, 64)
		require.Equal(t, "p000", api.Parameters[0])
		require.Equal(t, "p063", api.Parameters[63])
		require.Contains(t, strings.Join(lines, "\n"), "parameters: 36 of 100 backend-reported entries dropped")
	})

	t.Run("dots and dashes are names; nothing dropped means no trace line", func(t *testing.T) {
		api, lines := ComposeWithTrace(Evidence{Stack: "openrouter", ModelName: "x/y", Parameters: []string{"reasoning.effort", "top-k", "min_p", "seed"}})
		require.Equal(t, []string{"min_p", "reasoning.effort", "seed", "top-k"}, api.Parameters)
		require.NotContains(t, strings.Join(lines, "\n"), "backend-reported entries dropped")
	})

	t.Run("efforts: a none with a newline is not none", func(t *testing.T) {
		in := []string{"low", "none\n", "high", "low"}
		api, lines := ComposeWithTrace(Evidence{Stack: "litellm", ModelName: "claude-sonnet-4-5", GatewayReasoning: true, ReasoningEfforts: in})
		require.Equal(t, []string{"low", "high"}, api.Bindings[system.IntentReasoningEffort].EnumValues, "order preserved, junk and duplicate dropped")
		require.Nil(t, api.Bindings[system.IntentReasoningDisable], "no valid none: thinking cannot be switched off")
		require.Equal(t, "low", api.Bindings[system.IntentReasoningEnable].Value)
		require.Contains(t, strings.Join(lines, "\n"), "reasoning efforts: 2 of 4 backend-reported entries dropped (invalid name, duplicate or beyond the 16 cap); 2 kept")
		require.Equal(t, []string{"low", "none\n", "high", "low"}, in)
	})

	t.Run("cap of 16 efforts", func(t *testing.T) {
		in := make([]string, 0, 20)
		for i := 0; i < 20; i++ {
			in = append(in, fmt.Sprintf("e%02d", i))
		}
		api, lines := ComposeWithTrace(Evidence{Stack: "litellm", ModelName: "claude-sonnet-4-5", GatewayReasoning: true, ReasoningEfforts: in})
		require.Len(t, api.Bindings[system.IntentReasoningEffort].EnumValues, 16)
		require.Equal(t, "e00", api.Bindings[system.IntentReasoningEnable].Value)
		require.Contains(t, strings.Join(lines, "\n"), "reasoning efforts: 4 of 20 backend-reported entries dropped")
	})

	t.Run("a list that is all junk is an empty list", func(t *testing.T) {
		api := Compose(Evidence{Stack: "litellm", ModelName: "qwen3-32b", GatewayReasoning: true, Parameters: []string{"", " ", "a b"}, ReasoningEfforts: []string{"\t"}})
		require.Contains(t, api.Parameters, "messages", "the stack's documented list stands in for an empty import")
		require.Equal(t, stackBindings["litellm"][system.IntentReasoningEffort].EnumValues, api.Bindings[system.IntentReasoningEffort].EnumValues, "no valid efforts: the table enum stands")
	})
}
