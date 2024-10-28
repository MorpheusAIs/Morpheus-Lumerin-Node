package structs

import (
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/lib"
	"github.com/ethereum/go-ethereum/common"
)

type Provider struct {
	Address   common.Address
	Endpoint  string
	Stake     *lib.BigInt
	CreatedAt *lib.BigInt
	IsDeleted bool
}
