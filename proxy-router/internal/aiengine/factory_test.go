package aiengine

import (
	"testing"
	"time"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/lib"
	"github.com/stretchr/testify/require"
)

func TestApiAdapterFactoryAcceptsStackPresets(t *testing.T) {
	log := lib.NewTestLogger()
	for _, preset := range []string{"vllm", "sglang", "llamacpp", "ollama", "venice", "openrouter", "litellm", "openai"} {
		engine, ok := ApiAdapterFactory(preset, "m", "http://h/v1/chat/completions", "", nil, time.Second, log, nil)
		require.True(t, ok, preset)
		require.Equal(t, API_TYPE_OPENAI, engine.ApiType(), preset)
	}
	engine, ok := ApiAdapterFactory("anthropic", "m", "http://h/v1/messages", "", nil, time.Second, log, nil)
	require.True(t, ok)
	require.Equal(t, API_TYPE_CLAUDEAI, engine.ApiType())

	_, ok = ApiAdapterFactory("bogus", "m", "http://h", "", nil, time.Second, log, nil)
	require.False(t, ok)
}
