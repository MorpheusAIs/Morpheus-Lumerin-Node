package stratumv1_message

import (
	"encoding/json"

	"gitlab.com/TitanInd/proxy/proxy-router-v3/internal/resources/hashrate/proxy/interfaces"
)

// To enable multi-version needs to be >1, the number being how many bits of the version number you're allowing it to modify for ASICBOOST
// Message: {"id": 2510, "method": "mining.multi_version", "params": [1]}
const MethodMiningMultiVersion = "mining.multi_version"

type MiningMultiVersion struct {
	ID     int                       `json:"id"`
	Method string                    `json:"method,omitempty"`
	Params *miningMultiVersionParams `json:"params"`
}

type miningMultiVersionParams = [1]int

func NewMiningMultiVersion(version int) *MiningMultiVersion {
	return &MiningMultiVersion{
		Method: MethodMiningMultiVersion,
		Params: &miningMultiVersionParams{version},
	}
}

func ParseMiningMultiVersion(b []byte) (*MiningMultiVersion, error) {
	m := &MiningMultiVersion{}
	return m, json.Unmarshal(b, m)
}

func (m *MiningMultiVersion) GetID() int {
	return m.ID
}

func (m *MiningMultiVersion) SetID(ID int) {
	m.ID = ID
}

func (m *MiningMultiVersion) GetVersion() int {
	return m.Params[0]
}

func (m *MiningMultiVersion) SetVersion(version int) {
	m.Params[0] = version
}

func (m *MiningMultiVersion) Serialize() []byte {
	b, _ := json.Marshal(m)
	return b
}

var _ interfaces.MiningMessageGeneric = new(MiningMultiVersion)
