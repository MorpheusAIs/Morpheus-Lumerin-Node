package structs

import (
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/lib"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/videojob"
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

// VideoProviderGuide is one bid's provider and the video guide it advertises
// for the model, or why there is none.
type VideoProviderGuide struct {
	Provider          common.Address  `json:"provider"`
	BidID             common.Hash     `json:"bid_id"`
	PricePerSecondWei string          `json:"price_per_second_wei"`
	Guide             *videojob.Guide `json:"guide,omitempty"`
	Error             string          `json:"error,omitempty"`
}
