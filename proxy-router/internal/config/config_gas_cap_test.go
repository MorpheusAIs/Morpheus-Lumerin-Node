package config

import (
	"testing"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/lib"
	"github.com/stretchr/testify/require"
)

// The tx escalator's gas-price ceiling must be on even when the operator never
// sets ETH_NODE_MAX_GAS_PRICE_GWEI (#1011).
func TestLoadConfigMaxGasPriceGwei(t *testing.T) {
	valid, err := NewValidator()
	require.NoError(t, err)
	t.Setenv("BLOCKSCOUT_API_URL", "http://localhost:4000")

	t.Run("unset defaults to 1 gwei", func(t *testing.T) {
		var cfg Config
		args := []string{"proxy-router"}
		require.NoError(t, LoadConfig(&cfg, &args, valid))
		require.Equal(t, lib.DefaultMaxGasPriceGwei, cfg.Blockchain.EthMaxGasPriceGwei)
		require.Equal(t, int64(1_000_000_000), lib.MaxGasPriceFromGwei(cfg.Blockchain.EthMaxGasPriceGwei).Int64())
	})

	t.Run("fractional gwei is kept", func(t *testing.T) {
		t.Setenv("ETH_NODE_MAX_GAS_PRICE_GWEI", "0.25")
		var cfg Config
		args := []string{"proxy-router"}
		require.NoError(t, LoadConfig(&cfg, &args, valid))
		require.Equal(t, 0.25, cfg.Blockchain.EthMaxGasPriceGwei)
	})

	t.Run("-1 disables the cap", func(t *testing.T) {
		t.Setenv("ETH_NODE_MAX_GAS_PRICE_GWEI", "-1")
		var cfg Config
		args := []string{"proxy-router"}
		require.NoError(t, LoadConfig(&cfg, &args, valid))
		require.Equal(t, -1.0, cfg.Blockchain.EthMaxGasPriceGwei)
		require.Nil(t, lib.MaxGasPriceFromGwei(cfg.Blockchain.EthMaxGasPriceGwei))
	})

	t.Run("below -1 is rejected", func(t *testing.T) {
		t.Setenv("ETH_NODE_MAX_GAS_PRICE_GWEI", "-5")
		var cfg Config
		args := []string{"proxy-router"}
		require.Error(t, LoadConfig(&cfg, &args, valid))
	})

	t.Run("is public config", func(t *testing.T) {
		var cfg Config
		cfg.Blockchain.EthMaxGasPriceGwei = 2
		public := cfg.GetSanitized().(Config)
		require.Equal(t, 2.0, public.Blockchain.EthMaxGasPriceGwei)
	})
}
