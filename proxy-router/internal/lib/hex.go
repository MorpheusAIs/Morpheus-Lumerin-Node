package lib

import (
	"github.com/ethereum/go-ethereum/common"
)

// Arbitrary length hex string represented internally as a byte slice
type HexString []byte

func NewHexString(s string) HexString {
	return common.FromHex(s)
}

func (b *HexString) UnmarshalJSON(data []byte) error {
	str := string(data)
	if len(str) > 0 && str[0] == '"' {
		str = str[1 : len(str)-1]
	}
	d := common.FromHex(str)
	*b = d
	return nil
}

func (b HexString) String() string {
	return b.String()
}

func (b HexString) Hex() string {
	if len(b) == 0 {
		return ""
	}
	return BytesToString(b)
}

func (b HexString) MarshalJSON() ([]byte, error) {
	return []byte(`"` + b.String() + `"`), nil
}
