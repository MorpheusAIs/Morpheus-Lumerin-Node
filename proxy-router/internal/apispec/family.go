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
	{"cogito", "cogito"}, // before deepseek, llama, qwen: cogito-v2-preview-deepseek-671B-MoE
	{"qwq", "qwq"},
	{"qwen3.5", "qwen3.5"}, // before qwen3
	{"qwen3_5", "qwen3.5"},
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
	{"hermes", "hermes"},     // before llama and mistral: Hermes-3-Llama-3.1, Hermes-2-Pro-Mistral
	{"nemotron", "nemotron"}, // before llama: e.g. Llama-3.3-Nemotron
	{"llama", "llama"},
	{"magistral", "magistral"}, // before mistral
	{"mixtral", "mixtral"},     // before mistral
	{"mistral", "mistral"},
	{"gemma", "gemma"},
	{"phi-4-reasoning", "phi-4-reasoning"}, // before phi-
	{"phi4-reasoning", "phi-4-reasoning"},  // Ollama tag
	{"phi-", "phi"},
	{"phi3", "phi"}, // GGUF architecture names
	{"phi4", "phi"},
	{"olmo", "olmo"},
	{"granite", "granite"},
	{"hunyuan", "hunyuan"},
	{"exaone", "exaone"},
	{"kimi", "kimi"},
	{"mimo", "mimo"},
	{"minimax", "minimax"},
	{"arcee", "arcee"},
	{"trinity-large", "arcee"},
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

// Recognised but unbound (no documented switch, and not verifiably
// unconditional): magistral reasons through a recommended system prompt that
// the card does not bake into the template
// (https://huggingface.co/mistralai/Magistral-Small-2509); arcee's
// Trinity-Large-Thinking is covered by the "thinking" name rule and
// Trinity-Large-Preview has no reasoning mode
// (https://huggingface.co/arcee-ai/Trinity-Large-Thinking).
func bindingsForFamily(family, modelName string) (alwaysOn bool, b bindingSet) {
	if strings.Contains(strings.ToLower(modelName), "thinking") {
		return true, nil
	}

	switch family {
	case "qwen3", "qwen3.5", "glm", "hunyuan":
		return false, kwargBoolBindings("enable_thinking")
	case "deepseek-v3.1", "granite":
		return false, kwargBoolBindings("thinking")
	// phi-4-reasoning: the card documents a two-section response, a <think>
	// reasoning block followed by the solution, requested by the system
	// prompt its chat template applies, and no switch that turns it off
	// (https://huggingface.co/microsoft/Phi-4-reasoning).
	case "qwq", "deepseek-r1", "phi-4-reasoning":
		return true, nil
	case "gpt-oss":
		return false, effortKwargBindings()
	case "seed-oss":
		return false, budgetKwargBindings()
	case "nemotron":
		return nemotronBindings(modelName)
	case "cogito":
		return false, cogitoBindings()
	case "hermes":
		return hermesBindings(modelName)
	case "mimo":
		return mimoBindings(modelName)
	case "olmo":
		return olmoBindings(modelName)
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
		return false, thinkingModeKwargBindings()
	}
	return false, nil
}

// The M3 template compares thinking_mode against exactly "enabled",
// "disabled" and "adaptive":
// https://huggingface.co/MiniMaxAI/MiniMax-M3/raw/main/chat_template.jinja
func thinkingModeKwargBindings() bindingSet {
	param := "chat_template_kwargs.thinking_mode"
	hint := "enabled | disabled | adaptive; adaptive when unset"
	return bindingSet{
		system.IntentReasoningDisable: {Kind: system.BindingKindTemplateKwarg, Param: param, ParamType: "string", Value: "disabled", Hint: hint},
		system.IntentReasoningEnable:  {Kind: system.BindingKindTemplateKwarg, Param: param, ParamType: "string", Value: "enabled", Hint: hint},
	}
}

