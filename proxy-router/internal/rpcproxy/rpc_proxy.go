package rpcproxy

import (
	"context"

	"github.com/ethereum/go-ethereum/ethclient"
)

type RpcProxy struct {
	rpcClient *ethclient.Client
}

func NewRpcProxy(rpcClient *ethclient.Client) *RpcProxy {
	return &RpcProxy{
		rpcClient: rpcClient,
	}
}

func (rpcProxy *RpcProxy) GetLatestBlock(ctx context.Context) (uint64, error) {
	return rpcProxy.rpcClient.BlockNumber(ctx)
}
