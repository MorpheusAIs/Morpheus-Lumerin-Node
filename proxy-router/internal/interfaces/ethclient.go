package interfaces

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
)

type EthClient interface {
	ClientBackend
	ContractBackend
}

// ContractWaiter used to do a read/write operation on a contract and wait for the transaction to be mined
type ContractBackend interface {
	bind.ContractCaller
	bind.ContractFilterer
	bind.ContractTransactor
	bind.DeployBackend
}

// ClientBackend used to perform non-contract related read operations on blockchain
type ClientBackend interface {
	BlockNumber(context.Context) (uint64, error)
	BalanceAt(context.Context, common.Address, *big.Int) (*big.Int, error)
	ChainID(context.Context) (*big.Int, error)
}
