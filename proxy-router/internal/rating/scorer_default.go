package rating

import (
	"math"
	"math/big"
)

const (
	ScorerNameDefault = "default"
)

type ScorerDefault struct {
	params ScorerDefaultParams
}

func NewScorerDefault(weights ScorerDefaultParams) *ScorerDefault {
	return &ScorerDefault{params: weights}
}

func (r *ScorerDefault) GetScore(args *ScoreInput) float64 {
	tpsScore := r.params.Weights.TPS * tpsScore(args)
	ttftScore := r.params.Weights.TTFT * ttftScore(args)
	durationScore := r.params.Weights.Duration * durationScore(args)
	successScore := r.params.Weights.Success * successScore(args)
	stakeScore := r.params.Weights.Stake * stakeScore(args)

	totalScore := tpsScore + ttftScore + durationScore + successScore + stakeScore

	return priceAdjust(totalScore, args.PricePerSecond)
}

func tpsScore(args *ScoreInput) float64 {
	// normalize provider model tps value using stats from the model from other providers
	zIndex := normZIndex(args.ProviderModel.TpsScaled1000.Mean, args.Model.TpsScaled1000, int64(args.Model.Count))

	// cut off the values outside the range [-3, 3] and normalize to [0, 1]
	return normRange(zIndex, 3.0)
}

func ttftScore(args *ScoreInput) float64 {
	// normalize provider model ttft value using stats for this model from other providers
	zIndex := normZIndex(args.ProviderModel.TtftMs.Mean, args.Model.TtftMs, int64(args.Model.Count))

	// invert the value, because the higher the ttft, the worse the score
	zIndex = -zIndex

	// cut off the values outside the range [-3, 3] and normalize to [0, 1]
	return normRange(zIndex, 3.0)
}

func durationScore(args *ScoreInput) float64 {
	// normalize provider model duration value using stats for this model from other providers
	zIndex := normZIndex(int64(args.ProviderModel.TotalDuration), args.Model.TotalDuration, int64(args.Model.Count))

	// cut off the values outside the range [-3, 3] and normalize to [0, 1]
	return normRange(zIndex, 3.0)
}

func successScore(args *ScoreInput) float64 {
	// calculate the ratio of successful requests to total requests
	ratio := ratioScore(args.ProviderModel.SuccessCount, args.ProviderModel.TotalCount)

	// the higher the ratio, the better the score
	return math.Pow(ratio, 2)
}

func stakeScore(args *ScoreInput) float64 {
	// normalize provider stake value to the range [0, 10x min stake]
	return normMinMax(args.ProviderStake.Int64(), args.MinStake.Int64(), 10*args.MinStake.Int64())
}

func priceAdjust(score float64, pricePerSecond *big.Int) float64 {
	priceFloatDecimal, _ := pricePerSecond.Float64()

	// since the price is in decimals, we adjust it to have less of the exponent
	// TODO: consider removing it and using the price as is, since the exponent will be
	// the same for all providers and can be removed from the equation
	priceFloat := priceFloatDecimal / math.Pow10(18)

	// price cannot be 0 according to smart contract, so we can safely divide by it
	return score / priceFloat
}

var _ Scorer = &ScorerDefault{}
