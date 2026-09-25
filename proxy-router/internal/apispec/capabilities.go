package apispec

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/config"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/system"
)

const (
	SourceDeclared = "declared"
	// Sources of a capability the composer derived rather than read: the
	// thinking mode or a surviving reasoning.* binding proving reasoning.
	SourceDerivedThinkingMode = "derived:thinking.mode"
	SourceDerivedBindings     = "derived:bindings.reasoning"
)

// A capability layer is one precedence tier: declared > backend > template >
// registry. A field is set by the first layer that has it.
type capLayer struct {
	caps   system.ModelCapabilities
	source func(field string) string
}

func declaredLayer(d *config.DeclaredCapabilities) *capLayer {
	if d == nil {
		return nil
	}
	caps := system.ModelCapabilities{
		Context: d.Context, MaxOutput: d.MaxOutput, Tools: d.Tools, Structured: d.Structured,
		WebSearch: d.WebSearch, Embeddings: d.Embeddings,
	}
	if d.Modalities != nil {
		caps.Modalities = &system.Modalities{Input: d.Modalities.Input, Output: d.Modalities.Output}
	}
	return &capLayer{caps: caps, source: func(string) string { return SourceDeclared }}
}

func factsLayer(f *CapabilityFacts) *capLayer {
	if f == nil || f.Empty() {
		return nil
	}
	return &capLayer{caps: f.ModelCapabilities, source: func(field string) string { return f.Sources[field] }}
}

func registryLayer(r *RegistryFacts) *capLayer {
	if r == nil {
		return nil
	}
	src := r.Source()
	return &capLayer{caps: r.ModelCapabilities, source: func(string) string { return src }}
}

// capState is the capabilities block under construction plus the layer each
// field came from (rule 10 needs to know a served context is the backend's).
type capState struct {
	caps     system.ModelCapabilities
	evidence map[string]string
	layerOf  map[string]string
}

func (s *capState) set(field, value, source, layer string, tracef func(string, ...any)) {
	if s.evidence == nil {
		s.evidence = map[string]string{}
		s.layerOf = map[string]string{}
	}
	s.evidence[system.EvidenceKeyCapability(field)] = source
	s.layerOf[field] = layer
	tracef("capabilities.%s = %s (%s)", field, value, source)
}

func (s *capState) isSet(field string) bool {
	_, ok := s.layerOf[field]
	return ok
}

// fill applies one layer field by field to the fields still unset.
func (s *capState) fill(layer *capLayer, name string, tracef func(string, ...any)) {
	if layer == nil {
		return
	}
	src := layer.caps
	// Rule 2: served beats trained; a trained window alone stands in for the
	// served one under its own source, and is then not repeated as
	// contextTrained.
	promoted := false
	if !s.isSet(CapFieldContext) && src.Context == nil && src.ContextTrained != nil {
		src.Context = src.ContextTrained
		src.ContextTrained = nil
		promoted = true
	}
	for _, field := range capFields {
		if s.isSet(field) {
			continue
		}
		source := layer.source(field)
		if field == CapFieldContext && promoted {
			source = layer.source(CapFieldContextTrained)
		}
		if source == "" {
			continue
		}
		value, ok := s.take(field, &src, name, tracef)
		if !ok {
			continue
		}
		s.set(field, value, source, name, tracef)
	}
}

