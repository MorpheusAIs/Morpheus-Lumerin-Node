package main

import (
	"bytes"
	"os"
	"reflect"
	"strings"
	"testing"
)

func loadFixtures(t *testing.T) ([]byte, []byte) {
	t.Helper()
	md, err := os.ReadFile("testdata/modelsdev.json")
	if err != nil {
		t.Fatal(err)
	}
	ll, err := os.ReadFile("testdata/litellm.json")
	if err != nil {
		t.Fatal(err)
	}
	return md, ll
}

func TestMerge(t *testing.T) {
	md, ll := loadFixtures(t)
	got, err := merge(md, ll)
	if err != nil {
		t.Fatal(err)
	}
	var ids []string
	byID := map[string]entry{}
	for _, e := range got {
		ids = append(ids, e.ID)
		byID[e.ID] = e
	}
	// "openrouter/openai/gpt-4o" folds into gpt-4o, never "openai/gpt-4o";
	// the malformed models.dev rows are skipped, not fatal.
	wantIDs := []string{
		"accounts/fireworks/models/llama-v3p1-8b-instruct",
		"bar:free",
		"bare-litellm",
		"google/gemma-3-27b-it",
		"gpt-4o",
		"holo3:thinking",
		"meta-llama/llama-3-70b",
		"o1-pro",
	}
	if !reflect.DeepEqual(ids, wantIDs) {
		t.Fatalf("ids = %v, want %v", ids, wantIDs)
	}

	tr := true
	cases := []struct {
		id   string
		want entry
	}{
		// The vendor of record (openai, named by openrouter's prefix in both
		// sources) beats azure's own listing and openrouter's nested LiteLLM
		// key; LiteLLM adds only web search and caching, not openrouter's
		// reasoning flag or limits.
		{"gpt-4o", entry{ID: "gpt-4o", Context: 128000, Output: 16384, ToolCall: &tr, Structured: &tr, Attachment: &tr,
			WebSearch: &tr, PromptCache: &tr, In: []string{"text", "image", "pdf"}, Out: []string{"text"}, Src: srcBoth}},
		// models.dev's context wins over LiteLLM's; the missing output is filled.
		{"meta-llama/llama-3-70b", entry{ID: "meta-llama/llama-3-70b", Context: 8192, Output: 4096, ToolCall: &tr,
			PromptCache: &tr, In: []string{"text"}, Out: []string{"text"}, Src: srcBoth}},
		{"o1-pro", entry{ID: "o1-pro", Context: 200000, Output: 100000, Reasoning: &tr, Structured: &tr,
			In: []string{"text"}, Out: []string{"text"}, Src: srcLiteLLM}},
		{"accounts/fireworks/models/llama-v3p1-8b-instruct", entry{ID: "accounts/fireworks/models/llama-v3p1-8b-instruct",
			Context: 131072, ToolCall: &tr, Attachment: &tr,
			In: []string{"text", "audio", "video", "pdf"}, Out: []string{"text", "audio"}, Src: srcLiteLLM}},
		// A float that is an integer is accepted; "video" is no output modality.
		{"holo3:thinking", entry{ID: "holo3:thinking", Context: 32768, Reasoning: &tr,
			In: []string{"text"}, Out: []string{"text"}, Src: srcModelsDev}},
		// Out-of-range and fractional limits are dropped; ":temp=0.5" is a parameter.
		{"bar:free", entry{ID: "bar:free", In: []string{"text"}, Out: []string{"text"}, Src: srcModelsDev}},
	}
	for _, tc := range cases {
		got := byID[tc.id]
		got.prefix, got.vendor = "", ""
		if !sameEntry(got, tc.want) {
			t.Errorf("%s:\n got %s\nwant %s", tc.id, describe(got), describe(tc.want))
		}
	}
}

func sameEntry(a, b entry) bool {
	same := func(x, y *bool) bool { return isTrue(x) == isTrue(y) }
	return a.ID == b.ID && a.Context == b.Context && a.Output == b.Output &&
		same(a.Reasoning, b.Reasoning) && same(a.ToolCall, b.ToolCall) && same(a.Structured, b.Structured) &&
		same(a.WebSearch, b.WebSearch) && same(a.PromptCache, b.PromptCache) && same(a.Attachment, b.Attachment) &&
		reflect.DeepEqual(a.In, b.In) && reflect.DeepEqual(a.Out, b.Out) && a.Src == b.Src
}

func describe(e entry) string {
	var b bytes.Buffer
	writeEntry(&b, e)
	return strings.TrimSpace(b.String())
}

