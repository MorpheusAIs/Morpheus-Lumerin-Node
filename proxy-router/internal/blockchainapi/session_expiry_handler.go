package blockchainapi

import (
	"context"
	"math/big"
	"time"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/interfaces"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/lib"
	r "github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/repositories/registries"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/storages"
	"github.com/ethereum/go-ethereum/common"
)

const (
	rehydrationPageSize       uint8 = 255
	earlyTerminationThreshold       = 3 // consecutive all-closed pages before stopping
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

// rehydrateFromChain queries the blockchain for all sessions opened by this
// wallet and populates the local BadgerDB cache with any that are still
// unclosed. This allows the node to recover session state after a restart
// with ephemeral storage, enabling it to both close expired orphaned sessions
// and resume routing for sessions that are still active.
//
// Two optimizations are applied:
//   - Two-pass filtering: session data is fetched first and filtered for
//     ClosedAt == 0 before fetching the associated bid data, avoiding
//     unnecessary RPC calls for already-closed sessions.
//   - Early termination: when scanning in descending order (newest first),
//     the scan stops after encountering several consecutive pages of
//     entirely-closed sessions, since older sessions are overwhelmingly
//     likely to also be closed.
func (s *SessionExpiryHandler) rehydrateFromChain(ctx context.Context) {
	addr, err := s.getWalletAddress()
	if err != nil {
		s.log.Errorf("rehydration: cannot get wallet address: %s", err)
		return
	}

	walletAddr := common.HexToAddress(addr)
	now := time.Now().Unix()
	var offset uint64
	var rehydratedExpired, rehydratedActive, totalClosed int
	var consecutiveAllClosedPages int

	s.log.Infof("Rehydrating session cache from chain for %s", addr)

	for {
		unclosed, closedInPage, fetched, err := s.blockchainService.GetUnclosedUserSessions(
			ctx, walletAddr,
			new(big.Int).SetUint64(offset), rehydrationPageSize,
			r.OrderDESC,
		)
		if err != nil {
			s.log.Errorf("rehydration: error at offset %d: %s", offset, err)
			break
		}
		if fetched == 0 {
			break
		}

		totalClosed += closedInPage

		if len(unclosed) == 0 && closedInPage > 0 {
			consecutiveAllClosedPages++
			if consecutiveAllClosedPages >= earlyTerminationThreshold {
				s.log.Infof(
					"Rehydration: %d consecutive all-closed pages at offset %d, stopping scan early",
					earlyTerminationThreshold, offset,
				)
				break
			}
		} else {
			consecutiveAllClosedPages = 0
		}

		for _, ses := range unclosed {
			err := s.sessionStorage.AddSession(&storages.Session{
				Id:           ses.Id,
				UserAddr:     ses.User.Hex(),
				ProviderAddr: ses.Provider.Hex(),
				EndsAt:       ses.EndsAt,
				ModelID:      ses.ModelAgentId,
			})
			if err != nil {
				s.log.Warnf("rehydration: error caching session %s: %s", ses.Id, err)
				continue
			}

			if ses.EndsAt.Int64() < now {
				rehydratedExpired++
			} else {
				rehydratedActive++
			}
		}

		offset += uint64(fetched)
	}

	s.log.Infof(
		"Rehydration complete: %d expired-unclosed, %d active, %d closed (skipped)",
		rehydratedExpired, rehydratedActive, totalClosed,
	)
}

// Run starts the session autoclose process, checking every minute if any session has ended and closes it.
func (s *SessionExpiryHandler) Run(ctx context.Context) error {
	s.rehydrateFromChain(ctx)

	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	s.log.Info("Session autoclose started")
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			sessions, err := s.sessionStorage.GetSessions(ctx)
			if err != nil {
				s.log.Errorf("error reading sessions for expiry check: %s", err)
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
						if err := s.sessionStorage.RemoveSession(session.Id); err != nil {
							s.log.Warnf("error removing session %s from cache: %s", session.Id, err)
						}
						continue
					}
					if sessionData.ClosedAt.Int64() != 0 {
						s.log.Infof("Session %s already closed", session.Id)
						if err := s.sessionStorage.RemoveSession(session.Id); err != nil {
							s.log.Warnf("error removing session %s from cache: %s", session.Id, err)
						}
						continue
					}

					s.log.Infof("Closing session %s", session.Id)
					_, err = s.blockchainService.CloseSession(ctx, sessionId)
					if err != nil {
						s.log.Warnf("cannot close session: %s", err.Error())
						continue
					}
					if err := s.sessionStorage.RemoveSession(session.Id); err != nil {
						s.log.Warnf("error removing session %s from cache: %s", session.Id, err)
					}
				}
			}
		}
	}
}
