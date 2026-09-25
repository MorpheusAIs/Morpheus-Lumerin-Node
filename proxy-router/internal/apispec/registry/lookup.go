package registry

import (
	"strings"
	"sync"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/apispec"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/system"
)

// entry is one normalised model id of the snapshot. A false flag is
// indistinguishable from an absent one (LiteLLM omits what it does not
// know), so Lookup asserts only true values: a wrong capability is worse
// than a missing one.
type entry struct {
	ID          string   `json:"id"`
	Context     int64    `json:"context,omitempty"`
	Output      int64    `json:"output,omitempty"`
	Reasoning   bool     `json:"reasoning,omitempty"`
	ToolCall    bool     `json:"toolCall,omitempty"`
	Structured  bool     `json:"structured,omitempty"`
	WebSearch   bool     `json:"webSearch,omitempty"`
	PromptCache bool     `json:"promptCache,omitempty"`
	Attachment  bool     `json:"attachment,omitempty"`
	In          []string `json:"in,omitempty"`
	Out         []string `json:"out,omitempty"`
	Src         src      `json:"src"`
}

// src records which sources contributed to an entry; it is the RegistryFacts
// name of the evidence string "registry:<name>@<date>".
type src uint8

const (
	srcModelsDev src = iota + 1
	srcLiteLLM
	srcBoth
)

func (s src) name() string {
	switch s {
	case srcModelsDev:
		return "models.dev"
	case srcLiteLLM:
		return "litellm"
	case srcBoth:
		return "models.dev+litellm"
	}
	return "unknown"
}

// index answers exact ids and, for ids with a path, the segment after the
// last "/" (so "meta-llama/llama-3-70b" is found by "llama-3-70b" when no
// other id ends the same way).
type index struct {
	entries []entry
	exact   map[string]int
	tails   map[string][]int
}

func newIndex(entries []entry) *index {
	ix := &index{entries: entries, exact: make(map[string]int, len(entries)), tails: map[string][]int{}}
	for i, e := range entries {
		ix.exact[e.ID] = i
		if j := strings.LastIndexByte(e.ID, '/'); j >= 0 && j < len(e.ID)-1 {
			tail := e.ID[j+1:]
			ix.tails[tail] = append(ix.tails[tail], i)
		}
	}
	return ix
}

// lookup tries every candidate exactly before any suffix match, so a served
// id's exact hit beats a configured name's tail guess.
func (ix *index) lookup(candidates []string) (entry, bool) {
	keys := make([]string, 0, len(candidates))
	for _, c := range candidates {
		if k := normalizeCandidate(c); k != "" {
			keys = append(keys, k)
		}
	}
	for _, k := range keys {
		if i, ok := ix.exact[k]; ok {
			return ix.entries[i], true
		}
	}
	for _, k := range keys {
		tail := k
		if j := strings.LastIndexByte(k, '/'); j >= 0 {
			if j == len(k)-1 {
				continue
			}
			tail = k[j+1:]
			if i, ok := ix.exact[tail]; ok {
				return ix.entries[i], true
			}
		}
		if hits := ix.tails[tail]; len(hits) == 1 {
			return ix.entries[hits[0]], true
		}
	}
	return entry{}, false
}

// normalizeCandidate applies the generator's id rules that need no provider
// table: lowercase and no ":" segment carrying "=". A provider prefix on the
// candidate is handled by the tail pass of lookup.
func normalizeCandidate(id string) string {
	id = strings.ToLower(strings.TrimSpace(id))
	if !strings.Contains(id, "=") {
		return id
	}
	parts := strings.Split(id, ":")
	kept := parts[:1]
	for _, p := range parts[1:] {
		if !strings.Contains(p, "=") {
			kept = append(kept, p)
		}
	}
	return strings.Join(kept, ":")
}

var (
	snapshotOnce  sync.Once
	snapshotIndex *index
)

func snapshot() *index {
	snapshotOnce.Do(func() { snapshotIndex = newIndex(entries) })
	return snapshotIndex
}

// Lookup finds the snapshot entry for the first candidate that matches
// (callers pass the served id, the bare model name and the configured name,
// in that order): exact normalised id first across all candidates, then a
// unique match on the segment after the last "/". No fuzzy matching. The
// result is a fresh copy the caller may keep.
func Lookup(candidates ...string) (*apispec.RegistryFacts, bool) {
	e, ok := snapshot().lookup(candidates)
	if !ok {
		return nil, false
	}
	return facts(e, snapshotDate), true
}

// Date is the snapshot's generation date, YYYY-MM-DD.
func Date() string { return snapshotDate }

// Size is the number of ids in the snapshot.
func Size() int { return len(entries) }

func facts(e entry, date string) *apispec.RegistryFacts {
	f := &apispec.RegistryFacts{Name: e.Src.name(), Date: date}
	if e.Context > 0 {
		f.Context = ptr(e.Context)
	}
	if e.Output > 0 {
		f.MaxOutput = ptr(e.Output)
	}
	if e.Reasoning {
		f.Reasoning = ptr(true)
	}
	if e.ToolCall {
		f.Tools = ptr(true)
	}
	if e.Structured {
		f.Structured = []string{"json_schema"}
	}
	if e.WebSearch {
		f.WebSearch = ptr(true)
	}
	if e.PromptCache {
		f.PromptCache = ptr(true)
	}
	if len(e.In) > 0 || len(e.Out) > 0 {
		f.Modalities = &system.Modalities{Input: cloneStrings(e.In), Output: cloneStrings(e.Out)}
	}
	return f
}

func ptr[T any](v T) *T { return &v }

func cloneStrings(list []string) []string {
	if list == nil {
		return nil
	}
	return append([]string(nil), list...)
}
