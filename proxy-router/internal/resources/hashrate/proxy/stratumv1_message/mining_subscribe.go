package stratumv1_message

import (
	"encoding/json"

	"gitlab.com/TitanInd/proxy/proxy-router-v3/internal/resources/hashrate/proxy/interfaces"
)

// Message: {"id": 1, "method": "mining.subscribe", "params": ["cpuminer/2.5.1", "1"]}
const MethodMiningSubscribe = "mining.subscribe"

type MiningSubscribe struct {
	ID     int                    `json:"id"`
	Method string                 `json:"method,omitempty"`
	Params *miningSubscribeParams `json:"params"`
}

type miningSubscribeParams = [2]string

func NewMiningSubscribe(id int, name string, subscriptionId string) *MiningSubscribe {
	return &MiningSubscribe{
		ID:     id,
		Method: MethodMiningSubscribe,
		Params: &miningSubscribeParams{
			name, subscriptionId,
		},
	}
}

func ParseMiningSubscribe(b []byte) (*MiningSubscribe, error) {
	m := &MiningSubscribe{}
	return m, json.Unmarshal(b, m)
}

func (m *MiningSubscribe) GetID() int {
	return m.ID
}

func (m *MiningSubscribe) SetID(ID int) {
	m.ID = ID
}

func (m *MiningSubscribe) GetUseragent() string {
	return m.Params[0]
}

func (m *MiningSubscribe) SetUseragent(name string) {
	m.Params[0] = name
}

func (m *MiningSubscribe) GetWorkerNumber() string {
	return m.Params[1]
}

func (m *MiningSubscribe) SetWorkerNumber(name string) {
	m.Params[1] = name
}

func (m *MiningSubscribe) Serialize() []byte {
	b, _ := json.Marshal(m)
	return b
}

var _ interfaces.MiningMessageWithID = new(MiningSubscribe)
