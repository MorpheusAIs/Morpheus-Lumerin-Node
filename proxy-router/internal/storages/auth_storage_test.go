package storages

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetAgentTxOrder(t *testing.T) {
	authStorage := agentTxsFixture(t)

	txs, newCursor, err := authStorage.GetAgentTxs("testuser", nil, 10)
	require.NoError(t, err)

	require.Equal(t, []string{"0x0004", "0x0003", "0x0002", "0x0001"}, txs)
	require.Equal(t, []byte(nil), newCursor)
}

func TestGetAgentTxCursor(t *testing.T) {
	authStorage := agentTxsFixture(t)

	txs, nextCursor, err := authStorage.GetAgentTxs("testuser", nil, 2)
	require.NoError(t, err)

	require.Equal(t, 2, len(txs))
	require.NotNil(t, nextCursor)
	require.Equal(t, []string{"0x0004", "0x0003"}, txs)

	txs, nextCursor, err = authStorage.GetAgentTxs("testuser", nextCursor, 2)
	require.NoError(t, err)

	require.Equal(t, 2, len(txs))
	require.Nil(t, nextCursor)
	require.Equal(t, []string{"0x0002", "0x0001"}, txs)
}

func TestGetAgentTxDifferentUser(t *testing.T) {
	authStorage := agentTxsFixture(t)

	txs, _, err := authStorage.GetAgentTxs("testuser2", nil, 10)
	require.NoError(t, err)
	require.Equal(t, 0, len(txs))

	err = authStorage.SetAgentTx("0x0005", "testuser2", big.NewInt(5))
	require.NoError(t, err)

	txs, _, err = authStorage.GetAgentTxs("testuser2", nil, 10)
	require.NoError(t, err)
	require.Equal(t, 1, len(txs))
	require.Equal(t, "0x0005", txs[0])
}

func agentTxsFixture(t *testing.T) *AuthStorage {
	db := NewTestStorage()
	authStorage := NewAuthStorage(db)
	blocks := []*big.Int{big.NewInt(4), big.NewInt(1), big.NewInt(3), big.NewInt(2)}

	for _, block := range blocks {
		err := authStorage.SetAgentTx("0x000"+block.String(), "testuser", block)
		require.NoError(t, err)
	}
	return authStorage
}
