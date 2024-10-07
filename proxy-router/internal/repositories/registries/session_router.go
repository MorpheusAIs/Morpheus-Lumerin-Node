package registries

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/contracts/sessionrouter"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/lib"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

type SessionRouter struct {
	// config
	sessionRouterAddr common.Address

	// state
	nonce uint64
	srABI *abi.ABI

	// deps
	sessionRouter *sessionrouter.SessionRouter
	client        *ethclient.Client
	log           lib.ILogger
}

var closeReportAbi = []lib.AbiParameter{
	{Type: "bytes32"},
	{Type: "uint128"},
	{Type: "uint32"},
}

func NewSessionRouter(sessionRouterAddr common.Address, client *ethclient.Client, log lib.ILogger) *SessionRouter {
	sr, err := sessionrouter.NewSessionRouter(sessionRouterAddr, client)
	if err != nil {
		panic("invalid marketplace ABI")
	}
	srABI, err := sessionrouter.SessionRouterMetaData.GetAbi()
	if err != nil {
		panic("invalid marketplace ABI: " + err.Error())
	}
	return &SessionRouter{
		sessionRouter:     sr,
		sessionRouterAddr: sessionRouterAddr,
		client:            client,
		srABI:             srABI,
		log:               log,
	}
}

func (g *SessionRouter) OpenSession(opts *bind.TransactOpts, approval []byte, approvalSig []byte, stake *big.Int, privateKeyHex lib.HexString) (sessionID common.Hash, providerID common.Address, userID common.Address, err error) {
	sessionTx, err := g.sessionRouter.OpenSession(opts, stake, approval, approvalSig)
	if err != nil {
		return common.Hash{}, common.Address{}, common.Address{}, lib.TryConvertGethError(err)
	}

	// Wait for the transaction receipt
	receipt, err := bind.WaitMined(opts.Context, g.client, sessionTx)
	if err != nil {
		return common.Hash{}, common.Address{}, common.Address{}, lib.TryConvertGethError(err)
	}

	// Find the event log
	for _, log := range receipt.Logs {
		// Check if the log belongs to the OpenSession event
		event, err := g.sessionRouter.ParseSessionOpened(*log)
		if err == nil {
			return event.SessionId, event.ProviderId, event.User, nil
		}
	}

	return common.Hash{}, common.Address{}, common.Address{}, fmt.Errorf("OpenSession event not found in transaction logs")
}

func (g *SessionRouter) GetSession(ctx context.Context, sessionID common.Hash) (*sessionrouter.ISessionStorageSession, error) {
	session, err := g.sessionRouter.Sessions(&bind.CallOpts{Context: ctx}, sessionID)
	if err != nil {
		return nil, err
	}

	return &session, nil
}

func (g *SessionRouter) GetSessionsByProvider(ctx context.Context, providerAddr common.Address, offset *big.Int, limit uint8) ([]sessionrouter.ISessionStorageSession, error) {
	// sessions, err := g.sessionRouter.GetSessionsByProvider(&bind.CallOpts{Context: ctx}, providerAddr, offset, limit)
	// if err != nil {
	// 	return nil, lib.TryConvertGethError(err)
	// }
	// return sessions, nil
	return nil, fmt.Errorf("Not implemented")
}

func (g *SessionRouter) GetSessionsByUser(ctx context.Context, userAddr common.Address, offset *big.Int, limit uint8) ([]sessionrouter.ISessionStorageSession, error) {
	IDs, err := g.sessionRouter.GetSessionsByUser(&bind.CallOpts{Context: ctx}, userAddr, offset, big.NewInt(int64(limit)))
	if err != nil {
		return nil, lib.TryConvertGethError(err)
	}
	return g.getMultipleSessions(ctx, IDs)
}

func (g *SessionRouter) CloseSession(opts *bind.TransactOpts, sessionID common.Hash, report []byte, signedReport []byte, privateKeyHex lib.HexString) (common.Hash, error) {
	sessionTx, err := g.sessionRouter.CloseSession(opts, report, signedReport)
	if err != nil {
		return common.Hash{}, lib.TryConvertGethError(err)
	}

	// Wait for the transaction receipt
	_, err = bind.WaitMined(opts.Context, g.client, sessionTx)
	if err != nil {
		return common.Hash{}, err
	}

	return sessionTx.Hash(), nil
}

