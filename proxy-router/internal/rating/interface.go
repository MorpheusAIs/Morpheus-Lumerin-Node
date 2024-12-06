package rating

import (
	"math/big"

	s "github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/repositories/contracts/bindings/sessionrouter"
)

type Scorer interface {
	// GetScore returns the score of a bid. Returns -Inf if the bid should be skipped
	GetScore(args *ScoreInput) float64
}

// ScoreInput is a struct that holds the input data for the rating algorithm of a bid
type ScoreInput struct {
	ProviderModel  *s.IStatsStorageProviderModelStats // stats of the provider specific to the model
	Model          *s.IStatsStorageModelStats         // stats of the model across providers
	PricePerSecond *big.Int
	ProviderStake  *big.Int
	MinStake       *big.Int
}

func NewScoreArgs() *ScoreInput {
	return &ScoreInput{
		ProviderModel:  &s.IStatsStorageProviderModelStats{},
		Model:          &s.IStatsStorageModelStats{},
		ProviderStake:  big.NewInt(0),
		PricePerSecond: big.NewInt(0),
		MinStake:       big.NewInt(0),
	}
}
