package blockchainapi

import (
	"math/big"
	"testing"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/blockchainapi/scorer"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/lib"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/repositories/contracts/bindings/providerregistry"
	"github.com/stretchr/testify/require"
)

func TestRating(t *testing.T) {
	bidIds, bids, pmStats, mStats := sampleDataTPS()

	scoredBids := RateBids(bidIds, bids, pmStats, []providerregistry.IProviderStorageProvider{}, mStats, scorer.NewScorerMock(), big.NewInt(0), lib.NewTestLogger())

	for i := 1; i < len(scoredBids); i++ {
		require.GreaterOrEqual(t, scoredBids[i-1].Score, scoredBids[i].Score, "scoredBids not sorted")
	}
}
