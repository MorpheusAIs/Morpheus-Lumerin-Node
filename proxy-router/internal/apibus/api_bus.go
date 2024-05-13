package apibus

import (
	"context"
	"math/big"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/internal/aiengine"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/internal/proxyapi"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/internal/rpcproxy"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gin-gonic/gin"
)

// TODO: split implementations into separate client layer
type ApiBus struct {
	rpcProxy       *rpcproxy.RpcProxy
	aiEngine       *aiengine.AiEngine
	proxyRouterApi *proxyapi.ProxyRouterApi
}

func NewApiBus(rpcProxy *rpcproxy.RpcProxy, aiEngine *aiengine.AiEngine, proxyRouterApi *proxyapi.ProxyRouterApi) *ApiBus {
	return &ApiBus{
		rpcProxy:       rpcProxy,
		aiEngine:       aiEngine,
		proxyRouterApi: proxyRouterApi,
	}
}

// Proxy Router Api
func (apiBus *ApiBus) GetConfig(ctx context.Context) interface{} {
	return apiBus.proxyRouterApi.GetConfig(ctx)
}

func (apiBus *ApiBus) GetFiles(ctx *gin.Context) (int, interface{}) {
	return apiBus.proxyRouterApi.GetFiles(ctx)
}

func (apiBus *ApiBus) HealthCheck(ctx context.Context) interface{} {
	return apiBus.proxyRouterApi.HealthCheck(ctx)
}

func (apiBus *ApiBus) InitiateSession(ctx *gin.Context) (int, interface{}) {
	return apiBus.proxyRouterApi.InitiateSession(ctx)
}

func (apiBus *ApiBus) SendPrompt(ctx *gin.Context) (bool, int, interface{}) {
	return apiBus.proxyRouterApi.SendPrompt(ctx)
}

// AiEngine
func (apiBus *ApiBus) Prompt(ctx context.Context, req interface{}) (interface{}, error) {
	return apiBus.aiEngine.Prompt(ctx, req)
}

// AiEngine
func (apiBus *ApiBus) PromptStream(ctx context.Context, req interface{}, flush interface{}) (interface{}, error) {
	return apiBus.aiEngine.PromptStream(ctx, req, flush)
}

// RpcProxy
func (apiBus *ApiBus) GetLatestBlock(ctx context.Context) (uint64, error) {
	return apiBus.rpcProxy.GetLatestBlock(ctx)
}

func (apiBus *ApiBus) GetAllProviders(ctx context.Context) (int, gin.H) {
	return apiBus.rpcProxy.GetAllProviders(ctx)
}

func (apiBus *ApiBus) GetAllModels(ctx context.Context) (int, gin.H) {
	return apiBus.rpcProxy.GetAllModels(ctx)
}

func (apiBus *ApiBus) GetBidsByProvider(ctx context.Context, providerAddr string, offset *big.Int, limit uint8) (int, gin.H) {
	addr := common.HexToAddress(providerAddr)
	return apiBus.rpcProxy.GetBidsByProvider(ctx, addr, offset, limit)
}

func (apiBus *ApiBus) GetBidsByModelAgent(ctx context.Context, modelAgentId [32]byte, offset *big.Int, limit uint8) (int, gin.H) {
	return apiBus.rpcProxy.GetBidsByModelAgent(ctx, modelAgentId, offset, limit)
}

func (apiBus *ApiBus) OpenSession(ctx *gin.Context) (int, gin.H) {
	return apiBus.rpcProxy.OpenSession(ctx)
}

func (apiBus *ApiBus) CloseSession(ctx *gin.Context) (int, gin.H) {
	return apiBus.rpcProxy.CloseSession(ctx)
}
