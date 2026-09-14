package config

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// TestLoadConfigModelApiDetectEnabledFromEnv drives the real env-loading path
// cmd/main.go uses (LoadConfig, not a hand-built Config) to confirm
// MODEL_API_DETECT_ENABLED actually reaches Proxy.ModelApiDetectEnabled, and
// that leaving it unset still resolves to the SetDefaults default of true.
// BLOCKSCOUT_API_URL is set only because it is the one other required field
// with no default, so LoadConfig's validation step passes.
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
