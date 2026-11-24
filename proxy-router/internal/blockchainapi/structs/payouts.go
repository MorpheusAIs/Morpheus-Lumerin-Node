package structs

import (
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/lib"
	"github.com/ethereum/go-ethereum/common"
)

// PayoutTargetRequest redirects a provider's payouts. Target is a pointer so
// it must be sent: the zero address resets the target to the provider itself,
// which a request that left the field out must not do by accident.
type PayoutTargetRequest struct {
	Target *common.Address `json:"target" binding:"required" swaggertype:"string" example:"0x1234567890abcdef1234567890abcdef12345678"`
}

// ProviderPayoutsRes is where a provider's rewards go and what it is still
// owed. PendingProviderClaims is the debt to all providers together.
type ProviderPayoutsRes struct {
	Provider              common.Address `json:"provider"`
	PayoutTarget          common.Address `json:"payoutTarget"`
	Owed                  *lib.BigInt    `json:"owed" swaggertype:"string"`
	LifetimeClaimed       *lib.BigInt    `json:"lifetimeClaimed" swaggertype:"string"`
	LifetimeClaimedSince  *lib.BigInt    `json:"lifetimeClaimedSince" swaggertype:"string"`
	PendingProviderClaims *lib.BigInt    `json:"pendingProviderClaims" swaggertype:"string"`
}
