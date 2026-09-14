package aiengine

import (
	"testing"
	"time"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/lib"
	"github.com/stretchr/testify/require"
)

// apiType is the transport adapter and nothing else: each legacy value maps
// to its engine as it did before the presets work, and an apiStack preset
// name (vllm, anthropic, …) is not an adapter — the loader keeps those in
// ModelConfig.ApiStack, which never reaches the factory.
func TestApiAdapterFactoryLegacyApiTypesOnly(t *testing.T) {
	log := lib.NewTestLogger()
	for _, apiType := range []string{API_TYPE_OPENAI, API_TYPE_CLAUDEAI, API_TYPE_PRODIA_SD, API_TYPE_PRODIA_SDXL, API_TYPE_PRODIA_V2, API_TYPE_HYPERBOLIC_SD} {
		engine, ok := ApiAdapterFactory(apiType, "m", "http://h/v1", "", nil, time.Second, log, nil)
		require.True(t, ok, apiType)
		require.Equal(t, apiType, engine.ApiType(), apiType)
	}
	for _, notAnAdapter := range []string{"vllm", "sglang", "llamacpp", "ollama", "venice", "openrouter", "litellm", "anthropic", "bogus", ""} {
		engine, ok := ApiAdapterFactory(notAnAdapter, "m", "http://h/v1", "", nil, time.Second, log, nil)
		require.False(t, ok, notAnAdapter)
		require.Nil(t, engine, notAnAdapter)
	}
}
