package mobile

import (
	"context"
	"math/big"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/repositories/registries"
	"github.com/ethereum/go-ethereum/common"
)

// OpenSession opens a session with the best-rated provider for a model.
// duration is in seconds. Returns the session ID (tx hash).
func (s *SDK) OpenSession(ctx context.Context, modelID string, durationSec int64, directPayment bool) (string, error) {
	if err := s.checkClosed(); err != nil {
		return "", err
	}
	id := common.HexToHash(modelID)
	dur := big.NewInt(durationSec)
	txHash, err := s.blockchain.OpenSessionByModelId(ctx, id, dur, directPayment, true, common.Address{}, "")
	if err != nil {
		return "", err
	}
	return txHash.Hex(), nil
}

// CloseSession closes an active session. Returns the close tx hash.
func (s *SDK) CloseSession(ctx context.Context, sessionID string) (string, error) {
	if err := s.checkClosed(); err != nil {
		return "", err
	}
	id := common.HexToHash(sessionID)
	txHash, err := s.blockchain.CloseSession(ctx, id)
	if err != nil {
		return "", err
	}
	return txHash.Hex(), nil
}

// GetSession retrieves session details by ID.
func (s *SDK) GetSession(ctx context.Context, sessionID string) (*Session, error) {
	id := common.HexToHash(sessionID)
	sess, err := s.blockchain.GetSession(ctx, id)
	if err != nil {
		return nil, err
	}
	return sessionFromInternal(sess), nil
}

// GetSessionJSON returns session details as a JSON string (for FFI).
func (s *SDK) GetSessionJSON(ctx context.Context, sessionID string) (string, error) {
	sess, err := s.GetSession(ctx, sessionID)
	if err != nil {
		return "", err
	}
	return toJSON(sess)
}

// GetUnclosedUserSessions returns on-chain sessions for the wallet where ClosedAt == 0.
// Paginates newest-first until a short page or maxPages.
func (s *SDK) GetUnclosedUserSessions(ctx context.Context) ([]Session, error) {
	addr, err := s.getAddress()
	if err != nil {
		return nil, err
	}
	user := common.HexToAddress(addr)
	var out []Session
	offset := big.NewInt(0)
	order := registries.OrderDESC

	for page := 0; page < paginationMaxPages; page++ {
		unclosed, _, totalFetched, err := s.blockchain.GetUnclosedUserSessions(ctx, user, offset, paginationLimit, order)
		if err != nil {
			return nil, err
		}
		for _, ses := range unclosed {
			if ses == nil {
				continue
			}
			out = append(out, *sessionFromInternal(ses))
		}
		if totalFetched < int(paginationLimit) {
			break
		}
		offset = new(big.Int).Add(offset, big.NewInt(int64(totalFetched)))
	}
	if out == nil {
		out = []Session{}
	}
	return out, nil
}

// GetUnclosedUserSessionsJSON is for FFI / JSON consumers.
func (s *SDK) GetUnclosedUserSessionsJSON(ctx context.Context) (string, error) {
	list, err := s.GetUnclosedUserSessions(ctx)
	if err != nil {
		return "", err
	}
	return toJSON(list)
}
