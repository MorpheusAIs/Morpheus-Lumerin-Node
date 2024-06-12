package lib

import (
	"math/big"
	"math/rand"

	"github.com/ethereum/go-ethereum/common"
)

func GetRandomAddr() common.Address {
	return common.BigToAddress(big.NewInt(rand.Int63()))
}

// AddrShort returns a short representation of an address in "0x123..567" format
func AddrShort(addr string) string {
	return StrShortConf(addr, 5, 3)
}

func RemoveHexPrefix(s string) string {
	if len(s) >= 2 && s[0:2] == "0x" {
		return s[2:]
	}
	return s
}
