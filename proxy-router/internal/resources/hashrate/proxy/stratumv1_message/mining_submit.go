package stratumv1_message

import (
	"encoding/json"

	"gitlab.com/TitanInd/proxy/proxy-router-v3/internal/resources/hashrate/proxy/interfaces"
)

// Message: {"id": 4, "method": "mining.submit", "params": ["shev8.local", "620daf25f", "0000000000000000", "62cea7a6", "f9b40000"]}
const MethodMiningSubmit = "mining.submit"

type MiningSubmit struct {
	ID     int      `json:"id"`
	Method string   `json:"method,omitempty"`
	Params []string `json:"params"` // worker_name, job_id, extranonce2, ntime, nonce and optional version_bits (BIP_0310)
}

func NewMiningSubmit(workerName string, jobId string, extranonce2 string, ntime string, nonce string) *MiningSubmit {
	return &MiningSubmit{
		ID:     0,
		Method: MethodMiningSubmit,
		Params: []string{workerName, jobId, extranonce2, ntime, nonce},
	}
}

func ParseMiningSubmit(b []byte) (*MiningSubmit, error) {
	m := &MiningSubmit{}
	return m, json.Unmarshal(b, m)
}

func (m *MiningSubmit) GetID() int {
	return m.ID
}

func (m *MiningSubmit) SetID(ID int) {
	m.ID = ID
}

func (m *MiningSubmit) GetUserName() string {
	return m.Params[0]
}

func (m *MiningSubmit) SetUserName(name string) {
	m.Params[0] = name
}

func (m *MiningSubmit) GetJobId() string {
	return m.Params[1]
}

func (m *MiningSubmit) GetExtraNonce2() string {
	return m.Params[2]
}

func (m *MiningSubmit) SetExtraNonce2(xnonce2 string) {
	m.Params[2] = xnonce2
}

func (m *MiningSubmit) GetNtime() string {
	return m.Params[3]
}

func (m *MiningSubmit) GetNonce() string {
	return m.Params[4]
}

func (m *MiningSubmit) GetVmask() string {
	if len(m.Params) < 6 {
		return "00000000"
	}
	return m.Params[5]
}

func (m *MiningSubmit) Serialize() []byte {
	b, _ := json.Marshal(m)
	return b
}

var _ interfaces.MiningMessageWithID = new(MiningSubmit)