func (g *SessionRouter) GetProviderClaimableBalance(ctx context.Context, sessionId [32]byte) (*big.Int, error) {
	balance, err := g.sessionRouter.GetProviderClaimableBalance(&bind.CallOpts{Context: ctx}, sessionId)
	if err != nil {
		return nil, lib.TryConvertGethError(err)
	}
	return balance, nil
}

func (g *SessionRouter) ClaimProviderBalance(opts *bind.TransactOpts, sessionId [32]byte, amount *big.Int) (common.Hash, error) {
	tx, err := g.sessionRouter.ClaimProviderBalance(opts, sessionId, amount)
	if err != nil {
		return common.Hash{}, lib.TryConvertGethError(err)
	}

	// Wait for the transaction receipt
	_, err = bind.WaitMined(opts.Context, g.client, tx)
	if err != nil {
		return common.Hash{}, err
	}

	return tx.Hash(), nil
}

func (g *SessionRouter) GetTodaysBudget(ctx context.Context) (*big.Int, error) {
	timestamp := big.NewInt(time.Now().Unix())
	budget, err := g.sessionRouter.GetTodaysBudget(&bind.CallOpts{Context: ctx}, timestamp)
	if err != nil {
		return nil, lib.TryConvertGethError(err)
	}
	return budget, nil
}

func (g *SessionRouter) GetBidsWithRating(ctx context.Context, modelAgentID [32]byte, offset *big.Int, limit uint8) ([][32]byte, []sessionrouter.IBidStorageBid, []sessionrouter.IStatsStorageProviderModelStats, error) {
	return g.sessionRouter.GetActiveBidsRatingByModel(&bind.CallOpts{Context: ctx}, modelAgentID, offset, limit)
}

func (g *SessionRouter) GetAllBidsWithRating(ctx context.Context, modelAgentID [32]byte) ([][32]byte, []sessionrouter.IBidStorageBid, []sessionrouter.IStatsStorageProviderModelStats, error) {
	batchSize := uint8(255)
	return collectBids(ctx, modelAgentID, g.GetBidsWithRating, batchSize)
}

func (g *SessionRouter) GetModelStats(ctx context.Context, modelID [32]byte) (interface{}, error) {
	// return g.sessionRouter.GetModelStats(&bind.CallOpts{Context: ctx}, modelID)
	return nil, fmt.Errorf("not implemented")
}

func (g *SessionRouter) getMultipleSessions(ctx context.Context, IDs [][32]byte) ([]sessionrouter.ISessionStorageSession, error) {
	// todo: replace with multicall
	sessions := make([]sessionrouter.ISessionStorageSession, len(IDs))
	for i, id := range IDs {
		session, err := g.sessionRouter.Sessions(&bind.CallOpts{Context: ctx}, id)
		if err != nil {
			return nil, err
		}
		sessions[i] = session
	}
	return sessions, nil
}

func (g *SessionRouter) GetContractAddress() common.Address {
	return g.sessionRouterAddr
}

func (g *SessionRouter) GetABI() *abi.ABI {
	return g.srABI
}

type BidsGetter = func(ctx context.Context, modelAgentID [32]byte, offset *big.Int, limit uint8) ([][32]byte, []sessionrouter.IBidStorageBid, []sessionrouter.IStatsStorageProviderModelStats, error)

func collectBids(ctx context.Context, modelAgentID [32]byte, bidsGetter BidsGetter, batchSize uint8) ([][32]byte, []sessionrouter.IBidStorageBid, []sessionrouter.IStatsStorageProviderModelStats, error) {
	offset := big.NewInt(0)
	bids := make([]sessionrouter.IBidStorageBid, 0)
	ids := make([][32]byte, 0)
	providerModelStats := make([]sessionrouter.IStatsStorageProviderModelStats, 0)

	for {
		if ctx.Err() != nil {
			return nil, nil, nil, ctx.Err()
		}

		idsBatch, bidsBatch, providerModelStatsBatch, err := bidsGetter(ctx, modelAgentID, offset, batchSize)
		if err != nil {
			return nil, nil, nil, err
		}

		ids = append(ids, idsBatch...)
		bids = append(bids, bidsBatch...)
		providerModelStats = append(providerModelStats, providerModelStatsBatch...)

		if len(bidsBatch) < int(batchSize) {
			break
		}

		offset.Add(offset, big.NewInt(int64(batchSize)))
	}

	return ids, bids, providerModelStats, nil
}
