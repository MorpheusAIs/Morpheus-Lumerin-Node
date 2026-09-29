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
