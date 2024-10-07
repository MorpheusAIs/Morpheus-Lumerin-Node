// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package providerregistry

import (
	"errors"
	"math/big"
	"strings"

	ethereum "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/event"
)

// Reference imports to suppress errors if they are not otherwise used.
var (
	_ = errors.New
	_ = big.NewInt
	_ = strings.NewReader
	_ = ethereum.NotFound
	_ = bind.Bind
	_ = common.Big1
	_ = types.BloomLookup
	_ = event.NewSubscription
	_ = abi.ConvertType
)

// IBidStorageBid is an auto generated low-level Go binding around an user-defined struct.
type IBidStorageBid struct {
	Provider       common.Address
	ModelId        [32]byte
	PricePerSecond *big.Int
	Nonce          *big.Int
	CreatedAt      *big.Int
	DeletedAt      *big.Int
}

// IProviderStorageProvider is an auto generated low-level Go binding around an user-defined struct.
type IProviderStorageProvider struct {
	Endpoint          string
	Stake             *big.Int
	CreatedAt         *big.Int
	LimitPeriodEnd    *big.Int
	LimitPeriodEarned *big.Int
	IsDeleted         bool
}

// ProviderRegistryMetaData contains all meta data concerning the ProviderRegistry contract.
var ProviderRegistryMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[],\"name\":\"ErrNoStake\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"ErrNoWithdrawableStake\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"ErrProviderNotDeleted\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"NotOwnerOrProvider\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"account_\",\"type\":\"address\"}],\"name\":\"OwnableUnauthorizedAccount\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"ProviderHasActiveBids\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"ProviderNotFound\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"StakeTooLow\",\"type\":\"error\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"bytes32\",\"name\":\"storageSlot\",\"type\":\"bytes32\"}],\"name\":\"Initialized\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"provider\",\"type\":\"address\"}],\"name\":\"ProviderDeregistered\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"newStake\",\"type\":\"uint256\"}],\"name\":\"ProviderMinStakeUpdated\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"provider\",\"type\":\"address\"}],\"name\":\"ProviderRegisteredUpdated\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"provider\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"ProviderWithdrawnStake\",\"type\":\"event\"},{\"inputs\":[],\"name\":\"BID_STORAGE_SLOT\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"DIAMOND_OWNABLE_STORAGE_SLOT\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"PROVIDER_STORAGE_SLOT\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"__ProviderRegistry_init\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"bidId\",\"type\":\"bytes32\"}],\"name\":\"bids\",\"outputs\":[{\"components\":[{\"internalType\":\"address\",\"name\":\"provider\",\"type\":\"address\"},{\"internalType\":\"bytes32\",\"name\":\"modelId\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"pricePerSecond\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"nonce\",\"type\":\"uint256\"},{\"internalType\":\"uint128\",\"name\":\"createdAt\",\"type\":\"uint128\"},{\"internalType\":\"uint128\",\"name\":\"deletedAt\",\"type\":\"uint128\"}],\"internalType\":\"structIBidStorage.Bid\",\"name\":\"\",\"type\":\"tuple\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"provider\",\"type\":\"address\"}],\"name\":\"getProvider\",\"outputs\":[{\"components\":[{\"internalType\":\"string\",\"name\":\"endpoint\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"stake\",\"type\":\"uint256\"},{\"internalType\":\"uint128\",\"name\":\"createdAt\",\"type\":\"uint128\"},{\"internalType\":\"uint128\",\"name\":\"limitPeriodEnd\",\"type\":\"uint128\"},{\"internalType\":\"uint256\",\"name\":\"limitPeriodEarned\",\"type\":\"uint256\"},{\"internalType\":\"bool\",\"name\":\"isDeleted\",\"type\":\"bool\"}],\"internalType\":\"structIProviderStorage.Provider\",\"name\":\"\",\"type\":\"tuple\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getToken\",\"outputs\":[{\"internalType\":\"contractIERC20\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"provider_\",\"type\":\"address\"}],\"name\":\"isProviderExists\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"modelId_\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"offset_\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"limit_\",\"type\":\"uint256\"}],\"name\":\"modelActiveBids\",\"outputs\":[{\"internalType\":\"bytes32[]\",\"name\":\"\",\"type\":\"bytes32[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"modelId_\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"offset_\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"limit_\",\"type\":\"uint256\"}],\"name\":\"modelBids\",\"outputs\":[{\"internalType\":\"bytes32[]\",\"name\":\"\",\"type\":\"bytes32[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"owner\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"provider_\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"offset_\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"limit_\",\"type\":\"uint256\"}],\"name\":\"providerActiveBids\",\"outputs\":[{\"internalType\":\"bytes32[]\",\"name\":\"\",\"type\":\"bytes32[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"provider_\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"offset_\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"limit_\",\"type\":\"uint256\"}],\"name\":\"providerBids\",\"outputs\":[{\"internalType\":\"bytes32[]\",\"name\":\"\",\"type\":\"bytes32[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"provider_\",\"type\":\"address\"}],\"name\":\"providerDeregister\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"providerMinimumStake\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"providerAddress_\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount_\",\"type\":\"uint256\"},{\"internalType\":\"string\",\"name\":\"endpoint_\",\"type\":\"string\"}],\"name\":\"providerRegister\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"providerMinimumStake_\",\"type\":\"uint256\"}],\"name\":\"providerSetMinStake\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"provider_\",\"type\":\"address\"}],\"name\":\"providerWithdrawStake\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]",
}

// ProviderRegistryABI is the input ABI used to generate the binding from.
// Deprecated: Use ProviderRegistryMetaData.ABI instead.
var ProviderRegistryABI = ProviderRegistryMetaData.ABI

// ProviderRegistry is an auto generated Go binding around an Ethereum contract.
type ProviderRegistry struct {
	ProviderRegistryCaller     // Read-only binding to the contract
	ProviderRegistryTransactor // Write-only binding to the contract
	ProviderRegistryFilterer   // Log filterer for contract events
}

