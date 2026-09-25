package apispec

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/config"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/system"
)

// Runtime is the evidence one inference round trip (the health probe, the
// binding verifier, a capability probe) established about a model. It is
// merged onto the cached detection result by MergeRuntime (design contract
// §3 rule 11) and never written back into the detector's cache.
type Runtime struct {
	// ServedModelID is the model field of the completion; it refines the
	// family the way the composer does and is never published.
	ServedModelID string
	// ReasoningSeen is true when the response carried reasoning content
	// (aiengine.ResponseMeta.ReasoningSeen).
	ReasoningSeen bool
	// ReasoningDisabledInRequest is true when the request that produced the
	// response had the reasoning.disable binding applied; with ReasoningSeen
	// it proves the model reasons unconditionally.
	ReasoningDisabledInRequest bool
	// StackHints are stack names the transport suggested (litellm from
	// x-litellm-* headers); the first known one names the stack only when
	// detection left it empty.
	StackHints []string
	// ContextLimit is the served window read from a context-exceeded error;
	// ContextSource is its evidence entry (apiprobe.ContextSource*), default
	// "health-probe:context_length_exceeded". Fills capabilities.context only
	// when unset.
	ContextLimit  int64
	ContextSource string
	// RejectedParams are request parameters the backend refused
	// (apiprobe.ParseRejectedParams); every binding whose root param is
	// listed is dropped.
	RejectedParams []string
	// ToolsSupported is the tool-call probe's verdict; ToolsSource is its
	// evidence entry, default "verify:tools@<at>".
	ToolsSupported *bool
	ToolsSource    string
	// Verified maps an intent to the verifier's verdict on its binding.
	Verified map[string]VerifiedBinding
	// CapabilityFacts are further facts a probe established, each with its
	// own source (verify:json_schema@<at>, ...).
	CapabilityFacts CapabilityFacts
	// At is the unix time of the round trip; 0 means now.
	At int64
}

type VerifiedBinding struct {
	Verified bool
	At       int64
	// Reason is a trace note (what the probe saw); not published.
	Reason string
}

const (
	layerRuntime = "runtime"

	sourceContextFromError = "health-probe:context_length_exceeded"
	sourceReasoningSeen    = "health-probe:reasoning_seen"
	sourceStackHeaders     = "health-probe:headers"
	evidenceKeyStack       = "stack"
)

func verifySource(what string, at int64) string {
	return "verify:" + what + "@" + strconv.FormatInt(at, 10)
}

// MergeRuntime applies rt to api in place and returns the trace. Callers pass
// a Clone of the cached detection result so the cache stays free of runtime
// evidence. Precedence follows §0: declared facts are never overwritten, a
// runtime verdict replaces backend, template and registry facts, and the
// error-derived context window fills only an unset field. api == nil is a
// no-op, and an empty rt leaves api byte-identical.
func MergeRuntime(api *system.ModelApiSpec, rt Runtime) (trace []string) {
	if api == nil {
		return nil
	}
	tracef := func(format string, args ...any) {
		trace = append(trace, fmt.Sprintf(format, args...))
	}
	at := rt.At
	if at == 0 {
		at = time.Now().Unix()
	}
	evidence := func() map[string]string {
		if api.Evidence == nil {
			api.Evidence = map[string]string{}
		}
		return api.Evidence
	}

	if rt.ServedModelID != "" {
		if served := FamilyFromName(rt.ServedModelID); refinesFamily(api.ModelFamily, served) {
			api.ModelFamily = served
			tracef("family %q: from runtime-served model id", served)
		}
	}

	if api.Stack == "" && api.Via == "" {
		for _, hint := range rt.StackHints {
			if stack := StackFor(hint); stack != "" {
				api.Stack = stack
				evidence()[evidenceKeyStack] = sourceStackHeaders
				tracef("stack = %s (%s)", stack, sourceStackHeaders)
				break
			}
		}
	} else if len(rt.StackHints) > 0 {
		tracef("stack hints %v ignored — detection already named the stack", rt.StackHints)
	}

	dropRejectedBindings(api, rt.RejectedParams, at, tracef)

	for _, intent := range sortedVerified(rt.Verified) {
		v := rt.Verified[intent]
		b := api.Bindings[intent]
		if b == nil {
			tracef("verify: %s has no binding; verdict ignored", intent)
			continue
		}
		when := v.At
		if when == 0 {
			when = at
		}
		verified := v.Verified
		b.Verified = &verified
		b.VerifiedAt = when
		evidence()[system.EvidenceKeyBinding(intent)] = verifySource(intent, when)
		if v.Reason != "" {
			tracef("bindings.%s verified = %t (%s): %s", intent, verified, verifySource(intent, when), v.Reason)
		} else {
			tracef("bindings.%s verified = %t (%s)", intent, verified, verifySource(intent, when))
		}
	}

	var direct CapabilityFacts
	if rt.ContextLimit != 0 {
		source := rt.ContextSource
		if source == "" {
			source = sourceContextFromError
		}
		if !direct.SetContext(rt.ContextLimit, source) {
			tracef("capabilities.context: runtime value %d out of (0, %d); ignored", rt.ContextLimit, config.MaxCapabilityInt)
		}
	}
	if rt.ToolsSupported != nil {
		source := rt.ToolsSource
		if source == "" {
			source = verifySource(CapFieldTools, at)
		}
		direct.SetTools(*rt.ToolsSupported, source)
	}
	if rt.ReasoningSeen {
		if rt.ReasoningDisabledInRequest {
			direct.SetReasoning(true, verifySource("reasoning_seen_with_disable", at))
		} else {
			direct.SetReasoning(true, sourceReasoningSeen)
		}
	}
	mergeRuntimeCaps(api, &direct, map[string]bool{CapFieldContext: true}, tracef)
	mergeRuntimeCaps(api, &rt.CapabilityFacts, nil, tracef)

	if rt.ReasoningSeen && rt.ReasoningDisabledInRequest {
		source := verifySource("reasoning_seen_with_disable", at)
		switch {
		case api.Thinking == nil:
			api.Thinking = &system.ThinkingSpec{Mode: system.ThinkingModeAlwaysOn}
			tracef("thinking.mode = always_on (%s)", source)
		case api.Thinking.Mode != system.ThinkingModeAlwaysOn:
			tracef("thinking.mode: %s -> always_on (%s)", api.Thinking.Mode, source)
			api.Thinking.Mode = system.ThinkingModeAlwaysOn
		}
	}
	return trace
}

