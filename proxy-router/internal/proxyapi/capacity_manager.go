package proxyapi

import (
	"context"
	"time"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/blockchainapi/structs"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/config"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/lib"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/storages"
	"github.com/ethereum/go-ethereum/common"
)

// closedSessionCheckTimeout bounds the on-chain re-check that runs before a
// session request is refused for lack of capacity.
const closedSessionCheckTimeout = 5 * time.Second

type CapacityManagerInterface interface {
	HasCapacity(ctx context.Context, modelID string) bool
}

// SessionChainReader reads the current on-chain state of a session. Capacity
// managers use it to confirm that a session still holding a slot in local
// storage has not already been closed on chain (implemented by
// blockchainapi.BlockchainService).
type SessionChainReader interface {
	GetSession(ctx context.Context, sessionID common.Hash) (*structs.Session, error)
}

// slotHolder is a session counted against the model's concurrent slots, with
// the number of slots the count attributes to it.
type slotHolder struct {
	sessionID string
	slots     int
}

// releaseClosedSessions runs only when the local count says the model is full.
// The local count lags the chain until the SessionClosed event is processed,
// so it re-reads the counted sessions from chain and returns how many slots
// are held by sessions that are already closed there. A closed session is
// removed from local storage so the next request does not pay for the same
// read. A session whose chain read fails keeps its slot: an RPC error must
// never open an extra slot. It stops once `needed` slots are free.
func releaseClosedSessions(ctx context.Context, chain SessionChainReader, storage *storages.SessionStorage, holders []slotHolder, needed int, log lib.ILogger) int {
	if chain == nil || needed <= 0 {
		return 0
	}

	ctx, cancel := context.WithTimeout(ctx, closedSessionCheckTimeout)
	defer cancel()

	freed := 0
	for _, holder := range holders {
		if freed >= needed {
			break
		}

		ses, err := chain.GetSession(ctx, common.HexToHash(holder.sessionID))
		if err != nil {
			log.Warnf("error re-checking session %s on chain, keeping its slot: %s", holder.sessionID, err)
			continue
		}
		if ses == nil || ses.ClosedAt == nil || ses.ClosedAt.Sign() == 0 {
			continue
		}

		log.Infof("session %s is already closed on chain, releasing its slot", holder.sessionID)
		if err := storage.RemoveSession(holder.sessionID); err != nil {
			log.Warnf("error removing closed session %s from storage: %s", holder.sessionID, err)
		}
		freed += holder.slots
	}

	return freed
}

type SimpleCapacityManager struct {
	storage     *storages.SessionStorage
	chain       SessionChainReader
	modelConfig *config.ModelConfig
	log         lib.ILogger
}

func NewSimpleCapacityManager(modelConfig *config.ModelConfig, storage *storages.SessionStorage, chain SessionChainReader, log lib.ILogger) *SimpleCapacityManager {
	return &SimpleCapacityManager{
		storage:     storage,
		chain:       chain,
		modelConfig: modelConfig,
		log:         log,
	}
}

func (scm *SimpleCapacityManager) HasCapacity(ctx context.Context, modelID string) bool {
	if scm.modelConfig.ConcurrentSlots == 0 {
		return true
	}

	sessions, err := scm.storage.GetSessionsForModel(modelID)
	if err != nil {
		scm.log.Error(err)
		return false
	}

	holders := make([]slotHolder, 0, len(sessions))
	for _, session := range sessions {
		s, err := scm.storage.GetSession(session)
		if err != nil {
			scm.log.Warnf("error reading session %s: %s", session, err)
			continue
		}
		if s == nil {
			continue
		}
		if s.EndsAt.Int64() > time.Now().Unix() {
			holders = append(holders, slotHolder{sessionID: session, slots: 1})
		}
	}

	activeSessions := len(holders)
	if activeSessions < scm.modelConfig.ConcurrentSlots {
		return true
	}

	needed := activeSessions - scm.modelConfig.ConcurrentSlots + 1
	freed := releaseClosedSessions(ctx, scm.chain, scm.storage, holders, needed, scm.log)

	return activeSessions-freed < scm.modelConfig.ConcurrentSlots
}

type IdleTimeoutCapacityManager struct {
	storage     *storages.SessionStorage
	chain       SessionChainReader
	modelConfig *config.ModelConfig
	log         lib.ILogger
}

func NewIdleTimeoutCapacityManager(modelConfig *config.ModelConfig, storage *storages.SessionStorage, chain SessionChainReader, log lib.ILogger) *IdleTimeoutCapacityManager {
	return &IdleTimeoutCapacityManager{
		storage:     storage,
		chain:       chain,
		modelConfig: modelConfig,
		log:         log,
	}
}

func (idcm *IdleTimeoutCapacityManager) HasCapacity(ctx context.Context, modelID string) bool {
	IDLE_TIMEOUT := 15 * 60 // 15 min

	if idcm.modelConfig.ConcurrentSlots == 0 {
		return true
	}

	activities, err := idcm.storage.GetActivities(modelID)
	if err != nil {
		idcm.log.Error(err)
		return false
	}

	activeSessions := 0
	slotsBySession := make(map[string]int)
	holderOrder := make([]string, 0)
	for _, activity := range activities {
		session, err := idcm.storage.GetSession(activity.SessionID)
		if err != nil {
			idcm.log.Warnf("error reading session %s: %s", activity.SessionID, err)
			continue
		}
		if session == nil {
			continue
		}
		if session.EndsAt.Int64() < time.Now().Unix() {
			continue
		}
		if activity.EndTime > time.Now().Unix()-int64(IDLE_TIMEOUT) {
			activeSessions++
			if _, ok := slotsBySession[activity.SessionID]; !ok {
				holderOrder = append(holderOrder, activity.SessionID)
			}
			slotsBySession[activity.SessionID]++
		}
	}

	REMOTE_IDLE_TIMEOUT := 60 * 60 // 60 min
	err = idcm.storage.RemoveOldActivities(modelID, time.Now().Unix()-int64(REMOTE_IDLE_TIMEOUT))
	if err != nil {
		idcm.log.Warnf("Failed to remove old activities: %s", err)
	}

	if activeSessions < idcm.modelConfig.ConcurrentSlots {
		return true
	}

	holders := make([]slotHolder, 0, len(holderOrder))
	for _, sessionID := range holderOrder {
		holders = append(holders, slotHolder{sessionID: sessionID, slots: slotsBySession[sessionID]})
	}

	needed := activeSessions - idcm.modelConfig.ConcurrentSlots + 1
	freed := releaseClosedSessions(ctx, idcm.chain, idcm.storage, holders, needed, idcm.log)

	return activeSessions-freed < idcm.modelConfig.ConcurrentSlots
}

func CreateCapacityManager(modelConfig *config.ModelConfig, storage *storages.SessionStorage, chain SessionChainReader, log lib.ILogger) CapacityManagerInterface {
	switch modelConfig.CapacityPolicy {
	case "simple":
		return NewSimpleCapacityManager(modelConfig, storage, chain, log)
	case "idle_timeout":
		return NewIdleTimeoutCapacityManager(modelConfig, storage, chain, log)
	default:
		return NewSimpleCapacityManager(modelConfig, storage, chain, log)
	}
}
