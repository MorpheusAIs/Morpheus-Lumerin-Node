package structs

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

type Model struct {
	Id        common.Hash
	IpfsCID   common.Hash
	Fee       *big.Int `swaggertype:"integer"`
	Stake     *big.Int `swaggertype:"integer"`
	Owner     common.Address
	Name      string
	Tags      []string
	CreatedAt *big.Int `swaggertype:"integer"`
	IsDeleted bool
}
