// Package apispec composes a provider's API spec (system.ModelApiSpec) for a
// configured model from documentation-verified knowledge: the apiStack preset
// tables, the model-family defaults and — when the runtime detector
// (internal/apidetect) supplies it — backend evidence. No network access.
package apispec

import (
	"fmt"
	"sort"
	"strings"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/config"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/system"
)

// Evidence is everything the composer may know about one model backend.
// Build fills only the declared part (Stack from apiStack, ModelName,
// ModelFamily, Declared); the runtime detector fills the rest from probes.
type Evidence struct {
	// Stack is the apiStack preset or the detected stack / hosted vendor
	// name; "" means undetermined — then no bindings are advertised at all
	// (the wire vocabulary is unknown; it could be a gateway).
	Stack string
	// ModelName is the configured modelName (the local alias).
	ModelName string
	// ModelFamily is the explicit models-config modelFamily. When set it is
	// never overridden by evidence; evidence only fills an empty one.
	ModelFamily string
	// ServedModelID is the backend's own id/path for the served model
	// (SGLang model_path, TGI model_id, a /v1/models entry id, LiteLLM's
	// upstream model id).
	ServedModelID string
	// Architecture is the model architecture the backend reports directly
	// (Ollama general.architecture, LM Studio arch).
	Architecture string
	// ChatTemplate is the Jinja chat template source when the backend
	// exposes it (llama.cpp /props, Ollama /api/show).
	ChatTemplate string
	// OllamaThinking is true when Ollama lists "thinking" among the model's
	// capabilities.
	OllamaThinking bool
	// GatewayReasoning is true when a gateway or registry listing declares
	// that the model reasons (LiteLLM supports_reasoning, Venice
	// supportsReasoning, OpenRouter supported_parameters with "reasoning").
	GatewayReasoning bool
	// RegistryBindings are per-model bindings imported from a registry
	// listing; they win over every other binding source. No probe fills
	// them on this branch (OpenRouter's knobs live in the openrouter stack
	// table and are switched on by GatewayReasoning); kept for imports that
	// carry model-specific shapes.
	RegistryBindings map[string]*system.ParamBinding
	// Parameters is a per-model supported-parameter list imported from the
	// backend (LiteLLM supported_openai_params, OpenRouter
	// supported_parameters, Venice capabilities). It replaces the stack's
	// documented list.
	Parameters []string
	// ReasoningEfforts is LiteLLM's supported_reasoning_efforts for the
	// model group: the litellm reasoning.effort enum for this model. When it
	// lacks "none", thinking cannot be switched off through LiteLLM.
	ReasoningEfforts []string
	// Declared is true when Stack comes from models-config apiStack (Build)
	// rather than probing. A declared gateway preset merges its reasoning
	// knobs on family knowledge alone; a detected gateway only on probe
	// evidence, since a gateway may hide its upstream's knob.
	Declared bool
}

// Build composes the spec declared by cfg.ApiStack. It returns nil when
// cfg.ApiStack is empty or unknown: the declared api block is opt-in per
// model and apiType (the transport adapter) plays no part in it. DeclaredAt
// is left for the caller to stamp. Models without apiStack are handled by the
// runtime detector, which calls Compose with what it observed.
func Build(cfg config.ModelConfig) *system.ModelApiSpec {
	stack := StackFor(cfg.ApiStack)
	if stack == "" {
		return nil
	}
	return Compose(Evidence{Stack: stack, ModelName: cfg.ModelName, ModelFamily: cfg.ModelFamily, Declared: true})
}

// Compose turns evidence into the spec. Precedence: the explicit family wins,
// then backend architecture > served id > configured name (a more specific
// variant refines a generic one); bindings: registry import > Ollama
// capability > chat-template evidence > family default (gated for gateways /
// hosted vendors and skipped for an undetermined stack) > Ollama rewrite >
// stack tables (reasoning knobs only when the model is known to reason) >
// thinking mode. Returns nil when nothing at all is known.
func Compose(ev Evidence) *system.ModelApiSpec {
	api, _ := ComposeWithTrace(ev)
	return api
}

