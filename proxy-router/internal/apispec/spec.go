// Package apispec composes a provider-declared API spec (system.ModelApiSpec)
// for a configured model from static, documentation-verified knowledge: the
// apiType stack preset and the model family. No network access.
package apispec

import (
	"strings"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/config"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/system"
)

// Build composes the spec for cfg. It returns nil when the apiType has no
// chat API spec (image adapters, unknown values). DeclaredAt is left for the
// caller to stamp.
func Build(cfg config.ModelConfig) *system.ModelApiSpec {
	stack := StackFor(cfg.ApiType)
	if stack == "" {
		return nil
	}
	api := &system.ModelApiSpec{Stack: stack}

	family := strings.ToLower(strings.TrimSpace(cfg.ModelFamily))
	if family == "" {
		family = FamilyFromName(cfg.ModelName)
	}
	api.ModelFamily = family

	// Family layer: reasoning toggles. Skipped on gateway/hosted presets
	// (their own knobs come from the stack table) unless the family's native
	// vendor is this stack.
	alwaysOn, defaults := bindingsForFamily(family, cfg.ModelName)
	familyReasons := alwaysOn || len(defaults) > 0
	var bindings bindingSet
	switch {
	case defaults == nil:
	case gatewayStacks[stack] && familyNativeVendor[family] != stack:
	default:
		bindings = cloneSet(defaults)
	}

	if stack == "ollama" && bindings != nil {
		rewriteBindingsForOllama(bindings)
	}

	mergeStackTables(api, stack, bindings, familyReasons, alwaysOn)

	switch {
	case alwaysOn:
		api.Thinking = &system.ThinkingSpec{Mode: system.ThinkingModeAlwaysOn}
	case api.Bindings[system.IntentReasoningDisable] != nil, api.Bindings[system.IntentReasoningEnable] != nil:
		api.Thinking = &system.ThinkingSpec{Mode: system.ThinkingModeControllable}
	case api.Bindings[system.IntentReasoningEffort] != nil || api.Bindings[system.IntentReasoningBudget] != nil:
		api.Thinking = &system.ThinkingSpec{Mode: system.ThinkingModeTunable}
	}
	return api
}

// mergeStackTables sets api.Bindings to the family bindings plus the stack's
// documented bindings for every intent the family did not set. Reasoning
// intents from the stack merge only when the family is known to reason;
// for always-on families only reasoning.format merges. api.Parameters is
// the stack's documented list. Everything is copied so the shared tables
// are never aliased.
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
	if params := stackParameters[stack]; len(params) > 0 {
		api.Parameters = append([]string(nil), params...)
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
