package apispec

import (
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

func TestDescribeBinding(t *testing.T) {
	require.Equal(t, "unknown", DescribeBinding(nil))
	s := DescribeBinding(&system.ParamBinding{Kind: system.BindingKindBodyParam, Param: "thinking", ParamType: "object", Value: map[string]any{"type": "disabled"}, Hint: "h"})
	require.Equal(t, `body_param thinking (object, value={"type":"disabled"}) — h`, s)
	s = DescribeBinding(&system.ParamBinding{Kind: system.BindingKindTemplateKwarg, Param: "chat_template_kwargs.reasoning_effort", ParamType: "enum", EnumValues: []string{"low", "high"}})
	require.Equal(t, "template_kwarg chat_template_kwargs.reasoning_effort (enum: low|high)", s)
}
