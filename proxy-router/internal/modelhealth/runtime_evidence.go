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
	if len(entry.seen) == 0 && !hasVerifierFacts(&entry.rt) {
		delete(s.byModel, modelID)
		return apispec.Runtime{}, false
	}
	entry.rt.At = latest.Unix()
	return entry.rt, true
}

// forget drops the entry for a model that is no longer reported.
func (s *runtimeStore) forget(modelID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.byModel, modelID)
}

// hasVerifierFacts reports whether the entry carries facts the verifier
// wrote; those are not the harvest's to expire.
func hasVerifierFacts(rt *apispec.Runtime) bool {
	return len(rt.Verified) > 0 || rt.ToolsSupported != nil || !rt.CapabilityFacts.Empty()
}

func copyFact(dst, src *apispec.Runtime, fact runtimeFact) {
	switch fact {
	case factServedModelID:
		dst.ServedModelID = src.ServedModelID
	case factReasoningSeen:
		dst.ReasoningSeen = src.ReasoningSeen
		dst.ReasoningDisabledInRequest = src.ReasoningDisabledInRequest
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
