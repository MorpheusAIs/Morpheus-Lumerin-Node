package multicall

import (
	"context"
	"math/big"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/repositories/contracts/bindings/multicall3"
)

type MulticallBackend interface {
	Aggregate(ctx context.Context, calls []multicall3.Multicall3Call) (blockNumer *big.Int, returnData [][]byte, err error)
}
