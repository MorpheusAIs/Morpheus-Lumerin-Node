package structs

import (
	"math/big"
)

type Provider struct {
	Address   string
	Endpoint  string
	Stake     *big.Int
	Timestamp *big.Int
	IsDeleted bool
}
