package config

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// The inference-traffic probes cost tokens on the provider's backend, so
// every switch is off unless set.
func TestSetDefaultsInferenceProbeFlagsOff(t *testing.T) {
	cfg := newConfigForDefaults()
	cfg.SetDefaults()
	require.False(t, cfg.Proxy.ModelApiVerifyBindings)
	require.False(t, cfg.Proxy.ModelApiProbeCapabilities)

	cfg = newConfigForDefaults()
	cfg.Proxy.ModelApiVerifyBindings = true
	cfg.Proxy.ModelApiProbeCapabilities = true
	cfg.SetDefaults()
	require.True(t, cfg.Proxy.ModelApiVerifyBindings, "an explicit value is kept")
	require.True(t, cfg.Proxy.ModelApiProbeCapabilities)

	public := cfg.GetSanitized().(Config)
	require.True(t, public.Proxy.ModelApiVerifyBindings, "the switches are public config")
	require.True(t, public.Proxy.ModelApiProbeCapabilities)
}

func TestLoadConfigInferenceProbeFlagsFromEnv(t *testing.T) {
	valid, err := NewValidator()
	require.NoError(t, err)
	t.Setenv("BLOCKSCOUT_API_URL", "http://localhost:4000")

	t.Run("unset is off", func(t *testing.T) {
		var cfg Config
		args := []string{"proxy-router"}
		require.NoError(t, LoadConfig(&cfg, &args, valid))
		require.False(t, cfg.Proxy.ModelApiVerifyBindings)
		require.False(t, cfg.Proxy.ModelApiProbeCapabilities)
	})

	t.Run("set reads true", func(t *testing.T) {
		t.Setenv("MODEL_API_VERIFY_BINDINGS", "true")
		t.Setenv("MODEL_API_PROBE_CAPABILITIES", "true")
		var cfg Config
		args := []string{"proxy-router"}
		require.NoError(t, LoadConfig(&cfg, &args, valid))
		require.True(t, cfg.Proxy.ModelApiVerifyBindings)
		require.True(t, cfg.Proxy.ModelApiProbeCapabilities)
	})
}
