package blockchainapi

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/blockchainapi/structs"
	sessionrouter "github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/repositories/contracts/bindings/sessionrouter"
	r "github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/repositories/registries"
	"github.com/ethereum/go-ethereum/common"
)

const guardedSessionScanPageSize uint8 = 255

var ErrExistingLiveSession = errors.New("a live session already exists for this model")

// ExistingSessionError is returned by the opt-in guarded open flow. Keeping the
// session ID on a typed error lets the HTTP controller return a machine-readable
// conflict without changing the longstanding OpenSessionByModelId contract.
type ExistingSessionError struct {
	SessionID string
}

func (e *ExistingSessionError) Error() string {
	return fmt.Sprintf("%s: %s", ErrExistingLiveSession, e.SessionID)
}

func (e *ExistingSessionError) Unwrap() error {
	return ErrExistingLiveSession
}

type keyedLockEntry struct {
	mu   sync.Mutex
	refs int
}

// keyedLockSet serializes guarded opens for one wallet/model pair without
// making unrelated model purchases wait behind a blockchain transaction.
// Entries are reference-counted and removed after the last waiter leaves.
type keyedLockSet struct {
	mu      sync.Mutex
	entries map[string]*keyedLockEntry
}

func (s *keyedLockSet) lock(key string) func() {
	s.mu.Lock()
	if s.entries == nil {
		s.entries = make(map[string]*keyedLockEntry)
	}
	entry := s.entries[key]
	if entry == nil {
		entry = &keyedLockEntry{}
		s.entries[key] = entry
	}
	entry.refs++
	s.mu.Unlock()

	entry.mu.Lock()
	return func() {
		entry.mu.Unlock()
		s.mu.Lock()
		entry.refs--
		if entry.refs == 0 {
			delete(s.entries, key)
		}
		s.mu.Unlock()
	}
}

type liveSessionPageFetcher func(context.Context, *big.Int) (*structs.Session, int, error)

func isLiveUserSession(session sessionrouter.ISessionStorageSession, user common.Address, now *big.Int) (bool, error) {
	if session.User != user {
		return false, fmt.Errorf("session query returned a different user: got %s, want %s", session.User, user)
	}
	if session.ClosedAt == nil || session.EndsAt == nil {
		return false, errors.New("session query returned incomplete lifecycle data")
	}
	return session.ClosedAt.Sign() == 0 && session.EndsAt.Cmp(now) > 0, nil
}

// scanLiveSessionPages walks the append-only user session list in ascending
// order. Ascending pagination is deliberate: a newly appended session cannot
// shift an already-scanned page and make an entry disappear between calls.
func scanLiveSessionPages(ctx context.Context, fetch liveSessionPageFetcher) (*structs.Session, error) {
	offset := big.NewInt(0)
	var latest *structs.Session
	for {
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		match, fetched, err := fetch(ctx, new(big.Int).Set(offset))
		if err != nil {
			return nil, err
		}
		if fetched < 0 || fetched > int(guardedSessionScanPageSize) {
			return nil, fmt.Errorf("invalid guarded session page size: %d", fetched)
		}
		if match != nil {
			latest = match
		}
		if fetched < int(guardedSessionScanPageSize) {
			return latest, nil
		}

		offset.Add(offset, big.NewInt(int64(fetched)))
	}
}

func (s *BlockchainService) getLiveSessionForUserModelPage(
	ctx context.Context,
	user common.Address,
	modelID common.Hash,
	offset *big.Int,
) (*structs.Session, int, error) {
	ids, sessions, err := s.sessionRouter.GetSessionsByUser(
		ctx,
		user,
		offset,
		guardedSessionScanPageSize,
		r.OrderASC,
	)
	if err != nil {
		return nil, 0, err
	}
	if len(ids) != len(sessions) {
		return nil, 0, fmt.Errorf(
			"incomplete session data while checking existing sessions: got %d IDs and %d sessions",
			len(ids),
			len(sessions),
		)
	}

	totalFetched := len(sessions)
	now := big.NewInt(time.Now().Unix())
	liveIDs := make([][32]byte, 0, totalFetched)
	liveSessions := make([]sessionrouter.ISessionStorageSession, 0, totalFetched)
	liveBidIDs := make([][32]byte, 0, totalFetched)
	for i, session := range sessions {
		live, err := isLiveUserSession(session, user, now)
		if err != nil {
			return nil, totalFetched, err
		}
		if !live {
			continue
		}
		if session.BidId == ([32]byte{}) {
			return nil, totalFetched, errors.New("live session query returned an empty bid ID")
		}
		liveIDs = append(liveIDs, ids[i])
		liveSessions = append(liveSessions, session)
		liveBidIDs = append(liveBidIDs, session.BidId)
	}

	if len(liveBidIDs) == 0 {
		return nil, totalFetched, nil
	}

	_, bids, err := s.marketplace.GetMultipleBids(ctx, liveBidIDs)
	if err != nil {
		return nil, totalFetched, err
	}
	if len(bids) != len(liveSessions) {
		return nil, totalFetched, fmt.Errorf(
			"incomplete bid data while checking existing sessions: got %d, want %d",
			len(bids),
			len(liveSessions),
		)
	}

	var latest *structs.Session
	for i, bid := range bids {
		if common.Hash(bid.ModelId) == modelID {
			// Pages and entries are ascending, so retain the newest match while
			// continuing the stable scan. Older desktop builds could create
			// duplicate live sessions after a retry; resuming the latest avoids
			// selecting a nearly-expired predecessor.
			latest = mapSession(liveIDs[i], liveSessions[i], bid)
		}
	}

	return latest, totalFetched, nil
}

func (s *BlockchainService) findLiveSessionForUserModel(
	ctx context.Context,
	user common.Address,
	modelID common.Hash,
) (*structs.Session, error) {
	return scanLiveSessionPages(ctx, func(ctx context.Context, offset *big.Int) (*structs.Session, int, error) {
		return s.getLiveSessionForUserModelPage(ctx, user, modelID, offset)
	})
}

// OpenSessionByModelIdRejectExisting is the desktop-safe variant of
// OpenSessionByModelId. Existing callers retain the original method and its
// ability to deliberately open more than one session for a model.
func (s *BlockchainService) OpenSessionByModelIdRejectExisting(
	ctx context.Context,
	modelID common.Hash,
	duration *big.Int,
	directPayment bool,
	isFailoverEnabled bool,
	omitProvider common.Address,
	agentUsername string,
) (common.Hash, error) {
	user, err := s.GetMyAddress(ctx)
	if err != nil {
		return common.Hash{}, fmt.Errorf("get wallet address for existing-session check: %w", err)
	}

	unlock := s.sessionOpenLocks.lock(user.Hex() + ":" + modelID.Hex())
	defer unlock()
	if err := ctx.Err(); err != nil {
		return common.Hash{}, err
	}

	return s.openSessionByModelID(
		ctx,
		modelID,
		duration,
		directPayment,
		isFailoverEnabled,
		omitProvider,
		agentUsername,
		func(ctx context.Context, openingUser common.Address) error {
			if openingUser != user {
				return errors.New("wallet changed during existing-session check")
			}
			existing, err := s.findLiveSessionForUserModel(ctx, user, modelID)
			if err != nil {
				// A failed or partial query must never be interpreted as "no session".
				return fmt.Errorf("check existing live sessions: %w", err)
			}
			if existing != nil {
				return &ExistingSessionError{SessionID: existing.Id}
			}
			return nil
		},
	)
}
