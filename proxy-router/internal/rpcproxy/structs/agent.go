package structs

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

type Agent struct {
	AgentId   string
	Fee       big.Int
	Stake     big.Int
	Timestamp big.Int
	Owner     common.Address
	Name      string
	Tags      []string
}
