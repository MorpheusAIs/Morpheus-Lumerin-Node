package modelhealth

import (
	"strings"
	"sync"
	"time"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/aiengine"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/apiprobe"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/apispec"
)

// DefaultRuntimeTTL mirrors the detector's cache TTL (apidetect cacheTTL):
// runtime evidence is reused for as long as the detection result it is
// merged onto.
const DefaultRuntimeTTL = 30 * time.Minute

// litellmHeaderPrefix marks a LiteLLM proxy in front of the backend
// (x-litellm-version, x-litellm-model-id, ...):
// https://docs.litellm.ai/docs/proxy/response_headers
const litellmHeaderPrefix = "x-litellm-"

// The evidence entry MergeRuntime writes for a header-named stack
// (apispec/runtime.go sourceStackHeaders under evidenceKeyStack); the
// checker writes the same entry when it composes the spec itself.
const (
	evidenceKeyStack   = "stack"
	sourceStackHeaders = "health-probe:headers"
)

// runtimeFact names one harvested field of apispec.Runtime; each carries its
// own timestamp because the probe establishes them independently (a 400
// names the window, a 200 the served id) and a later answer that lacks a
// fact does not refute it.
type runtimeFact int

const (
	factServedModelID runtimeFact = iota
	factReasoningSeen
	factStackHints
	factContext
	factRejectedParams
)

type runtimeEntry struct {
	rt apispec.Runtime
	// seen is when each harvested fact was last established; a fact older
	// than the TTL is cleared before the entry is merged.
	seen map[runtimeFact]time.Time
	// verify holds what the verifier and the capability probes established
	// (Verified, ToolsSupported, CapabilityFacts, RejectedParams, reasoning
	// seen with the disable applied); they are established together, so one
	// timestamp expires them together. verifyKey names the spec they were
	// established for (verifyKey in checker.go): a changed spec is verified
	// again before the TTL.
	verify     apispec.Runtime
	verifiedAt time.Time
	verifyKey  string
}

// runtimeStore keeps the runtime evidence per model id (design contract §5)
// so a sweep whose probe yields nothing re-merges what earlier round trips
// established, fact by fact, until a fact is older than the detector TTL:
// the detection result it complements is refreshed on the same schedule,
// and a fact kept longer would outlive a backend change the detector has
// already seen.
type runtimeStore struct {
	ttl time.Duration
	now func() time.Time

	mu      sync.Mutex
	byModel map[string]*runtimeEntry
}

func newRuntimeStore(ttl time.Duration) *runtimeStore {
	if ttl <= 0 {
		ttl = DefaultRuntimeTTL
	}
	return &runtimeStore{ttl: ttl, now: time.Now, byModel: make(map[string]*runtimeEntry)}
}

// harvest records the facts the health probe's response carries (a nil or
// unfilled meta carries none) and returns the runtime to merge for the
// model: every fact still within the TTL. ok is false when nothing is left
// to merge.
func (s *runtimeStore) harvest(modelID string, meta *aiengine.ResponseMeta) (rt apispec.Runtime, ok bool) {
	now := s.now()
	s.mu.Lock()
	defer s.mu.Unlock()

	entry := s.byModel[modelID]
	if fresh, facts := harvestRuntime(meta, now); len(facts) > 0 {
		if entry == nil {
			entry = &runtimeEntry{seen: map[runtimeFact]time.Time{}}
			s.byModel[modelID] = entry
		}
		for _, fact := range facts {
			copyFact(&entry.rt, &fresh, fact)
			entry.seen[fact] = now
		}
	}
	if entry == nil {
		return apispec.Runtime{}, false
	}
	latest := time.Time{}
	for fact, at := range entry.seen {
		if now.Sub(at) > s.ttl {
			clearFact(&entry.rt, fact)
			delete(entry.seen, fact)
			continue
		}
		if at.After(latest) {
			latest = at
		}
	}
	if !entry.verifiedAt.IsZero() && now.Sub(entry.verifiedAt) > s.ttl {
		entry.verify = apispec.Runtime{}
		entry.verifiedAt = time.Time{}
		entry.verifyKey = ""
	}
	if len(entry.seen) == 0 && !hasVerifierFacts(&entry.verify) {
		delete(s.byModel, modelID)
		return apispec.Runtime{}, false
	}
	// At reaches the wire through the verify:rejected:<param>@<at> evidence
	// MergeRuntime writes for a refused parameter and through the
	// verify:reasoning_seen_with_disable@<at> source, so it follows the
	// timestamp of the fact that needs it: stamped from the newest fact, a
	// later served id would rewrite the string and reset declaredAt although
	// nothing about the backend changed. An entry kept alive only by other
	// verifier facts leaves At at 0, which MergeRuntime reads as now.
	switch rejectedAt := entry.seen[factRejectedParams]; {
	case !rejectedAt.IsZero():
		entry.rt.At = rejectedAt.Unix()
	case len(entry.verify.RejectedParams) > 0 || entry.verify.ReasoningDisabledInRequest:
		entry.rt.At = entry.verifiedAt.Unix()
	case !latest.IsZero():
		entry.rt.At = latest.Unix()
	default:
		entry.rt.At = 0
	}
	rt = entry.rt
	overlayVerify(&rt, &entry.verify)
	return rt, true
}

