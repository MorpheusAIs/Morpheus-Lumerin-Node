package registries

import (
	"context"
	"fmt"
	"math/big"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/contracts/sessionrouter"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/internal/interfaces"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/internal/lib"
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

func (g *SessionRouter) OpenSession(ctx *bind.TransactOpts, bidId [32]byte, stake *big.Int) (string, error) {
	sessionTx, err := g.sessionRouter.OpenSession(ctx, bidId, stake)
	if err != nil {
		return "", err
	}

	// Wait for the transaction receipt
	receipt, err := bind.WaitMined(context.Background(), g.client, sessionTx)
	if err != nil {
		return "", err
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

func (g *SessionRouter) GetContractAddress() common.Address {
	return g.sessionRouterAddr
}

func (g *SessionRouter) GetABI() *abi.ABI {
	return g.srABI
}