// take copies one field from src when it is present and valid; the returned
// string is the trace rendering.
func (s *capState) take(field string, src *system.ModelCapabilities, layer string, tracef func(string, ...any)) (string, bool) {
	takeInt := func(dst **int64, v *int64) (string, bool) {
		if v == nil {
			return "", false
		}
		if !config.ValidCapabilityInt(*v) {
			tracef("capabilities.%s: %s value %d out of (0, %d); ignored", field, layer, *v, config.MaxCapabilityInt)
			return "", false
		}
		n := *v
		*dst = &n
		return strconv.FormatInt(n, 10), true
	}
	takeBool := func(dst **bool, v *bool) (string, bool) {
		if v == nil {
			return "", false
		}
		b := *v
		*dst = &b
		return strconv.FormatBool(b), true
	}
	takeList := func(dst *[]string, v []string, valid func(string) bool) (string, bool) {
		kept := filterVocabulary(v, valid)
		if kept == nil {
			if len(v) > 0 {
				tracef("capabilities.%s: %s value %v has no known entry; ignored", field, layer, v)
			}
			return "", false
		}
		if len(kept) != len(v) {
			tracef("capabilities.%s: %s value %v narrowed to the known entries %v", field, layer, v, kept)
		}
		*dst = kept
		return strings.Join(kept, "|"), true
	}
	dst := &s.caps
	switch field {
	case CapFieldContext:
		return takeInt(&dst.Context, src.Context)
	case CapFieldContextTrained:
		return takeInt(&dst.ContextTrained, src.ContextTrained)
	case CapFieldMaxOutput:
		return takeInt(&dst.MaxOutput, src.MaxOutput)
	case CapFieldModalitiesInput:
		if src.Modalities == nil {
			return "", false
		}
		var list []string
		text, ok := takeList(&list, src.Modalities.Input, config.IsModalityInput)
		if ok {
			s.modalities().Input = list
		}
		return text, ok
	case CapFieldModalitiesOutput:
		if src.Modalities == nil {
			return "", false
		}
		var list []string
		text, ok := takeList(&list, src.Modalities.Output, config.IsModalityOutput)
		if ok {
			s.modalities().Output = list
		}
		return text, ok
	case CapFieldTools:
		return takeBool(&dst.Tools, src.Tools)
	case CapFieldToolsParallel:
		return takeBool(&dst.ToolsParallel, src.ToolsParallel)
	case CapFieldStructured:
		return takeList(&dst.Structured, src.Structured, config.IsStructuredKind)
	case CapFieldReasoning:
		return takeBool(&dst.Reasoning, src.Reasoning)
	case CapFieldWebSearch:
		return takeBool(&dst.WebSearch, src.WebSearch)
	case CapFieldEmbeddings:
		return takeBool(&dst.Embeddings, src.Embeddings)
	case CapFieldPromptCache:
		return takeBool(&dst.PromptCache, src.PromptCache)
	case CapFieldQuantization:
		if src.Quantization == "" {
			return "", false
		}
		if !nameRe.MatchString(src.Quantization) {
			tracef("capabilities.quantization: %s value is not a plain label; ignored", layer)
			return "", false
		}
		dst.Quantization = src.Quantization
		return src.Quantization, true
	}
	return "", false
}

func (s *capState) modalities() *system.Modalities {
	if s.caps.Modalities == nil {
		s.caps.Modalities = &system.Modalities{}
	}
	return s.caps.Modalities
}

// Rule 3: reasoning is true when the mode is always_on or a reasoning.*
// binding survived every filter; false only when a layer reported false.
func (s *capState) deriveReasoning(alwaysOn bool, bindings map[string]*system.ParamBinding, tracef func(string, ...any)) {
	if s.isSet(CapFieldReasoning) {
		return
	}
	t := true
	switch {
	case alwaysOn:
		s.caps.Reasoning = &t
		s.set(CapFieldReasoning, "true", SourceDerivedThinkingMode, "derived", tracef)
	case hasReasoningBindings(bindings):
		s.caps.Reasoning = &t
		s.set(CapFieldReasoning, "true", SourceDerivedBindings, "derived", tracef)
	}
}

// Rule 10, second half: a served window far below the registry's documented
// one is worth a look (Ollama num_ctx defaults, --max-model-len), but is not
// an error.
func (s *capState) noteContextGap(registry *RegistryFacts, tracef func(string, ...any)) {
	if registry == nil || registry.Context == nil || !config.ValidCapabilityInt(*registry.Context) {
		return
	}
	if s.layerOf[CapFieldContext] != layerBackend || s.caps.Context == nil {
		return
	}
	if served, documented := *s.caps.Context, *registry.Context; served*4 < documented {
		tracef("context: served window %d is far below the registry's %d", served, documented)
	}
}

