package structs

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

type Model struct {
	Id        string
	IpfsCID   string
	Fee       *big.Int
	Stake     *big.Int
	Owner     common.Address
	Name      string
	Tags      []string
	CreatedAt *big.Int
	IsDeleted bool
}
