package proxyapi

import (
	"context"
	"errors"
	"math/big"
	"testing"
	"time"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/blockchainapi/structs"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/config"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/lib"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/storages"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
)

func TestIdleTimeoutCapacityManager_HasCapacity(t *testing.T) {
	now := time.Now().Unix()
	storage := storages.NewTestStorage()
	sessionStorage := storages.NewSessionStorage(storage)

	session := &storages.Session{
		Id:       "0xsession1",
		ModelID:  "0xmodel1",
		EndsAt:   big.NewInt(now + 3600),
		UserAddr: "0xuser",
	}
	require.NoError(t, sessionStorage.AddSession(session))

	modelCfg := &config.ModelConfig{
		ConcurrentSlots: 1,
		CapacityPolicy:  "idle_timeout",
	}
	manager := NewIdleTimeoutCapacityManager(modelCfg, sessionStorage, nil, &lib.LoggerMock{})

	// Recent activity should count as active, filling the single slot.
	require.NoError(t, sessionStorage.AddActivity(session.ModelID, &storages.PromptActivity{
		SessionID: session.Id,
		StartTime: now - 30,
		EndTime:   now - 10,
	}))
	require.False(t, manager.HasCapacity(context.Background(), session.ModelID))

	// Remove old activity and add stale one that is outside 15m idle timeout.
	require.NoError(t, sessionStorage.RemoveOldActivities(session.ModelID, now))
	require.NoError(t, sessionStorage.AddActivity(session.ModelID, &storages.PromptActivity{
		SessionID: session.Id,
		StartTime: now - 2000,
		EndTime:   now - 1800,
	}))
	require.True(t, manager.HasCapacity(context.Background(), session.ModelID))
}

// fakeSessionChain serves on-chain session state for the capacity re-check.
type fakeSessionChain struct {
	closedAt map[common.Hash]int64
	err      error
	calls    map[common.Hash]int
}

func newFakeSessionChain() *fakeSessionChain {
	return &fakeSessionChain{
		closedAt: make(map[common.Hash]int64),
		calls:    make(map[common.Hash]int),
	}
}

func (f *fakeSessionChain) GetSession(ctx context.Context, sessionID common.Hash) (*structs.Session, error) {
	f.calls[sessionID]++
	if f.err != nil {
		return nil, f.err
	}
	return &structs.Session{
		Id:       sessionID.Hex(),
		ClosedAt: big.NewInt(f.closedAt[sessionID]),
	}, nil
}

func (f *fakeSessionChain) totalCalls() int {
	total := 0
	for _, n := range f.calls {
		total += n
	}
	return total
}

var capacityPolicies = []string{"simple", "idle_timeout"}

// addOpenSession stores a session that is still open locally (EndsAt in the
// future) and records a recent prompt on it, so both capacity policies count
// it against the model's slots.
func addOpenSession(t *testing.T, sessionStorage *storages.SessionStorage, sessionID common.Hash, modelID string) {
	t.Helper()
	now := time.Now().Unix()
	require.NoError(t, sessionStorage.AddSession(&storages.Session{
		Id:       sessionID.Hex(),
		ModelID:  modelID,
		EndsAt:   big.NewInt(now + 3600),
		UserAddr: "0xuser",
	}))
	require.NoError(t, sessionStorage.AddActivity(modelID, &storages.PromptActivity{
		SessionID: sessionID.Hex(),
		StartTime: now - 30,
		EndTime:   now - 10,
	}))
}

func TestCapacityManager_LocalOpenChainClosedAllows(t *testing.T) {
	for _, policy := range capacityPolicies {
		t.Run(policy, func(t *testing.T) {
			sessionStorage := storages.NewSessionStorage(storages.NewTestStorage())
			modelID := common.HexToHash("0xaa").Hex()
			sessionID := common.HexToHash("0x01")
			addOpenSession(t, sessionStorage, sessionID, modelID)

			chain := newFakeSessionChain()
			chain.closedAt[sessionID] = time.Now().Unix()

			modelCfg := &config.ModelConfig{ConcurrentSlots: 1, CapacityPolicy: policy}
			manager := CreateCapacityManager(modelCfg, sessionStorage, chain, &lib.LoggerMock{})

			require.True(t, manager.HasCapacity(context.Background(), modelID))
			require.Equal(t, 1, chain.calls[sessionID])

			// The closed session is removed from local storage, so the next
			// request does not read it from chain again.
			ses, err := sessionStorage.GetSession(sessionID.Hex())
			require.NoError(t, err)
			require.Nil(t, ses)

			require.True(t, manager.HasCapacity(context.Background(), modelID))
			require.Equal(t, 1, chain.calls[sessionID])
		})
	}
}

func TestCapacityManager_ChainOpenKeepsSlot(t *testing.T) {
	for _, policy := range capacityPolicies {
		t.Run(policy, func(t *testing.T) {
			sessionStorage := storages.NewSessionStorage(storages.NewTestStorage())
			modelID := common.HexToHash("0xaa").Hex()
			sessionID := common.HexToHash("0x01")
			addOpenSession(t, sessionStorage, sessionID, modelID)

			chain := newFakeSessionChain()

			modelCfg := &config.ModelConfig{ConcurrentSlots: 1, CapacityPolicy: policy}
			manager := CreateCapacityManager(modelCfg, sessionStorage, chain, &lib.LoggerMock{})

			require.False(t, manager.HasCapacity(context.Background(), modelID))
			require.Equal(t, 1, chain.calls[sessionID])

			ses, err := sessionStorage.GetSession(sessionID.Hex())
			require.NoError(t, err)
			require.NotNil(t, ses)
		})
	}
}

