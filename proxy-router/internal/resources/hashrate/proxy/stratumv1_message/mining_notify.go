package stratumv1_message

import (
	"encoding/json"
	"log"

	"gitlab.com/TitanInd/proxy/proxy-router-v3/internal/lib"
	"gitlab.com/TitanInd/proxy/proxy-router-v3/internal/resources/hashrate/proxy/interfaces"
)

// Message: {"id":null,"method":"mining.notify","params":["620e41a18","b56266ef4c94ba61562510b7656d132cacc928c50008488c0000000000000000","01000000010000000000000000000000000000000000000000000000000000000000000000ffffffff4b03795d0bfabe6d6d021f1a6edc2237ed5d6b5ce13c9b8516dcae143a5cf8a373fa0406c7ae06fb260100000000000000","181ae420062f736c7573682f00000000030f1b0a27000000001976a9147c154ed1dc59609e3d26abb2df2ea3d587cd8c4188ac00000000000000002c6a4c2952534b424c4f434b3ad6b225f8545c851f458a3f603c9a5bd63959ab646f977ff2fd8d1e2f004427dc0000000000000000266a24aa21a9ed46c37118ea57f6c2152b2f156e4df28edb997e21ea8fe5754e9a869f0287cfb800000000",["bed26bac18890c62ab48bdd913ab2b648326286607b3a159987cf36b6fe55d7e","8c39bd3ac4aeedc7b3c354008692ccb5e758a9cd9b1888a72090268365d9fbf0","00dca1b0b193f0298f155d32a9c04a79a49a2617853b787a79adc942cae74fed","fbb3b3a6bf5710f885fd377a2fde24fbb795933c9e6ceea67f91f1d90c532be2","ddd51d322b9c61621f762002dc179de1c24f4454e17943de172e24e9ad4be942","6fcffd6b0ebd01f15f57cb6ab3fabe151d757f1f71b45ef420b97b5fac0cc670","c28b6ef7c87ff5982bec0eaccb1855fff78397b2c732030dc792636a9492ba7c","afe4f418f45c78c36a848930b56323749fb7da453d557aa41c820a3170f7d20a","c61bab8479a8ada4c9761be0ce82e3183e7405b57b31925e0e411e73376fb22e","224b536c03f1379f708172970fca51160192bd8eded9cf578f7f5c3c8795eb33","fe4af4678dd3f66946738a1479627683e461bb832629a01978b4f2165490f460"],"20000004","1709a7af","62cea7e2",false]}
const MethodMiningNotify = "mining.notify"

type MiningNotify struct {
	ID     *int               `json:"id"` // always null
	Method string             `json:"method"`
	Params [9]json.RawMessage `json:"params"`
}

func ParseMiningNotify(b []byte) (*MiningNotify, error) {
	m := &MiningNotify{}
	return m, json.Unmarshal(b, m)
}

// jobID := n[0].(string)
// prevblock := n[1].(string)
// gen1 := n[2].(string)
// gen2 := n[3].(string)
// merkel := n[4].([]interface{})
// version := n[5].(string)
// nbits := n[6].(string)
// ntime := n[7].(string)
// clean := n[8].(bool)

func (m *MiningNotify) GetJobID() string {
	return lib.MustUnmarshallString(m.Params[0])
}

func (m *MiningNotify) SetJobID(ID string) {
	m.Params[0], _ = json.Marshal(ID)
}

func (m *MiningNotify) GetPrevBlockHash() string {
	return lib.MustUnmarshallString(m.Params[1])
}

func (m *MiningNotify) SetPrevBlockHash(hash string) {
	m.Params[1], _ = json.Marshal(hash)
}

func (m *MiningNotify) GetGen1() string {
	return lib.MustUnmarshallString(m.Params[2])
}

func (m *MiningNotify) SetGen1(gen1 string) {
	m.Params[2], _ = json.Marshal(gen1)
}

func (m *MiningNotify) GetGen2() string {
	return lib.MustUnmarshallString(m.Params[3])
}

func (m *MiningNotify) GetMerkel() []interface{} {
	merkel := []interface{}{}
	err := json.Unmarshal(m.Params[4], &merkel)
	if err != nil {
		log.Println(err)
	}
	return merkel
}

func (m *MiningNotify) GetVersion() string {
	return lib.MustUnmarshallString(m.Params[5])

}

func (m *MiningNotify) GetNbits() string {
	return lib.MustUnmarshallString(m.Params[6])

}

func (m *MiningNotify) GetNtime() string {
	return lib.MustUnmarshallString(m.Params[7])
}

func (m *MiningNotify) GetCleanJobs() bool {
	var cleanJobs bool
	err := json.Unmarshal(m.Params[8], &cleanJobs)
	if err != nil {
		log.Println(err)
	}
	return cleanJobs
}

func (m *MiningNotify) SetCleanJobs(shouldClean bool) {
	m.Params[8], _ = json.Marshal(shouldClean)
}

func (m *MiningNotify) Serialize() []byte {
	b, _ := json.Marshal(m)
	return b
}

func (m *MiningNotify) Copy() *MiningNotify {
	res, _ := ParseMiningNotify(m.Serialize())
	return res
}

var _ interfaces.MiningMessageGeneric = new(MiningNotify)
