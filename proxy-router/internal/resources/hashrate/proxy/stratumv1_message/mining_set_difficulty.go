package stratumv1_message

import (
	"encoding/json"

	"gitlab.com/TitanInd/proxy/proxy-router-v3/internal/resources/hashrate/proxy/interfaces"
)

// Message: {"id":null,"method":"mining.set_difficulty","params":[8192.1234]}
const MethodMiningSetDifficulty = "mining.set_difficulty"

type MiningSetDifficulty struct {
	Method string                     `json:"method,omitempty"`
	Params *miningSetDifficultyParams `json:"params"`
}

type miningSetDifficultyParams = [1]float64

func NewMiningSetDifficulty(difficulty float64) *MiningSetDifficulty {
	return &MiningSetDifficulty{
		Method: MethodMiningSetDifficulty,
		Params: &miningSetDifficultyParams{difficulty},
	}
}

func ParseMiningSetDifficulty(b []byte) (*MiningSetDifficulty, error) {
	m := &MiningSetDifficulty{}
	return m, json.Unmarshal(b, m)
}

func (m *MiningSetDifficulty) GetDifficulty() float64 {
	return m.Params[0]
}

func (m *MiningSetDifficulty) SetDifficulty(difficulty float64) {
	m.Params[0] = difficulty
}

func (m *MiningSetDifficulty) Serialize() []byte {
	b, _ := json.Marshal(m)
	return b
}

func (m *MiningSetDifficulty) Copy() *MiningSetDifficulty {
	return &MiningSetDifficulty{
		Method: MethodMiningSetDifficulty,
		Params: &miningSetDifficultyParams{m.GetDifficulty()},
	}
}

var _ interfaces.MiningMessageGeneric = new(MiningSetDifficulty)
