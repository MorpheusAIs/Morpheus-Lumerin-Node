package modelsfile

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/config"
	"github.com/stretchr/testify/require"
)

func writeTemp(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "models-config.json")
	require.NoError(t, os.WriteFile(path, []byte(content), 0644))
	return path
}

func TestLoadV2Format(t *testing.T) {
	path := writeTemp(t, `{"models": [
		{"modelId": "0x01", "modelName": "qwen3-32b", "apiType": "openai", "apiUrl": "http://a:8000/v1"},
		{"modelId": "0x02", "modelName": "llama-3.3-70b", "apiType": "openai", "apiUrl": "http://b:8000/v1"}
	]}`)
	entries, err := Load(path)
	require.NoError(t, err)
	require.Len(t, entries, 2)
	require.Equal(t, "0x01", entries[0].ID)
	require.Equal(t, "qwen3-32b", entries[0].Cfg.ModelName)
}

func TestLoadLegacyFormat(t *testing.T) {
	path := writeTemp(t, `{
		"0xbb": {"modelName": "deepseek-r1", "apiType": "openai", "apiUrl": "http://c:8000/v1"},
		"0xaa": {"modelName": "gpt-oss-120b", "apiType": "openai", "apiUrl": "http://d:8000/v1"}
	}`)
	entries, err := Load(path)
	require.NoError(t, err)
	require.Len(t, entries, 2)
	require.Equal(t, "0xaa", entries[0].ID)
}

func TestFilter(t *testing.T) {
	entries := []Entry{
		{ID: "0x01", Cfg: config.ModelConfig{ModelName: "Qwen3-32B"}},
		{ID: "0xabc", Cfg: config.ModelConfig{ModelName: "llama-3.3-70b"}},
	}
	require.Len(t, Filter(entries, ""), 2)
	require.Equal(t, "0x01", Filter(entries, "qwen")[0].ID)
	require.Equal(t, "0xabc", Filter(entries, "0xAB")[0].ID)
	require.Empty(t, Filter(entries, "mistral"))
}