// ProviderRegistryCaller is an auto generated read-only Go binding around an Ethereum contract.
type ProviderRegistryCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ProviderRegistryTransactor is an auto generated write-only Go binding around an Ethereum contract.
type ProviderRegistryTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ProviderRegistryFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type ProviderRegistryFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ProviderRegistrySession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type ProviderRegistrySession struct {
	Contract     *ProviderRegistry // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// ProviderRegistryCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type ProviderRegistryCallerSession struct {
	Contract *ProviderRegistryCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts           // Call options to use throughout this session
}

// ProviderRegistryTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type ProviderRegistryTransactorSession struct {
	Contract     *ProviderRegistryTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts           // Transaction auth options to use throughout this session
}

// ProviderRegistryRaw is an auto generated low-level Go binding around an Ethereum contract.
type ProviderRegistryRaw struct {
	Contract *ProviderRegistry // Generic contract binding to access the raw methods on
}

// ProviderRegistryCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type ProviderRegistryCallerRaw struct {
	Contract *ProviderRegistryCaller // Generic read-only contract binding to access the raw methods on
}

// ProviderRegistryTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type ProviderRegistryTransactorRaw struct {
	Contract *ProviderRegistryTransactor // Generic write-only contract binding to access the raw methods on
}

// NewProviderRegistry creates a new instance of ProviderRegistry, bound to a specific deployed contract.
func NewProviderRegistry(address common.Address, backend bind.ContractBackend) (*ProviderRegistry, error) {
	contract, err := bindProviderRegistry(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &ProviderRegistry{ProviderRegistryCaller: ProviderRegistryCaller{contract: contract}, ProviderRegistryTransactor: ProviderRegistryTransactor{contract: contract}, ProviderRegistryFilterer: ProviderRegistryFilterer{contract: contract}}, nil
}

// NewProviderRegistryCaller creates a new read-only instance of ProviderRegistry, bound to a specific deployed contract.
func NewProviderRegistryCaller(address common.Address, caller bind.ContractCaller) (*ProviderRegistryCaller, error) {
	contract, err := bindProviderRegistry(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &ProviderRegistryCaller{contract: contract}, nil
}

// NewProviderRegistryTransactor creates a new write-only instance of ProviderRegistry, bound to a specific deployed contract.
func NewProviderRegistryTransactor(address common.Address, transactor bind.ContractTransactor) (*ProviderRegistryTransactor, error) {
	contract, err := bindProviderRegistry(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &ProviderRegistryTransactor{contract: contract}, nil
}

// NewProviderRegistryFilterer creates a new log filterer instance of ProviderRegistry, bound to a specific deployed contract.
func NewProviderRegistryFilterer(address common.Address, filterer bind.ContractFilterer) (*ProviderRegistryFilterer, error) {
	contract, err := bindProviderRegistry(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &ProviderRegistryFilterer{contract: contract}, nil
}

// bindProviderRegistry binds a generic wrapper to an already deployed contract.
func bindProviderRegistry(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := ProviderRegistryMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_ProviderRegistry *ProviderRegistryRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _ProviderRegistry.Contract.ProviderRegistryCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_ProviderRegistry *ProviderRegistryRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _ProviderRegistry.Contract.ProviderRegistryTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_ProviderRegistry *ProviderRegistryRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _ProviderRegistry.Contract.ProviderRegistryTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_ProviderRegistry *ProviderRegistryCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _ProviderRegistry.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_ProviderRegistry *ProviderRegistryTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _ProviderRegistry.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_ProviderRegistry *ProviderRegistryTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _ProviderRegistry.Contract.contract.Transact(opts, method, params...)
}

// BIDSTORAGESLOT is a free data retrieval call binding the contract method 0x4fa816f2.
//
// Solidity: function BID_STORAGE_SLOT() view returns(bytes32)
func (_ProviderRegistry *ProviderRegistryCaller) BIDSTORAGESLOT(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _ProviderRegistry.contract.Call(opts, &out, "BID_STORAGE_SLOT")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// BIDSTORAGESLOT is a free data retrieval call binding the contract method 0x4fa816f2.
//
// Solidity: function BID_STORAGE_SLOT() view returns(bytes32)
func (_ProviderRegistry *ProviderRegistrySession) BIDSTORAGESLOT() ([32]byte, error) {
	return _ProviderRegistry.Contract.BIDSTORAGESLOT(&_ProviderRegistry.CallOpts)
}

// BIDSTORAGESLOT is a free data retrieval call binding the contract method 0x4fa816f2.
//
// Solidity: function BID_STORAGE_SLOT() view returns(bytes32)
func (_ProviderRegistry *ProviderRegistryCallerSession) BIDSTORAGESLOT() ([32]byte, error) {
	return _ProviderRegistry.Contract.BIDSTORAGESLOT(&_ProviderRegistry.CallOpts)
}

// DIAMONDOWNABLESTORAGESLOT is a free data retrieval call binding the contract method 0x4ac3371e.
//
// Solidity: function DIAMOND_OWNABLE_STORAGE_SLOT() view returns(bytes32)
func (_ProviderRegistry *ProviderRegistryCaller) DIAMONDOWNABLESTORAGESLOT(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _ProviderRegistry.contract.Call(opts, &out, "DIAMOND_OWNABLE_STORAGE_SLOT")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// DIAMONDOWNABLESTORAGESLOT is a free data retrieval call binding the contract method 0x4ac3371e.
//
// Solidity: function DIAMOND_OWNABLE_STORAGE_SLOT() view returns(bytes32)
func (_ProviderRegistry *ProviderRegistrySession) DIAMONDOWNABLESTORAGESLOT() ([32]byte, error) {
	return _ProviderRegistry.Contract.DIAMONDOWNABLESTORAGESLOT(&_ProviderRegistry.CallOpts)
}

// DIAMONDOWNABLESTORAGESLOT is a free data retrieval call binding the contract method 0x4ac3371e.
//
// Solidity: function DIAMOND_OWNABLE_STORAGE_SLOT() view returns(bytes32)
func (_ProviderRegistry *ProviderRegistryCallerSession) DIAMONDOWNABLESTORAGESLOT() ([32]byte, error) {
	return _ProviderRegistry.Contract.DIAMONDOWNABLESTORAGESLOT(&_ProviderRegistry.CallOpts)
}

// PROVIDERSTORAGESLOT is a free data retrieval call binding the contract method 0x490713b1.
//
// Solidity: function PROVIDER_STORAGE_SLOT() view returns(bytes32)
func (_ProviderRegistry *ProviderRegistryCaller) PROVIDERSTORAGESLOT(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _ProviderRegistry.contract.Call(opts, &out, "PROVIDER_STORAGE_SLOT")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// PROVIDERSTORAGESLOT is a free data retrieval call binding the contract method 0x490713b1.
//
// Solidity: function PROVIDER_STORAGE_SLOT() view returns(bytes32)
func (_ProviderRegistry *ProviderRegistrySession) PROVIDERSTORAGESLOT() ([32]byte, error) {
	return _ProviderRegistry.Contract.PROVIDERSTORAGESLOT(&_ProviderRegistry.CallOpts)
}

// PROVIDERSTORAGESLOT is a free data retrieval call binding the contract method 0x490713b1.
//
// Solidity: function PROVIDER_STORAGE_SLOT() view returns(bytes32)
func (_ProviderRegistry *ProviderRegistryCallerSession) PROVIDERSTORAGESLOT() ([32]byte, error) {
	return _ProviderRegistry.Contract.PROVIDERSTORAGESLOT(&_ProviderRegistry.CallOpts)
}

// Bids is a free data retrieval call binding the contract method 0x8f98eeda.
//
// Solidity: function bids(bytes32 bidId) view returns((address,bytes32,uint256,uint256,uint128,uint128))
func (_ProviderRegistry *ProviderRegistryCaller) Bids(opts *bind.CallOpts, bidId [32]byte) (IBidStorageBid, error) {
	var out []interface{}
	err := _ProviderRegistry.contract.Call(opts, &out, "bids", bidId)

	if err != nil {
		return *new(IBidStorageBid), err
	}

	out0 := *abi.ConvertType(out[0], new(IBidStorageBid)).(*IBidStorageBid)

	return out0, err

}

// Bids is a free data retrieval call binding the contract method 0x8f98eeda.
//
// Solidity: function bids(bytes32 bidId) view returns((address,bytes32,uint256,uint256,uint128,uint128))
func (_ProviderRegistry *ProviderRegistrySession) Bids(bidId [32]byte) (IBidStorageBid, error) {
	return _ProviderRegistry.Contract.Bids(&_ProviderRegistry.CallOpts, bidId)
}

// Bids is a free data retrieval call binding the contract method 0x8f98eeda.
//
// Solidity: function bids(bytes32 bidId) view returns((address,bytes32,uint256,uint256,uint128,uint128))
func (_ProviderRegistry *ProviderRegistryCallerSession) Bids(bidId [32]byte) (IBidStorageBid, error) {
	return _ProviderRegistry.Contract.Bids(&_ProviderRegistry.CallOpts, bidId)
}

// GetProvider is a free data retrieval call binding the contract method 0x55f21eb7.
//
// Solidity: function getProvider(address provider) view returns((string,uint256,uint128,uint128,uint256,bool))
func (_ProviderRegistry *ProviderRegistryCaller) GetProvider(opts *bind.CallOpts, provider common.Address) (IProviderStorageProvider, error) {
	var out []interface{}
	err := _ProviderRegistry.contract.Call(opts, &out, "getProvider", provider)

	if err != nil {
		return *new(IProviderStorageProvider), err
	}

	out0 := *abi.ConvertType(out[0], new(IProviderStorageProvider)).(*IProviderStorageProvider)

	return out0, err

}

// GetProvider is a free data retrieval call binding the contract method 0x55f21eb7.
//
// Solidity: function getProvider(address provider) view returns((string,uint256,uint128,uint128,uint256,bool))
func (_ProviderRegistry *ProviderRegistrySession) GetProvider(provider common.Address) (IProviderStorageProvider, error) {
	return _ProviderRegistry.Contract.GetProvider(&_ProviderRegistry.CallOpts, provider)
}

// GetProvider is a free data retrieval call binding the contract method 0x55f21eb7.
//
// Solidity: function getProvider(address provider) view returns((string,uint256,uint128,uint128,uint256,bool))
func (_ProviderRegistry *ProviderRegistryCallerSession) GetProvider(provider common.Address) (IProviderStorageProvider, error) {
	return _ProviderRegistry.Contract.GetProvider(&_ProviderRegistry.CallOpts, provider)
}

// GetToken is a free data retrieval call binding the contract method 0x21df0da7.
//
// Solidity: function getToken() view returns(address)
func (_ProviderRegistry *ProviderRegistryCaller) GetToken(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _ProviderRegistry.contract.Call(opts, &out, "getToken")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// GetToken is a free data retrieval call binding the contract method 0x21df0da7.
//
// Solidity: function getToken() view returns(address)
func (_ProviderRegistry *ProviderRegistrySession) GetToken() (common.Address, error) {
	return _ProviderRegistry.Contract.GetToken(&_ProviderRegistry.CallOpts)
}

// GetToken is a free data retrieval call binding the contract method 0x21df0da7.
//
// Solidity: function getToken() view returns(address)
func (_ProviderRegistry *ProviderRegistryCallerSession) GetToken() (common.Address, error) {
	return _ProviderRegistry.Contract.GetToken(&_ProviderRegistry.CallOpts)
}

// IsProviderExists is a free data retrieval call binding the contract method 0x41876cc4.
//
// Solidity: function isProviderExists(address provider_) view returns(bool)
func (_ProviderRegistry *ProviderRegistryCaller) IsProviderExists(opts *bind.CallOpts, provider_ common.Address) (bool, error) {
	var out []interface{}
	err := _ProviderRegistry.contract.Call(opts, &out, "isProviderExists", provider_)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsProviderExists is a free data retrieval call binding the contract method 0x41876cc4.
//
// Solidity: function isProviderExists(address provider_) view returns(bool)
func (_ProviderRegistry *ProviderRegistrySession) IsProviderExists(provider_ common.Address) (bool, error) {
	return _ProviderRegistry.Contract.IsProviderExists(&_ProviderRegistry.CallOpts, provider_)
}

// IsProviderExists is a free data retrieval call binding the contract method 0x41876cc4.
//
// Solidity: function isProviderExists(address provider_) view returns(bool)
func (_ProviderRegistry *ProviderRegistryCallerSession) IsProviderExists(provider_ common.Address) (bool, error) {
	return _ProviderRegistry.Contract.IsProviderExists(&_ProviderRegistry.CallOpts, provider_)
}

// ModelActiveBids is a free data retrieval call binding the contract method 0x3fd8e5e3.
//
// Solidity: function modelActiveBids(bytes32 modelId_, uint256 offset_, uint256 limit_) view returns(bytes32[])
func (_ProviderRegistry *ProviderRegistryCaller) ModelActiveBids(opts *bind.CallOpts, modelId_ [32]byte, offset_ *big.Int, limit_ *big.Int) ([][32]byte, error) {
	var out []interface{}
	err := _ProviderRegistry.contract.Call(opts, &out, "modelActiveBids", modelId_, offset_, limit_)

	if err != nil {
		return *new([][32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([][32]byte)).(*[][32]byte)

	return out0, err

}

// ModelActiveBids is a free data retrieval call binding the contract method 0x3fd8e5e3.
//
// Solidity: function modelActiveBids(bytes32 modelId_, uint256 offset_, uint256 limit_) view returns(bytes32[])
func (_ProviderRegistry *ProviderRegistrySession) ModelActiveBids(modelId_ [32]byte, offset_ *big.Int, limit_ *big.Int) ([][32]byte, error) {
	return _ProviderRegistry.Contract.ModelActiveBids(&_ProviderRegistry.CallOpts, modelId_, offset_, limit_)
}

// ModelActiveBids is a free data retrieval call binding the contract method 0x3fd8e5e3.
//
// Solidity: function modelActiveBids(bytes32 modelId_, uint256 offset_, uint256 limit_) view returns(bytes32[])
func (_ProviderRegistry *ProviderRegistryCallerSession) ModelActiveBids(modelId_ [32]byte, offset_ *big.Int, limit_ *big.Int) ([][32]byte, error) {
	return _ProviderRegistry.Contract.ModelActiveBids(&_ProviderRegistry.CallOpts, modelId_, offset_, limit_)
}

// ModelBids is a free data retrieval call binding the contract method 0x5954d1b3.
//
// Solidity: function modelBids(bytes32 modelId_, uint256 offset_, uint256 limit_) view returns(bytes32[])
func (_ProviderRegistry *ProviderRegistryCaller) ModelBids(opts *bind.CallOpts, modelId_ [32]byte, offset_ *big.Int, limit_ *big.Int) ([][32]byte, error) {
	var out []interface{}
	err := _ProviderRegistry.contract.Call(opts, &out, "modelBids", modelId_, offset_, limit_)

	if err != nil {
		return *new([][32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([][32]byte)).(*[][32]byte)

	return out0, err

}

// ModelBids is a free data retrieval call binding the contract method 0x5954d1b3.
//
// Solidity: function modelBids(bytes32 modelId_, uint256 offset_, uint256 limit_) view returns(bytes32[])
func (_ProviderRegistry *ProviderRegistrySession) ModelBids(modelId_ [32]byte, offset_ *big.Int, limit_ *big.Int) ([][32]byte, error) {
	return _ProviderRegistry.Contract.ModelBids(&_ProviderRegistry.CallOpts, modelId_, offset_, limit_)
}

// ModelBids is a free data retrieval call binding the contract method 0x5954d1b3.
//
// Solidity: function modelBids(bytes32 modelId_, uint256 offset_, uint256 limit_) view returns(bytes32[])
func (_ProviderRegistry *ProviderRegistryCallerSession) ModelBids(modelId_ [32]byte, offset_ *big.Int, limit_ *big.Int) ([][32]byte, error) {
	return _ProviderRegistry.Contract.ModelBids(&_ProviderRegistry.CallOpts, modelId_, offset_, limit_)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_ProviderRegistry *ProviderRegistryCaller) Owner(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _ProviderRegistry.contract.Call(opts, &out, "owner")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_ProviderRegistry *ProviderRegistrySession) Owner() (common.Address, error) {
	return _ProviderRegistry.Contract.Owner(&_ProviderRegistry.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_ProviderRegistry *ProviderRegistryCallerSession) Owner() (common.Address, error) {
	return _ProviderRegistry.Contract.Owner(&_ProviderRegistry.CallOpts)
}

// ProviderActiveBids is a free data retrieval call binding the contract method 0x6dd7d31c.
//
// Solidity: function providerActiveBids(address provider_, uint256 offset_, uint256 limit_) view returns(bytes32[])
func (_ProviderRegistry *ProviderRegistryCaller) ProviderActiveBids(opts *bind.CallOpts, provider_ common.Address, offset_ *big.Int, limit_ *big.Int) ([][32]byte, error) {
	var out []interface{}
	err := _ProviderRegistry.contract.Call(opts, &out, "providerActiveBids", provider_, offset_, limit_)

	if err != nil {
		return *new([][32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([][32]byte)).(*[][32]byte)

	return out0, err

}

// ProviderActiveBids is a free data retrieval call binding the contract method 0x6dd7d31c.
//
// Solidity: function providerActiveBids(address provider_, uint256 offset_, uint256 limit_) view returns(bytes32[])
func (_ProviderRegistry *ProviderRegistrySession) ProviderActiveBids(provider_ common.Address, offset_ *big.Int, limit_ *big.Int) ([][32]byte, error) {
	return _ProviderRegistry.Contract.ProviderActiveBids(&_ProviderRegistry.CallOpts, provider_, offset_, limit_)
}

// ProviderActiveBids is a free data retrieval call binding the contract method 0x6dd7d31c.
//
// Solidity: function providerActiveBids(address provider_, uint256 offset_, uint256 limit_) view returns(bytes32[])
func (_ProviderRegistry *ProviderRegistryCallerSession) ProviderActiveBids(provider_ common.Address, offset_ *big.Int, limit_ *big.Int) ([][32]byte, error) {
	return _ProviderRegistry.Contract.ProviderActiveBids(&_ProviderRegistry.CallOpts, provider_, offset_, limit_)
}

// ProviderBids is a free data retrieval call binding the contract method 0x22fbda9a.
//
// Solidity: function providerBids(address provider_, uint256 offset_, uint256 limit_) view returns(bytes32[])
func (_ProviderRegistry *ProviderRegistryCaller) ProviderBids(opts *bind.CallOpts, provider_ common.Address, offset_ *big.Int, limit_ *big.Int) ([][32]byte, error) {
	var out []interface{}
	err := _ProviderRegistry.contract.Call(opts, &out, "providerBids", provider_, offset_, limit_)

	if err != nil {
		return *new([][32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([][32]byte)).(*[][32]byte)

	return out0, err

}

// ProviderBids is a free data retrieval call binding the contract method 0x22fbda9a.
//
// Solidity: function providerBids(address provider_, uint256 offset_, uint256 limit_) view returns(bytes32[])
func (_ProviderRegistry *ProviderRegistrySession) ProviderBids(provider_ common.Address, offset_ *big.Int, limit_ *big.Int) ([][32]byte, error) {
	return _ProviderRegistry.Contract.ProviderBids(&_ProviderRegistry.CallOpts, provider_, offset_, limit_)
}

// ProviderBids is a free data retrieval call binding the contract method 0x22fbda9a.
//
// Solidity: function providerBids(address provider_, uint256 offset_, uint256 limit_) view returns(bytes32[])
func (_ProviderRegistry *ProviderRegistryCallerSession) ProviderBids(provider_ common.Address, offset_ *big.Int, limit_ *big.Int) ([][32]byte, error) {
	return _ProviderRegistry.Contract.ProviderBids(&_ProviderRegistry.CallOpts, provider_, offset_, limit_)
}

// ProviderMinimumStake is a free data retrieval call binding the contract method 0x9476c58e.
//
// Solidity: function providerMinimumStake() view returns(uint256)
func (_ProviderRegistry *ProviderRegistryCaller) ProviderMinimumStake(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _ProviderRegistry.contract.Call(opts, &out, "providerMinimumStake")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// ProviderMinimumStake is a free data retrieval call binding the contract method 0x9476c58e.
//
// Solidity: function providerMinimumStake() view returns(uint256)
func (_ProviderRegistry *ProviderRegistrySession) ProviderMinimumStake() (*big.Int, error) {
	return _ProviderRegistry.Contract.ProviderMinimumStake(&_ProviderRegistry.CallOpts)
}

// ProviderMinimumStake is a free data retrieval call binding the contract method 0x9476c58e.
//
// Solidity: function providerMinimumStake() view returns(uint256)
func (_ProviderRegistry *ProviderRegistryCallerSession) ProviderMinimumStake() (*big.Int, error) {
	return _ProviderRegistry.Contract.ProviderMinimumStake(&_ProviderRegistry.CallOpts)
}

// ProviderRegistryInit is a paid mutator transaction binding the contract method 0x5c7ce38b.
//
// Solidity: function __ProviderRegistry_init() returns()
func (_ProviderRegistry *ProviderRegistryTransactor) ProviderRegistryInit(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _ProviderRegistry.contract.Transact(opts, "__ProviderRegistry_init")
}

// ProviderRegistryInit is a paid mutator transaction binding the contract method 0x5c7ce38b.
//
// Solidity: function __ProviderRegistry_init() returns()
func (_ProviderRegistry *ProviderRegistrySession) ProviderRegistryInit() (*types.Transaction, error) {
	return _ProviderRegistry.Contract.ProviderRegistryInit(&_ProviderRegistry.TransactOpts)
}

// ProviderRegistryInit is a paid mutator transaction binding the contract method 0x5c7ce38b.
//
// Solidity: function __ProviderRegistry_init() returns()
func (_ProviderRegistry *ProviderRegistryTransactorSession) ProviderRegistryInit() (*types.Transaction, error) {
	return _ProviderRegistry.Contract.ProviderRegistryInit(&_ProviderRegistry.TransactOpts)
}

// ProviderDeregister is a paid mutator transaction binding the contract method 0x2ca36c49.
//
// Solidity: function providerDeregister(address provider_) returns()
func (_ProviderRegistry *ProviderRegistryTransactor) ProviderDeregister(opts *bind.TransactOpts, provider_ common.Address) (*types.Transaction, error) {
	return _ProviderRegistry.contract.Transact(opts, "providerDeregister", provider_)
}

// ProviderDeregister is a paid mutator transaction binding the contract method 0x2ca36c49.
//
// Solidity: function providerDeregister(address provider_) returns()
func (_ProviderRegistry *ProviderRegistrySession) ProviderDeregister(provider_ common.Address) (*types.Transaction, error) {
	return _ProviderRegistry.Contract.ProviderDeregister(&_ProviderRegistry.TransactOpts, provider_)
}

// ProviderDeregister is a paid mutator transaction binding the contract method 0x2ca36c49.
//
// Solidity: function providerDeregister(address provider_) returns()
func (_ProviderRegistry *ProviderRegistryTransactorSession) ProviderDeregister(provider_ common.Address) (*types.Transaction, error) {
	return _ProviderRegistry.Contract.ProviderDeregister(&_ProviderRegistry.TransactOpts, provider_)
}

// ProviderRegister is a paid mutator transaction binding the contract method 0x365700cb.
//
// Solidity: function providerRegister(address providerAddress_, uint256 amount_, string endpoint_) returns()
func (_ProviderRegistry *ProviderRegistryTransactor) ProviderRegister(opts *bind.TransactOpts, providerAddress_ common.Address, amount_ *big.Int, endpoint_ string) (*types.Transaction, error) {
	return _ProviderRegistry.contract.Transact(opts, "providerRegister", providerAddress_, amount_, endpoint_)
}

// ProviderRegister is a paid mutator transaction binding the contract method 0x365700cb.
//
// Solidity: function providerRegister(address providerAddress_, uint256 amount_, string endpoint_) returns()
func (_ProviderRegistry *ProviderRegistrySession) ProviderRegister(providerAddress_ common.Address, amount_ *big.Int, endpoint_ string) (*types.Transaction, error) {
	return _ProviderRegistry.Contract.ProviderRegister(&_ProviderRegistry.TransactOpts, providerAddress_, amount_, endpoint_)
}

// ProviderRegister is a paid mutator transaction binding the contract method 0x365700cb.
//
// Solidity: function providerRegister(address providerAddress_, uint256 amount_, string endpoint_) returns()
func (_ProviderRegistry *ProviderRegistryTransactorSession) ProviderRegister(providerAddress_ common.Address, amount_ *big.Int, endpoint_ string) (*types.Transaction, error) {
	return _ProviderRegistry.Contract.ProviderRegister(&_ProviderRegistry.TransactOpts, providerAddress_, amount_, endpoint_)
}

// ProviderSetMinStake is a paid mutator transaction binding the contract method 0x0b7f94d6.
//
// Solidity: function providerSetMinStake(uint256 providerMinimumStake_) returns()
func (_ProviderRegistry *ProviderRegistryTransactor) ProviderSetMinStake(opts *bind.TransactOpts, providerMinimumStake_ *big.Int) (*types.Transaction, error) {
	return _ProviderRegistry.contract.Transact(opts, "providerSetMinStake", providerMinimumStake_)
}

// ProviderSetMinStake is a paid mutator transaction binding the contract method 0x0b7f94d6.
//
// Solidity: function providerSetMinStake(uint256 providerMinimumStake_) returns()
func (_ProviderRegistry *ProviderRegistrySession) ProviderSetMinStake(providerMinimumStake_ *big.Int) (*types.Transaction, error) {
	return _ProviderRegistry.Contract.ProviderSetMinStake(&_ProviderRegistry.TransactOpts, providerMinimumStake_)
}

// ProviderSetMinStake is a paid mutator transaction binding the contract method 0x0b7f94d6.
//
// Solidity: function providerSetMinStake(uint256 providerMinimumStake_) returns()
func (_ProviderRegistry *ProviderRegistryTransactorSession) ProviderSetMinStake(providerMinimumStake_ *big.Int) (*types.Transaction, error) {
	return _ProviderRegistry.Contract.ProviderSetMinStake(&_ProviderRegistry.TransactOpts, providerMinimumStake_)
}

// ProviderWithdrawStake is a paid mutator transaction binding the contract method 0x8209d9ed.
//
// Solidity: function providerWithdrawStake(address provider_) returns()
func (_ProviderRegistry *ProviderRegistryTransactor) ProviderWithdrawStake(opts *bind.TransactOpts, provider_ common.Address) (*types.Transaction, error) {
	return _ProviderRegistry.contract.Transact(opts, "providerWithdrawStake", provider_)
}

// ProviderWithdrawStake is a paid mutator transaction binding the contract method 0x8209d9ed.
//
// Solidity: function providerWithdrawStake(address provider_) returns()
func (_ProviderRegistry *ProviderRegistrySession) ProviderWithdrawStake(provider_ common.Address) (*types.Transaction, error) {
	return _ProviderRegistry.Contract.ProviderWithdrawStake(&_ProviderRegistry.TransactOpts, provider_)
}

// ProviderWithdrawStake is a paid mutator transaction binding the contract method 0x8209d9ed.
//
// Solidity: function providerWithdrawStake(address provider_) returns()
func (_ProviderRegistry *ProviderRegistryTransactorSession) ProviderWithdrawStake(provider_ common.Address) (*types.Transaction, error) {
	return _ProviderRegistry.Contract.ProviderWithdrawStake(&_ProviderRegistry.TransactOpts, provider_)
}

// ProviderRegistryInitializedIterator is returned from FilterInitialized and is used to iterate over the raw logs and unpacked data for Initialized events raised by the ProviderRegistry contract.
type ProviderRegistryInitializedIterator struct {
	Event *ProviderRegistryInitialized // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *ProviderRegistryInitializedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ProviderRegistryInitialized)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(ProviderRegistryInitialized)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *ProviderRegistryInitializedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ProviderRegistryInitializedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ProviderRegistryInitialized represents a Initialized event raised by the ProviderRegistry contract.
type ProviderRegistryInitialized struct {
	StorageSlot [32]byte
	Raw         types.Log // Blockchain specific contextual infos
}

// FilterInitialized is a free log retrieval operation binding the contract event 0xdc73717d728bcfa015e8117438a65319aa06e979ca324afa6e1ea645c28ea15d.
//
// Solidity: event Initialized(bytes32 storageSlot)
func (_ProviderRegistry *ProviderRegistryFilterer) FilterInitialized(opts *bind.FilterOpts) (*ProviderRegistryInitializedIterator, error) {

	logs, sub, err := _ProviderRegistry.contract.FilterLogs(opts, "Initialized")
	if err != nil {
		return nil, err
	}
	return &ProviderRegistryInitializedIterator{contract: _ProviderRegistry.contract, event: "Initialized", logs: logs, sub: sub}, nil
}

// WatchInitialized is a free log subscription operation binding the contract event 0xdc73717d728bcfa015e8117438a65319aa06e979ca324afa6e1ea645c28ea15d.
//
// Solidity: event Initialized(bytes32 storageSlot)
func (_ProviderRegistry *ProviderRegistryFilterer) WatchInitialized(opts *bind.WatchOpts, sink chan<- *ProviderRegistryInitialized) (event.Subscription, error) {

	logs, sub, err := _ProviderRegistry.contract.WatchLogs(opts, "Initialized")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ProviderRegistryInitialized)
				if err := _ProviderRegistry.contract.UnpackLog(event, "Initialized", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseInitialized is a log parse operation binding the contract event 0xdc73717d728bcfa015e8117438a65319aa06e979ca324afa6e1ea645c28ea15d.
//
// Solidity: event Initialized(bytes32 storageSlot)
func (_ProviderRegistry *ProviderRegistryFilterer) ParseInitialized(log types.Log) (*ProviderRegistryInitialized, error) {
	event := new(ProviderRegistryInitialized)
	if err := _ProviderRegistry.contract.UnpackLog(event, "Initialized", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ProviderRegistryProviderDeregisteredIterator is returned from FilterProviderDeregistered and is used to iterate over the raw logs and unpacked data for ProviderDeregistered events raised by the ProviderRegistry contract.
type ProviderRegistryProviderDeregisteredIterator struct {
	Event *ProviderRegistryProviderDeregistered // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *ProviderRegistryProviderDeregisteredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ProviderRegistryProviderDeregistered)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(ProviderRegistryProviderDeregistered)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *ProviderRegistryProviderDeregisteredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ProviderRegistryProviderDeregisteredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ProviderRegistryProviderDeregistered represents a ProviderDeregistered event raised by the ProviderRegistry contract.
type ProviderRegistryProviderDeregistered struct {
	Provider common.Address
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterProviderDeregistered is a free log retrieval operation binding the contract event 0xf04091b4a187e321a42001e46961e45b6a75b203fc6fb766b7e05505f6080abb.
//
// Solidity: event ProviderDeregistered(address indexed provider)
func (_ProviderRegistry *ProviderRegistryFilterer) FilterProviderDeregistered(opts *bind.FilterOpts, provider []common.Address) (*ProviderRegistryProviderDeregisteredIterator, error) {

	var providerRule []interface{}
	for _, providerItem := range provider {
		providerRule = append(providerRule, providerItem)
	}

	logs, sub, err := _ProviderRegistry.contract.FilterLogs(opts, "ProviderDeregistered", providerRule)
	if err != nil {
		return nil, err
	}
	return &ProviderRegistryProviderDeregisteredIterator{contract: _ProviderRegistry.contract, event: "ProviderDeregistered", logs: logs, sub: sub}, nil
}

// WatchProviderDeregistered is a free log subscription operation binding the contract event 0xf04091b4a187e321a42001e46961e45b6a75b203fc6fb766b7e05505f6080abb.
//
// Solidity: event ProviderDeregistered(address indexed provider)
func (_ProviderRegistry *ProviderRegistryFilterer) WatchProviderDeregistered(opts *bind.WatchOpts, sink chan<- *ProviderRegistryProviderDeregistered, provider []common.Address) (event.Subscription, error) {

	var providerRule []interface{}
	for _, providerItem := range provider {
		providerRule = append(providerRule, providerItem)
	}

	logs, sub, err := _ProviderRegistry.contract.WatchLogs(opts, "ProviderDeregistered", providerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ProviderRegistryProviderDeregistered)
				if err := _ProviderRegistry.contract.UnpackLog(event, "ProviderDeregistered", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseProviderDeregistered is a log parse operation binding the contract event 0xf04091b4a187e321a42001e46961e45b6a75b203fc6fb766b7e05505f6080abb.
//
// Solidity: event ProviderDeregistered(address indexed provider)
func (_ProviderRegistry *ProviderRegistryFilterer) ParseProviderDeregistered(log types.Log) (*ProviderRegistryProviderDeregistered, error) {
	event := new(ProviderRegistryProviderDeregistered)
	if err := _ProviderRegistry.contract.UnpackLog(event, "ProviderDeregistered", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ProviderRegistryProviderMinStakeUpdatedIterator is returned from FilterProviderMinStakeUpdated and is used to iterate over the raw logs and unpacked data for ProviderMinStakeUpdated events raised by the ProviderRegistry contract.
type ProviderRegistryProviderMinStakeUpdatedIterator struct {
	Event *ProviderRegistryProviderMinStakeUpdated // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *ProviderRegistryProviderMinStakeUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ProviderRegistryProviderMinStakeUpdated)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(ProviderRegistryProviderMinStakeUpdated)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *ProviderRegistryProviderMinStakeUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ProviderRegistryProviderMinStakeUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ProviderRegistryProviderMinStakeUpdated represents a ProviderMinStakeUpdated event raised by the ProviderRegistry contract.
type ProviderRegistryProviderMinStakeUpdated struct {
	NewStake *big.Int
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterProviderMinStakeUpdated is a free log retrieval operation binding the contract event 0x1ee852018221ad3e0f9b96b4d6f870d0e1393c4060f626c01ad1e09a1917d818.
//
// Solidity: event ProviderMinStakeUpdated(uint256 newStake)
func (_ProviderRegistry *ProviderRegistryFilterer) FilterProviderMinStakeUpdated(opts *bind.FilterOpts) (*ProviderRegistryProviderMinStakeUpdatedIterator, error) {

	logs, sub, err := _ProviderRegistry.contract.FilterLogs(opts, "ProviderMinStakeUpdated")
	if err != nil {
		return nil, err
	}
	return &ProviderRegistryProviderMinStakeUpdatedIterator{contract: _ProviderRegistry.contract, event: "ProviderMinStakeUpdated", logs: logs, sub: sub}, nil
}

// WatchProviderMinStakeUpdated is a free log subscription operation binding the contract event 0x1ee852018221ad3e0f9b96b4d6f870d0e1393c4060f626c01ad1e09a1917d818.
//
// Solidity: event ProviderMinStakeUpdated(uint256 newStake)
func (_ProviderRegistry *ProviderRegistryFilterer) WatchProviderMinStakeUpdated(opts *bind.WatchOpts, sink chan<- *ProviderRegistryProviderMinStakeUpdated) (event.Subscription, error) {

	logs, sub, err := _ProviderRegistry.contract.WatchLogs(opts, "ProviderMinStakeUpdated")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ProviderRegistryProviderMinStakeUpdated)
				if err := _ProviderRegistry.contract.UnpackLog(event, "ProviderMinStakeUpdated", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseProviderMinStakeUpdated is a log parse operation binding the contract event 0x1ee852018221ad3e0f9b96b4d6f870d0e1393c4060f626c01ad1e09a1917d818.
//
// Solidity: event ProviderMinStakeUpdated(uint256 newStake)
func (_ProviderRegistry *ProviderRegistryFilterer) ParseProviderMinStakeUpdated(log types.Log) (*ProviderRegistryProviderMinStakeUpdated, error) {
	event := new(ProviderRegistryProviderMinStakeUpdated)
	if err := _ProviderRegistry.contract.UnpackLog(event, "ProviderMinStakeUpdated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ProviderRegistryProviderRegisteredUpdatedIterator is returned from FilterProviderRegisteredUpdated and is used to iterate over the raw logs and unpacked data for ProviderRegisteredUpdated events raised by the ProviderRegistry contract.
type ProviderRegistryProviderRegisteredUpdatedIterator struct {
	Event *ProviderRegistryProviderRegisteredUpdated // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *ProviderRegistryProviderRegisteredUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ProviderRegistryProviderRegisteredUpdated)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(ProviderRegistryProviderRegisteredUpdated)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *ProviderRegistryProviderRegisteredUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ProviderRegistryProviderRegisteredUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ProviderRegistryProviderRegisteredUpdated represents a ProviderRegisteredUpdated event raised by the ProviderRegistry contract.
type ProviderRegistryProviderRegisteredUpdated struct {
	Provider common.Address
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterProviderRegisteredUpdated is a free log retrieval operation binding the contract event 0xe041bebe929cc665c6c558e3ad7913156fef3abc77ac6a2a4f0182e6dcb11193.
//
// Solidity: event ProviderRegisteredUpdated(address indexed provider)
func (_ProviderRegistry *ProviderRegistryFilterer) FilterProviderRegisteredUpdated(opts *bind.FilterOpts, provider []common.Address) (*ProviderRegistryProviderRegisteredUpdatedIterator, error) {

	var providerRule []interface{}
	for _, providerItem := range provider {
		providerRule = append(providerRule, providerItem)
	}

	logs, sub, err := _ProviderRegistry.contract.FilterLogs(opts, "ProviderRegisteredUpdated", providerRule)
	if err != nil {
		return nil, err
	}
	return &ProviderRegistryProviderRegisteredUpdatedIterator{contract: _ProviderRegistry.contract, event: "ProviderRegisteredUpdated", logs: logs, sub: sub}, nil
}

// WatchProviderRegisteredUpdated is a free log subscription operation binding the contract event 0xe041bebe929cc665c6c558e3ad7913156fef3abc77ac6a2a4f0182e6dcb11193.
//
// Solidity: event ProviderRegisteredUpdated(address indexed provider)
func (_ProviderRegistry *ProviderRegistryFilterer) WatchProviderRegisteredUpdated(opts *bind.WatchOpts, sink chan<- *ProviderRegistryProviderRegisteredUpdated, provider []common.Address) (event.Subscription, error) {

	var providerRule []interface{}
	for _, providerItem := range provider {
		providerRule = append(providerRule, providerItem)
	}

	logs, sub, err := _ProviderRegistry.contract.WatchLogs(opts, "ProviderRegisteredUpdated", providerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ProviderRegistryProviderRegisteredUpdated)
				if err := _ProviderRegistry.contract.UnpackLog(event, "ProviderRegisteredUpdated", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseProviderRegisteredUpdated is a log parse operation binding the contract event 0xe041bebe929cc665c6c558e3ad7913156fef3abc77ac6a2a4f0182e6dcb11193.
//
// Solidity: event ProviderRegisteredUpdated(address indexed provider)
func (_ProviderRegistry *ProviderRegistryFilterer) ParseProviderRegisteredUpdated(log types.Log) (*ProviderRegistryProviderRegisteredUpdated, error) {
	event := new(ProviderRegistryProviderRegisteredUpdated)
	if err := _ProviderRegistry.contract.UnpackLog(event, "ProviderRegisteredUpdated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ProviderRegistryProviderWithdrawnStakeIterator is returned from FilterProviderWithdrawnStake and is used to iterate over the raw logs and unpacked data for ProviderWithdrawnStake events raised by the ProviderRegistry contract.
type ProviderRegistryProviderWithdrawnStakeIterator struct {
	Event *ProviderRegistryProviderWithdrawnStake // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *ProviderRegistryProviderWithdrawnStakeIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ProviderRegistryProviderWithdrawnStake)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(ProviderRegistryProviderWithdrawnStake)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *ProviderRegistryProviderWithdrawnStakeIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ProviderRegistryProviderWithdrawnStakeIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ProviderRegistryProviderWithdrawnStake represents a ProviderWithdrawnStake event raised by the ProviderRegistry contract.
type ProviderRegistryProviderWithdrawnStake struct {
	Provider common.Address
	Amount   *big.Int
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterProviderWithdrawnStake is a free log retrieval operation binding the contract event 0x51bcac309509d42be030eac333c3d18a8a05c7b400560ce45b122bba5877c76d.
//
// Solidity: event ProviderWithdrawnStake(address indexed provider, uint256 amount)
func (_ProviderRegistry *ProviderRegistryFilterer) FilterProviderWithdrawnStake(opts *bind.FilterOpts, provider []common.Address) (*ProviderRegistryProviderWithdrawnStakeIterator, error) {

	var providerRule []interface{}
	for _, providerItem := range provider {
		providerRule = append(providerRule, providerItem)
	}

	logs, sub, err := _ProviderRegistry.contract.FilterLogs(opts, "ProviderWithdrawnStake", providerRule)
	if err != nil {
		return nil, err
	}
	return &ProviderRegistryProviderWithdrawnStakeIterator{contract: _ProviderRegistry.contract, event: "ProviderWithdrawnStake", logs: logs, sub: sub}, nil
}

// WatchProviderWithdrawnStake is a free log subscription operation binding the contract event 0x51bcac309509d42be030eac333c3d18a8a05c7b400560ce45b122bba5877c76d.
//
// Solidity: event ProviderWithdrawnStake(address indexed provider, uint256 amount)
func (_ProviderRegistry *ProviderRegistryFilterer) WatchProviderWithdrawnStake(opts *bind.WatchOpts, sink chan<- *ProviderRegistryProviderWithdrawnStake, provider []common.Address) (event.Subscription, error) {

	var providerRule []interface{}
	for _, providerItem := range provider {
		providerRule = append(providerRule, providerItem)
	}

	logs, sub, err := _ProviderRegistry.contract.WatchLogs(opts, "ProviderWithdrawnStake", providerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ProviderRegistryProviderWithdrawnStake)
				if err := _ProviderRegistry.contract.UnpackLog(event, "ProviderWithdrawnStake", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseProviderWithdrawnStake is a log parse operation binding the contract event 0x51bcac309509d42be030eac333c3d18a8a05c7b400560ce45b122bba5877c76d.
//
// Solidity: event ProviderWithdrawnStake(address indexed provider, uint256 amount)
func (_ProviderRegistry *ProviderRegistryFilterer) ParseProviderWithdrawnStake(log types.Log) (*ProviderRegistryProviderWithdrawnStake, error) {
	event := new(ProviderRegistryProviderWithdrawnStake)
	if err := _ProviderRegistry.contract.UnpackLog(event, "ProviderWithdrawnStake", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