// mergeRuntimeCaps copies facts' set fields into api.Capabilities through the
// composer's validation. fillOnly names the fields that never overwrite an
// existing value; every other field overwrites unless it is declared.
func mergeRuntimeCaps(api *system.ModelApiSpec, facts *CapabilityFacts, fillOnly map[string]bool, tracef func(string, ...any)) {
	if facts == nil || facts.Empty() {
		return
	}
	st := capState{evidence: map[string]string{}, layerOf: map[string]string{}}
	if api.Capabilities != nil {
		st.caps = *api.Capabilities.Clone()
	}
	for key, source := range api.Evidence {
		if field, ok := strings.CutPrefix(key, "capabilities."); ok {
			st.evidence[key] = source
			st.layerOf[field] = layerFromSource(source)
		}
	}
	changed := false
	for _, field := range capFields {
		source := facts.Sources[field]
		if source == "" {
			continue
		}
		existing := api.Evidence[system.EvidenceKeyCapability(field)]
		switch {
		case st.layerOf[field] == layerDeclared:
			tracef("capabilities.%s: runtime value ignored — declared (%s)", field, existing)
			continue
		case fillOnly[field] && st.isSet(field):
			tracef("capabilities.%s: runtime value ignored — already set (%s)", field, existing)
			continue
		}
		value, ok := st.take(field, &facts.ModelCapabilities, layerRuntime, tracef)
		if !ok {
			continue
		}
		st.set(field, value, source, layerRuntime, tracef)
		changed = true
	}
	if !changed {
		return
	}
	caps := st.caps
	api.Capabilities = &caps
	if api.Evidence == nil {
		api.Evidence = map[string]string{}
	}
	for key, source := range st.evidence {
		api.Evidence[key] = source
	}
}

func layerFromSource(source string) string {
	switch {
	case source == SourceDeclared:
		return layerDeclared
	case strings.HasPrefix(source, "registry:"):
		return layerRegistry
	case strings.HasPrefix(source, "verify:"), strings.HasPrefix(source, "health-probe:"):
		return layerRuntime
	}
	return layerBackend
}

// dropRejectedBindings removes every binding whose root request parameter
// the backend refused, records why under the binding's evidence key, and
// keeps thinking.mode consistent with the bindings that remain (the
// composer's rule; always_on is a model fact and stays).
func dropRejectedBindings(api *system.ModelApiSpec, rejected []string, at int64, tracef func(string, ...any)) {
	if len(rejected) == 0 || len(api.Bindings) == 0 {
		return
	}
	rejectedSet := make(map[string]bool, len(rejected))
	for _, p := range rejected {
		rejectedSet[p] = true
	}
	dropped := false
	for _, intent := range sortedIntents(api.Bindings) {
		b := api.Bindings[intent]
		root := b.Param
		if i := strings.IndexByte(root, '.'); i >= 0 {
			root = root[:i]
		}
		if root == "" || !rejectedSet[root] {
			continue
		}
		delete(api.Bindings, intent)
		source := verifySource("rejected:"+root, at)
		if api.Evidence == nil {
			api.Evidence = map[string]string{}
		}
		api.Evidence[system.EvidenceKeyBinding(intent)] = source
		tracef("bindings: %s dropped — backend rejected parameter %s (%s)", intent, root, source)
		dropped = true
	}
	if !dropped {
		return
	}
	if len(api.Bindings) == 0 {
		api.Bindings = nil
	}
	if api.Thinking == nil || api.Thinking.Mode == system.ThinkingModeAlwaysOn {
		return
	}
	mode := ""
	switch {
	case api.Bindings[system.IntentReasoningDisable] != nil, api.Bindings[system.IntentReasoningEnable] != nil:
		mode = system.ThinkingModeControllable
	case api.Bindings[system.IntentReasoningEffort] != nil || api.Bindings[system.IntentReasoningBudget] != nil:
		mode = system.ThinkingModeTunable
	}
	switch {
	case mode == "":
		tracef("thinking: %s dropped — no reasoning binding remains", api.Thinking.Mode)
		api.Thinking = nil
		if api.Evidence != nil {
			delete(api.Evidence, system.EvidenceKeyThinkingDefault)
		}
	case mode != api.Thinking.Mode:
		tracef("thinking.mode: %s -> %s after the drop", api.Thinking.Mode, mode)
		api.Thinking.Mode = mode
	}
}

func sortedVerified(v map[string]VerifiedBinding) []string {
	out := make([]string, 0, len(v))
	for intent := range v {
		out = append(out, intent)
	}
	sort.Strings(out)
	return out
}
