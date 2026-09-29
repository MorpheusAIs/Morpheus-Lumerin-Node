package config

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// BLOCKSCOUT_API_URL: the only other required field with no default.
func TestLoadConfigModelApiDetectEnabledFromEnv(t *testing.T) {
	valid, err := NewValidator()
	require.NoError(t, err)
	t.Setenv("BLOCKSCOUT_API_URL", "http://localhost:4000")

	t.Setenv("MODEL_API_DETECT_ENABLED", "false")
	var cfg Config
	args := []string{"proxy-router"}
	require.NoError(t, LoadConfig(&cfg, &args, valid))
	require.NotNil(t, cfg.Proxy.ModelApiDetectEnabled.Bool)
	require.False(t, *cfg.Proxy.ModelApiDetectEnabled.Bool)
}

func TestLoadConfigModelApiDetectEnabledDefaultsTrueWhenUnset(t *testing.T) {
	valid, err := NewValidator()
	require.NoError(t, err)
	t.Setenv("BLOCKSCOUT_API_URL", "http://localhost:4000")

	var cfg Config
	args := []string{"proxy-router"}
	require.NoError(t, LoadConfig(&cfg, &args, valid))
	require.NotNil(t, cfg.Proxy.ModelApiDetectEnabled.Bool)
	require.True(t, *cfg.Proxy.ModelApiDetectEnabled.Bool)
}
