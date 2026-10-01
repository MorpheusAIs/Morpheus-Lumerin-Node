package proxyapi

import (
	"crypto/ecdsa"
	"testing"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/lib"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/storages"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/require"
)

func newTestController(t *testing.T) (*MORRPCController, *storages.SessionStorage) {
	t.Helper()
	ss := storages.NewSessionStorage(storages.NewTestStorage())
	c := &MORRPCController{sessionStorage: ss}
	return c, ss
}

func TestSessionUserPubKeyAcceptsMatchingRecord(t *testing.T) {
	c, ss := newTestController(t)

	key, err := crypto.GenerateKey()
	require.NoError(t, err)
	pub := crypto.FromECDSAPub(key.Public().(*ecdsa.PublicKey))
	addr := crypto.PubkeyToAddress(key.PublicKey)

	require.NoError(t, ss.AddUser(&storages.User{Addr: addr.Hex(), PubKey: lib.HexString(pub).Hex()}))

	user, pubHex, err := c.sessionUserPubKey(addr)
	require.NoError(t, err)
	require.NotNil(t, user)
	require.Equal(t, lib.HexString(pub).Hex(), pubHex.Hex())
}

func TestSessionUserPubKeyRejectsMismatchedRecord(t *testing.T) {
	c, ss := newTestController(t)

	victim, err := crypto.GenerateKey()
	require.NoError(t, err)
	victimAddr := crypto.PubkeyToAddress(victim.PublicKey)

	other, err := crypto.GenerateKey()
	require.NoError(t, err)
	otherPub := crypto.FromECDSAPub(other.Public().(*ecdsa.PublicKey))

	// a stale record where the stored key does not derive to the address
	require.NoError(t, ss.AddUser(&storages.User{Addr: victimAddr.Hex(), PubKey: lib.HexString(otherPub).Hex()}))

	_, _, err = c.sessionUserPubKey(victimAddr)
	require.Error(t, err)
	require.Contains(t, err.Error(), "does not match")
}

func TestSessionUserPubKeyMissingUser(t *testing.T) {
	c, _ := newTestController(t)
	key, err := crypto.GenerateKey()
	require.NoError(t, err)
	_, _, err = c.sessionUserPubKey(crypto.PubkeyToAddress(key.PublicKey))
	require.Error(t, err)
}
