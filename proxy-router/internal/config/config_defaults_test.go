package config

import (
	"testing"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/lib"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
)

// SetDefaults dereferences the pointer fields the env loader allocates, so a
// hand-built Config must allocate them too.
func newConfigForDefaults() *Config {
	cfg := &Config{}
	cfg.Blockchain.Multicall3Addr = &common.Address{}
	cfg.Proxy.StoreChatContext = &lib.Bool{}
	cfg.Proxy.ForwardChatContext = &lib.Bool{}
	cfg.Proxy.ModelApiDetectEnabled = &lib.Bool{}
	return cfg
}

// R6: MODEL_API_DETECT_ENABLED defaults to true and an explicit false survives.
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
