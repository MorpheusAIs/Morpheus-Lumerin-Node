package storages

import (
	"math/big"
	"sync"
	"testing"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/lib"
	"github.com/stretchr/testify/require"
)

func TestDecreaseAllowanceRejectsWhenBalanceIsTooSmall(t *testing.T) {
	authStorage := agentAllowanceFixture(t, "100")

	err := authStorage.DecreaseAllowance("testuser", "eth", bigInt("100"))
	require.NoError(t, err)

	err = authStorage.DecreaseAllowance("testuser", "eth", bigInt("1"))
	require.Error(t, err)
	require.Contains(t, err.Error(), "not enough allowance")

	user, err := authStorage.GetAgentUser("testuser")
	require.NoError(t, err)
	eth := user.Allowances["eth"]
	require.Equal(t, "0", eth.String())
}

func TestDecreaseAllowanceIsAtomicAcrossOverlappingDebits(t *testing.T) {
	authStorage := agentAllowanceFixture(t, "10")

	var (
		wg        sync.WaitGroup
		mu        sync.Mutex
		successes int
	)
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := authStorage.DecreaseAllowance("testuser", "eth", bigInt("1"))
			if err == nil {
				mu.Lock()
				successes++
				mu.Unlock()
			}
		}()
	}
	wg.Wait()

	require.Equal(t, 10, successes)
	user, err := authStorage.GetAgentUser("testuser")
	require.NoError(t, err)
	eth := user.Allowances["eth"]
	require.Equal(t, "0", eth.String())
}

func TestIncreaseAllowanceRestoresAHeldDebit(t *testing.T) {
	authStorage := agentAllowanceFixture(t, "5")

	require.NoError(t, authStorage.DecreaseAllowance("testuser", "eth", bigInt("5")))
	require.NoError(t, authStorage.IncreaseAllowance("testuser", "eth", bigInt("5")))

	user, err := authStorage.GetAgentUser("testuser")
	require.NoError(t, err)
	eth := user.Allowances["eth"]
	require.Equal(t, "5", eth.String())
}

func agentAllowanceFixture(t *testing.T, amount string) *AuthStorage {
	t.Helper()
	db := NewTestStorage()
	authStorage := NewAuthStorage(db)
	err := authStorage.AddAuthRequest(&AgentUser{
		Username: "testuser",
		Allowances: map[string]lib.BigInt{
			"eth": bigInt(amount),
		},
		IsConfirmed: true,
	})
	require.NoError(t, err)
	return authStorage
}

func bigInt(value string) lib.BigInt {
	n, ok := new(big.Int).SetString(value, 10)
	if !ok {
		panic("invalid fixture amount")
	}
	return lib.BigInt{Int: *n}
}

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
