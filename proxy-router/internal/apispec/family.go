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

// knownFamilies is every label FamilyFromName can produce: the familyRules
// table plus o-series (matched by oSeriesRe). Built from the table so the
// two can't drift.
var knownFamilies = func() map[string]bool {
	m := map[string]bool{"o-series": true}
	for _, r := range familyRules {
		m[r.family] = true
	}
	return m
}()

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
// variant of a hybrid family reasons unconditionally, and gemma, kimi,
// minimax and exaone only bind for the members whose official docs describe
// a reasoning mode. alwaysOn is reported separately since it is a mode with
// no bindings at all.
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

// containsAny reports whether the lowercased name contains any of subs.
func containsAny(name string, subs ...string) bool {
	n := strings.ToLower(name)
	for _, sub := range subs {
		if strings.Contains(n, sub) {
			return true
		}
	}
	return false
}

// gemmaBindings: only Gemma 4 has a thinking mode, and it is off unless the
// enable_thinking chat-template kwarg is passed. Verified against the Google
// model card https://huggingface.co/google/gemma-4-31B-it ("To enable
// reasoning, set `enable_thinking=True`"), its chat template
// https://huggingface.co/google/gemma-4-31B-it/raw/main/chat_template.jinja
// (`enable_thinking | default(false)`),
// https://ai.google.dev/gemma/docs/capabilities/thinking and the vLLM note at
// https://docs.vllm.ai/en/latest/features/reasoning_outputs/ ("Gemma 4
// reasoning is disabled by default; to enable it, pass enable_thinking=True
// in your chat_template_kwargs"). Gemma 3 and older have no thinking mode
// (the Gemma 4 card's own comparison column is "Gemma 3 27B (no think)" and
// https://ai.google.dev/gemma/docs/core/model_card_3 documents none), so
// they get no bindings.
//
// gemma4Re matches the Gemma 4 generation token (gemma-4-31b-it,
// gemma-4-26b-a4b-it, gemma4:31b, gemma_4_12b) and not a 4B size token:
// google/medgemma-4b-it "is built based on Gemma 3"
// (https://huggingface.co/google/medgemma-4b-it) and has no thinking mode,
// so a plain "gemma-4" substring test would bind it wrongly.
var gemma4Re = regexp.MustCompile(`gemma[-_]?4(?:[-_:.]|$)`)

func gemmaBindings(modelName string) (alwaysOn bool, b bindingSet) {
	if gemma4Re.MatchString(strings.ToLower(modelName)) {
		return false, kwargBoolBindings("enable_thinking")
	}
	return false, nil
}

// kimiBindings covers Moonshot's Kimi K2 / K3 line, per the official model
// cards:
//   - Kimi-K3 (https://huggingface.co/moonshotai/Kimi-K3): "Kimi K3 always
//     has thinking enabled, and will return reasoning_content" — always on.
//     (Its encoding_k3.py carries a thinking flag, but no deploy guide
//     documents it as a chat-template kwarg, so no toggle is advertised.)
//   - Kimi-K2-Thinking (https://huggingface.co/moonshotai/Kimi-K2-Thinking):
//     a thinking model whose template and deploy guide expose no toggle —
//     always on; the generic "thinking" name rule above already handles it.
//   - Kimi-K2.7-Code (https://huggingface.co/moonshotai/Kimi-K2.7-Code):
//     "forces thinking and preserve_thinking as True. [...] Instant mode is
//     not supported." — always on.
//   - Kimi-K2.5 / Kimi-K2.6 (https://huggingface.co/moonshotai/Kimi-K2.5,
//     https://huggingface.co/moonshotai/Kimi-K2.6): thinking on by default;
//     "To use instant mode, you need to pass {'chat_template_kwargs':
//     {"thinking": False}}" — the same boolean kwarg shape as deepseek-v3.1.
//   - Kimi-K2-Instruct (https://huggingface.co/moonshotai/Kimi-K2-Instruct):
//     "a reflex-grade model without long thinking" — no bindings.
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

