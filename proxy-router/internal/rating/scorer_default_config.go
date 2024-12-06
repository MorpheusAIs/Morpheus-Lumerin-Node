package rating

import (
	"encoding/json"
	"errors"
)

var ErrInvalidWeights = errors.New("weights must sum to 1")

type ScorerDefaultParams struct {
	Weights struct {
		TPS      float64 `json:"tps"`
		TTFT     float64 `json:"ttft"`
		Duration float64 `json:"duration"`
		Success  float64 `json:"success"`
		Stake    float64 `json:"stake"`
	} `json:"weights"`
}

func (w *ScorerDefaultParams) Validate() bool {
	return w.Weights.TPS+w.Weights.TTFT+w.Weights.Duration+w.Weights.Success+w.Weights.Stake == 1
}

func NewScorerDefaultFromJSON(data json.RawMessage) (*ScorerDefault, error) {
	var params ScorerDefaultParams
	err := json.Unmarshal(data, &params)
	if err != nil {
		return nil, err
	}
	if !params.Validate() {
		return nil, ErrInvalidWeights
	}
	return &ScorerDefault{params: params}, nil
}
