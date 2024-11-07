package wallet

import (
	"testing"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/lib"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
)

func TestIssue172(t *testing.T) {
	// https://github.com/btcsuite/btcutil/pull/182
	// makes sure that the private key is derived correctly
	seed := common.FromHex("00000000000000500000000000000000")
	path, err := accounts.ParseDerivationPath("m/44'/60'/0'/0/0")
	require.NoError(t, err)
	expectedAddr := common.HexToAddress("0xe39Be4d7E9D91D14e837589F3027798f3911A83c")

	wallet, err := newWallet(seed)
	require.NoError(t, err)

	prkey, err := wallet.DerivePrivateKey(path)
	require.NoError(t, err)

	addr, err := lib.PrivKeyToAddr(prkey)
	require.NoError(t, err)

	require.Equal(t, expectedAddr, addr)
}
