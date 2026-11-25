package proxyapi

import (
	"context"
	"math/big"
	"net/http"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/blockchainapi/structs"
	"github.com/ethereum/go-ethereum/common"
)

type ResponderFlusher interface {
	http.ResponseWriter
	http.Flusher
}

type BidGetter interface {
	GetBidByID(ctx context.Context, ID common.Hash) (*structs.Bid, error)
}

type SessionService interface {
	OpenSessionByModelId(ctx context.Context, modelID common.Hash, duration *big.Int, isDirectPayment, fromPool, isFailoverEnabled bool, omitProvider common.Address, agentUsername string) (common.Hash, error)
	// CloseSessionReportingPool closes the session and reports whether the
	// close the chain accepted was the pool close.
	CloseSessionReportingPool(ctx context.Context, sessionID common.Hash) (txHash common.Hash, fromPool bool, err error)
}
