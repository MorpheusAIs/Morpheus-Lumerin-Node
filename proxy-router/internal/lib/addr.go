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
