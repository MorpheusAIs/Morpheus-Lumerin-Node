// Package apispec composes a model's API spec (system.ModelApiSpec) from
// preset tables, family defaults and detector evidence.
package apispec

import (
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/config"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/system"
)

type Evidence struct {
	// "" is undetermined: the wire vocabulary is unknown (it could be a
	// gateway), so no bindings are advertised.
	Stack            string
	Via              string
	ModelName        string
	ModelFamily      string
	ServedModelID    string
	Architecture     string
	ChatTemplate     string
	OllamaThinking   bool
	GatewayReasoning bool
	RegistryBindings map[string]*system.ParamBinding
	Parameters       []string
	ReasoningEfforts []string
	LiteLLMSeen      bool
	// nil means the list could not be read; an empty list is a known, empty list.
	LiteLLMSupportedParams []string
	Declared               bool
}

func Build(cfg config.ModelConfig) *system.ModelApiSpec {
	stack := StackFor(cfg.ApiStack)
	if stack == "" {
		return nil
	}
	return Compose(Evidence{Stack: stack, ModelName: cfg.ModelName, ModelFamily: cfg.ModelFamily, Declared: true})
}

func Compose(ev Evidence) *system.ModelApiSpec {
	api, _ := ComposeWithTrace(ev)
	return api
}

