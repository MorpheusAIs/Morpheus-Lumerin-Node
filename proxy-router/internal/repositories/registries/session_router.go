package registries

import (
	"context"
	"encoding/hex"
	"math/big"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/contracts/modelregistry"
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
	srABI, err := modelregistry.ModelRegistryMetaData.GetAbi()
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

	sessionId := hex.EncodeToString(sessionTx.Data())

	return sessionId, nil
}

func (g *SessionRouter) GetSession(ctx context.Context, sessionId string) (*sessionrouter.Session, error) {
	id, err := hex.DecodeString(sessionId)
	if err != nil {
		return nil, err
	}

	var idBytes [32]byte
	copy(idBytes[:], id)

	session, err := g.sessionRouter.GetSession(&bind.CallOpts{Context: ctx}, idBytes)
	if err != nil {
		return nil, err
	}

	return &session, nil
}
