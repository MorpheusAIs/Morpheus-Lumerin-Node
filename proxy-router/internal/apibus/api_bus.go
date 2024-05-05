package apibus

import (
	"context"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/internal/aiengine"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/internal/proxyapi"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/internal/rpcproxy"
	"github.com/gin-gonic/gin"
)

//TODO: split implementations into separate client layer
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
