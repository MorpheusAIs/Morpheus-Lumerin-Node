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
	OpenSessionByModelId(ctx context.Context, modelID common.Hash, duration *big.Int, isFailoverEnabled bool, omitProvider common.Address) (common.Hash, error)
	CloseSession(ctx context.Context, sessionID common.Hash) (common.Hash, error)
}
