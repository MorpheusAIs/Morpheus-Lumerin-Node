package registry

import (
	"encoding/json"
	"os"
	"reflect"
	"regexp"
	"sort"
	"strings"
	"testing"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/apispec"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/config"
)

func fixtureIndex(t *testing.T) *index {
	t.Helper()
	raw, err := os.ReadFile("testdata/entries.json")
	if err != nil {
		t.Fatal(err)
	}
	var list []entry
	if err := json.Unmarshal(raw, &list); err != nil {
		t.Fatal(err)
	}
	return newIndex(list)
}

func TestLookupOrder(t *testing.T) {
	ix := fixtureIndex(t)
	cases := []struct {
		name       string
		candidates []string
		want       string
	}{
		{"exact", []string{"gpt-4o"}, "gpt-4o"},
		{"case-insensitive", []string{"GPT-4o"}, "gpt-4o"},
		{"parameter segment stripped", []string{"gpt-4o:temp=0"}, "gpt-4o"},
		{"served id with provider prefix", []string{"openai/gpt-4o"}, "gpt-4o"},
		{"gateway prefix on an org id", []string{"together_ai/meta-llama/llama-3-70b"}, "meta-llama/llama-3-70b"},
		{"bare name finds the unique org id", []string{"llama-3-70b"}, "meta-llama/llama-3-70b"},
		{"exact beats a unique tail", []string{"qwen3-235b-a22b"}, "qwen3-235b-a22b"},
		{"org id exact", []string{"qwen/qwen3-235b-a22b"}, "qwen/qwen3-235b-a22b"},
		{"every candidate is tried exactly before any tail", []string{"org/gpt-4o", "meta-llama/llama-3-70b"}, "meta-llama/llama-3-70b"},
		{"first candidate wins the exact pass", []string{"o1-pro", "gpt-4o"}, "o1-pro"},
		{"blank candidates skipped", []string{"", "  ", "gpt-4o"}, "gpt-4o"},
		{"ambiguous tail", []string{"dup"}, ""},
		{"ambiguous tail behind a prefix", []string{"x/dup"}, ""},
		{"no fuzzy", []string{"gpt-4o-mini", "gpt4o"}, ""},
		{"trailing slash", []string{"gpt-4o/"}, ""},
		{"nothing", nil, ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			e, ok := ix.lookup(tc.candidates)
			if ok != (tc.want != "") || e.ID != tc.want {
				t.Fatalf("lookup(%v) = (%q, %v), want %q", tc.candidates, e.ID, ok, tc.want)
			}
		})
	}
}

func TestFacts(t *testing.T) {
	ix := fixtureIndex(t)
	tr := true
	get := func(id string) *apispec.RegistryFacts {
		e, ok := ix.lookup([]string{id})
		if !ok {
			t.Fatalf("%s missing from the fixture", id)
		}
		return facts(e, "2026-09-25")
	}

	got := get("gpt-4o")
	if got.Name != "models.dev+litellm" || got.Date != "2026-09-25" || got.Source() != "registry:models.dev+litellm@2026-09-25" {
		t.Fatalf("name/date/source = %q %q %q", got.Name, got.Date, got.Source())
	}
	if *got.Context != 128000 || *got.MaxOutput != 16384 || !*got.Tools || !*got.WebSearch || !*got.PromptCache ||
		!reflect.DeepEqual(got.Structured, []string{"json_schema"}) ||
		!reflect.DeepEqual(got.Modalities.Input, []string{"text", "image", "pdf"}) ||
		!reflect.DeepEqual(got.Modalities.Output, []string{"text"}) {
		t.Fatalf("gpt-4o facts = %+v", got.ModelCapabilities)
	}
	if got.Reasoning != nil || got.ToolsParallel != nil || got.Embeddings != nil || got.ContextTrained != nil || got.Quantization != "" {
		t.Fatalf("fields the snapshot cannot know must stay nil: %+v", got.ModelCapabilities)
	}

	// A false flag is unknown, never asserted.
	got = get("flagless")
	if got.Name != "litellm" || *got.Context != 4096 || got.MaxOutput != nil || got.Tools != nil || got.Reasoning != nil ||
		got.Structured != nil || got.WebSearch != nil || got.PromptCache != nil {
		t.Fatalf("flagless facts = %+v", got.ModelCapabilities)
	}

	got = get("o1-pro")
	if got.Name != "litellm" || !reflect.DeepEqual(got.Reasoning, &tr) || got.Tools != nil {
		t.Fatalf("o1-pro facts = %+v", got.ModelCapabilities)
	}
	if get("meta-llama/llama-3-70b").Name != "models.dev" {
		t.Fatal("a models.dev-only entry must be named models.dev")
	}

	// The result is the caller's copy.
	got = get("gpt-4o")
	got.Modalities.Input[0] = "mutated"
	*got.Context = 1
	if again := get("gpt-4o"); again.Modalities.Input[0] != "text" || *again.Context != 128000 {
		t.Fatal("facts must not alias the snapshot")
	}
}

func TestSnapshot(t *testing.T) {
	if !regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`).MatchString(Date()) {
		t.Fatalf("Date() = %q", Date())
	}
	if Size() < 1000 {
		t.Fatalf("Size() = %d, the snapshot looks truncated", Size())
	}
	// Invariants the generator must hold so the composer takes every field.
	if !sort.SliceIsSorted(entries, func(i, j int) bool { return entries[i].ID < entries[j].ID }) {
		t.Fatal("entries are not sorted by id")
	}
	seen := map[string]bool{}
	for _, e := range entries {
		if e.ID == "" || seen[e.ID] || e.ID != strings.ToLower(e.ID) || strings.Contains(e.ID, "=") {
			t.Fatalf("bad or duplicate id %q", e.ID)
		}
		seen[e.ID] = true
		if e.Src < srcModelsDev || e.Src > srcBoth {
			t.Fatalf("%s: src %d", e.ID, e.Src)
		}
		for _, v := range []int64{e.Context, e.Output} {
			if v != 0 && !config.ValidCapabilityInt(v) {
				t.Fatalf("%s: limit %d out of range", e.ID, v)
			}
		}
		for _, m := range e.In {
			if !config.IsModalityInput(m) {
				t.Fatalf("%s: input modality %q", e.ID, m)
			}
		}
		for _, m := range e.Out {
			if !config.IsModalityOutput(m) {
				t.Fatalf("%s: output modality %q", e.ID, m)
			}
		}
	}

	for _, id := range []string{"gpt-4o", "claude-sonnet-4-5"} {
		f, ok := Lookup(id)
		if !ok || f.Context == nil || *f.Context <= 0 || f.Date != Date() || !strings.HasPrefix(f.Name, "models.dev") {
			t.Fatalf("Lookup(%q) = %+v, %v", id, f, ok)
		}
	}
	if f, ok := Lookup("claude-sonnet-4-5"); !ok || f.Reasoning == nil || !*f.Reasoning || f.Tools == nil || !*f.Tools {
		t.Fatalf("claude-sonnet-4-5 must carry reasoning and tools: %+v", f)
	}
	if _, ok := Lookup("no-such-model-xyz"); ok {
		t.Fatal("unknown id found")
	}
}

func BenchmarkLookup(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Lookup("openai/gpt-4o", "gpt-4o", "my-gpt")
	}
}
