package blockchainapi

import (
	"math"
	"sort"

	m "github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/contracts/marketplace"
	"github.com/ethereum/go-ethereum/common"
)

type ScoredBid struct {
	ID    common.Hash
	Bid   m.Bid
	Score float64
}

func rateBids(bidIds [][32]byte, bids []m.Bid, pmStats []m.ProviderModelStats, mStats m.ModelStats) []ScoredBid {

	scoredBids := make([]ScoredBid, len(bids))

	for i := range bids {
		scoredBid := ScoredBid{
			ID:    bidIds[i],
			Bid:   bids[i],
			Score: getScore(bids[i], pmStats[i], mStats),
		}
		scoredBids[i] = scoredBid
	}

	sort.Slice(scoredBids, func(i, j int) bool {
		return scoredBids[i].Score < scoredBids[j].Score
	})

	return scoredBids
}

func getScore(bid m.Bid, pmStats m.ProviderModelStats, mStats m.ModelStats) float64 {
	tpsWeight, ttftWeight, durationWeight, successWeight := 0.25, 0.25, 0.25, 0.25
	count := int64(mStats.Count)

	tpsScore := tpsWeight * getZIndexScore(pmStats.TpsScaled1000.Mean, mStats.TpsScaled1000, count, 3.0)
	ttftScore := ttftWeight * getZIndexScore(pmStats.TtftMs.Mean, mStats.TtftMs, count, 3.0)
	durationScore := durationWeight * getZIndexScore(int64(pmStats.TotalDuration), mStats.TotalDuration, count, 3.0)
	successScore := successWeight * math.Pow(float64(pmStats.SuccessCount)/float64(count), 2)

	priceFloatDecimal, _ := bid.PricePerSecond.Float64()
	priceFloat := priceFloatDecimal / math.Pow10(18)

	return (tpsScore + ttftScore + durationScore + successScore) / priceFloat
}

func getZIndexScore(pmMean int64, mSD m.LibSDSD, obsNum int64, normSigma float64) float64 {
	// TODO: consider variance(SD) of provider model stats
	score := float64(pmMean-mSD.Mean) / getSD(mSD, obsNum)
	return cutRange01((score + normSigma) / 2 * normSigma)
}

func getSD(sd m.LibSDSD, obsNum int64) float64 {
	return math.Sqrt(getVariance(sd, obsNum))
}

func getVariance(sd m.LibSDSD, obsNum int64) float64 {
	return float64(sd.SqSum) / float64(obsNum-1)
}

func cutRange01(val float64) float64 {
	if val > 1 {
		return 1
	}
	if val < 0 {
		return 0
	}
	return val
}
