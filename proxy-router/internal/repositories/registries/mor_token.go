package registries

import (
	"context"
	"math/big"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/contracts/morpheustoken"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/internal/interfaces"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/internal/lib"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

type MorToken struct {
	// config
	morTokenAddr common.Address

	// state
	nonce  uint64
	mutex  lib.Mutex
	morABI *abi.ABI

	// deps
	mor    *morpheustoken.MorpheusToken
	client *ethclient.Client
	log    interfaces.ILogger
}

func NewMorToken(morTokenAddr common.Address, client *ethclient.Client, log interfaces.ILogger) *MorToken {
	mor, err := morpheustoken.NewMorpheusToken(morTokenAddr, client)
	if err != nil {
		panic("invalid mor ABI")
	}
	morABI, err := morpheustoken.MorpheusTokenMetaData.GetAbi()
	if err != nil {
		panic("invalid mpr ABI: " + err.Error())
	}
	return &MorToken{
		mor:          mor,
		morTokenAddr: morTokenAddr,
		client:       client,
		morABI:       morABI,
		mutex:        lib.NewMutex(),
		log:          log,
	}
}

func (g *MorToken) GetBalance(ctx context.Context, account common.Address) (*big.Int, error) {
	return g.mor.BalanceOf(&bind.CallOpts{Context: ctx}, account)
}

func (g *MorToken) GetAllowance(ctx context.Context, owner common.Address, spender common.Address) (*big.Int, error) {
	return g.mor.Allowance(&bind.CallOpts{Context: ctx}, owner, spender)
}

func (g *MorToken) Approve(ctx context.Context, spender common.Address, amount *big.Int) (*bind.TransactOpts, error) {
	return nil, nil
}