// ComposeWithTrace is Compose plus the human-readable decisions that
// produced the spec (which evidence decided the family, the bindings and the
// thinking knob), for diagnostics traces.
func ComposeWithTrace(ev Evidence) (*system.ModelApiSpec, []string) {
	var lines []string
	tracef := func(format string, args ...any) { lines = append(lines, fmt.Sprintf(format, args...)) }

	stack := ev.Stack
	api := &system.ModelApiSpec{Stack: stack}

	// Family: the explicit declaration is never overridden; otherwise direct
	// backend evidence beats name heuristics and the backend's own model id
	// beats the local alias. Only canonical families are published: an
	// architecture that maps to no family never short-circuits the names.
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

	// The name the family heuristics run against is the most
	// backend-specific one available.
	name := ev.ModelName
	if ev.ServedModelID != "" {
		name = ev.ServedModelID
	}

	// Bindings precedence: registry import > Ollama capability > chat
	// template evidence > family default. alwaysOn marks families that
	// reason unconditionally (no toggle can exist for them).
	alwaysOn := false
	var bindings bindingSet
	if len(ev.RegistryBindings) > 0 {
		bindings = cloneSet(ev.RegistryBindings)
		tracef("bindings: from the %s model listing", stack)
	}
	if len(bindings) == 0 && ev.OllamaThinking {
		// The capability only says the model thinks. Whether it can be
		// toggled is a family fact: always-on families stay always-on, and
		// effort-only families (gpt-oss) ignore Ollama's boolean toggle and
		// take levels instead.
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
		if bindings = bindingsFromTemplate(ev.ChatTemplate); bindings != nil {
			tracef("bindings: from chat template source -> %s", describeBindings(bindings))
		}
	}
	// familyReasons: the family is known to reason (a default set exists or
	// it is always-on), whether or not the default applies to this stack.
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
			// Family defaults describe either HF chat-template kwargs (only
			// honored by self-hosted engines) or the family's native vendor
			// API. On any other hosted vendor or gateway they are wrong
			// request shapes; gateways contribute their own knobs via the
			// stack table.
			tracef("bindings: family default for %q skipped — %q does not accept it", family, stack)
		default:
			bindings = cloneSet(defaults)
			tracef("bindings: family default for %q (model %q) -> %s", family, name, describeBindings(bindings))
		}
	}

	// Ollama ignores chat_template_kwargs on its API: its own knobs are what
	// count regardless of what the template suggests.
	if stack == "ollama" && bindings != nil {
		before := describeBindings(bindings)
		rewriteBindingsForOllama(bindings)
		if after := describeBindings(bindings); after != before {
			tracef("bindings: rewritten onto ollama's /v1 reasoning_effort (chat_template_kwargs is not honored by ollama) -> %s", after)
		}
	}

	// Stack-documented remaps fill every intent direct evidence did not
	// establish. Reasoning knobs merge only when the model is known to
	// reason: direct evidence, or — for a declared stack — family knowledge
	// alone (a detected gateway may hide its upstream's knob, so probe
	// evidence is required there).
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
	if len(api.Bindings) == 0 && !alwaysOn {
		tracef("bindings: no remappable knobs known for this backend")
	}

	// LiteLLM's per-model supported_reasoning_efforts narrows the litellm
	// reasoning.effort enum; without "none" thinking cannot be switched off
	// through LiteLLM, so the reasoning.disable knob (value none) goes. The
	// reasoning.enable knob takes the first supported effort other than
	// none (LiteLLM's order) and goes too when there is no such effort.
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

	// Thinking mode summary is derived from what the bindings can express.
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

// hasReasoningBindings reports whether any reasoning.* intent is bound.
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

// mergeStackTables sets api.Bindings to the family bindings plus the stack's
// documented bindings for every intent the family did not set. Reasoning
// intents from the stack merge only when the family is known to reason;
// for always-on families only reasoning.format merges. api.Parameters is
// the stack's documented list unless a per-model list was already imported
// (registry/gateway listings are authoritative). Everything is copied so the
// shared tables are never aliased.
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

func cloneSet(b bindingSet) bindingSet {
	out := bindingSet{}
	for intent, pb := range b {
		out[intent] = cloneBinding(pb)
	}
	return out
}

func cloneBinding(b *system.ParamBinding) *system.ParamBinding {
	c := *b
	if b.EnumValues != nil {
		c.EnumValues = append([]string(nil), b.EnumValues...)
	}
	if m, ok := b.Value.(map[string]any); ok {
		mc := make(map[string]any, len(m))
		for k, v := range m {
			mc[k] = v
		}
		c.Value = mc
	}
	return &c
}
