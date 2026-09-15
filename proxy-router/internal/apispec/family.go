package apispec

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/system"
)

// Ordered: first match wins.
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

var oSeriesRe = regexp.MustCompile(`^o[0-9]+(-|$)`)

// Built from familyRules so the two cannot drift.
var knownFamilies = func() map[string]bool {
	m := map[string]bool{"o-series": true}
	for _, r := range familyRules {
		m[r.family] = true
	}
	return m
}()

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
	case "gemma":
		return gemmaBindings(modelName)
	case "kimi":
		return kimiBindings(modelName)
	case "minimax":
		return minimaxBindings(modelName)
	case "exaone":
		return exaoneBindings(modelName)
	}
	return false, nil
}

func containsAny(name string, subs ...string) bool {
	n := strings.ToLower(name)
	for _, sub := range subs {
		if strings.Contains(n, sub) {
			return true
		}
	}
	return false
}

// Gemma 4 only, off by default: https://huggingface.co/google/gemma-4-31B-it,
// https://ai.google.dev/gemma/docs/capabilities/thinking. Gemma 3 has no
// thinking mode: https://ai.google.dev/gemma/docs/core/model_card_3.
// gemma4Re must not match a 4B size token: google/medgemma-4b-it is Gemma 3
// based (https://huggingface.co/google/medgemma-4b-it).
var gemma4Re = regexp.MustCompile(`gemma[-_]?4(?:[-_:.]|$)`)

func gemmaBindings(modelName string) (alwaysOn bool, b bindingSet) {
	if gemma4Re.MatchString(strings.ToLower(modelName)) {
		return false, kwargBoolBindings("enable_thinking")
	}
	return false, nil
}

// https://huggingface.co/moonshotai/Kimi-K3 (always on; no documented kwarg toggle)
// https://huggingface.co/moonshotai/Kimi-K2.7-Code (always on)
// https://huggingface.co/moonshotai/Kimi-K2.5, https://huggingface.co/moonshotai/Kimi-K2.6 (thinking kwarg, on by default)
// https://huggingface.co/moonshotai/Kimi-K2-Instruct (no thinking mode)
func kimiBindings(modelName string) (alwaysOn bool, b bindingSet) {
	switch {
	case containsAny(modelName, "kimi-k3", "kimi_k3"):
		return true, nil
	case containsAny(modelName, "k2.7-code"):
		return true, nil
	case containsAny(modelName, "k2.5", "k2.6"):
		return false, kwargBoolBindings("thinking")
	}
	return false, nil
}

// https://huggingface.co/MiniMaxAI/MiniMax-M2, https://platform.minimax.io/docs/api-reference/text-chat-openai (M2.x: no off switch)
// https://huggingface.co/MiniMaxAI/MiniMax-M3/raw/main/chat_template.jinja (thinking_mode; adaptive when unset)
// https://huggingface.co/MiniMaxAI/MiniMax-Text-01 (no thinking mode)
// https://huggingface.co/MiniMaxAI/MiniMax-M1-80k (emits <think>; no source says whether it can be disabled, so unbound)
func minimaxBindings(modelName string) (alwaysOn bool, b bindingSet) {
	switch {
	case containsAny(modelName, "minimax-m2"):
		return true, nil
	case containsAny(modelName, "minimax-m3"):
		param := "chat_template_kwargs.thinking_mode"
		hint := "enabled | disabled | adaptive; adaptive when unset"
		return false, bindingSet{
			system.IntentReasoningDisable: {Kind: system.BindingKindTemplateKwarg, Param: param, ParamType: "string", Value: "disabled", Hint: hint},
			system.IntentReasoningEnable:  {Kind: system.BindingKindTemplateKwarg, Param: param, ParamType: "string", Value: "enabled", Hint: hint},
		}
	}
	return false, nil
}

// https://huggingface.co/LGAI-EXAONE/EXAONE-Deep-32B (always on)
// https://huggingface.co/LGAI-EXAONE/EXAONE-4.0-32B, https://huggingface.co/LGAI-EXAONE/EXAONE-4.0.1-32B (enable_thinking, off by default)
// https://huggingface.co/LGAI-EXAONE/EXAONE-3.5-32B-Instruct (no reasoning mode)
func exaoneBindings(modelName string) (alwaysOn bool, b bindingSet) {
	switch {
	case containsAny(modelName, "exaone-deep"):
		return true, nil
	case containsAny(modelName, "exaone-4", "exaone4", "exaone_4"):
		return false, kwargBoolBindings("enable_thinking")
	}
	return false, nil
}

func IsKnownFamily(name string) bool {
	return knownFamilies[strings.ToLower(strings.TrimSpace(name))]
}

// Minor is capped at two digits: three or more is a date
// (claude-sonnet-4-5-20250929).
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

// https://platform.claude.com/docs/en/build-with-claude/thinking
// https://platform.claude.com/docs/en/build-with-claude/effort
func claudeBindings(modelName string) (alwaysOn bool, b bindingSet) {
	name := strings.ToLower(modelName)
	disable := &system.ParamBinding{Kind: system.BindingKindBodyParam, Param: "thinking", ParamType: "object", Value: map[string]any{"type": "disabled"}}
	effortAll := &system.ParamBinding{Kind: system.BindingKindBodyParam, Param: "output_config.effort", ParamType: "enum", EnumValues: []string{"low", "medium", "high", "xhigh", "max"}, Hint: "xhigh/max availability varies per model"}
	enabledWithBudget := &system.ParamBinding{Kind: system.BindingKindBodyParam, Param: "thinking", ParamType: "object", Value: map[string]any{"type": "enabled", "budget_tokens": 1024}, Hint: "budget_tokens >= 1024 and below max_tokens"}
	budget := &system.ParamBinding{Kind: system.BindingKindBodyParam, Param: "thinking.budget_tokens", ParamType: "number", Hint: "minimum 1024; must be below max_tokens"}

	if strings.Contains(name, "fable") || strings.Contains(name, "mythos") {
		// thinking.type is rejected and only effort is tunable, so reported as
		// tunable rather than always_on.
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

func refinesFamily(base, candidate string) bool {
	if candidate == "" {
		return false
	}
	return base == "" || strings.HasPrefix(candidate, base+"-")
}

var jinjaStmtRe = regexp.MustCompile(`(?s)\{%.*?%\}`)

var bareThinkingRe = regexp.MustCompile(`(^|[^a-zA-Z0-9_])thinking($|[^a-zA-Z0-9_])`)

func bindingsFromTemplate(tpl string) bindingSet {
	if tpl == "" {
		return nil
	}
	if strings.Contains(tpl, "enable_thinking") {
		return kwargBoolBindings("enable_thinking")
	}
	if strings.Contains(tpl, "reasoning_effort") {
		return effortKwargBindings()
	}
	if strings.Contains(tpl, "thinking_budget") {
		return budgetKwargBindings()
	}
	for _, stmt := range jinjaStmtRe.FindAllString(tpl, -1) {
		if bareThinkingRe.MatchString(stmt) {
			return kwargBoolBindings("thinking")
		}
	}
	return nil
}
