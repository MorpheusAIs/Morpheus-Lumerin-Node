package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
)

// Wire vocabularies shared by the declared override and the composed api
// block (system.Modalities / ModelCapabilities.Structured). They live here
// because config cannot import system.
var (
	ModalityInputs  = []string{"text", "image", "audio", "video", "pdf"}
	ModalityOutputs = []string{"text", "image", "audio", "embeddings"}
	StructuredKinds = []string{"json_object", "json_schema", "grammar", "regex", "choice"}
)

// Integers reported by backends or declared are accepted only in (0, 1<<31).
const MaxCapabilityInt = int64(1) << 31

func ValidCapabilityInt(v int64) bool { return v > 0 && v < MaxCapabilityInt }

func IsModalityInput(m string) bool  { return containsName(ModalityInputs, m) }
func IsModalityOutput(m string) bool { return containsName(ModalityOutputs, m) }
func IsStructuredKind(s string) bool { return containsName(StructuredKinds, s) }

func containsName(list []string, s string) bool {
	for _, v := range list {
		if v == s {
			return true
		}
	}
	return false
}

// DeclaredCapabilities is the models-config.json "capabilities" override: the
// wire subset a provider can state without probing (no contextTrained,
// reasoning, promptCache, quantization, toolsParallel). Declared values win
// field by field over detected and registry facts; their evidence is
// "declared".
type DeclaredCapabilities struct {
	Context    *int64              `json:"context,omitempty"`
	MaxOutput  *int64              `json:"maxOutput,omitempty"`
	Modalities *DeclaredModalities `json:"modalities,omitempty"`
	Tools      *bool               `json:"tools,omitempty"`
	Structured []string            `json:"structured,omitempty"`
	WebSearch  *bool               `json:"webSearch,omitempty"`
	Embeddings *bool               `json:"embeddings,omitempty"`

	// parseErr keeps a malformed block from failing the whole load: the
	// loader reports it through Validate and clears the block for that model.
	parseErr error
}

type DeclaredModalities struct {
	Input  []string `json:"input,omitempty"`
	Output []string `json:"output,omitempty"`
}

// UnmarshalJSON decodes strictly (unknown keys and type mismatches are
// errors) but records the error instead of returning it, so one model's bad
// block never empties a serving node (same policy as apiStack).
func (d *DeclaredCapabilities) UnmarshalJSON(raw []byte) error {
	type plain DeclaredCapabilities
	var p plain
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&p); err != nil {
		*d = DeclaredCapabilities{parseErr: err}
		return nil
	}
	*d = DeclaredCapabilities(p)
	return nil
}

func (d *DeclaredCapabilities) Clone() *DeclaredCapabilities {
	if d == nil {
		return nil
	}
	c := *d
	if d.Context != nil {
		v := *d.Context
		c.Context = &v
	}
	if d.MaxOutput != nil {
		v := *d.MaxOutput
		c.MaxOutput = &v
	}
	if d.Tools != nil {
		v := *d.Tools
		c.Tools = &v
	}
	if d.WebSearch != nil {
		v := *d.WebSearch
		c.WebSearch = &v
	}
	if d.Embeddings != nil {
		v := *d.Embeddings
		c.Embeddings = &v
	}
	if d.Structured != nil {
		c.Structured = append([]string(nil), d.Structured...)
	}
	if d.Modalities != nil {
		m := DeclaredModalities{}
		if d.Modalities.Input != nil {
			m.Input = append([]string(nil), d.Modalities.Input...)
		}
		if d.Modalities.Output != nil {
			m.Output = append([]string(nil), d.Modalities.Output...)
		}
		c.Modalities = &m
	}
	return &c
}

func ValidateDeclaredCapabilities(modelID string, caps *DeclaredCapabilities) error {
	if caps == nil {
		return nil
	}
	if caps.parseErr != nil {
		return fmt.Errorf("model %s: invalid capabilities: %w", modelID, caps.parseErr)
	}
	if caps.Context != nil && !ValidCapabilityInt(*caps.Context) {
		return fmt.Errorf("model %s: invalid capabilities: context must be in (0, %d), got %d", modelID, MaxCapabilityInt, *caps.Context)
	}
	if caps.MaxOutput != nil && !ValidCapabilityInt(*caps.MaxOutput) {
		return fmt.Errorf("model %s: invalid capabilities: maxOutput must be in (0, %d), got %d", modelID, MaxCapabilityInt, *caps.MaxOutput)
	}
	if caps.Modalities != nil {
		if err := validateVocabulary("modalities.input", caps.Modalities.Input, ModalityInputs); err != nil {
			return fmt.Errorf("model %s: invalid capabilities: %w", modelID, err)
		}
		if err := validateVocabulary("modalities.output", caps.Modalities.Output, ModalityOutputs); err != nil {
			return fmt.Errorf("model %s: invalid capabilities: %w", modelID, err)
		}
	}
	if err := validateVocabulary("structured", caps.Structured, StructuredKinds); err != nil {
		return fmt.Errorf("model %s: invalid capabilities: %w", modelID, err)
	}
	return nil
}

func validateVocabulary(field string, values, vocabulary []string) error {
	seen := make(map[string]bool, len(values))
	for _, v := range values {
		if !containsName(vocabulary, v) {
			return fmt.Errorf("%s: unknown value %q (allowed: %s)", field, v, strings.Join(vocabulary, ", "))
		}
		if seen[v] {
			return fmt.Errorf("%s: duplicate value %q", field, v)
		}
		seen[v] = true
	}
	return nil
}
