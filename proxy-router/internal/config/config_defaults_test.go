package config

import (
	"testing"
	"time"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/lib"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
)

// SetDefaults dereferences these; a hand-built Config must allocate them.
func newConfigForDefaults() *Config {
	cfg := &Config{}
	cfg.Blockchain.Multicall3Addr = &common.Address{}
	cfg.Proxy.StoreChatContext = &lib.Bool{}
	cfg.Proxy.ForwardChatContext = &lib.Bool{}
	cfg.Proxy.ModelApiDetectEnabled = &lib.Bool{}
	return cfg
}

func TestSetDefaultsModelApiDetectEnabled(t *testing.T) {
	cfg := newConfigForDefaults()
	cfg.SetDefaults()
	require.NotNil(t, cfg.Proxy.ModelApiDetectEnabled.Bool)
	require.True(t, *cfg.Proxy.ModelApiDetectEnabled.Bool)

	off := false
	cfg = newConfigForDefaults()
	cfg.Proxy.ModelApiDetectEnabled = &lib.Bool{Bool: &off}
	cfg.SetDefaults()
	require.False(t, *cfg.Proxy.ModelApiDetectEnabled.Bool)

	public := cfg.GetSanitized().(Config)
	require.NotNil(t, public.Proxy.ModelApiDetectEnabled)
	require.False(t, *public.Proxy.ModelApiDetectEnabled.Bool, "the switch is public config")
}

func TestSetDefaultsModelApiDetectTTL(t *testing.T) {
	t.Run("unset follows MODEL_HEALTH_CHECK_INTERVAL", func(t *testing.T) {
		cfg := newConfigForDefaults()
		cfg.Proxy.ModelHealthCheckInterval = 2 * time.Hour
		cfg.SetDefaults()
		require.Equal(t, 2*time.Hour, cfg.Proxy.ModelApiDetectTTL)
	})

	t.Run("both unset is one hour", func(t *testing.T) {
		cfg := newConfigForDefaults()
		cfg.SetDefaults()
		require.Equal(t, time.Hour, cfg.Proxy.ModelHealthCheckInterval)
		require.Equal(t, time.Hour, cfg.Proxy.ModelApiDetectTTL)
	})

	t.Run("values under the floor are clamped", func(t *testing.T) {
		cfg := newConfigForDefaults()
		cfg.Proxy.ModelApiDetectTTL = 10 * time.Second
		cfg.SetDefaults()
		require.Equal(t, MinModelApiDetectTTL, cfg.Proxy.ModelApiDetectTTL)
	})

	t.Run("explicit values are kept and public", func(t *testing.T) {
		cfg := newConfigForDefaults()
		cfg.Proxy.ModelApiDetectTTL = 5 * time.Minute
		cfg.SetDefaults()
		require.Equal(t, 5*time.Minute, cfg.Proxy.ModelApiDetectTTL)
		public := cfg.GetSanitized().(Config)
		require.Equal(t, 5*time.Minute, public.Proxy.ModelApiDetectTTL)
	})
}
