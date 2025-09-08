package blockchainapi

import (
	"context"
	"time"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/interfaces"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/lib"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/storages"
)

type SessionExpiryHandler struct {
	blockchainService *BlockchainService
	sessionStorage    *storages.SessionStorage
	wallet            interfaces.Wallet
	log               lib.ILogger
}

func NewSessionExpiryHandler(blockchainService *BlockchainService, sessionStorage *storages.SessionStorage, wallet interfaces.Wallet, log lib.ILogger) *SessionExpiryHandler {
	return &SessionExpiryHandler{
		blockchainService: blockchainService,
		sessionStorage:    sessionStorage,
		wallet:            wallet,
		log:               log.Named("SESSION_CLOSER"),
	}
}

func (s *SessionExpiryHandler) getWalletAddress() (string, error) {
	privateKey, err := s.wallet.GetPrivateKey()
	if err != nil {
		return "", err
	}

	addr, err := lib.PrivKeyBytesToAddr(privateKey)
	if err != nil {
		return "", err
	}

	return addr.Hex(), nil
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

			addr, err := s.getWalletAddress()
			if err != nil {
				s.log.Error(err)
				continue
			}
			for _, session := range sessions {
				if session.UserAddr == addr && session.EndsAt != nil && session.EndsAt.Int64() < time.Now().Unix() {
					sessionId, err := lib.HexToHash(session.Id)
					if err != nil {
						s.log.Error(err)
						continue
					}
					sessionData, err := s.blockchainService.GetSession(ctx, sessionId)
					if err != nil {
						s.log.Error(err)
						continue
					}
					if sessionData.ClosedAt.Int64() != 0 {
						s.log.Infof("Session %s already closed", session.Id)
						s.sessionStorage.RemoveSession(session.Id)
						continue
					}

					s.log.Infof("Closing session %s", session.Id)
					_, err = s.blockchainService.CloseSession(ctx, sessionId)
					if err != nil {
						s.log.Warnf("cannot close session: %s", err.Error())
						continue
					}
					s.sessionStorage.RemoveSession(session.Id)
				}
			}
		}
	}
}
