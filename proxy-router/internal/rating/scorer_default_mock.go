package rating

func ScorerDefaultParamsMock() ScorerDefaultParams {
	var sc = ScorerDefaultParams{}
	sc.Weights.Duration = 0.24
	sc.Weights.Stake = 0.12
	sc.Weights.Success = 0.32
	sc.Weights.TPS = 0.24
	sc.Weights.TTFT = 0.08
	return sc
}