func ComposeWithTrace(ev Evidence) (*system.ModelApiSpec, []string) {
	var lines []string
	tracef := func(format string, args ...any) { lines = append(lines, fmt.Sprintf(format, args...)) }

	ev.Parameters = sanitizeNames(ev.Parameters, maxParameters, "parameters", tracef)
	ev.ReasoningEfforts = sanitizeNames(ev.ReasoningEfforts, maxReasoningEfforts, "reasoning efforts", tracef)
	if ev.LiteLLMSupportedParams != nil {
		// Keep it a known list: nil means unread.
		supported := sanitizeNames(ev.LiteLLMSupportedParams, maxParameters, "litellm supported params", tracef)
		if supported == nil {
			supported = []string{}
		}
		ev.LiteLLMSupportedParams = supported
	}

	stack := ev.Stack
	api := &system.ModelApiSpec{Stack: stack, Via: ev.Via, Source: system.ApiSpecSourceDetected}
	if ev.Declared {
		api.Source = system.ApiSpecSourceDeclared
	}

	family := strings.ToLower(strings.TrimSpace(ev.ModelFamily))
	if family != "" {
		tracef("family %q: declared in models-config", family)
	} else {
		family = FamilyFromName(ev.Architecture)
		if family != "" {
			tracef("family %q: from backend-reported architecture %q", family, ev.Architecture)
		} else if ev.Architecture != "" {
			tracef("architecture %q maps to no known family; falling back to names", ev.Architecture)
		}
		if served := FamilyFromName(ev.ServedModelID); refinesFamily(family, served) {
			family = served
			tracef("family %q: from backend-served model id %q", family, ev.ServedModelID)
		}
		if named := FamilyFromName(ev.ModelName); refinesFamily(family, named) {
			family = named
			tracef("family %q: from configured model name %q", family, ev.ModelName)
		}
		if family == "" {
			tracef("family: undetermined (no architecture, served id or name match)")
		}
	}
	api.ModelFamily = family

	name := ev.ModelName
	if ev.ServedModelID != "" {
		name = ev.ServedModelID
	}

	alwaysOn := false
	var bindings bindingSet
	if len(ev.RegistryBindings) > 0 {
		bindings = cloneSet(ev.RegistryBindings)
		tracef("bindings: from the %s model listing", stack)
	}
	if len(bindings) == 0 && ev.OllamaThinking {
		// The capability only says the model thinks; whether it can be toggled is
		// a family fact.
		famAlwaysOn, fam := bindingsForFamily(family, name)
		switch {
		case famAlwaysOn:
			alwaysOn = true
			tracef("bindings: ollama reports the 'thinking' capability; family %q reasons unconditionally", family)
		case fam != nil && fam[system.IntentReasoningDisable] == nil && fam[system.IntentReasoningEffort] != nil:
			bindings = cloneSet(fam)
			tracef("bindings: ollama reports the 'thinking' capability; family %q takes effort levels, not a toggle", family)
		default:
			bindings = ollamaThinkBindings()
			tracef("bindings: ollama reports the 'thinking' capability -> /v1 reasoning_effort (none = off)")
		}
	}
	if len(bindings) == 0 && !alwaysOn {
		if stack == "" {
			if bindingsFromTemplate(ev.ChatTemplate) != nil {
				tracef("bindings: chat template evidence skipped — stack undetermined (unknown wire vocabulary)")
			}
		} else if bindings = bindingsFromTemplate(ev.ChatTemplate); bindings != nil {
			tracef("bindings: from chat template source -> %s", describeBindings(bindings))
		}
	}
	familyReasons := false
	if len(bindings) == 0 && !alwaysOn {
		var defaults bindingSet
		alwaysOn, defaults = bindingsForFamily(family, name)
		familyReasons = alwaysOn || len(defaults) > 0
		switch {
		case defaults == nil && alwaysOn:
			tracef("thinking: family %q (model %q) reasons unconditionally", family, name)
		case defaults == nil:
		case stack == "":
			tracef("bindings: family default for %q skipped — stack undetermined (unknown wire vocabulary)", family)
		case gatedStack(stack) && familyNativeVendor[family] != stack:
			tracef("bindings: family default for %q skipped — %q does not accept it", family, stack)
		default:
			bindings = cloneSet(defaults)
			tracef("bindings: family default for %q (model %q) -> %s", family, name, describeBindings(bindings))
		}
	}

	if stack == "ollama" && bindings != nil {
		before := describeBindings(bindings)
		rewriteBindingsForOllama(bindings)
		if after := describeBindings(bindings); after != before {
			tracef("bindings: rewritten onto ollama's /v1 reasoning_effort (chat_template_kwargs is not honored by ollama) -> %s", after)
		}
	}

	// A detected gateway may hide its upstream's knob, so family knowledge alone
	// counts as reasoning evidence only for a declared stack.
	reasoningKnown := alwaysOn || ev.GatewayReasoning || hasReasoningBindings(bindings) || (ev.Declared && familyReasons)
	if len(ev.Parameters) > 0 {
		api.Parameters = append([]string(nil), ev.Parameters...)
		sort.Strings(api.Parameters)
	}
	before := len(bindings)
	mergeStackTables(api, stack, bindings, reasoningKnown, alwaysOn)
	if added := len(api.Bindings) - before; added > 0 {
		tracef("bindings: +%d from the %s stack table (documented request params)", added, stack)
	}
	if len(ev.Parameters) == 0 && len(api.Parameters) > 0 {
		tracef("parameters: %d standard params from the %s stack table", len(api.Parameters), stack)
	}
	if ev.LiteLLMSeen {
		filterForLiteLLM(api, ev.LiteLLMSupportedParams, tracef)
	}
	if len(api.Bindings) == 0 && !alwaysOn {
		tracef("bindings: no remappable knobs known for this backend")
	}

	if stack == "litellm" && len(ev.ReasoningEfforts) > 0 {
		if effort := api.Bindings[system.IntentReasoningEffort]; effort != nil {
			effort.EnumValues = append([]string(nil), ev.ReasoningEfforts...)
			tracef("bindings: reasoning.effort levels from litellm supported_reasoning_efforts -> %s", strings.Join(effort.EnumValues, "|"))
			if !containsString(ev.ReasoningEfforts, "none") {
				if disable := api.Bindings[system.IntentReasoningDisable]; disable != nil && disable.Value == "none" {
					delete(api.Bindings, system.IntentReasoningDisable)
					tracef("bindings: reasoning.disable dropped — litellm does not list none among the supported efforts")
				}
			}
			if enable := api.Bindings[system.IntentReasoningEnable]; enable != nil {
				level := ""
				for _, e := range ev.ReasoningEfforts {
					if e != "none" {
						level = e
						break
					}
				}
				if level == "" {
					delete(api.Bindings, system.IntentReasoningEnable)
					tracef("bindings: reasoning.enable dropped — litellm lists no supported effort other than none")
				} else {
					enable.Value = level
					tracef("bindings: reasoning.enable value from litellm supported_reasoning_efforts -> %s", level)
				}
			}
		}
	}

	switch {
	case alwaysOn:
		api.Thinking = &system.ThinkingSpec{Mode: system.ThinkingModeAlwaysOn}
	case api.Bindings[system.IntentReasoningDisable] != nil, api.Bindings[system.IntentReasoningEnable] != nil:
		api.Thinking = &system.ThinkingSpec{Mode: system.ThinkingModeControllable}
	case api.Bindings[system.IntentReasoningEffort] != nil || api.Bindings[system.IntentReasoningBudget] != nil:
		api.Thinking = &system.ThinkingSpec{Mode: system.ThinkingModeTunable}
	}

	if stack == "" && api.ModelFamily == "" && api.Thinking == nil && len(api.Bindings) == 0 && len(api.Parameters) == 0 {
		tracef("nothing known about this backend: no api block")
		return nil, lines
	}
	return api, lines
}

// Largest documented lists: ~22 OpenAI params, ~30 per OpenRouter entry,
// 7 LiteLLM efforts.
const (
	maxParameters       = 64
	maxReasoningEfforts = 16
)

var nameRe = regexp.MustCompile(`^[A-Za-z0-9_.-]{1,64}$`)

func sanitizeNames(list []string, limit int, what string, tracef func(string, ...any)) []string {
	if len(list) == 0 {
		return nil
	}
	seen := make(map[string]bool, len(list))
	var out []string
	for _, name := range list {
		if len(out) >= limit || seen[name] || !nameRe.MatchString(name) {
			continue
		}
		seen[name] = true
		out = append(out, name)
	}
	if dropped := len(list) - len(out); dropped > 0 {
		tracef("%s: %d of %d backend-reported entries dropped (invalid name, duplicate or beyond the %d cap); %d kept", what, dropped, len(list), limit, len(out))
	}
	return out
}