// minimaxBindings: MiniMax-M2 and every M2.x (M2.1, M2.5, M2.7, -highspeed)
// are interleaved-thinking models with no off switch. Verified against
// https://huggingface.co/MiniMaxAI/MiniMax-M2 ("MiniMax-M2 is an interleaved
// thinking model [...] Do not remove the <think>...</think> part"; its chat
// template pre-fills "<think>\n" unconditionally and reads no kwarg) and
// https://platform.minimax.io/docs/api-reference/text-chat-openai ("For M2.x
// models, thinking cannot be disabled").
// MiniMax-M3 (https://huggingface.co/MiniMaxAI/MiniMax-M3) is hybrid: "M3
// supports three reasoning modes through the `thinking` parameter:
// enabled / adaptive / disabled". Its chat template
// (https://huggingface.co/MiniMaxAI/MiniMax-M3/raw/main/chat_template.jinja)
// reads the thinking_mode kwarg ("enabled" | "disabled" | "adaptive", and
// takes the adaptive branch when it is undefined), and the platform docs
// expose the same switch as thinking.type ("When omitted, adaptive thinking
// is enabled by default"; "disabled: Skip thinking for MiniMax-M3 and answer
// directly"). MiniMax-Text-01
// (https://huggingface.co/MiniMaxAI/MiniMax-Text-01) has no thinking mode.
// MiniMax-M1 (https://huggingface.co/MiniMaxAI/MiniMax-M1-80k) emits <think>
// blocks, but no official source says whether that can be turned off, so it
// is deliberately left without bindings rather than guessed.
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

// exaoneBindings covers LG AI Research's EXAONE line, per the official
// LGAI-EXAONE model cards:
//   - EXAONE Deep (https://huggingface.co/LGAI-EXAONE/EXAONE-Deep-32B): the
//     template unconditionally opens <thought> ("Ensure the model starts
//     with `<thought>\n` for reasoning steps") and documents no toggle —
//     always on.
//   - EXAONE 4.0 / 4.0.1 (https://huggingface.co/LGAI-EXAONE/EXAONE-4.0-32B,
//     https://huggingface.co/LGAI-EXAONE/EXAONE-4.0.1-32B): "You can activate
//     reasoning mode by using the `enable_thinking=True` argument"; the chat
//     template emits a closed, empty <think> block unless enable_thinking is
//     true, so the default is off.
//   - EXAONE 3.5 and older
//     (https://huggingface.co/LGAI-EXAONE/EXAONE-3.5-32B-Instruct): no
//     reasoning mode.
//
// Like the gemma rule, the 4.x token also accepts the hyphen-less Ollama-tag
// and GGUF-architecture spellings (exaone4:32b, exaone4, exaone_4).
func exaoneBindings(modelName string) (alwaysOn bool, b bindingSet) {
	switch {
	case containsAny(modelName, "exaone-deep"):
		return true, nil
	case containsAny(modelName, "exaone-4", "exaone4", "exaone_4"):
		return false, kwargBoolBindings("enable_thinking")
	}
	return false, nil
}

// IsKnownFamily reports whether name is a family label FamilyFromName can
// produce (the familyRules table plus o-series), so an explicit modelFamily
// never warns when the inferred value would not. Labels without bindings
// (llama, mistral, phi...) are known too: Build accepts any value and
// reports it as modelFamily, the warning only flags labels that no name
// would ever infer. Comparison is case-insensitive to match Build's own
// normalization of ModelFamily.
func IsKnownFamily(name string) bool {
	return knownFamilies[strings.ToLower(strings.TrimSpace(name))]
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
		// Always reasons (thinking.type enabled/disabled are rejected) but
		// effort is tunable, so it is reported as tunable with only the
		// effort binding.
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

// refinesFamily reports whether candidate is a more specific variant of
// base (e.g. "deepseek-r1" refines "deepseek"), or base is unknown.
func refinesFamily(base, candidate string) bool {
	if candidate == "" {
		return false
	}
	return base == "" || strings.HasPrefix(candidate, base+"-")
}

// jinjaStmtRe extracts {% ... %} statement blocks from a chat template.
var jinjaStmtRe = regexp.MustCompile(`(?s)\{%.*?%\}`)

// bareThinkingRe matches the word `thinking` used as a template variable.
var bareThinkingRe = regexp.MustCompile(`(^|[^a-zA-Z0-9_])thinking($|[^a-zA-Z0-9_])`)

// bindingsFromTemplate derives the reasoning bindings from chat template
// source — the template is the authoritative definition of which kwargs it
// honors.
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
