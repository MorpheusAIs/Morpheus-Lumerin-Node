package stratumv1_message

import (
	"encoding/json"
	"fmt"

	"gitlab.com/TitanInd/proxy/proxy-router-v3/internal/resources/hashrate/proxy/interfaces"
)

// Message: {"id":47,"result":true,"error":null}
// Message: {"id":1,"result":[[["mining.set_difficulty","1"],["mining.notify","1"]],"06650601bd171b",8],"error":null}
// Message: {"id":4,"result":null,"error":[-5,"Too low difficulty",null]}
type MiningResult struct {
	ID     int               `json:"id"`
	Result json.RawMessage   `json:"result"`
	Error  MiningResultError `json:"error"`
}

type MiningResultError = []json.RawMessage // data of different types: int and string

func ParseMiningResult(b []byte) (*MiningResult, error) {
	m := &MiningResult{}
	return m, json.Unmarshal(b, m)
}

func NewMiningResultSuccess(ID int) *MiningResult {
	return &MiningResult{
		ID:     ID,
		Result: []byte("true"),
	}
}

// TODO: verify error codes
func NewMiningResultJobNotFound(ID int) *MiningResult {
	return &MiningResult{
		ID: ID,
		Error: MiningResultError{
			json.RawMessage(`"21"`),
			json.RawMessage(`"Job not found"`),
		},
	}
}

func NewMiningResultLowDifficulty(ID int) *MiningResult {
	return &MiningResult{
		ID: ID,
		Error: MiningResultError{
			json.RawMessage(`"-5"`),
			json.RawMessage(`"Too low difficulty"`),
		},
	}
}

func NewMiningResultDuplicatedShare(ID int) *MiningResult {
	return &MiningResult{
		ID: ID,
		Error: MiningResultError{
			json.RawMessage(`"22"`),
			json.RawMessage(`"Duplicate share"`),
		},
	}
}

func (m *MiningResult) GetID() int {
	return m.ID
}

func (m *MiningResult) SetID(ID int) {
	m.ID = ID
}

func (m *MiningResult) IsError() bool {
	return m.Error != nil
}

// Returns unparsed error field (json)
// TODO: parse error code and message correctly
func (m *MiningResult) GetError() string {
	b, _ := json.Marshal(m.Error)
	return string(b)
}

func (m *MiningResult) Serialize() []byte {
	b, err := json.Marshal(m)
	if err != nil {
		panic(err)
	}
	return b
}

func (m *MiningResult) Copy() *MiningResult {
	res, _ := ParseMiningResult(m.Serialize())
	return res
}

var _ interfaces.MiningMessageGeneric = new(MiningResult)

// Parses Subscribe result message
//
// Message: {"id":1,"result":[[["mining.set_difficulty","1"],["mining.notify","1"]],"06650601bd171b",8],"error":null}
func ParseExtranonceSubscribeResult(m *MiningResult) (extranonce string, extranonceSize int, err error) {
	data := [3]interface{}{}

	err = json.Unmarshal(m.Result, &data)
	if err != nil {
		return "", 0, fmt.Errorf("cannot unmarhal subscribe response %s %w", string(m.Result), err)
	}

	return data[1].(string), int(data[2].(float64)), nil
}
