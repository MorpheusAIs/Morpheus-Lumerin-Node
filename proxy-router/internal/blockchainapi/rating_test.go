package blockchainapi

import (
	"math"
	"testing"

	m "github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/contracts/marketplace"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/lib"
	"github.com/stretchr/testify/require"
)

func TestRating(t *testing.T) {
	bidIds, bids, pmStats, mStats := sampleDataTPS()

	scoredBids := rateBids(bidIds, bids, pmStats, mStats, lib.NewTestLogger())

	for i := 1; i < len(scoredBids); i++ {
		require.GreaterOrEqual(t, scoredBids[i-1].Score, scoredBids[i].Score, "scoredBids not sorted")
	}
}

func TestGetScoreZeroObservations(t *testing.T) {
	_, bids, _, _ := sampleDataTPS()
	score := getScore(bids[0], m.ProviderModelStats{}, m.ModelStats{})
	if math.IsNaN(score) {
		require.Fail(t, "score is NaN")
	}
}

func TestGetScoreSingleObservation(t *testing.T) {
	_, bids, _, _ := sampleDataTPS()
	pmStats := m.ProviderModelStats{
		TpsScaled1000: m.LibSDSD{Mean: 1000, SqSum: 0},
		TtftMs:        m.LibSDSD{Mean: 1000, SqSum: 0},
		TotalDuration: 1000,
		SuccessCount:  1,
		TotalCount:    1,
	}
	mStats := m.ModelStats{
		TpsScaled1000: m.LibSDSD{Mean: 1000, SqSum: 0},
		TtftMs:        m.LibSDSD{Mean: 1000, SqSum: 0},
		TotalDuration: m.LibSDSD{Mean: 1000, SqSum: 0},
		Count:         1,
	}
	score := getScore(bids[0], pmStats, mStats)
	if math.IsNaN(score) {
		require.Fail(t, "score is NaN")
	}
}
