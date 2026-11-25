package blockchainapi

import (
	"context"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
)

// Pool sessions are stake-only: every open and quote that asks for direct
// payment from the pool is refused before the wallet lock, any chain read or
// a provider approval. The service has no dependencies, so touching one panics.
func TestPoolOpenRefusesDirectPayment(t *testing.T) {
	s, ctx, duration := &BlockchainService{}, context.Background(), big.NewInt(600)
	cases := map[string]func() error{
		"approval": func() error {
			_, err := s.OpenSession(ctx, nil, nil, duration, true, true, "", false)
			return err
		},
		"bid": func() error {
			_, err := s.OpenSessionByBidId(ctx, common.Hash{}, duration, true, true, "")
			return err
		},
		"model": func() error {
			_, err := s.OpenSessionByModelId(ctx, common.Hash{}, duration, true, true, true, common.Address{}, "")
			return err
		},
		"estimate": func() error {
			_, err := s.EstimateOpenSessionStake(ctx, common.Hash{}, common.Hash{}, duration, true, true)
			return err
		},
	}
	for name, open := range cases {
		t.Run(name, func(t *testing.T) {
			require.ErrorIs(t, open(), ErrPoolStakeOnly)
		})
	}
}

// The router refuses a settlement exactly when settleExpiredSession would
// revert: the session is closed or settled, or chain time is before
// endsAt + settlementGrace.
func TestCheckSettleable(t *testing.T) {
	endsAt, grace := big.NewInt(1_000_000), big.NewInt(86_400)
	settleableAt := endsAt.Int64() + grace.Int64()

	cases := []struct {
		name     string
		settled  bool
		closedAt int64
		now      int64
		wantErr  bool
	}{
		{"already settled", true, settleableAt, settleableAt + 10, true},
		{"already closed", false, endsAt.Int64() - 10, settleableAt + 10, true},
		{"one second before the grace ends", false, 0, settleableAt - 1, true},
		{"when the grace ends", false, 0, settleableAt, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := checkSettleable(tc.settled, big.NewInt(tc.closedAt), endsAt, grace, big.NewInt(tc.now))
			if !tc.wantErr {
				require.NoError(t, err)
				return
			}
			var precondition *SettlePreconditionError
			require.ErrorAs(t, err, &precondition)
		})
	}
}
