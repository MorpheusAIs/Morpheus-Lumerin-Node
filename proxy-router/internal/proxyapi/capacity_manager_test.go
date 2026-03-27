package proxyapi

import (
	"math/big"
	"testing"
	"time"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/config"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/lib"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/storages"
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
	manager := NewIdleTimeoutCapacityManager(modelCfg, sessionStorage, &lib.LoggerMock{})

	// Recent activity should count as active, filling the single slot.
	require.NoError(t, sessionStorage.AddActivity(session.ModelID, &storages.PromptActivity{
		SessionID: session.Id,
		StartTime: now - 30,
		EndTime:   now - 10,
	}))
	require.False(t, manager.HasCapacity(session.ModelID))

	// Remove old activity and add stale one that is outside 15m idle timeout.
	require.NoError(t, sessionStorage.RemoveOldActivities(session.ModelID, now))
	require.NoError(t, sessionStorage.AddActivity(session.ModelID, &storages.PromptActivity{
		SessionID: session.Id,
		StartTime: now - 2000,
		EndTime:   now - 1800,
	}))
	require.True(t, manager.HasCapacity(session.ModelID))
}