func TestCapacityManager_ChainReadErrorFailsClosed(t *testing.T) {
	for _, policy := range capacityPolicies {
		t.Run(policy, func(t *testing.T) {
			sessionStorage := storages.NewSessionStorage(storages.NewTestStorage())
			modelID := common.HexToHash("0xaa").Hex()
			sessionID := common.HexToHash("0x01")
			addOpenSession(t, sessionStorage, sessionID, modelID)

			chain := newFakeSessionChain()
			chain.err = errors.New("rpc unavailable")

			modelCfg := &config.ModelConfig{ConcurrentSlots: 1, CapacityPolicy: policy}
			manager := CreateCapacityManager(modelCfg, sessionStorage, chain, &lib.LoggerMock{})

			require.False(t, manager.HasCapacity(context.Background(), modelID))

			ses, err := sessionStorage.GetSession(sessionID.Hex())
			require.NoError(t, err)
			require.NotNil(t, ses)
		})
	}
}

func TestCapacityManager_NoChainReadWhenSlotFree(t *testing.T) {
	for _, policy := range capacityPolicies {
		t.Run(policy, func(t *testing.T) {
			sessionStorage := storages.NewSessionStorage(storages.NewTestStorage())
			modelID := common.HexToHash("0xaa").Hex()
			addOpenSession(t, sessionStorage, common.HexToHash("0x01"), modelID)

			chain := newFakeSessionChain()

			modelCfg := &config.ModelConfig{ConcurrentSlots: 2, CapacityPolicy: policy}
			manager := CreateCapacityManager(modelCfg, sessionStorage, chain, &lib.LoggerMock{})

			require.True(t, manager.HasCapacity(context.Background(), modelID))
			require.Equal(t, 0, chain.totalCalls())
		})
	}
}

func TestCapacityManager_OnlyClosedSessionsFreeSlots(t *testing.T) {
	for _, policy := range capacityPolicies {
		t.Run(policy, func(t *testing.T) {
			sessionStorage := storages.NewSessionStorage(storages.NewTestStorage())
			modelID := common.HexToHash("0xaa").Hex()
			openID := common.HexToHash("0x01")
			closedID := common.HexToHash("0x02")
			addOpenSession(t, sessionStorage, openID, modelID)
			addOpenSession(t, sessionStorage, closedID, modelID)

			chain := newFakeSessionChain()
			chain.closedAt[closedID] = time.Now().Unix()

			modelCfg := &config.ModelConfig{ConcurrentSlots: 2, CapacityPolicy: policy}
			manager := CreateCapacityManager(modelCfg, sessionStorage, chain, &lib.LoggerMock{})

			require.True(t, manager.HasCapacity(context.Background(), modelID))

			ses, err := sessionStorage.GetSession(openID.Hex())
			require.NoError(t, err)
			require.NotNil(t, ses)

			ses, err = sessionStorage.GetSession(closedID.Hex())
			require.NoError(t, err)
			require.Nil(t, ses)
		})
	}
}

// With concurrentSlots: 1, a consumer that closes its session early and
// reopens on the same bid right away must get the slot as soon as the close
// is on chain, without waiting for the SessionClosed event.
func TestCapacityManager_SingleSlotEarlyCloseImmediateReopen(t *testing.T) {
	for _, policy := range capacityPolicies {
		t.Run(policy, func(t *testing.T) {
			sessionStorage := storages.NewSessionStorage(storages.NewTestStorage())
			modelID := common.HexToHash("0xaa").Hex()
			firstID := common.HexToHash("0x01")
			secondID := common.HexToHash("0x02")
			addOpenSession(t, sessionStorage, firstID, modelID)

			chain := newFakeSessionChain()
			modelCfg := &config.ModelConfig{ConcurrentSlots: 1, CapacityPolicy: policy}

			// First session is still open on chain: the slot is taken.
			manager := CreateCapacityManager(modelCfg, sessionStorage, chain, &lib.LoggerMock{})
			require.False(t, manager.HasCapacity(context.Background(), modelID))

			// The consumer closes early; the close is mined but the
			// SessionClosed event has not been processed, so local storage
			// still holds the first session.
			chain.closedAt[firstID] = time.Now().Unix()
			ses, err := sessionStorage.GetSession(firstID.Hex())
			require.NoError(t, err)
			require.NotNil(t, ses)

			// The immediate reopen is accepted.
			manager = CreateCapacityManager(modelCfg, sessionStorage, chain, &lib.LoggerMock{})
			require.True(t, manager.HasCapacity(context.Background(), modelID))
			require.Equal(t, 2, chain.calls[firstID])

			// The new session takes the slot; the first one is not read again.
			addOpenSession(t, sessionStorage, secondID, modelID)
			manager = CreateCapacityManager(modelCfg, sessionStorage, chain, &lib.LoggerMock{})
			require.False(t, manager.HasCapacity(context.Background(), modelID))
			require.Equal(t, 2, chain.calls[firstID])
			require.Equal(t, 1, chain.calls[secondID])
		})
	}
}
