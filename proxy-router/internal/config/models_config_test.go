package config

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestModelConfigV2ParsesModelFamilyAndPresetApiType(t *testing.T) {
	raw := `{"models":[{"modelId":"0x01","modelName":"qwen3-235b","apiType":"vllm","modelFamily":"qwen3","apiUrl":"http://h:8000/v1/chat/completions"}]}`
	var v2 ModelConfigsV2
	require.NoError(t, json.Unmarshal([]byte(raw), &v2))
	require.Len(t, v2.Models, 1)
	require.Equal(t, "vllm", v2.Models[0].ApiType)
	require.Equal(t, "qwen3", v2.Models[0].ModelFamily)
}

func TestModelConfigModelFamilyIsOptional(t *testing.T) {
	var cfg ModelConfig
	require.NoError(t, json.Unmarshal([]byte(`{"modelName":"m","apiType":"openai","apiUrl":"http://h/v1"}`), &cfg))
	require.Equal(t, "", cfg.ModelFamily)
}