func TestNormalize(t *testing.T) {
	prefixes := map[string]bool{"openai": true, "anthropic": true, "openrouter": true, "fireworks-ai": true, "together-ai": true}
	cases := []struct {
		in, id, prefix string
	}{
		{"GPT-4o", "gpt-4o", ""},
		{"openai/gpt-4o", "gpt-4o", "openai"},
		{"fireworks_ai/accounts/fireworks/models/x", "accounts/fireworks/models/x", "fireworks-ai"},
		{"meta-llama/Llama-3-70b", "meta-llama/llama-3-70b", ""},
		{"openrouter/anthropic/claude-sonnet-4-5", "claude-sonnet-4-5", "anthropic"},
		{"together_ai/meta-llama/Llama-3-70b", "meta-llama/llama-3-70b", "together-ai"},
		{"openrouter/openai/anthropic/openai/x", "openai/x", "anthropic"},
		{"openai/", "", ""},
		{"llama3:8b-q4_K_M", "llama3:8b-q4_k_m", ""},
		{"deepseek/deepseek-r1:free", "deepseek/deepseek-r1:free", ""},
		{"claude:thinking=high:free", "claude:free", ""},
		{"x:a=1", "x", ""},
		{"  ", "", ""},
		{"/", "", ""},
	}
	for _, tc := range cases {
		id, prefix := normalize(tc.in, prefixes)
		if id != tc.id || prefix != tc.prefix {
			t.Errorf("normalize(%q) = (%q, %q), want (%q, %q)", tc.in, id, prefix, tc.id, tc.prefix)
		}
	}
}

func TestRenderDeterministicAndCapped(t *testing.T) {
	md, ll := loadFixtures(t)
	entries, err := merge(md, ll)
	if err != nil {
		t.Fatal(err)
	}
	const date = "2026-09-25"
	full, dropped, err := render(entries, date, maxSource)
	if err != nil || dropped.count() != 0 {
		t.Fatalf("render: err=%v dropped=%+v", err, dropped)
	}
	again, _, _ := render(entries, date, maxSource)
	if !bytes.Equal(full, again) {
		t.Fatal("render is not deterministic")
	}
	for _, want := range []string{
		"// Code generated by go run ./gen -date=2026-09-25; DO NOT EDIT.",
		"// Snapshot 2026-09-25 of",
		"//   " + modelsDevURL + " (MIT, https://github.com/anomalyco/models.dev)",
		"//   " + liteLLMURL + " (MIT, https://github.com/BerriAI/litellm)",
		"// Entries: 8 (models.dev only 3, both 2, litellm only 3).",
		"// Dropped for the size cap: none.",
		"package registry",
		`const snapshotDate = "2026-09-25"`,
		"var entries = []entry{",
	} {
		if !strings.Contains(string(full), want) {
			t.Errorf("missing %q in output:\n%s", want, full)
		}
	}

	// One byte under the full size forces the first drop (LiteLLM-only).
	capped, dropped, err := render(entries, date, len(full)-1)
	if err != nil {
		t.Fatal(err)
	}
	if dropped.liteLLMOnly != 3 || dropped.flagless != 0 {
		t.Fatalf("dropped = %+v, want 3 litellm-only", dropped)
	}
	if !strings.Contains(string(capped), "// Entries: 5 (models.dev only 3, both 2, litellm only 0).") ||
		!strings.Contains(string(capped), "// Dropped for the size cap: 3 litellm-only entries, 0 entries stating no capability.") ||
		strings.Contains(string(capped), `"o1-pro"`) {
		t.Errorf("unexpected capped output:\n%s", capped)
	}

	// One byte under that forces the second drop (no capability stated).
	both, dropped, err := render(entries, date, len(capped)-1)
	if err != nil {
		t.Fatal(err)
	}
	if dropped.liteLLMOnly != 3 || dropped.flagless != 2 {
		t.Fatalf("dropped = %+v, want 3 litellm-only and 2 flagless", dropped)
	}
	for _, id := range []string{`"bar:free"`, `"google/gemma-3-27b-it"`} {
		if strings.Contains(string(both), id) {
			t.Errorf("flagless %s survived the cap", id)
		}
	}
	for _, id := range []string{`"gpt-4o"`, `"holo3:thinking"`, `"meta-llama/llama-3-70b"`} {
		if !strings.Contains(string(both), id) {
			t.Errorf("%s missing after the cap", id)
		}
	}

	if _, _, err := render(entries, date, 1); err == nil {
		t.Fatal("an impossible cap must fail, not write a truncated snapshot")
	}
}
