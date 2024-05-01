package rpcproxy

import (
	"context"
	"math/big"

	constants "github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/internal"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/internal/interfaces"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/internal/repositories/registries"
	"github.com/gin-gonic/gin"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

type RpcProxy struct {
	rpcClient        *ethclient.Client
	providerRegistry *registries.ProviderRegistry
	modelRegistry    *registries.ModelRegistry
	marketplace      *registries.Marketplace
}

func NewRpcProxy(rpcClient *ethclient.Client, diamonContractAddr common.Address, log interfaces.ILogger) *RpcProxy {
	providerRegistry := registries.NewProviderRegistry(diamonContractAddr, rpcClient, log)
	modelRegistry := registries.NewModelRegistry(diamonContractAddr, rpcClient, log)
	marketplace := registries.NewMarketplace(diamonContractAddr, rpcClient, log)
	return &RpcProxy{
		rpcClient:        rpcClient,
		providerRegistry: providerRegistry,
		modelRegistry:    modelRegistry,
		marketplace:      marketplace,
	}
}

func (rpcProxy *RpcProxy) GetLatestBlock(ctx context.Context) (uint64, error) {
	return rpcProxy.rpcClient.BlockNumber(ctx)
}

func (rpcProxy *RpcProxy) GetAllProviders(ctx context.Context) (int, gin.H) {
	addrs, providers, err := rpcProxy.providerRegistry.GetAllProviders(ctx)
	if err != nil {
		return constants.HTTP_STATUS_BAD_REQUEST, gin.H{"error": err.Error()}
	}
	return constants.HTTP_STATUS_OK, gin.H{"addresses": addrs, "providers": providers}
}

func (rpcProxy *RpcProxy) GetAllModels(ctx context.Context) (int, gin.H) {
	models, err := rpcProxy.modelRegistry.GetAllProviders(ctx)
	if err != nil {
		return constants.HTTP_STATUS_BAD_REQUEST, gin.H{"error": err.Error()}
	}
	return constants.HTTP_STATUS_OK, gin.H{"models": models}
}

func (rpcProxy *RpcProxy) GetBidsByProvider(ctx context.Context, providerAddr common.Address, offset *big.Int, limit uint8) (int, gin.H) {
	bids, err := rpcProxy.marketplace.GetBidsByProvider(ctx, providerAddr, offset, limit)
	if err != nil {
		return constants.HTTP_STATUS_BAD_REQUEST, gin.H{"error": err.Error()}
	}
	return constants.HTTP_STATUS_OK, gin.H{"bids": bids}
}

func (rpcProxy *RpcProxy) GetBidsByModelAgent(ctx context.Context, modelId [32]byte, offset *big.Int, limit uint8) (int, gin.H) {
	bids, err := rpcProxy.marketplace.GetBidsByModelAgent(ctx, modelId, offset, limit)
	if err != nil {
		return constants.HTTP_STATUS_BAD_REQUEST, gin.H{"error": err.Error()}
	}
	return constants.HTTP_STATUS_OK, gin.H{"bids": bids}
}
