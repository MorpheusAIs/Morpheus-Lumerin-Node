package lib

import (
	"crypto/ecdsa"
	"testing"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/require"
)

func TestPubKeyBytesToAddrMatchesWallet(t *testing.T) {
	key, err := crypto.GenerateKey()
	require.NoError(t, err)
	pub := crypto.FromECDSAPub(key.Public().(*ecdsa.PublicKey))
	addr, err := PubKeyBytesToAddr(pub)
	require.NoError(t, err)
	require.Equal(t, crypto.PubkeyToAddress(key.PublicKey), addr)
}

func TestPubKeyBytesToAddrRejectsGarbage(t *testing.T) {
	_, err := PubKeyBytesToAddr([]byte("not-a-key"))
	require.Error(t, err)
}
