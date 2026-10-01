package proxyapi

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"testing"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/lib"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/storages"
	"github.com/ethereum/go-ethereum/common"
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

type fakeAuthResolver struct {
	signers map[common.Address]common.Address
	err     error
}

func (f fakeAuthResolver) AuthorizedSigner(_ context.Context, provider common.Address) (common.Address, error) {
	if f.err != nil {
		return common.Address{}, f.err
	}
	if s, ok := f.signers[provider]; ok {
		return s, nil
	}
	return provider, nil // EOA signs as itself
}

func TestServesProvider(t *testing.T) {
	me := common.HexToAddress("0x1111111111111111111111111111111111111111")
	contract := common.HexToAddress("0xc0dec0dec0dec0dec0dec0dec0dec0dec0dec0de")
	other := common.HexToAddress("0x2222222222222222222222222222222222222222")
	ctx := context.Background()

	// no resolver: only sessions opened directly against my address
	c := &MORRPCController{providerAddr: me}
	require.True(t, c.servesProvider(ctx, me))
	require.False(t, c.servesProvider(ctx, contract))
	require.False(t, c.servesProvider(ctx, other))

	// resolver: contract whose owner() is me is served; foreign EOA/contract is not
	c.SetProviderAuthResolver(fakeAuthResolver{signers: map[common.Address]common.Address{
		contract: me,
		other:    other,
	}})
	require.True(t, c.servesProvider(ctx, me))
	require.True(t, c.servesProvider(ctx, contract))
	require.False(t, c.servesProvider(ctx, other))

	// resolver error fails closed for anything that is not my own address
	c.SetProviderAuthResolver(fakeAuthResolver{err: errors.New("rpc down")})
	require.True(t, c.servesProvider(ctx, me))
	require.False(t, c.servesProvider(ctx, contract))

	// unknown node key fails closed everywhere
	z := &MORRPCController{}
	z.SetProviderAuthResolver(fakeAuthResolver{signers: map[common.Address]common.Address{contract: common.Address{}}})
	require.False(t, z.servesProvider(ctx, common.Address{}))
	require.False(t, z.servesProvider(ctx, contract))
}

func TestSessionUserPubKeyMissingUser(t *testing.T) {
	c, _ := newTestController(t)
	key, err := crypto.GenerateKey()
	require.NoError(t, err)
	_, _, err = c.sessionUserPubKey(crypto.PubkeyToAddress(key.PublicKey))
	require.Error(t, err)
}
