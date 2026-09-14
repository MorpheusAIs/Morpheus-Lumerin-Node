package blockchainapi

import (
	"math/big"
	"testing"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/lib"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/rating"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/repositories/contracts/bindings/providerregistry"
	"github.com/stretchr/testify/require"
)

// providersFor builds a provider slice index-aligned with the sample bids.
//
// The previous version of this test passed an empty slice here, which made
// rateBids index past the end and panic — so the test never actually asserted
// anything about ordering, it just crashed the package.
func providersFor(n int) []providerregistry.IProviderStorageProvider {
	providers := make([]providerregistry.IProviderStorageProvider, n)
	for i := range providers {
		providers[i] = providerregistry.IProviderStorageProvider{
			Stake: big.NewInt(int64(100 * (i + 1))),
		}
	}
	return providers
}

func TestRating(t *testing.T) {
	bidIds, bids, pmStats, mStats := sampleDataTPS()

	bs := BlockchainService{
		rating: rating.NewRating(rating.NewScorerMock(), nil, nil, lib.NewTestLogger()),
	}

	scoredBids := bs.rateBids(
		bidIds, bids, pmStats,
		providersFor(len(bids)),
		mStats, big.NewInt(0), lib.NewTestLogger(),
	)

	// Guard against a hollow pass: with an empty result the ordering loop below
	// never executes and the test asserts nothing.
	require.Len(t, scoredBids, len(bids), "every bid should be scored")

	for i := 1; i < len(scoredBids); i++ {
		require.GreaterOrEqual(t, scoredBids[i-1].Score, scoredBids[i].Score, "scoredBids not sorted")
	}
}

// rateBids assembles its inputs from several independent on-chain reads. If any
// of them comes back short the slices are no longer index-aligned, which used
// to panic the daemon. It must degrade instead.
func TestRatingHandlesMismatchedInputLengths(t *testing.T) {
	bidIds, bids, pmStats, mStats := sampleDataTPS()

	bs := BlockchainService{
		rating: rating.NewRating(rating.NewScorerMock(), nil, nil, lib.NewTestLogger()),
	}

	for _, tc := range []struct {
		name      string
		providers int
	}{
		{"no providers", 0},
		{"one provider short", len(bids) - 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			require.NotPanics(t, func() {
				scored := bs.rateBids(
					bidIds, bids, pmStats,
					providersFor(tc.providers),
					mStats, big.NewInt(0), lib.NewTestLogger(),
				)
				require.LessOrEqual(t, len(scored), tc.providers)
			})
		})
	}
}
