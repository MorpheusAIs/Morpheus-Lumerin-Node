package apispec

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/system"
)

// familyRule maps a substring (or regexp for the basename) of the lowercased
// model name to a canonical family. Order matters: first match wins.
type familyRule struct {
	substr string
	family string
}

var familyRules = []familyRule{
	{"qwq", "qwq"},
	{"qwen3", "qwen3"},
	{"qwen2", "qwen2.5"},
	{"deepseek-r1", "deepseek-r1"},
	{"deepseek_r1", "deepseek-r1"},
	{"deepseek-reasoner", "deepseek-r1"},
	{"deepseek-v3.1", "deepseek-v3.1"},
	{"deepseek-v3-1", "deepseek-v3.1"},
	{"deepseek_v3.1", "deepseek-v3.1"},
	{"deepseek-chat-v3.1", "deepseek-v3.1"},
	{"deepseek", "deepseek"},
	{"chatglm", "glm"},
	{"glm", "glm"},
	{"nemotron", "nemotron"}, // before llama: e.g. Llama-3.3-Nemotron
	{"llama", "llama"},
	{"mixtral", "mixtral"}, // before mistral
	{"mistral", "mistral"},
	{"gemma", "gemma"},
	{"phi-", "phi"},
	{"phi3", "phi"}, // GGUF architecture names
	{"phi4", "phi"},
	{"olmo", "olmo"},
	{"granite", "granite"},
	{"hunyuan", "hunyuan"},
	{"exaone", "exaone"},
	{"kimi", "kimi"},
	{"minimax", "minimax"},
	{"seed-oss", "seed-oss"},
	{"gpt-oss", "gpt-oss"}, // before gpt-
	{"gptoss", "gpt-oss"},  // GGUF architecture name
	{"gpt-", "gpt"},
	{"claude", "claude"},
	{"grok", "grok"},
	{"gemini", "gemini"},
	{"command", "command"},
}

// oSeriesRe matches OpenAI reasoning-series basenames: o1, o3-mini, o4-mini...
var oSeriesRe = regexp.MustCompile(`^o[0-9]+(-|$)`)

// FamilyFromName infers the canonical model family from a model name or
// registry id (e.g. "Qwen/Qwen3-235B-A22B", "deepseek-r1:70b").
func FamilyFromName(name string) string {
	n := strings.ToLower(strings.TrimSpace(name))
	if n == "" {
		return ""
	}
	base := n
	if i := strings.LastIndex(base, "/"); i >= 0 {
		base = base[i+1:]
	}
	if oSeriesRe.MatchString(base) {
		return "o-series"
	}
	for _, r := range familyRules {
		if strings.Contains(n, r.substr) {
			return r.family
		}
	}
	return ""
}

// kwargBoolBindings maps reasoning.disable/enable onto a boolean
// chat-template kwarg (vLLM / SGLang / llama.cpp convention).
func kwargBoolBindings(kwarg string) bindingSet {
	param := "chat_template_kwargs." + kwarg
	return bindingSet{
		system.IntentReasoningDisable: {Kind: system.BindingKindTemplateKwarg, Param: param, ParamType: "boolean", Value: false},
		system.IntentReasoningEnable:  {Kind: system.BindingKindTemplateKwarg, Param: param, ParamType: "boolean", Value: true},
	}
}

func effortKwargBindings() bindingSet {
	return bindingSet{
		system.IntentReasoningEffort: {
			Kind:       system.BindingKindTemplateKwarg,
			Param:      "chat_template_kwargs.reasoning_effort",
			ParamType:  "enum",
			EnumValues: []string{"low", "medium", "high"},
		},
	}
}

func budgetKwargBindings() bindingSet {
	param := "chat_template_kwargs.thinking_budget"
	return bindingSet{
		system.IntentReasoningBudget:  {Kind: system.BindingKindTemplateKwarg, Param: param, ParamType: "number"},
		system.IntentReasoningDisable: {Kind: system.BindingKindTemplateKwarg, Param: param, ParamType: "number", Value: 0},
	}
}

