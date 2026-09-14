package apispec

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/system"
)

// DescribeBinding renders one ParamBinding as a human-readable line, shared
// by trace output and diagnostics tooling.
func DescribeBinding(b *system.ParamBinding) string {
	if b == nil {
		return "unknown"
	}
	s := b.Kind
	if b.Param != "" {
		s += " " + b.Param
	}
	var details []string
	if b.ParamType != "" {
		if len(b.EnumValues) > 0 {
			details = append(details, b.ParamType+": "+strings.Join(b.EnumValues, "|"))
		} else {
			details = append(details, b.ParamType)
		}
	}
	if b.Value != nil {
		details = append(details, "value="+formatValue(b.Value))
	}
	if len(details) > 0 {
		s += " (" + strings.Join(details, ", ") + ")"
	}
	if b.Hint != "" {
		s += " — " + b.Hint
	}
	return s
}

// formatValue renders a binding value the way it appears on the wire.
func formatValue(v any) string {
	switch v.(type) {
	case map[string]any, []any, []string:
		if raw, err := json.Marshal(v); err == nil {
			return string(raw)
		}
	}
	return fmt.Sprintf("%v", v)
}

// describeBindings renders a whole binding set compactly, intents sorted.
func describeBindings(b map[string]*system.ParamBinding) string {
	intents := make([]string, 0, len(b))
	for intent := range b {
		intents = append(intents, intent)
	}
	sort.Strings(intents)
	parts := make([]string, 0, len(intents))
	for _, intent := range intents {
		parts = append(parts, intent+": "+DescribeBinding(b[intent]))
	}
	return strings.Join(parts, "; ")
}
