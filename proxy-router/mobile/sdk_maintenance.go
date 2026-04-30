package mobile

import (
	"context"
	"math/big"
	"time"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/lib"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/repositories/registries"
	"github.com/ethereum/go-ethereum/common"
)

// DefaultMaintenanceInterval is how often the background loop checks for expired sessions.
// Only naturally-expired sessions need this — provider-initiated closes already refund MOR
// on-chain immediately. This is a housekeeping sweep, not time-critical.
const DefaultMaintenanceInterval = 15 * time.Minute

// SetMaintenanceInterval restarts the session maintenance loop with a new interval.
// Pass 0 to disable automatic session closing entirely.
func (s *SDK) SetMaintenanceInterval(d time.Duration) {
	s.maintMu.Lock()
	defer s.maintMu.Unlock()

	if s.maintCancel != nil {
		s.maintCancel()
		s.maintCancel = nil
	}
	s.maintWg.Wait()

	if d > 0 {
		ctx, cancel := context.WithCancel(context.Background())
		s.maintCancel = cancel
		s.maintWg.Add(1)
		go s.runSessionMaintenance(ctx, d)
	}
}

func (s *SDK) startSessionMaintenance() {
	s.maintMu.Lock()
	defer s.maintMu.Unlock()

	ctx, cancel := context.WithCancel(context.Background())
	s.maintCancel = cancel
	s.maintWg.Add(1)
	go s.runSessionMaintenance(ctx, DefaultMaintenanceInterval)
}

// runSessionMaintenance periodically checks for expired on-chain sessions and auto-closes them
// to reclaim the user's locked MOR stake. It also detects provider-initiated closures by
// comparing against the on-chain state.
func (s *SDK) runSessionMaintenance(ctx context.Context, interval time.Duration) {
	defer s.maintWg.Done()
	log := s.log.Named("SESSION_MAINT")
	log.Infof("session maintenance started (interval %s)", interval)

	select {
	case <-ctx.Done():
		return
	case <-time.After(maintenanceStartupGrace):
	}

	s.runMaintenanceTick(ctx, log)

	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			log.Info("session maintenance stopped")
			return
		case <-ticker.C:
			s.runMaintenanceTick(ctx, log)
		}
	}
}

func (s *SDK) runMaintenanceTick(ctx context.Context, log lib.ILogger) {
	addr, err := s.getAddress()
	if err != nil {
		log.Warnf("cannot get wallet address: %s", err)
		return
	}
	user := common.HexToAddress(addr)
	now := time.Now().Unix()

	var offset big.Int
	var closed, alreadyClosed, active int

	for page := 0; page < paginationMaxPages; page++ {
		unclosed, _, totalFetched, err := s.blockchain.GetUnclosedUserSessions(ctx, user, &offset, paginationLimit, registries.OrderDESC)
		if err != nil {
			log.Warnf("error fetching sessions page %d: %s", page, err)
			break
		}
		if totalFetched == 0 {
			break
		}
		for _, ses := range unclosed {
			if ses == nil {
				continue
			}
			if ses.User.Hex() != addr {
				continue
			}
			endsAt := ses.EndsAt.Int64()
			if endsAt > 0 && endsAt < now {
				sessionHash := common.HexToHash(ses.Id)
				chainSes, err := s.blockchain.GetSession(ctx, sessionHash)
				if err != nil {
					log.Warnf("cannot fetch session %s from chain: %s", ses.Id, err)
					continue
				}
				if chainSes.ClosedAt.Int64() != 0 {
					alreadyClosed++
					continue
				}
				log.Infof("auto-closing expired session %s (ended %ds ago)", ses.Id, now-endsAt)
				_, err = s.blockchain.CloseSession(ctx, sessionHash)
				if err != nil {
					log.Warnf("failed to close session %s: %s", ses.Id, err)
					continue
				}
				closed++
			} else {
				active++
			}
		}
		offset.Add(&offset, big.NewInt(int64(totalFetched)))
	}

	if closed > 0 || alreadyClosed > 0 {
		log.Infof("maintenance: closed %d expired sessions, %d already closed, %d active", closed, alreadyClosed, active)
	}
}
