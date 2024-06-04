package registries

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/Lumerin-protocol/Morpheus-Lumerin-Node/proxy-router/internal/contracts/sessionrouter"
	"github.com/Lumerin-protocol/Morpheus-Lumerin-Node/proxy-router/internal/internal/interfaces"
	"github.com/Lumerin-protocol/Morpheus-Lumerin-Node/proxy-router/internal/internal/lib"
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
	mutex lib.Mutex
	srABI *abi.ABI

	// deps
	sessionRouter *sessionrouter.SessionRouter
	client        *ethclient.Client
	log           interfaces.ILogger
}

var closeReportAbi = []lib.AbiParameter{
	{Type: "bytes32"},
	{Type: "uint128"},
	{Type: "uint32"},
}

func NewSessionRouter(sessionRouterAddr common.Address, client *ethclient.Client, log interfaces.ILogger) *SessionRouter {
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
		mutex:             lib.NewMutex(),
		log:               log,
	}
}

func (g *SessionRouter) OpenSession(ctx *bind.TransactOpts, approval []byte, approvalSig []byte, stake *big.Int, privateKeyHex string) (string, error) {
	sessionTx, err := g.sessionRouter.OpenSession(ctx, stake, approval, approvalSig)
	if err != nil {
		return "", lib.TryConvertGethError(err, sessionrouter.SessionRouterMetaData)
	}

	// Wait for the transaction receipt
	receipt, err := bind.WaitMined(context.Background(), g.client, sessionTx)
	if err != nil {
		return "", lib.TryConvertGethError(err, sessionrouter.SessionRouterMetaData)
	}

	// Find the event log
	for _, log := range receipt.Logs {
		// Check if the log belongs to the OpenSession event
		event, err := g.sessionRouter.ParseSessionOpened(*log)
		if err != nil {
			continue // not our event, skip it
		}

		// Convert the sessionId to string
		sessionId := lib.BytesToString(event.SessionId[:])
		return sessionId, nil
	}

	return "", fmt.Errorf("OpenSession event not found in transaction logs")
}

func (g *SessionRouter) GetSession(ctx context.Context, sessionId string) (*sessionrouter.Session, error) {
	id := common.FromHex(sessionId)

	session, err := g.sessionRouter.GetSession(&bind.CallOpts{Context: ctx}, [32]byte(id))
	if err != nil {
		return nil, err
	}

	return &session, nil
}

func (g *SessionRouter) GetSessionsByProvider(ctx context.Context, providerAddr common.Address, offset *big.Int, limit uint8) ([]sessionrouter.Session, error) {
	sessions, err := g.sessionRouter.GetSessionsByProvider(&bind.CallOpts{Context: ctx}, providerAddr, offset, limit)
	if err != nil {
		return nil, lib.TryConvertGethError(err, sessionrouter.SessionRouterMetaData)
	}
	return sessions, nil
}

func (g *SessionRouter) GetSessionsByUser(ctx context.Context, userAddr common.Address, offset *big.Int, limit uint8) ([]sessionrouter.Session, error) {
	sessions, err := g.sessionRouter.GetSessionsByUser(&bind.CallOpts{Context: ctx}, userAddr, offset, limit)
	if err != nil {
		return nil, lib.TryConvertGethError(err, sessionrouter.SessionRouterMetaData)
	}
	return sessions, nil
}

func (g *SessionRouter) CloseSession(ctx *bind.TransactOpts, sessionId string, privateKeyHex string) (string, error) {
	id := [32]byte(common.FromHex(sessionId))

	ips := uint32(1)
	timestamp := big.NewInt(time.Now().UnixMilli())
	report, err := lib.EncodeAbiParameters(closeReportAbi, []interface{}{id, timestamp, ips})
	if err != nil {
		return "", err
	}

	signature, err := lib.SignEthMessage(report, privateKeyHex)
	if err != nil {
		return "", err
	}

	sessionTx, err := g.sessionRouter.CloseSession(ctx, report, signature)
	if err != nil {
		return "", lib.TryConvertGethError(err, sessionrouter.SessionRouterMetaData)
	}

	// Wait for the transaction receipt
	receipt, err := bind.WaitMined(context.Background(), g.client, sessionTx)
	if err != nil {
		fmt.Println(err)
		return "", err
	}

	// Find the event log
	for _, log := range receipt.Logs {
		// Check if the log belongs to the CloseSession event
		event, err := g.sessionRouter.ParseSessionClosed(*log)
		if err != nil {
			continue // not our event, skip it
		}

		// Convert the sessionId to string
		sessionId := lib.BytesToString(event.SessionId[:])
		return sessionId, nil
	}

	return "", fmt.Errorf("CloseSession event not found in transaction logs")
}

func (g *SessionRouter) GetProviderClaimableBalance(ctx context.Context, sessionId string) (*big.Int, error) {
	id := [32]byte(common.FromHex(sessionId))
	balance, err := g.sessionRouter.GetProviderClaimableBalance(&bind.CallOpts{Context: ctx}, id)
	if err != nil {
		return nil, lib.TryConvertGethError(err, sessionrouter.SessionRouterMetaData)
	}
	return balance, nil
}

func (g *SessionRouter) ClaimProviderBalance(ctx *bind.TransactOpts, sessionId string, amount *big.Int, to common.Address) (string, error) {
	id := [32]byte(common.FromHex(sessionId))
	tx, err := g.sessionRouter.ClaimProviderBalance(ctx, id, amount, to)
	if err != nil {
		return "", lib.TryConvertGethError(err, sessionrouter.SessionRouterMetaData)
	}

	// Wait for the transaction receipt
	_, err = bind.WaitMined(context.Background(), g.client, tx)
	if err != nil {
		return "", err
	}

	return tx.Hash().String(), nil
}

func (g *SessionRouter) GetTodaysBudget(ctx context.Context) (*big.Int, error) {
	timestamp := big.NewInt(time.Now().Unix())
	budget, err := g.sessionRouter.GetTodaysBudget(&bind.CallOpts{Context: ctx}, timestamp)
	if err != nil {
		return nil, lib.TryConvertGethError(err, sessionrouter.SessionRouterMetaData)
	}
	return budget, nil
}

func (g *SessionRouter) GetContractAddress() common.Address {
	return g.sessionRouterAddr
}

func (g *SessionRouter) GetABI() *abi.ABI {
	return g.srABI
}
