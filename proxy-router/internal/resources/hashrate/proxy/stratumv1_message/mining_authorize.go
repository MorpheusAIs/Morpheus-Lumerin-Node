package stratumv1_message

import (
	"encoding/json"

	"gitlab.com/TitanInd/proxy/proxy-router-v3/internal/resources/hashrate/proxy/interfaces"
)

// Message: {"id": 2, "method": "mining.authorize", "params": ["workername", "password"]}
const MethodMiningAuthorize = "mining.authorize"

type MiningAuthorize struct {
	ID     int                    `json:"id"`
	Method string                 `json:"method,omitempty"`
	Params *miningAuthorizeParams `json:"params"`
}

type miningAuthorizeParams = [2]string

func NewMiningAuthorize(ID int, minerID string, password string) *MiningAuthorize {
	return &MiningAuthorize{
		ID:     ID,
		Method: MethodMiningAuthorize,
		Params: &miningAuthorizeParams{minerID, password},
	}
}

func ParseMiningAuthorize(b []byte) (*MiningAuthorize, error) {
	m := &MiningAuthorize{}
	return m, json.Unmarshal(b, m)
}

func (m *MiningAuthorize) GetID() int {
	return m.ID
}

func (m *MiningAuthorize) SetID(ID int) {
	m.ID = ID
}

func (m *MiningAuthorize) GetUserName() string {
	return m.Params[0]
}

func (m *MiningAuthorize) SetUserName(ID string) {
	m.Params[0] = ID
}

func (m *MiningAuthorize) GetPassword() string {
	return m.Params[1]
}

func (m *MiningAuthorize) SetPassword(pwd string) {
	m.Params[1] = pwd
}

func (m *MiningAuthorize) Serialize() []byte {
	b, _ := json.Marshal(m)
	return b
}

var _ interfaces.MiningMessageWithID = new(MiningAuthorize)