func (s *capState) result() (*system.ModelCapabilities, map[string]string) {
	if len(s.evidence) == 0 {
		return nil, nil
	}
	caps := s.caps
	return &caps, s.evidence
}

const (
	layerDeclared = "declared"
	layerBackend  = "backend"
	layerTemplate = "template"
	layerRegistry = "registry"
)

func copyCapField(dst, src *system.ModelCapabilities, field string) {
	switch field {
	case CapFieldContext:
		dst.Context = cloneInt64(src.Context)
	case CapFieldContextTrained:
		dst.ContextTrained = cloneInt64(src.ContextTrained)
	case CapFieldMaxOutput:
		dst.MaxOutput = cloneInt64(src.MaxOutput)
	case CapFieldModalitiesInput:
		if src.Modalities != nil {
			if dst.Modalities == nil {
				dst.Modalities = &system.Modalities{}
			}
			dst.Modalities.Input = append([]string(nil), src.Modalities.Input...)
		}
	case CapFieldModalitiesOutput:
		if src.Modalities != nil {
			if dst.Modalities == nil {
				dst.Modalities = &system.Modalities{}
			}
			dst.Modalities.Output = append([]string(nil), src.Modalities.Output...)
		}
	case CapFieldTools:
		dst.Tools = cloneBool(src.Tools)
	case CapFieldToolsParallel:
		dst.ToolsParallel = cloneBool(src.ToolsParallel)
	case CapFieldStructured:
		dst.Structured = append([]string(nil), src.Structured...)
	case CapFieldReasoning:
		dst.Reasoning = cloneBool(src.Reasoning)
	case CapFieldWebSearch:
		dst.WebSearch = cloneBool(src.WebSearch)
	case CapFieldEmbeddings:
		dst.Embeddings = cloneBool(src.Embeddings)
	case CapFieldPromptCache:
		dst.PromptCache = cloneBool(src.PromptCache)
	case CapFieldQuantization:
		dst.Quantization = src.Quantization
	}
}

func clearCapField(caps *system.ModelCapabilities, field string) {
	switch field {
	case CapFieldContext:
		caps.Context = nil
	case CapFieldContextTrained:
		caps.ContextTrained = nil
	case CapFieldMaxOutput:
		caps.MaxOutput = nil
	case CapFieldModalitiesInput:
		if caps.Modalities != nil {
			caps.Modalities.Input = nil
		}
	case CapFieldModalitiesOutput:
		if caps.Modalities != nil {
			caps.Modalities.Output = nil
		}
	case CapFieldTools:
		caps.Tools = nil
	case CapFieldToolsParallel:
		caps.ToolsParallel = nil
	case CapFieldStructured:
		caps.Structured = nil
	case CapFieldReasoning:
		caps.Reasoning = nil
	case CapFieldWebSearch:
		caps.WebSearch = nil
	case CapFieldEmbeddings:
		caps.Embeddings = nil
	case CapFieldPromptCache:
		caps.PromptCache = nil
	case CapFieldQuantization:
		caps.Quantization = ""
	}
	if caps.Modalities != nil && caps.Modalities.Input == nil && caps.Modalities.Output == nil {
		caps.Modalities = nil
	}
}

func emptyCapabilities(c *system.ModelCapabilities) bool {
	return c.Context == nil && c.ContextTrained == nil && c.MaxOutput == nil && c.Modalities == nil &&
		c.Tools == nil && c.ToolsParallel == nil && c.Structured == nil && c.Reasoning == nil &&
		c.WebSearch == nil && c.Embeddings == nil && c.PromptCache == nil && c.Quantization == ""
}

func cloneInt64(p *int64) *int64 {
	if p == nil {
		return nil
	}
	v := *p
	return &v
}

