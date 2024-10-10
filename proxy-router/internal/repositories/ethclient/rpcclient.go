package ethclient

import (
	"context"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/interfaces"
	"github.com/ethereum/go-ethereum/rpc"
)

type RPCClient interface {
	Close()
	BatchCallContext(context.Context, []rpc.BatchElem) error
	CallContext(ctx context.Context, result interface{}, method string, args ...interface{}) error
	EthSubscribe(ctx context.Context, channel interface{}, args ...interface{}) (*rpc.ClientSubscription, error)
}

type RPCClientModifiable interface {
	RPCClient
	interfaces.RPCEndpoints
}
