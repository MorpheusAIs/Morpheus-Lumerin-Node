package structs

import (
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/lib"
	"github.com/ethereum/go-ethereum/common"
)

type Bid struct {
	Id             common.Hash
	Provider       common.Address
	ModelAgentId   common.Hash
	PricePerSecond *lib.BigInt `swaggertype:"integer"`
	Nonce          *lib.BigInt `swaggertype:"integer"`
	CreatedAt      *lib.BigInt `swaggertype:"integer"`
	DeletedAt      *lib.BigInt `swaggertype:"integer"`
}

type ScoredBid struct {
	ID    common.Hash
	Bid   Bid
	Score float64
}
