package blockchainapi

import (
	"math/big"
	"testing"

	pr "github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/repositories/contracts/bindings/providerregistry"
	sr "github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/repositories/contracts/bindings/sessionrouter"
	"github.com/stretchr/testify/require"
)

// The claimable figure is what claimForProvider accrues now: the reward ends at
// the earlier of closedAt and endsAt, the closing day's share waits for that UTC
// day to end, and the provider's reward limiter caps the rest.
func TestProviderClaimable(t *testing.T) {
	const day = int64(secondsInDay)
	const midnight = int64(1_700_006_400) // a UTC day start
	session := func(openedAt, endsAt, closedAt, withdrawn int64) *sr.ISessionStorageSession {
		return &sr.ISessionStorageSession{OpenedAt: big.NewInt(openedAt), EndsAt: big.NewInt(endsAt), ClosedAt: big.NewInt(closedAt), ProviderWithdrawnAmount: big.NewInt(withdrawn)}
	}
	provider := func(stake, earned, periodEnd int64) *pr.IProviderStorageProvider {
		return &pr.IProviderStorageProvider{Stake: big.NewInt(stake), LimitPeriodEarned: big.NewInt(earned), LimitPeriodEnd: big.NewInt(periodEnd)}
	}
	unlimited := provider(1_000_000, 0, midnight+365*day)

	tests := []struct {
		name     string
		session  *sr.ISessionStorageSession
		provider *pr.IProviderStorageProvider
		now      int64
		want     int64
		wantErr  bool
	}{
		{"late close read on its closing day: ends at endsAt, never negative", session(midnight-2000, midnight-1000, midnight+300, 10_000), unlimited, midnight + 500, 0, false},
		{"disputed early close: the closing day's share is held", session(midnight-1000, midnight+5000, midnight+3000, 0), unlimited, midnight + 4000, 10_000, false},
		{"disputed early close: paid once that day ends", session(midnight-1000, midnight+5000, midnight+3000, 0), unlimited, midnight + day, 40_000, false},
		{"limiter caps at stake minus earned until its period ends", session(midnight+1000, midnight+2000, 0, 0), provider(4_000, 1_500, midnight+day), midnight + day, 2_500, false},
		{"limiter period over: earned starts again", session(midnight+1000, midnight+2000, 0, 0), provider(4_000, 1_500, midnight+10), midnight + 2000, 4_000, false},
		{"not ended yet", session(midnight+1000, midnight+2000, 0, 0), unlimited, midnight + 1999, 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := providerClaimable(tt.session, big.NewInt(10), tt.provider, tt.now)
			if tt.wantErr {
				require.ErrorIs(t, err, errSessionNotEnded)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.want, got.Int64())
		})
	}
}
