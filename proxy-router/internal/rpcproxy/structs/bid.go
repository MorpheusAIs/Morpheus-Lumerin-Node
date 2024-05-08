package structs

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

type Bid struct {
	Id             string
	Provider       common.Address
	ModelAgentId   string
	PricePerSecond *big.Int
	Nonce          *big.Int
	CreatedAt      *big.Int
	DeletedAt      *big.Int
}