// Nemotron 3: enable_thinking kwarg, on by default
// (https://huggingface.co/nvidia/NVIDIA-Nemotron-3-Nano-30B-A3B-BF16).
// Nemotron Nano v2 and Llama-3.3-Nemotron-Super v1.5: "/no_think" in the
// system prompt, reasoning on by default
// (https://huggingface.co/nvidia/NVIDIA-Nemotron-Nano-9B-v2,
// https://huggingface.co/nvidia/Llama-3_3-Nemotron-Super-49B-v1_5).
// Llama-3.x-Nemotron v1: "detailed thinking on|off" in the system prompt
// (https://huggingface.co/nvidia/Llama-3_3-Nemotron-Super-49B-v1).
var nemotronNoThinkRe = regexp.MustCompile(`nemotron[-_]nano[-_]\d+b[-_]v2|nemotron.*v1[-_.]5`)

func nemotronBindings(modelName string) (alwaysOn bool, b bindingSet) {
	name := strings.ToLower(modelName)
	switch {
	case containsAny(name, "nemotron-3", "nemotron_3", "nemotron3"):
		return false, kwargBoolBindings("enable_thinking")
	case nemotronNoThinkRe.MatchString(name):
		return false, bindingSet{
			system.IntentReasoningDisable: {Kind: system.BindingKindSystemPrompt, Hint: `put "/no_think" in the system prompt`},
			system.IntentReasoningEnable:  {Kind: system.BindingKindSystemPrompt, Hint: `put "/think" in the system prompt (the default)`},
		}
	}
	return false, bindingSet{
		system.IntentReasoningDisable: {Kind: system.BindingKindSystemPrompt, Hint: `put "detailed thinking off" in the system prompt`},
		system.IntentReasoningEnable:  {Kind: system.BindingKindSystemPrompt, Hint: `put "detailed thinking on" in the system prompt`},
	}
}

// Cogito v2: off by default; enable_thinking kwarg, or the system prompt
// "Enable deep thinking subroutine." with the reply prefilled "<think>\n"
// (https://huggingface.co/deepcogito/cogito-v2-preview-llama-70B).
func cogitoBindings() bindingSet {
	b := kwargBoolBindings("enable_thinking")
	b[system.IntentReasoningEnable].Hint = `or put "Enable deep thinking subroutine." in the system prompt and prefill the reply with "<think>\n"`
	return b
}

// Hermes 4: hybrid, off unless the thinking kwarg is true or the deep-thinking
// system prompt is used (https://huggingface.co/NousResearch/Hermes-4-70B,
// https://huggingface.co/NousResearch/Hermes-4-70B/raw/main/chat_template.jinja).
// Hermes 2/3 have no thinking mode.
func hermesBindings(modelName string) (alwaysOn bool, b bindingSet) {
	if containsAny(modelName, "hermes-4", "hermes_4", "hermes4") {
		return false, kwargBoolBindings("thinking")
	}
	return false, nil
}

// MiMo-V2-Flash: enable_thinking kwarg
// (https://huggingface.co/XiaomiMiMo/MiMo-V2-Flash). MiMo-VL-7B-RL-2508
// toggles with "/no_think" appended to the user message
// (https://huggingface.co/XiaomiMiMo/MiMo-VL-7B-RL-2508), which no binding
// kind expresses; MiMo-7B-RL documents no toggle
// (https://huggingface.co/XiaomiMiMo/MiMo-7B-RL). Both stay unbound.
func mimoBindings(modelName string) (alwaysOn bool, b bindingSet) {
	if containsAny(modelName, "mimo-v2", "mimo_v2") {
		return false, kwargBoolBindings("enable_thinking")
	}
	return false, nil
}

// Olmo 3 Think: the template opens every reply with <think> and exposes no
// switch (https://huggingface.co/allenai/Olmo-3-32B-Think,
// https://huggingface.co/allenai/Olmo-3-32B-Think/raw/main/chat_template.jinja).
// Olmo 3 Instruct and earlier Olmo have no thinking mode.
var olmo3ThinkRe = regexp.MustCompile(`olmo[-_]?3.*think`)

func olmoBindings(modelName string) (alwaysOn bool, b bindingSet) {
	if olmo3ThinkRe.MatchString(strings.ToLower(modelName)) {
		return true, nil
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
	if strings.Contains(tpl, "thinking_mode") {
		return thinkingModeKwargBindings()
	}
	for _, stmt := range jinjaStmtRe.FindAllString(tpl, -1) {
		if bareThinkingRe.MatchString(stmt) {
			return kwargBoolBindings("thinking")
		}
	}
	return nil
}
