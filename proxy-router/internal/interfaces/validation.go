package interfaces

import (
	"context"
	"math/big"
	"time"
)

type Validation interface {
	ValidateEthResourse(ctx context.Context, url string, chainId *big.Int, timeout time.Duration) error
}
