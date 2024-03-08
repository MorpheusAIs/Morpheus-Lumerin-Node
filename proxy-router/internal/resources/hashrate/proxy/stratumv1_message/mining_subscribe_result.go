package stratumv1_message

import (
	"encoding/json"
	"fmt"

	"gitlab.com/TitanInd/proxy/proxy-router-v3/internal/resources/hashrate/proxy/interfaces"
)

// Message: {"id":1,"result":[[["mining.set_difficulty","1"],["mining.notify","1"]],"06650601bd171b",8],"error":null}
type MiningSubscribeResult struct {
	ID     int                         `json:"id"`
	Result miningSubscribeResultResult `json:"result"`
	Error  miningSubscribeResultError  `json:"error"`
}

type miningSubscribeResultResult = [3]interface{}
type miningSubscribeResultError = interface{} // null

func ParseMiningSubscribeResult(b []byte) (*MiningSubscribeResult, error) {
	m := &MiningSubscribeResult{}
	if err := json.Unmarshal(b, m); err != nil {
		return nil, err
	}
	if _, ok := m.Result[1].(string); !ok {
		return nil, fmt.Errorf("invalid extranonce format")
	}
	if _, ok := m.Result[2].(string); !ok {
		return nil, fmt.Errorf("invalid extranonce size")
	}
	return m, nil
}

func ToMiningSubscribeResult(m *MiningResult) (*MiningSubscribeResult, error) {
	result := &miningSubscribeResultResult{}
	err := json.Unmarshal(m.Result, result)
	if err != nil {
		return nil, err
	}
	return &MiningSubscribeResult{
		ID:     m.ID,
		Result: *result,
		Error:  m.Error,
	}, nil
}

func NewMiningSubscribeResult(ID int, extranonce1 string, size int) *MiningSubscribeResult {
	data := [2][2]string{{"mining.set_difficulty", ""}, {"mining.notify", ""}}
	result := [3]interface{}{data, extranonce1, size}
	return &MiningSubscribeResult{
		ID:     ID,
		Result: result,
		Error:  nil,
	}
}

func (m *MiningSubscribeResult) GetID() int {
	return m.ID
}

func (m *MiningSubscribeResult) SetID(ID int) {
	m.ID = ID
}

func (m *MiningSubscribeResult) IsError() bool {
	return false
}

// Returns unparsed error field (json)
// TODO: parse error code and message correctly
func (m *MiningSubscribeResult) GetError() string {
	return ""
}

func (m *MiningSubscribeResult) GetExtranonce() (extranonce string, size int) {
	return m.Result[1].(string), int(m.Result[2].(float64))
}

func (m *MiningSubscribeResult) Serialize() []byte {
	b, _ := json.Marshal(m)
	return b
}

var _ interfaces.MiningMessageGeneric = new(MiningSubscribeResult)
