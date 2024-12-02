package rating

import (
	"math"
	"sort"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/lib"
	"github.com/ethereum/go-ethereum/common"
)

// Filters bids based on the config, uses the scorer to rate the bids and sorts them
type Rating struct {
	scorer            Scorer
	providerAllowList map[common.Address]struct{}
}

func (r *Rating) RateBids(scoreInputs []RatingInput, log lib.ILogger) []RatingRes {
	scoredBids := make([]RatingRes, 0)

	for _, input := range scoreInputs {
		score := r.scorer.GetScore(&input.ScoreInput)
		if !r.isAllowed(input.ProviderID) {
			log.Warnf("provider %s is not in the allow list, skipping", input.ProviderID.String())
			continue
		}

		if math.IsNaN(score) || math.IsInf(score, 0) {
			log.Warnf("provider score is not valid %d for %+v), skipping", score, input)
			continue
		}

		scoredBid := RatingRes{
			BidID: input.BidID,
			Score: score,
		}
		scoredBids = append(scoredBids, scoredBid)
	}

	sort.Slice(scoredBids, func(i, j int) bool {
		return scoredBids[i].Score > scoredBids[j].Score
	})

	return scoredBids
}

func (r *Rating) isAllowed(provider common.Address) bool {
	if len(r.providerAllowList) == 0 {
		return true
	}
	_, ok := r.providerAllowList[provider]
	return ok
}

type RatingInput struct {
	ScoreInput
	BidID      common.Hash
	ModelID    common.Hash
	ProviderID common.Address
}

type RatingRes struct {
	BidID common.Hash
	Score float64
}
