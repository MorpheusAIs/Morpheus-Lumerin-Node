package structs

import (
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/lib"
	"github.com/ethereum/go-ethereum/common"
)

type Bid struct {
	Id             common.Hash
	Provider       common.Address
	ModelAgentId   common.Hash
	PricePerSecond *lib.BigInt
	Nonce          *lib.BigInt
	CreatedAt      *lib.BigInt
	DeletedAt      *lib.BigInt
}

type ScoredBid struct {
	ID    common.Hash
	Bid   Bid
	Score float64
}
