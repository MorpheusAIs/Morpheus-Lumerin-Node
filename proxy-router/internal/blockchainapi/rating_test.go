package blockchainapi

import (
	"math/big"
	"testing"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/lib"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/rating"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/repositories/contracts/bindings/providerregistry"
	"github.com/stretchr/testify/require"
)

func TestRating(t *testing.T) {
	bidIds, bids, pmStats, mStats := sampleDataTPS()

	bs := BlockchainService{
		rating: rating.NewRating(rating.NewScorerMock(), nil, lib.NewTestLogger()),
	}

	scoredBids := bs.rateBids(bidIds, bids, pmStats, []providerregistry.IProviderStorageProvider{}, mStats, big.NewInt(0), lib.NewTestLogger())

	for i := 1; i < len(scoredBids); i++ {
		require.GreaterOrEqual(t, scoredBids[i-1].Score, scoredBids[i].Score, "scoredBids not sorted")
	}
}
