package rating

type ScorerMock struct {
}

func NewScorerMock() *ScorerMock {
	return &ScorerMock{}
}

func (m *ScorerMock) GetScore(args *ScoreInput) float64 {
	return 0.5
}