func cloneBool(p *bool) *bool {
	if p == nil {
		return nil
	}
	v := *p
	return &v
}

// IsDeclaredOrRegistrySource is the default keep rule of
// ClearDetectedCapabilities: facts that did not come from the dropped
// stack's endpoints.
func IsDeclaredOrRegistrySource(source string) bool {
	return source == SourceDeclared || strings.HasPrefix(source, "registry:")
}

// ClearDetectedCapabilities is the capabilities half of the transport-
// consistency drop (rule 12): every capability, thinking.default and binding
// evidence entry whose source keep rejects is removed together with its
// value. keep nil means IsDeclaredOrRegistrySource. Bindings themselves are
// the caller's to drop; their evidence keys are cleared here so none dangles.
func ClearDetectedCapabilities(api *system.ModelApiSpec, keep func(source string) bool) {
	if api == nil || len(api.Evidence) == 0 {
		return
	}
	if keep == nil {
		keep = IsDeclaredOrRegistrySource
	}
	keys := make([]string, 0, len(api.Evidence))
	for k := range api.Evidence {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, key := range keys {
		if keep(api.Evidence[key]) {
			continue
		}
		switch {
		case strings.HasPrefix(key, "capabilities."):
			if api.Capabilities != nil {
				clearCapField(api.Capabilities, strings.TrimPrefix(key, "capabilities."))
			}
		case key == system.EvidenceKeyThinkingDefault:
			if api.Thinking != nil {
				api.Thinking.Default = ""
			}
		}
		delete(api.Evidence, key)
	}
	if api.Capabilities != nil && emptyCapabilities(api.Capabilities) {
		api.Capabilities = nil
	}
	if len(api.Evidence) == 0 {
		api.Evidence = nil
	}
}

// Rule 6: bindings from Anthropic's reported capabilities, replacing the
// name-regex guess. Sources: https://platform.claude.com/docs/en/api/models/retrieve,
// https://platform.claude.com/docs/en/build-with-claude/thinking,
// https://platform.claude.com/docs/en/build-with-claude/effort
func anthropicBindings(f *AnthropicFacts) bindingSet {
	b := bindingSet{}
	if f.ThinkingSupported {
		b[system.IntentReasoningDisable] = &system.ParamBinding{Kind: system.BindingKindBodyParam, Param: "thinking", ParamType: "object", Value: map[string]any{"type": "disabled"}}
		switch {
		case containsString(f.ThinkingTypes, "adaptive"):
			b[system.IntentReasoningEnable] = &system.ParamBinding{Kind: system.BindingKindBodyParam, Param: "thinking", ParamType: "object", Value: map[string]any{"type": "adaptive"}}
		case containsString(f.ThinkingTypes, "enabled"):
			b[system.IntentReasoningEnable] = &system.ParamBinding{Kind: system.BindingKindBodyParam, Param: "thinking", ParamType: "object", Value: map[string]any{"type": "enabled", "budget_tokens": 1024}, Hint: "budget_tokens >= 1024 and below max_tokens"}
			b[system.IntentReasoningBudget] = &system.ParamBinding{Kind: system.BindingKindBodyParam, Param: "thinking.budget_tokens", ParamType: "number", Hint: "minimum 1024; must be below max_tokens"}
		}
	}
	if levels := filterVocabulary(f.EffortLevels, nameRe.MatchString); levels != nil {
		b[system.IntentReasoningEffort] = &system.ParamBinding{Kind: system.BindingKindBodyParam, Param: "output_config.effort", ParamType: "enum", EnumValues: levels}
	}
	return b
}

func describeAnthropicFacts(f *AnthropicFacts) string {
	return fmt.Sprintf("thinking.supported=%t types=%v effort=%v structured_outputs=%t", f.ThinkingSupported, f.ThinkingTypes, f.EffortLevels, f.Structured)
}

// Rule 7: on, off or a level name.
func validThinkingDefault(v string) bool {
	return v == "on" || v == "off" || nameRe.MatchString(v)
}
