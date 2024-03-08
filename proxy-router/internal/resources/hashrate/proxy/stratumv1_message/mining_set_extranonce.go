package stratumv1_message

import (
	"encoding/json"

	"gitlab.com/TitanInd/proxy/proxy-router-v3/internal/resources/hashrate/proxy/interfaces"
)

// Message: {"id":null,"method":"mining.set_difficulty","params":[8192]}
const MethodMiningSetExtranonce = "mining.set_extranonce"

type MiningSetExtranonce struct {
	Method string                     `json:"method,omitempty"`
	Params *miningSetExtranonceParams `json:"params"`
}

type miningSetExtranonceParams = [2]interface{}

func NewMiningSetExtranonce(extranonce string, size int) *MiningSetExtranonce {
	return &MiningSetExtranonce{
		Method: MethodMiningSetExtranonce,
		Params: &miningSetExtranonceParams{extranonce, size},
	}
}

func ParseMiningSetExtranonce(b []byte) (*MiningSetExtranonce, error) {
	m := &MiningSetExtranonce{}
	return m, json.Unmarshal(b, m)
}

func (m *MiningSetExtranonce) GetExtranonce() (extranonce1 string, extranonce2size int) {
	return m.Params[0].(string), int(m.Params[1].(float64)) // observed that extranonce2size returned as float, TODO: check if it is correct
}

func (m *MiningSetExtranonce) SetExtranonce(extranonce string, size int) {
	m.Params[0], m.Params[1] = extranonce, size
}

func (m *MiningSetExtranonce) Serialize() []byte {
	b, _ := json.Marshal(m)
	return b
}

var _ interfaces.MiningMessageGeneric = new(MiningSetExtranonce)
