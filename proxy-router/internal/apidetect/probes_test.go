package apidetect

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBareModelID(t *testing.T) {
	cases := []struct {
		name string
		id   string
		want string
	}{
		{"venice inline parameter", "deepseek-v4-pro:include_venice_system_prompt=false", "deepseek-v4-pro"},
		{"multiple chained parameters", "m:a=1:b=2", "m"},
		{"ollama tag unchanged", "gpt-oss:120b", "gpt-oss:120b"},
		{"ollama tag with extra segment unchanged", "qwen3:8b:tee", "qwen3:8b:tee"},
		{"trailing empty segment unchanged", "x:", "x:"},
		{"empty", "", ""},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			require.Equal(t, c.want, bareModelID(c.id))
		})
	}
}

func TestMatchModelEntryMatchesByBareID(t *testing.T) {
	data := []any{
		map[string]any{"id": "deepseek-v4-pro"},
		map[string]any{"id": "some-other-model"},
	}
	entry := matchModelEntry(data, "deepseek-v4-pro:include_venice_system_prompt=false")
	require.NotNil(t, entry)
	id, _ := entry["id"].(string)
	require.Equal(t, "deepseek-v4-pro", id)
}

func TestFindModelEntryBySpelling(t *testing.T) {
	listing := func(ids ...string) []any {
		out := make([]any, 0, len(ids))
		for _, id := range ids {
			out = append(out, map[string]any{"id": id})
		}
		return out
	}
	for _, tc := range []struct {
		name       string
		data       []any
		model      string
		want       string
		bySpelling bool
	}{
		{"dot read as dash, inline parameters dropped", listing("google-gemma-4-31b-it", "llama-3.3-70b"), "google.gemma-4-31b-it:include_venice_system_prompt=false", "google-gemma-4-31b-it", true},
		{"an exact id wins over a respelling", listing("llama-3-3-70b", "llama-3.3-70b"), "llama-3.3-70b", "llama-3.3-70b", false},
		{"org-prefixed id", listing("google/gemma-4-31b-it", "meta-llama/llama-3.3-70b"), "Gemma_4_31B_it", "google/gemma-4-31b-it", true},
		{"two respellings are ambiguous", listing("qwen3.5-9b", "qwen3-5-9b"), "Qwen3_5_9B", "", false},
		{"a different id stays unmatched", listing("mistral-small-3-2-24b-instruct", "llama-3.3-70b"), "mistral-31-24b", "", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			entry, bySpelling := findModelEntry(tc.data, tc.model)
			if tc.want == "" {
				require.Nil(t, entry)
				return
			}
			require.NotNil(t, entry)
			require.Equal(t, tc.want, entry["id"])
			require.Equal(t, tc.bySpelling, bySpelling)
		})
	}
}
