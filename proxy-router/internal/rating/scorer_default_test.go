package rating

import (
	"fmt"
	"math"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseConfig(t *testing.T) {
	cfg := `{"weights":{"tps":0.24,"ttft":0.08,"duration":0.24,"success":0.32,"stake":0.12}}`
	k, err := NewScorerDefaultFromJSON([]byte(cfg))
	require.NoError(t, err)
	fmt.Printf("%+v", k)
}

func TestZeroObservations(t *testing.T) {
	sc := NewScorerDefault(ScorerDefaultParamsMock())
	args := NewScoreArgs()
	args.PricePerSecond.SetUint64(1)

	score := sc.GetScore(args)

	require.NotEqual(t, math.NaN(), score)
	require.NotEqual(t, math.Inf(0), score)
	require.NotEqual(t, math.Inf(-1), score)
}

func TestPriceIncrease(t *testing.T) {
	sc := NewScorerDefault(ScorerDefaultParamsMock())
	args := NewScoreArgs()
	args.PricePerSecond.SetUint64(1)

	score1 := sc.GetScore(args)

	args.PricePerSecond.SetUint64(2)
	score2 := sc.GetScore(args)

	require.Greater(t, score1, score2)
}

func TestFirstTTFTObservation(t *testing.T) {
	// when there is only one observation, of in rare case few with same values
	// the score should be the same (because standard deviation is 0)
	sc := NewScorerDefault(ScorerDefaultParamsMock())
	args := NewScoreArgs()
	args.PricePerSecond.SetUint64(10000000000000000)

	args.ProviderModel.TtftMs.Mean = 100
	args.ProviderModel.TotalCount = 1
	args.ProviderModel.SuccessCount = 1
	args.Model.Count = 1
	score1 := sc.GetScore(args)

	args.ProviderModel.TtftMs.Mean = 200
	args.ProviderModel.TotalCount = 1
	args.ProviderModel.SuccessCount = 1
	args.Model.Count = 1
	score2 := sc.GetScore(args)

	require.Equal(t, score1, score2)
}

func TestTPSImpact(t *testing.T) {
	sc := NewScorerDefault(ScorerDefaultParamsMock())
	args := NewScoreArgs()
	args.PricePerSecond.SetUint64(10000000000000000)

	args.Model.Count = 2
	args.Model.TpsScaled1000.Mean = 100
	args.Model.TpsScaled1000.SqSum = 1000

	args.ProviderModel.TpsScaled1000.Mean = 100
	args.ProviderModel.TotalCount = 1
	args.ProviderModel.SuccessCount = 1
	score1 := sc.GetScore(args)

	args.ProviderModel.TpsScaled1000.Mean = 150
	args.ProviderModel.TotalCount = 1
	args.ProviderModel.SuccessCount = 1
	score2 := sc.GetScore(args)

	require.Less(t, score1, score2)
}

func TestTTFTImpact(t *testing.T) {
	// when there is more than one observation, larger ttft should produce lower score
	sc := NewScorerDefault(ScorerDefaultParamsMock())
	args := NewScoreArgs()
	args.PricePerSecond.SetUint64(10000000000000000)

	args.Model.Count = 2
	args.Model.TtftMs.Mean = 100
	args.Model.TtftMs.SqSum = 1000

	args.ProviderModel.TtftMs.Mean = 100
	args.ProviderModel.TotalCount = 1
	args.ProviderModel.SuccessCount = 1
	score1 := sc.GetScore(args)

	args.ProviderModel.TtftMs.Mean = 150
	args.ProviderModel.TotalCount = 1
	args.ProviderModel.SuccessCount = 1
	score2 := sc.GetScore(args)

	require.Greater(t, score1, score2)
}

func TestSuccessScoreImpact(t *testing.T) {
	sc := NewScorerDefault(ScorerDefaultParamsMock())
	args := NewScoreArgs()
	args.PricePerSecond.SetUint64(10000000000000000)
	args.Model.Count = 2

	args.ProviderModel.TotalCount = 2
	args.ProviderModel.SuccessCount = 1
	score1 := sc.GetScore(args)

	args.ProviderModel.SuccessCount = 2
	score2 := sc.GetScore(args)

	require.Less(t, score1, score2)
}

func TestStakeScoreImpact(t *testing.T) {
	sc := NewScorerDefault(ScorerDefaultParamsMock())
	args := NewScoreArgs()
	args.PricePerSecond.SetUint64(10000000000000000)
	args.MinStake.SetUint64(1)

	args.ProviderStake.SetUint64(2)
	score1 := sc.GetScore(args)

	args.ProviderStake.SetUint64(3)
	score2 := sc.GetScore(args)

	require.Less(t, score1, score2)
}

func TestStakeScoreImpactLimit(t *testing.T) {
	// stake score should be capped at 10x min stake
	sc := NewScorerDefault(ScorerDefaultParamsMock())
	args := NewScoreArgs()
	args.PricePerSecond.SetUint64(10000000000000000)
	minStake := uint64(10)
	args.MinStake.SetUint64(minStake)

	args.ProviderStake.SetUint64(10 * minStake)
	score1 := sc.GetScore(args)

	args.ProviderStake.SetUint64(11 * minStake)
	score2 := sc.GetScore(args)

	require.Equal(t, score1, score2)
}

func TestPriceImpact(t *testing.T) {
	sc := NewScorerDefault(ScorerDefaultParamsMock())
	args := NewScoreArgs()
	args.PricePerSecond.SetUint64(10000000000000000)

	score1 := sc.GetScore(args)

	args.PricePerSecond.SetUint64(20000000000000000)
	score2 := sc.GetScore(args)

	require.Greater(t, score1, score2)
}
