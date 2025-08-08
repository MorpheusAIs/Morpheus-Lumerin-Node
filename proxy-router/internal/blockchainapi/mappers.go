package blockchainapi

import (
	"encoding/hex"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/blockchainapi/structs"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/lib"
	m "github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/repositories/contracts/bindings/marketplace"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/repositories/contracts/bindings/modelregistry"
	pr "github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/repositories/contracts/bindings/providerregistry"
	s "github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/repositories/contracts/bindings/sessionrouter"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/repositories/registries"
	"github.com/ethereum/go-ethereum/common"
)

func mapBids(bidIDs [][32]byte, bids []m.IBidStorageBid) []*structs.Bid {
	result := make([]*structs.Bid, len(bidIDs))
	for i, value := range bids {
		result[i] = mapBid(bidIDs[i], value)
	}
	return result
}

func mapBid(bidID common.Hash, bid m.IBidStorageBid) *structs.Bid {
	return &structs.Bid{
		Id:             bidID,
		ModelAgentId:   bid.ModelId,
		Provider:       bid.Provider,
		Nonce:          &lib.BigInt{Int: *bid.Nonce},
		CreatedAt:      &lib.BigInt{Int: *bid.CreatedAt},
		DeletedAt:      &lib.BigInt{Int: *bid.DeletedAt},
		PricePerSecond: &lib.BigInt{Int: *bid.PricePerSecond},
	}
}

func mapSessions(sessionIDs [][32]byte, sessions []s.ISessionStorageSession, bids []m.IBidStorageBid) []*structs.Session {
	result := make([]*structs.Session, len(sessionIDs))
	for i := 0; i < len(sessionIDs); i++ {
		result[i] = mapSession(sessionIDs[i], sessions[i], bids[i])
	}
	return result
}

func mapSession(ID common.Hash, ses s.ISessionStorageSession, bid m.IBidStorageBid) *structs.Session {
	return &structs.Session{
		Id:                      lib.BytesToString(ID[:]),
		Provider:                bid.Provider,
		User:                    ses.User,
		ModelAgentId:            lib.BytesToString(bid.ModelId[:]),
		BidID:                   lib.BytesToString(ses.BidId[:]),
		Stake:                   ses.Stake,
		PricePerSecond:          bid.PricePerSecond,
		CloseoutReceipt:         hex.EncodeToString(ses.CloseoutReceipt),
		CloseoutType:            ses.CloseoutType,
		ProviderWithdrawnAmount: ses.ProviderWithdrawnAmount,
		OpenedAt:                ses.OpenedAt,
		EndsAt:                  ses.EndsAt,
		ClosedAt:                ses.ClosedAt,
	}
}

func mapModels(ids [][32]byte, models []modelregistry.IModelStorageModel) []*structs.Model {
	result := make([]*structs.Model, len(ids))
	for i, value := range models {
		result[i] = mapModel(ids[i], value)
	}
	return result
}

func mapModel(id [32]byte, model modelregistry.IModelStorageModel) *structs.Model {
	var modelType structs.ModelType = "UNKNOWN" // Default type
	for _, tag := range model.Tags {
		if tag == "STT" {
			modelType = structs.ModelTypeSTT
			break
		} else if tag == "TTS" {
			modelType = structs.ModelTypeTTS
			break
		} else if tag == "EMBEDDING" {
			modelType = structs.ModelTypeEMBEDDING
			break
		} else if tag == "LLM" {
			modelType = structs.ModelTypeLLM
			break
		}
	}
	return &structs.Model{
		Id:        id,
		IpfsCID:   model.IpfsCID,
		Fee:       model.Fee,
		Stake:     model.Stake,
		Owner:     model.Owner,
		Name:      model.Name,
		Tags:      model.Tags,
		CreatedAt: model.CreatedAt,
		IsDeleted: model.IsDeleted,
		ModelType: modelType,
	}
}

func mapProviders(addrs []common.Address, providers []pr.IProviderStorageProvider) []*structs.Provider {
	result := make([]*structs.Provider, len(addrs))
	for i, value := range providers {
		result[i] = mapProvider(addrs[i], value)
	}
	return result
}

func mapProvider(addr common.Address, provider pr.IProviderStorageProvider) *structs.Provider {
	return &structs.Provider{
		Address:   addr,
		Endpoint:  provider.Endpoint,
		Stake:     &lib.BigInt{Int: *provider.Stake},
		IsDeleted: provider.IsDeleted,
		CreatedAt: &lib.BigInt{Int: *provider.CreatedAt},
	}
}

func mapOrder(order string) registries.Order {
	if order == "ASC" {
		return registries.OrderASC
	}
	return registries.OrderDESC
}
