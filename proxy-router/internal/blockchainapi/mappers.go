package blockchainapi

import (
	"encoding/hex"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/contracts/marketplace"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/contracts/sessionrouter"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/blockchainapi/structs"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/lib"
	"github.com/ethereum/go-ethereum/common"
)

func mapBids(bidIDs [][32]byte, bids []marketplace.Bid) []*structs.Bid {
	result := make([]*structs.Bid, len(bidIDs))
	for i, value := range bids {
		result[i] = mapBid(bidIDs[i], value)
	}
	return result
}

func mapBid(bidID common.Hash, bid marketplace.Bid) *structs.Bid {
	return &structs.Bid{
		Id:             bidID,
		ModelAgentId:   bid.ModelAgentId,
		Provider:       bid.Provider,
		Nonce:          &lib.BigInt{Int: *bid.Nonce},
		CreatedAt:      &lib.BigInt{Int: *bid.CreatedAt},
		DeletedAt:      &lib.BigInt{Int: *bid.DeletedAt},
		PricePerSecond: &lib.BigInt{Int: *bid.PricePerSecond},
	}
}

func mapSessions(sessions []sessionrouter.Session) []*structs.Session {
	result := make([]*structs.Session, len(sessions))
	for i, value := range sessions {
		result[i] = mapSession(value)
	}
	return result
}

func mapSession(s sessionrouter.Session) *structs.Session {
	return &structs.Session{
		Id:                      lib.BytesToString(s.Id[:]),
		Provider:                s.Provider,
		User:                    s.User,
		ModelAgentId:            lib.BytesToString(s.ModelAgentId[:]),
		BidID:                   lib.BytesToString(s.BidID[:]),
		Stake:                   s.Stake,
		PricePerSecond:          s.PricePerSecond,
		CloseoutReceipt:         hex.EncodeToString(s.CloseoutReceipt),
		CloseoutType:            s.CloseoutType,
		ProviderWithdrawnAmount: s.ProviderWithdrawnAmount,
		OpenedAt:                s.OpenedAt,
		EndsAt:                  s.EndsAt,
		ClosedAt:                s.ClosedAt,
	}
}