func hasReasoningBindings(b bindingSet) bool {
	for intent := range b {
		if strings.HasPrefix(intent, "reasoning.") {
			return true
		}
	}
	return false
}

func containsString(list []string, s string) bool {
	for _, v := range list {
		if v == s {
			return true
		}
	}
	return false
}

func mergeStackTables(api *system.ModelApiSpec, stack string, family bindingSet, reasoningKnown, alwaysOn bool) {
	out := bindingSet{}
	for intent, b := range family {
		out[intent] = cloneBinding(b)
	}
	for intent, b := range stackBindings[stack] {
		if b == nil {
			continue
		}
		if _, exists := out[intent]; exists {
			continue
		}
		if strings.HasPrefix(intent, "reasoning.") {
			if !reasoningKnown {
				continue
			}
			if alwaysOn && intent != system.IntentReasoningFormat {
				continue
			}
		}
		out[intent] = cloneBinding(b)
	}
	if len(out) > 0 {
		api.Bindings = out
	}
	if len(api.Parameters) == 0 {
		if params := stackParameters[stack]; len(params) > 0 {
			api.Parameters = append([]string(nil), params...)
		}
	}
}

// LiteLLM forwards params it does not recognise verbatim as provider kwargs
// but validates standard OpenAI chat-completions params against its
// per-provider map and rejects unlisted ones (litellm.UnsupportedParamsError).
// With supported nil (unread) only reasoning_effort, the demonstrated
// rejection, is treated as not forwarded.
//
// anthropic's documented list is Messages vocabulary: intersecting it with
// LiteLLM's OpenAI-shaped list would keep only names that collide (model,
// messages, stream, ...) and drop forwarded ones spelled differently
// (stop_sequences vs. stop), so supported is reported verbatim instead.
// https://docs.litellm.ai/docs/completion/provider_specific_params
// https://docs.litellm.ai/docs/completion/drop_params
func filterForLiteLLM(api *system.ModelApiSpec, supported []string, tracef func(string, ...any)) {
	standard := make(map[string]bool, len(openaiChatParameters))
	for _, p := range openaiChatParameters {
		standard[p] = true
	}
	known := supported != nil
	listed := make(map[string]bool, len(supported))
	for _, p := range supported {
		listed[p] = true
	}
	if !known {
		tracef("litellm supported_openai_params unavailable: only reasoning_effort is treated as not forwarded (the documented rejection)")
	}
	forwards := func(root string) bool {
		if !standard[root] {
			return true
		}
		if known {
			return listed[root]
		}
		return root != "reasoning_effort"
	}

	intents := make([]string, 0, len(api.Bindings))
	for intent := range api.Bindings {
		intents = append(intents, intent)
	}
	sort.Strings(intents)
	for _, intent := range intents {
		root, _, _ := strings.Cut(api.Bindings[intent].Param, ".")
		if forwards(root) {
			continue
		}
		delete(api.Bindings, intent)
		tracef("bindings: %s dropped — litellm does not forward %s for this model", intent, root)
	}
	if len(api.Bindings) == 0 {
		api.Bindings = nil
	}

	if api.Stack == "anthropic" {
		if !known {
			api.Parameters = nil
			tracef("parameters: litellm supported_openai_params unavailable — anthropic's documented list is Messages vocabulary and cannot be intersected with it; no parameters reported")
			return
		}
		kept := append([]string(nil), supported...)
		sort.Strings(kept)
		if len(kept) == 0 {
			kept = nil
		}
		api.Parameters = kept
		tracef("parameters: litellm's supported_openai_params reported verbatim (%d) — anthropic's documented list is Messages vocabulary and is not intersected with it", len(kept))
		return
	}

	kept := make([]string, 0, len(api.Parameters))
	for _, p := range api.Parameters {
		if (known && listed[p]) || (!known && p != "reasoning_effort") {
			kept = append(kept, p)
		}
	}
	if dropped := len(api.Parameters) - len(kept); dropped > 0 {
		if known {
			tracef("parameters: %d of %d dropped — not in litellm's supported_openai_params for this model; %d kept", dropped, len(api.Parameters), len(kept))
		} else {
			tracef("parameters: reasoning_effort dropped — litellm supported_openai_params unavailable")
		}
	}
	sort.Strings(kept)
	if len(kept) == 0 {
		kept = nil
	}
	api.Parameters = kept
}

func cloneSet(b bindingSet) bindingSet {
	out := bindingSet{}
	for intent, pb := range b {
		out[intent] = cloneBinding(pb)
	}
	return out
}

func cloneBinding(b *system.ParamBinding) *system.ParamBinding {
	return b.Clone()
}
