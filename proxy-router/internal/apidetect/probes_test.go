package apidetect

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// TestBareModelID: Venice appends inline "key=value" parameters to a model
// id as extra ":"-separated segments — observed live behind a LiteLLM proxy
// as "deepseek-v4-pro:include_venice_system_prompt=false" — possibly several
// chained. Only segments containing "=" are stripped; an Ollama-style tag
// carries no "=" and must survive unchanged.
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

// matchModelEntry must find a listing entry by the bare id when the exact
// suffixed name has no match and the listing has more than one entry, so the
// single-entry-list fallback cannot mask this path — the shape of Venice's
// real listing (100+ models) reached through a LiteLLM deployment whose
// upstream model id carries an inline parameter.
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