// bindingsForFamily returns the family's default reasoning bindings, used
// when the backend exposes no direct evidence (no chat template, no
// capability list). The model name refines the family answer: a "-thinking"
// variant of a hybrid family reasons unconditionally. alwaysOn is reported
// separately since it is a mode with no bindings at all.
func bindingsForFamily(family, modelName string) (alwaysOn bool, b bindingSet) {
	if strings.Contains(strings.ToLower(modelName), "thinking") {
		return true, nil
	}

	switch family {
	case "qwen3", "glm", "hunyuan":
		return false, kwargBoolBindings("enable_thinking")
	case "deepseek-v3.1", "granite":
		return false, kwargBoolBindings("thinking")
	case "qwq", "deepseek-r1":
		return true, nil
	case "gpt-oss":
		return false, effortKwargBindings()
	case "seed-oss":
		return false, budgetKwargBindings()
	case "nemotron":
		return false, bindingSet{
			system.IntentReasoningDisable: {Kind: system.BindingKindSystemPrompt, Hint: `put "detailed thinking off" in the system prompt`},
			system.IntentReasoningEnable:  {Kind: system.BindingKindSystemPrompt, Hint: `put "detailed thinking on" in the system prompt`},
		}
	case "claude":
		return claudeBindings(modelName)
	case "o-series":
		return false, bindingSet{
			system.IntentReasoningEffort: {Kind: system.BindingKindBodyParam, Param: "reasoning_effort", ParamType: "enum", EnumValues: []string{"minimal", "low", "medium", "high"}},
		}
	case "gemini":
		param := "generationConfig.thinkingConfig.thinkingBudget"
		return false, bindingSet{
			system.IntentReasoningBudget:  {Kind: system.BindingKindBodyParam, Param: param, ParamType: "number"},
			system.IntentReasoningDisable: {Kind: system.BindingKindBodyParam, Param: param, ParamType: "number", Value: 0},
		}
	}
	return false, nil
}

// claudeVersionRe extracts the first major[.minor] number pair from a Claude
// model name (claude-opus-4-6, claude-sonnet-4-5-20250929, claude-3-7-sonnet,
// claude-opus-5). A minor of three or more digits is a date, not a version.
var claudeVersionRe = regexp.MustCompile(`(\d+)(?:[-.](\d{1,2}))?(?:[-.]|$)`)

func claudeVersion(name string) (major, minor int, ok bool) {
	base := strings.ToLower(name)
	if i := strings.LastIndex(base, "/"); i >= 0 {
		base = base[i+1:]
	}
	m := claudeVersionRe.FindStringSubmatch(base)
	if m == nil {
		return 0, 0, false
	}
	major, _ = strconv.Atoi(m[1])
	if m[2] != "" {
		minor, _ = strconv.Atoi(m[2])
	}
	return major, minor, true
}

// claudeBindings returns the Anthropic Messages API reasoning knobs for a
// Claude model generation. Verified against
// https://platform.claude.com/docs/en/build-with-claude/thinking and
// https://platform.claude.com/docs/en/build-with-claude/effort:
//   - Fable / Mythos: always-on adaptive reasoning; only effort is tunable.
//   - 4.6 and newer (incl. 5.x): thinking {type: adaptive|disabled},
//     output_config.effort; budget_tokens is gone.
//   - 4.5: thinking {type: enabled, budget_tokens}|{type: disabled},
//     effort low|medium|high.
//   - older: thinking {type: enabled, budget_tokens}|{type: disabled}.
func claudeBindings(modelName string) (alwaysOn bool, b bindingSet) {
	name := strings.ToLower(modelName)
	disable := &system.ParamBinding{Kind: system.BindingKindBodyParam, Param: "thinking", ParamType: "object", Value: map[string]any{"type": "disabled"}}
	effortAll := &system.ParamBinding{Kind: system.BindingKindBodyParam, Param: "output_config.effort", ParamType: "enum", EnumValues: []string{"low", "medium", "high", "xhigh", "max"}, Hint: "xhigh/max availability varies per model"}
	enabledWithBudget := &system.ParamBinding{Kind: system.BindingKindBodyParam, Param: "thinking", ParamType: "object", Value: map[string]any{"type": "enabled", "budget_tokens": 1024}, Hint: "budget_tokens >= 1024 and below max_tokens"}
	budget := &system.ParamBinding{Kind: system.BindingKindBodyParam, Param: "thinking.budget_tokens", ParamType: "number", Hint: "minimum 1024; must be below max_tokens"}

	if strings.Contains(name, "fable") || strings.Contains(name, "mythos") {
		// Reasons unconditionally; thinking.type enabled/disabled are rejected.
		return false, bindingSet{system.IntentReasoningEffort: effortAll}
	}

	major, minor, ok := claudeVersion(name)
	switch {
	case !ok || major >= 5 || (major == 4 && minor >= 6):
		disable.Hint = "Opus 5 / Sonnet 5 reject it at effort xhigh/max"
		return false, bindingSet{
			system.IntentReasoningDisable: disable,
			system.IntentReasoningEnable:  {Kind: system.BindingKindBodyParam, Param: "thinking", ParamType: "object", Value: map[string]any{"type": "adaptive"}},
			system.IntentReasoningEffort:  effortAll,
		}
	case major == 4 && minor == 5:
		return false, bindingSet{
			system.IntentReasoningDisable: disable,
			system.IntentReasoningEnable:  enabledWithBudget,
			system.IntentReasoningBudget:  budget,
			system.IntentReasoningEffort:  {Kind: system.BindingKindBodyParam, Param: "output_config.effort", ParamType: "enum", EnumValues: []string{"low", "medium", "high"}},
		}
	default:
		return false, bindingSet{
			system.IntentReasoningDisable: disable,
			system.IntentReasoningEnable:  enabledWithBudget,
			system.IntentReasoningBudget:  budget,
		}
	}
}
