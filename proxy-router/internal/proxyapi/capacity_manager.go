package proxyapi

import (
	"time"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/config"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/lib"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/storages"
)

type CapacityManagerInterface interface {
	HasCapacity(modelID string) bool
}

type CapacityManager struct {
}

type SimpleCapacityManager struct {
	storage     *storages.SessionStorage
	modelConfig *config.ModelConfig
	log         lib.ILogger
}

func NewSimpleCapacityManager(modelConfig *config.ModelConfig, storage *storages.SessionStorage, log lib.ILogger) *SimpleCapacityManager {
	return &SimpleCapacityManager{
		storage:     storage,
		modelConfig: modelConfig,
		log:         log,
	}
}

func (scm *SimpleCapacityManager) HasCapacity(modelID string) bool {
	if scm.modelConfig.ConcurrentSlots == 0 {
		return true
	}

	sessions, err := scm.storage.GetSessionsForModel(modelID)
	if err != nil {
		scm.log.Error(err)
		return false
	}

	activeSessions := 0
	for _, session := range sessions {
		s, ok := scm.storage.GetSession(session)
		if !ok {
			continue
		}
		if s.EndsAt.Int64() > time.Now().Unix() {
			activeSessions++
		}
	}

	return activeSessions < scm.modelConfig.ConcurrentSlots
}

type IdleTimeoutCapacityManager struct {
	storage     *storages.SessionStorage
	modelConfig *config.ModelConfig
	log         lib.ILogger
}

func NewIdleTimeoutCapacityManager(modelConfig *config.ModelConfig, storage *storages.SessionStorage, log lib.ILogger) *IdleTimeoutCapacityManager {
	return &IdleTimeoutCapacityManager{
		storage:     storage,
		modelConfig: modelConfig,
		log:         log,
	}
}

func (idcm *IdleTimeoutCapacityManager) HasCapacity(modelID string) bool {
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
	for _, activity := range activities {
		session, ok := idcm.storage.GetSession(activity.SessionID)
		if !ok {
			continue
		}
		if session.EndsAt.Int64() < time.Now().Unix() {
			continue
		}
		if activity.EndTime > time.Now().Unix()-int64(IDLE_TIMEOUT) {
			activeSessions++
		}
	}

	REMOTE_IDLE_TIMEOUT := 60 * 60 // 60 min
	err = idcm.storage.RemoveOldActivities(modelID, time.Now().Unix()-int64(REMOTE_IDLE_TIMEOUT))
	if err != nil {
		idcm.log.Warnf("Failed to remove old activities: %s", err)
	}

	return activeSessions < idcm.modelConfig.ConcurrentSlots
}

func CreateCapacityManager(modelConfig *config.ModelConfig, storage *storages.SessionStorage, log lib.ILogger) CapacityManagerInterface {
	switch modelConfig.CapacityPolicy {
	case "simple":
		return NewSimpleCapacityManager(modelConfig, storage, log)
	case "idle_timeout":
		return NewIdleTimeoutCapacityManager(modelConfig, storage, log)
	default:
		return NewSimpleCapacityManager(modelConfig, storage, log)
	}
}
