package contracts

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
)

type DataAccess interface {
	// string used as address cause different blockchains may return different type of addresses
	GetContractsIDs() ([]string, error)
	GetContract(contractID string) (interface{}, error)
	CloseContract(contractID string, meta interface{}) error

	OnNewContract(func(contractID string)) CloseListererFunc
	OnContractUpdated(contractID string, cb func()) CloseListererFunc
	OnContractClosed(contractID string, cb func()) CloseListererFunc
}

type EthereumClient interface {
	bind.ContractBackend
	bind.DeployBackend
	ChainID(ctx context.Context) (*big.Int, error)
	BalanceAt(ctx context.Context, addr common.Address, blockNumber *big.Int) (*big.Int, error)
	StorageAt(ctx context.Context, account common.Address, key common.Hash, blockNumber *big.Int) ([]byte, error)
}

type CloseListererFunc = func()
