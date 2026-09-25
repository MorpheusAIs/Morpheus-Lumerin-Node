package apispec

import (
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/config"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/system"
)

// Wire field names of system.ModelCapabilities; the evidence key is
// system.EvidenceKeyCapability(field).
const (
	CapFieldContext          = "context"
	CapFieldContextTrained   = "contextTrained"
	CapFieldMaxOutput        = "maxOutput"
	CapFieldModalitiesInput  = "modalities.input"
	CapFieldModalitiesOutput = "modalities.output"
	CapFieldTools            = "tools"
	CapFieldToolsParallel    = "toolsParallel"
	CapFieldStructured       = "structured"
	CapFieldReasoning        = "reasoning"
	CapFieldWebSearch        = "webSearch"
	CapFieldEmbeddings       = "embeddings"
	CapFieldPromptCache      = "promptCache"
	CapFieldQuantization     = "quantization"
)

// Composition order of the fields; also the trace order.
var capFields = []string{
	CapFieldContext, CapFieldContextTrained, CapFieldMaxOutput,
	CapFieldModalitiesInput, CapFieldModalitiesOutput,
	CapFieldTools, CapFieldToolsParallel, CapFieldStructured,
	CapFieldReasoning, CapFieldWebSearch, CapFieldEmbeddings, CapFieldPromptCache,
	CapFieldQuantization,
}

// CapabilityFacts carries backend- or vendor-reported capability values with,
// per field, the source that becomes the evidence entry
// ("<stack>:<endpoint>.<field>", e.g. "vllm:/v1/models.max_model_len").
// The zero value is empty and usable; setters validate (integers in
// (0, 1<<31), lists against the wire vocabularies) and return whether the
// value was accepted. A later Set of the same field overwrites: the probe
// decides its own precedence between endpoints.
type CapabilityFacts struct {
	system.ModelCapabilities
	Sources map[string]string
}

func (c *CapabilityFacts) record(field, source string) {
	if c.Sources == nil {
		c.Sources = map[string]string{}
	}
	c.Sources[field] = source
}

func (c *CapabilityFacts) Empty() bool {
	return len(c.Sources) == 0
}

func (c *CapabilityFacts) SetContext(v int64, source string) bool {
	if !config.ValidCapabilityInt(v) {
		return false
	}
	c.Context = &v
	c.record(CapFieldContext, source)
	return true
}

func (c *CapabilityFacts) SetContextTrained(v int64, source string) bool {
	if !config.ValidCapabilityInt(v) {
		return false
	}
	c.ContextTrained = &v
	c.record(CapFieldContextTrained, source)
	return true
}

func (c *CapabilityFacts) SetMaxOutput(v int64, source string) bool {
	if !config.ValidCapabilityInt(v) {
		return false
	}
	c.MaxOutput = &v
	c.record(CapFieldMaxOutput, source)
	return true
}

func (c *CapabilityFacts) SetTools(v bool, source string) {
	c.Tools = &v
	c.record(CapFieldTools, source)
}

func (c *CapabilityFacts) SetToolsParallel(v bool, source string) {
	c.ToolsParallel = &v
	c.record(CapFieldToolsParallel, source)
}

func (c *CapabilityFacts) SetReasoning(v bool, source string) {
	c.Reasoning = &v
	c.record(CapFieldReasoning, source)
}

func (c *CapabilityFacts) SetWebSearch(v bool, source string) {
	c.WebSearch = &v
	c.record(CapFieldWebSearch, source)
}

func (c *CapabilityFacts) SetEmbeddings(v bool, source string) {
	c.Embeddings = &v
	c.record(CapFieldEmbeddings, source)
}

func (c *CapabilityFacts) SetPromptCache(v bool, source string) {
	c.PromptCache = &v
	c.record(CapFieldPromptCache, source)
}

// SetInputModalities replaces the list with the vocabulary-filtered, deduped
// values (order kept); nothing valid leaves the field untouched.
func (c *CapabilityFacts) SetInputModalities(list []string, source string) bool {
	kept := filterVocabulary(list, config.IsModalityInput)
	if kept == nil {
		return false
	}
	c.modalities().Input = kept
	c.record(CapFieldModalitiesInput, source)
	return true
}

func (c *CapabilityFacts) SetOutputModalities(list []string, source string) bool {
	kept := filterVocabulary(list, config.IsModalityOutput)
	if kept == nil {
		return false
	}
	c.modalities().Output = kept
	c.record(CapFieldModalitiesOutput, source)
	return true
}

// AddInputModality appends one value when it is in the vocabulary and not
// yet listed; the source of the last accepted add is the field's source.
func (c *CapabilityFacts) AddInputModality(m, source string) bool {
	if !config.IsModalityInput(m) {
		return false
	}
	mod := c.modalities()
	if !containsString(mod.Input, m) {
		mod.Input = append(mod.Input, m)
	}
	c.record(CapFieldModalitiesInput, source)
	return true
}

func (c *CapabilityFacts) AddOutputModality(m, source string) bool {
	if !config.IsModalityOutput(m) {
		return false
	}
	mod := c.modalities()
	if !containsString(mod.Output, m) {
		mod.Output = append(mod.Output, m)
	}
	c.record(CapFieldModalitiesOutput, source)
	return true
}

func (c *CapabilityFacts) modalities() *system.Modalities {
	if c.Modalities == nil {
		c.Modalities = &system.Modalities{}
	}
	return c.Modalities
}

func (c *CapabilityFacts) SetStructured(list []string, source string) bool {
	kept := filterVocabulary(list, config.IsStructuredKind)
	if kept == nil {
		return false
	}
	c.Structured = kept
	c.record(CapFieldStructured, source)
	return true
}

// SetQuantization keeps the backend's own label when it is a plain name
// (fp8, Q4_K_M, bf16, MXFP4).
func (c *CapabilityFacts) SetQuantization(label, source string) bool {
	if !nameRe.MatchString(label) {
		return false
	}
	c.Quantization = label
	c.record(CapFieldQuantization, source)
	return true
}

// FillFrom copies the fields of other that are still unset here (one carrier
// per host; the caller orders hops by precedence).
func (c *CapabilityFacts) FillFrom(other *CapabilityFacts) {
	if other == nil {
		return
	}
	for field, source := range other.Sources {
		if _, set := c.Sources[field]; set {
			continue
		}
		copyCapField(&c.ModelCapabilities, &other.ModelCapabilities, field)
		c.record(field, source)
	}
}

func filterVocabulary(list []string, valid func(string) bool) []string {
	var kept []string
	for _, v := range list {
		if valid(v) && !containsString(kept, v) {
			kept = append(kept, v)
		}
	}
	return kept
}

// AnthropicFacts are the model's capabilities as reported by Anthropic's
// GET /v1/models/{id} (https://platform.claude.com/docs/en/api/models/retrieve):
// capabilities.thinking{supported, types{adaptive, enabled}}, effort{low..max},
// structured_outputs. Present facts replace the name-regex guess in
// claudeBindings.
type AnthropicFacts struct {
	ThinkingSupported bool
	// ThinkingTypes is the subset of adaptive, enabled with supported: true.
	ThinkingTypes []string
	// EffortLevels is the subset of low, medium, high, xhigh, max with
	// supported: true, in that order.
	EffortLevels []string
	// Structured is structured_outputs.supported.
	Structured bool
	// Source is the evidence entry for everything derived from these facts,
	// e.g. "anthropic:/v1/models/{id}.capabilities".
	Source string
}
