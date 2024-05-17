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
	"github.com/ethereum/go-ethereum/crypto"
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
		return "", lib.TryConvertGethError(err, sessionrouter.SessionRouterMetaData)
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

func (g *SessionRouter) CloseSession(ctx *bind.TransactOpts, sessionId string, encodedReport string, privateKeyHex string) (string, error) {
	id := [32]byte(common.FromHex(sessionId))
	report := common.FromHex(encodedReport)

	privateKey, err := crypto.HexToECDSA(privateKeyHex)
	if err != nil {
		return "", err
	}
	hash := crypto.Keccak256Hash(report)

	prefixStr := fmt.Sprintf("\x19Ethereum Signed Message:\n%d", len(hash.Bytes()))
	message := append([]byte(prefixStr), hash.Bytes()...)
	resultHash := crypto.Keccak256Hash(message)

	signature, err := crypto.Sign(resultHash.Bytes(), privateKey)
	if err != nil {
		return "", err
	}

	// https://github.com/ethereum/go-ethereum/blob/44a50c9f96386f44a8682d51cf7500044f6cbaea/internal/ethapi/api.go#L580
	signature[64] += 27 // Transform V from 0/1 to 27/28

	sessionTx, err := g.sessionRouter.CloseSession(ctx, id, []byte(encodedReport), signature)
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

func (g *SessionRouter) GetContractAddress() common.Address {
	return g.sessionRouterAddr
}

func (g *SessionRouter) GetABI() *abi.ABI {
	return g.srABI
}
