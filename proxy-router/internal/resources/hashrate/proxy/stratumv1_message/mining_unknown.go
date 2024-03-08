package stratumv1_message

import (
	"encoding/json"

	"gitlab.com/TitanInd/proxy/proxy-router-v3/internal/resources/hashrate/proxy/interfaces"
)

type MiningUnknown struct {
	ID     int             `json:"id"`
	Method string          `json:"method,omitempty"`
	Params json.RawMessage `json:"params"`
}

func ParseMiningUnknown(b []byte) (*MiningUnknown, error) {
	m := &MiningUnknown{}
	return m, json.Unmarshal(b, m)
}

func (m *MiningUnknown) GetID() int {
	return m.ID
}

func (m *MiningUnknown) SetID(ID int) {
	m.ID = ID
}

func (m *MiningUnknown) Serialize() []byte {
	b, _ := json.Marshal(m)
	return b
}

var _ interfaces.MiningMessageGeneric = new(MiningUnknown)