// overlayVerify lays the verifier's facts over the harvested ones. The maps
// and slices of v stay shared with the store (MergeRuntime only reads them);
// RejectedParams is copied before the union so the store's slice is never
// appended to.
func overlayVerify(rt *apispec.Runtime, v *apispec.Runtime) {
	if len(v.Verified) > 0 {
		rt.Verified = v.Verified
	}
	if v.ToolsSupported != nil {
		rt.ToolsSupported, rt.ToolsSource = v.ToolsSupported, v.ToolsSource
	}
	if !v.CapabilityFacts.Empty() {
		rt.CapabilityFacts = v.CapabilityFacts
	}
	if v.ReasoningSeen {
		rt.ReasoningSeen = true
		rt.ReasoningDisabledInRequest = rt.ReasoningDisabledInRequest || v.ReasoningDisabledInRequest
	}
	if len(v.RejectedParams) > 0 {
		rt.RejectedParams = append([]string(nil), rt.RejectedParams...)
		for _, p := range v.RejectedParams {
			if !containsParam(rt.RejectedParams, p) {
				rt.RejectedParams = append(rt.RejectedParams, p)
			}
		}
	}
}

func containsParam(list []string, p string) bool {
	for _, have := range list {
		if have == p {
			return true
		}
	}
	return false
}

// verifyDue reports whether the verifier should run for the model: never
// yet, for another spec (key), or last more than the TTL ago.
func (s *runtimeStore) verifyDue(modelID, key string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	entry := s.byModel[modelID]
	if entry == nil || entry.verifyKey != key || entry.verifiedAt.IsZero() {
		return true
	}
	return s.now().Sub(entry.verifiedAt) > s.ttl
}

// recordVerify stores one verifier run's facts for the model under key. A
// run within the TTL happens only for a changed spec (key): the binding
// verdicts and rejections of the old spec go, since they describe bindings
// that no longer exist, while the tool and capability facts describe the
// backend and carry over unless the new run answered them itself (a
// question the merged spec already answered is not asked again). A run with
// no facts still counts as done, so a backend that answers nothing usable is
// asked once per TTL, not once per sweep.
func (s *runtimeStore) recordVerify(modelID, key string, facts apispec.Runtime) {
	s.mu.Lock()
	defer s.mu.Unlock()
	entry := s.byModel[modelID]
	if entry == nil {
		entry = &runtimeEntry{seen: map[runtimeFact]time.Time{}}
		s.byModel[modelID] = entry
	}
	prev := entry.verify
	if facts.ToolsSupported == nil && prev.ToolsSupported != nil {
		facts.ToolsSupported, facts.ToolsSource = prev.ToolsSupported, prev.ToolsSource
	}
	for field, source := range prev.CapabilityFacts.Sources {
		if _, answered := facts.CapabilityFacts.Sources[field]; answered {
			continue
		}
		switch field {
		case apispec.CapFieldStructured:
			facts.CapabilityFacts.SetStructured(prev.CapabilityFacts.Structured, source)
		case apispec.CapFieldModalitiesInput:
			facts.CapabilityFacts.SetInputModalities(prev.CapabilityFacts.Modalities.Input, source)
		}
	}
	entry.verify = facts
	entry.verifiedAt = s.now()
	entry.verifyKey = key
}

// forget drops the entry for a model that is no longer reported.
func (s *runtimeStore) forget(modelID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.byModel, modelID)
}

// hasVerifierFacts reports whether the verify block carries facts; those
// expire with verifiedAt, not fact by fact.
func hasVerifierFacts(rt *apispec.Runtime) bool {
	return len(rt.Verified) > 0 || rt.ToolsSupported != nil || !rt.CapabilityFacts.Empty() || len(rt.RejectedParams) > 0 || rt.ReasoningDisabledInRequest
}

func copyFact(dst, src *apispec.Runtime, fact runtimeFact) {
	switch fact {
	case factServedModelID:
		dst.ServedModelID = src.ServedModelID
	case factReasoningSeen:
		// ReasoningDisabledInRequest is the verifier's fact (the health
		// probe sends no reasoning.disable binding); a plain probe that sees
		// reasoning content neither establishes nor refutes it.
		dst.ReasoningSeen = src.ReasoningSeen
	case factStackHints:
		dst.StackHints = src.StackHints
	case factContext:
		dst.ContextLimit = src.ContextLimit
		dst.ContextSource = src.ContextSource
	case factRejectedParams:
		dst.RejectedParams = src.RejectedParams
	}
}

func clearFact(rt *apispec.Runtime, fact runtimeFact) {
	copyFact(rt, &apispec.Runtime{}, fact)
}

// harvestRuntime reads the §5 harvest out of one probe response and names
// the facts it established. A meta the adapter never filled (no response, a
// non-OpenAI adapter) and a 2xx without a model id or reasoning content
// establish none.
func harvestRuntime(meta *aiengine.ResponseMeta, now time.Time) (rt apispec.Runtime, facts []runtimeFact) {
	if meta == nil {
		return rt, nil
	}
	rt.At = now.Unix()
	if rt.ServedModelID = meta.Model; rt.ServedModelID != "" {
		facts = append(facts, factServedModelID)
	}
	if rt.ReasoningSeen = meta.ReasoningSeen; rt.ReasoningSeen {
		facts = append(facts, factReasoningSeen)
	}
	for name := range meta.Headers {
		if strings.HasPrefix(name, litellmHeaderPrefix) {
			rt.StackHints = []string{"litellm"}
			facts = append(facts, factStackHints)
			break
		}
	}
	if meta.StatusCode >= 400 && len(meta.ErrorBody) > 0 {
		if limit, source, found := apiprobe.ParseContextLimit(meta.ErrorBody, meta.StatusCode); found {
			rt.ContextLimit, rt.ContextSource = limit, source
			facts = append(facts, factContext)
		}
		if params, found := apiprobe.ParseRejectedParams(meta.ErrorBody); found {
			rt.RejectedParams = params
			facts = append(facts, factRejectedParams)
		}
	}
	return rt, facts
}
