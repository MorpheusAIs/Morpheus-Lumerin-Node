package blockchainapi

import (
	"context"
	"time"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/lib"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/storages"
)

type SessionExpiryHandler struct {
	blockchainService *BlockchainService
	sessionStorage    *storages.SessionStorage
	log               lib.ILogger
}

func NewSessionExpiryHandler(blockchainService *BlockchainService, sessionStorage *storages.SessionStorage, log lib.ILogger) *SessionExpiryHandler {
	return &SessionExpiryHandler{
		blockchainService: blockchainService,
		sessionStorage:    sessionStorage,
		log:               log,
	}
}

// Run starts the session autoclose process, checking every minute if any session has ended and closes it.
func (s *SessionExpiryHandler) Run(ctx context.Context) error {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	s.log.Info("Session autoclose started")
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			sessions, err := s.sessionStorage.GetSessions()
			if err != nil {
				s.log.Error(err)
				continue
			}
			for _, session := range sessions {
				if session.EndsAt.Int64() < time.Now().Unix() {
					sessionId, err := lib.HexToHash(session.Id)
					if err != nil {
						s.log.Error(err)
						continue
					}
					_, err = s.blockchainService.CloseSession(ctx, sessionId)
					if err != nil {
						s.log.Error(err)
						continue
					}
					s.sessionStorage.RemoveSession(session.Id)
				}
			}
		}
	}
}
