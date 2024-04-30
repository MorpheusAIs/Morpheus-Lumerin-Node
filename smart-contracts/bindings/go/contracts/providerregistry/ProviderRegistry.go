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

// ProviderRegistryProvider is an auto generated low-level Go binding around an user-defined struct.
type ProviderRegistryProvider struct {
	Endpoint  string
	Stake     *big.Int
	Timestamp *big.Int
	IsDeleted bool
}

// ProviderRegistryMetaData contains all meta data concerning the ProviderRegistry contract.
var ProviderRegistryMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[],\"name\":\"KeyExists\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"KeyNotFound\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"NotSenderOrOwner\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"StakeTooLow\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"ZeroKey\",\"type\":\"error\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"provider\",\"type\":\"address\"}],\"name\":\"Deregistered\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint8\",\"name\":\"version\",\"type\":\"uint8\"}],\"name\":\"Initialized\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"newStake\",\"type\":\"uint256\"}],\"name\":\"MinStakeUpdated\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"previousOwner\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"OwnershipTransferred\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"provider\",\"type\":\"address\"}],\"name\":\"RegisteredUpdated\",\"type\":\"event\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"addr\",\"type\":\"address\"}],\"name\":\"deregister\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"addr\",\"type\":\"address\"}],\"name\":\"exists\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getAll\",\"outputs\":[{\"internalType\":\"address[]\",\"name\":\"\",\"type\":\"address[]\"},{\"components\":[{\"internalType\":\"string\",\"name\":\"endpoint\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"stake\",\"type\":\"uint256\"},{\"internalType\":\"uint128\",\"name\":\"timestamp\",\"type\":\"uint128\"},{\"internalType\":\"bool\",\"name\":\"isDeleted\",\"type\":\"bool\"}],\"internalType\":\"structProviderRegistry.Provider[]\",\"name\":\"\",\"type\":\"tuple[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"index\",\"type\":\"uint256\"}],\"name\":\"getByIndex\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"addr\",\"type\":\"address\"},{\"components\":[{\"internalType\":\"string\",\"name\":\"endpoint\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"stake\",\"type\":\"uint256\"},{\"internalType\":\"uint128\",\"name\":\"timestamp\",\"type\":\"uint128\"},{\"internalType\":\"bool\",\"name\":\"isDeleted\",\"type\":\"bool\"}],\"internalType\":\"structProviderRegistry.Provider\",\"name\":\"provider\",\"type\":\"tuple\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getCount\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"count\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getIds\",\"outputs\":[{\"internalType\":\"address[]\",\"name\":\"\",\"type\":\"address[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_token\",\"type\":\"address\"}],\"name\":\"initialize\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"name\":\"map\",\"outputs\":[{\"internalType\":\"string\",\"name\":\"endpoint\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"stake\",\"type\":\"uint256\"},{\"internalType\":\"uint128\",\"name\":\"timestamp\",\"type\":\"uint128\"},{\"internalType\":\"bool\",\"name\":\"isDeleted\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"minStake\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"owner\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"providers\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"addr\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"addStake\",\"type\":\"uint256\"},{\"internalType\":\"string\",\"name\":\"endpoint\",\"type\":\"string\"}],\"name\":\"register\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"renounceOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"_minStake\",\"type\":\"uint256\"}],\"name\":\"setMinStake\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"token\",\"outputs\":[{\"internalType\":\"contractERC20\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"transferOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]",
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

// Exists is a free data retrieval call binding the contract method 0xf6a3d24e.
//
// Solidity: function exists(address addr) view returns(bool)
func (_ProviderRegistry *ProviderRegistryCaller) Exists(opts *bind.CallOpts, addr common.Address) (bool, error) {
	var out []interface{}
	err := _ProviderRegistry.contract.Call(opts, &out, "exists", addr)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// Exists is a free data retrieval call binding the contract method 0xf6a3d24e.
//
// Solidity: function exists(address addr) view returns(bool)
func (_ProviderRegistry *ProviderRegistrySession) Exists(addr common.Address) (bool, error) {
	return _ProviderRegistry.Contract.Exists(&_ProviderRegistry.CallOpts, addr)
}

// Exists is a free data retrieval call binding the contract method 0xf6a3d24e.
//
// Solidity: function exists(address addr) view returns(bool)
func (_ProviderRegistry *ProviderRegistryCallerSession) Exists(addr common.Address) (bool, error) {
	return _ProviderRegistry.Contract.Exists(&_ProviderRegistry.CallOpts, addr)
}

// GetAll is a free data retrieval call binding the contract method 0x53ed5143.
//
// Solidity: function getAll() view returns(address[], (string,uint256,uint128,bool)[])
func (_ProviderRegistry *ProviderRegistryCaller) GetAll(opts *bind.CallOpts) ([]common.Address, []ProviderRegistryProvider, error) {
	var out []interface{}
	err := _ProviderRegistry.contract.Call(opts, &out, "getAll")

	if err != nil {
		return *new([]common.Address), *new([]ProviderRegistryProvider), err
	}

	out0 := *abi.ConvertType(out[0], new([]common.Address)).(*[]common.Address)
	out1 := *abi.ConvertType(out[1], new([]ProviderRegistryProvider)).(*[]ProviderRegistryProvider)

	return out0, out1, err

}

// GetAll is a free data retrieval call binding the contract method 0x53ed5143.
//
// Solidity: function getAll() view returns(address[], (string,uint256,uint128,bool)[])
func (_ProviderRegistry *ProviderRegistrySession) GetAll() ([]common.Address, []ProviderRegistryProvider, error) {
	return _ProviderRegistry.Contract.GetAll(&_ProviderRegistry.CallOpts)
}

// GetAll is a free data retrieval call binding the contract method 0x53ed5143.
//
// Solidity: function getAll() view returns(address[], (string,uint256,uint128,bool)[])
func (_ProviderRegistry *ProviderRegistryCallerSession) GetAll() ([]common.Address, []ProviderRegistryProvider, error) {
	return _ProviderRegistry.Contract.GetAll(&_ProviderRegistry.CallOpts)
}

// GetByIndex is a free data retrieval call binding the contract method 0x2d883a73.
//
// Solidity: function getByIndex(uint256 index) view returns(address addr, (string,uint256,uint128,bool) provider)
func (_ProviderRegistry *ProviderRegistryCaller) GetByIndex(opts *bind.CallOpts, index *big.Int) (struct {
	Addr     common.Address
	Provider ProviderRegistryProvider
}, error) {
	var out []interface{}
	err := _ProviderRegistry.contract.Call(opts, &out, "getByIndex", index)

	outstruct := new(struct {
		Addr     common.Address
		Provider ProviderRegistryProvider
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.Addr = *abi.ConvertType(out[0], new(common.Address)).(*common.Address)
	outstruct.Provider = *abi.ConvertType(out[1], new(ProviderRegistryProvider)).(*ProviderRegistryProvider)

	return *outstruct, err

}

// GetByIndex is a free data retrieval call binding the contract method 0x2d883a73.
//
// Solidity: function getByIndex(uint256 index) view returns(address addr, (string,uint256,uint128,bool) provider)
func (_ProviderRegistry *ProviderRegistrySession) GetByIndex(index *big.Int) (struct {
	Addr     common.Address
	Provider ProviderRegistryProvider
}, error) {
	return _ProviderRegistry.Contract.GetByIndex(&_ProviderRegistry.CallOpts, index)
}

// GetByIndex is a free data retrieval call binding the contract method 0x2d883a73.
//
// Solidity: function getByIndex(uint256 index) view returns(address addr, (string,uint256,uint128,bool) provider)
func (_ProviderRegistry *ProviderRegistryCallerSession) GetByIndex(index *big.Int) (struct {
	Addr     common.Address
	Provider ProviderRegistryProvider
}, error) {
	return _ProviderRegistry.Contract.GetByIndex(&_ProviderRegistry.CallOpts, index)
}

// GetCount is a free data retrieval call binding the contract method 0xa87d942c.
//
// Solidity: function getCount() view returns(uint256 count)
func (_ProviderRegistry *ProviderRegistryCaller) GetCount(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _ProviderRegistry.contract.Call(opts, &out, "getCount")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetCount is a free data retrieval call binding the contract method 0xa87d942c.
//
// Solidity: function getCount() view returns(uint256 count)
func (_ProviderRegistry *ProviderRegistrySession) GetCount() (*big.Int, error) {
	return _ProviderRegistry.Contract.GetCount(&_ProviderRegistry.CallOpts)
}

// GetCount is a free data retrieval call binding the contract method 0xa87d942c.
//
// Solidity: function getCount() view returns(uint256 count)
func (_ProviderRegistry *ProviderRegistryCallerSession) GetCount() (*big.Int, error) {
	return _ProviderRegistry.Contract.GetCount(&_ProviderRegistry.CallOpts)
}

// GetIds is a free data retrieval call binding the contract method 0x2b105663.
//
// Solidity: function getIds() view returns(address[])
func (_ProviderRegistry *ProviderRegistryCaller) GetIds(opts *bind.CallOpts) ([]common.Address, error) {
	var out []interface{}
	err := _ProviderRegistry.contract.Call(opts, &out, "getIds")

	if err != nil {
		return *new([]common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new([]common.Address)).(*[]common.Address)

	return out0, err

}

// GetIds is a free data retrieval call binding the contract method 0x2b105663.
//
// Solidity: function getIds() view returns(address[])
func (_ProviderRegistry *ProviderRegistrySession) GetIds() ([]common.Address, error) {
	return _ProviderRegistry.Contract.GetIds(&_ProviderRegistry.CallOpts)
}

// GetIds is a free data retrieval call binding the contract method 0x2b105663.
//
// Solidity: function getIds() view returns(address[])
func (_ProviderRegistry *ProviderRegistryCallerSession) GetIds() ([]common.Address, error) {
	return _ProviderRegistry.Contract.GetIds(&_ProviderRegistry.CallOpts)
}

// Map is a free data retrieval call binding the contract method 0xb721ef6e.
//
// Solidity: function map(address ) view returns(string endpoint, uint256 stake, uint128 timestamp, bool isDeleted)
func (_ProviderRegistry *ProviderRegistryCaller) Map(opts *bind.CallOpts, arg0 common.Address) (struct {
	Endpoint  string
	Stake     *big.Int
	Timestamp *big.Int
	IsDeleted bool
}, error) {
	var out []interface{}
	err := _ProviderRegistry.contract.Call(opts, &out, "map", arg0)

	outstruct := new(struct {
		Endpoint  string
		Stake     *big.Int
		Timestamp *big.Int
		IsDeleted bool
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.Endpoint = *abi.ConvertType(out[0], new(string)).(*string)
	outstruct.Stake = *abi.ConvertType(out[1], new(*big.Int)).(**big.Int)
	outstruct.Timestamp = *abi.ConvertType(out[2], new(*big.Int)).(**big.Int)
	outstruct.IsDeleted = *abi.ConvertType(out[3], new(bool)).(*bool)

	return *outstruct, err

}

// Map is a free data retrieval call binding the contract method 0xb721ef6e.
//
// Solidity: function map(address ) view returns(string endpoint, uint256 stake, uint128 timestamp, bool isDeleted)
func (_ProviderRegistry *ProviderRegistrySession) Map(arg0 common.Address) (struct {
	Endpoint  string
	Stake     *big.Int
	Timestamp *big.Int
	IsDeleted bool
}, error) {
	return _ProviderRegistry.Contract.Map(&_ProviderRegistry.CallOpts, arg0)
}

// Map is a free data retrieval call binding the contract method 0xb721ef6e.
//
// Solidity: function map(address ) view returns(string endpoint, uint256 stake, uint128 timestamp, bool isDeleted)
func (_ProviderRegistry *ProviderRegistryCallerSession) Map(arg0 common.Address) (struct {
	Endpoint  string
	Stake     *big.Int
	Timestamp *big.Int
	IsDeleted bool
}, error) {
	return _ProviderRegistry.Contract.Map(&_ProviderRegistry.CallOpts, arg0)
}

// MinStake is a free data retrieval call binding the contract method 0x375b3c0a.
//
// Solidity: function minStake() view returns(uint256)
func (_ProviderRegistry *ProviderRegistryCaller) MinStake(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _ProviderRegistry.contract.Call(opts, &out, "minStake")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// MinStake is a free data retrieval call binding the contract method 0x375b3c0a.
//
// Solidity: function minStake() view returns(uint256)
func (_ProviderRegistry *ProviderRegistrySession) MinStake() (*big.Int, error) {
	return _ProviderRegistry.Contract.MinStake(&_ProviderRegistry.CallOpts)
}

// MinStake is a free data retrieval call binding the contract method 0x375b3c0a.
//
// Solidity: function minStake() view returns(uint256)
func (_ProviderRegistry *ProviderRegistryCallerSession) MinStake() (*big.Int, error) {
	return _ProviderRegistry.Contract.MinStake(&_ProviderRegistry.CallOpts)
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

// Providers is a free data retrieval call binding the contract method 0x50f3fc81.
//
// Solidity: function providers(uint256 ) view returns(address)
func (_ProviderRegistry *ProviderRegistryCaller) Providers(opts *bind.CallOpts, arg0 *big.Int) (common.Address, error) {
	var out []interface{}
	err := _ProviderRegistry.contract.Call(opts, &out, "providers", arg0)

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Providers is a free data retrieval call binding the contract method 0x50f3fc81.
//
// Solidity: function providers(uint256 ) view returns(address)
func (_ProviderRegistry *ProviderRegistrySession) Providers(arg0 *big.Int) (common.Address, error) {
	return _ProviderRegistry.Contract.Providers(&_ProviderRegistry.CallOpts, arg0)
}

// Providers is a free data retrieval call binding the contract method 0x50f3fc81.
//
// Solidity: function providers(uint256 ) view returns(address)
func (_ProviderRegistry *ProviderRegistryCallerSession) Providers(arg0 *big.Int) (common.Address, error) {
	return _ProviderRegistry.Contract.Providers(&_ProviderRegistry.CallOpts, arg0)
}

// Token is a free data retrieval call binding the contract method 0xfc0c546a.
//
// Solidity: function token() view returns(address)
func (_ProviderRegistry *ProviderRegistryCaller) Token(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _ProviderRegistry.contract.Call(opts, &out, "token")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Token is a free data retrieval call binding the contract method 0xfc0c546a.
//
// Solidity: function token() view returns(address)
func (_ProviderRegistry *ProviderRegistrySession) Token() (common.Address, error) {
	return _ProviderRegistry.Contract.Token(&_ProviderRegistry.CallOpts)
}

// Token is a free data retrieval call binding the contract method 0xfc0c546a.
//
// Solidity: function token() view returns(address)
func (_ProviderRegistry *ProviderRegistryCallerSession) Token() (common.Address, error) {
	return _ProviderRegistry.Contract.Token(&_ProviderRegistry.CallOpts)
}

// Deregister is a paid mutator transaction binding the contract method 0x84ac33ec.
//
// Solidity: function deregister(address addr) returns()
func (_ProviderRegistry *ProviderRegistryTransactor) Deregister(opts *bind.TransactOpts, addr common.Address) (*types.Transaction, error) {
	return _ProviderRegistry.contract.Transact(opts, "deregister", addr)
}

// Deregister is a paid mutator transaction binding the contract method 0x84ac33ec.
//
// Solidity: function deregister(address addr) returns()
func (_ProviderRegistry *ProviderRegistrySession) Deregister(addr common.Address) (*types.Transaction, error) {
	return _ProviderRegistry.Contract.Deregister(&_ProviderRegistry.TransactOpts, addr)
}

// Deregister is a paid mutator transaction binding the contract method 0x84ac33ec.
//
// Solidity: function deregister(address addr) returns()
func (_ProviderRegistry *ProviderRegistryTransactorSession) Deregister(addr common.Address) (*types.Transaction, error) {
	return _ProviderRegistry.Contract.Deregister(&_ProviderRegistry.TransactOpts, addr)
}

// Initialize is a paid mutator transaction binding the contract method 0xc4d66de8.
//
// Solidity: function initialize(address _token) returns()
func (_ProviderRegistry *ProviderRegistryTransactor) Initialize(opts *bind.TransactOpts, _token common.Address) (*types.Transaction, error) {
	return _ProviderRegistry.contract.Transact(opts, "initialize", _token)
}

// Initialize is a paid mutator transaction binding the contract method 0xc4d66de8.
//
// Solidity: function initialize(address _token) returns()
func (_ProviderRegistry *ProviderRegistrySession) Initialize(_token common.Address) (*types.Transaction, error) {
	return _ProviderRegistry.Contract.Initialize(&_ProviderRegistry.TransactOpts, _token)
}

// Initialize is a paid mutator transaction binding the contract method 0xc4d66de8.
//
// Solidity: function initialize(address _token) returns()
func (_ProviderRegistry *ProviderRegistryTransactorSession) Initialize(_token common.Address) (*types.Transaction, error) {
	return _ProviderRegistry.Contract.Initialize(&_ProviderRegistry.TransactOpts, _token)
}

// Register is a paid mutator transaction binding the contract method 0xf11b1b88.
//
// Solidity: function register(address addr, uint256 addStake, string endpoint) returns()
func (_ProviderRegistry *ProviderRegistryTransactor) Register(opts *bind.TransactOpts, addr common.Address, addStake *big.Int, endpoint string) (*types.Transaction, error) {
	return _ProviderRegistry.contract.Transact(opts, "register", addr, addStake, endpoint)
}

// Register is a paid mutator transaction binding the contract method 0xf11b1b88.
//
// Solidity: function register(address addr, uint256 addStake, string endpoint) returns()
func (_ProviderRegistry *ProviderRegistrySession) Register(addr common.Address, addStake *big.Int, endpoint string) (*types.Transaction, error) {
	return _ProviderRegistry.Contract.Register(&_ProviderRegistry.TransactOpts, addr, addStake, endpoint)
}

// Register is a paid mutator transaction binding the contract method 0xf11b1b88.
//
// Solidity: function register(address addr, uint256 addStake, string endpoint) returns()
func (_ProviderRegistry *ProviderRegistryTransactorSession) Register(addr common.Address, addStake *big.Int, endpoint string) (*types.Transaction, error) {
	return _ProviderRegistry.Contract.Register(&_ProviderRegistry.TransactOpts, addr, addStake, endpoint)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_ProviderRegistry *ProviderRegistryTransactor) RenounceOwnership(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _ProviderRegistry.contract.Transact(opts, "renounceOwnership")
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_ProviderRegistry *ProviderRegistrySession) RenounceOwnership() (*types.Transaction, error) {
	return _ProviderRegistry.Contract.RenounceOwnership(&_ProviderRegistry.TransactOpts)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_ProviderRegistry *ProviderRegistryTransactorSession) RenounceOwnership() (*types.Transaction, error) {
	return _ProviderRegistry.Contract.RenounceOwnership(&_ProviderRegistry.TransactOpts)
}

// SetMinStake is a paid mutator transaction binding the contract method 0x8c80fd90.
//
// Solidity: function setMinStake(uint256 _minStake) returns()
func (_ProviderRegistry *ProviderRegistryTransactor) SetMinStake(opts *bind.TransactOpts, _minStake *big.Int) (*types.Transaction, error) {
	return _ProviderRegistry.contract.Transact(opts, "setMinStake", _minStake)
}

// SetMinStake is a paid mutator transaction binding the contract method 0x8c80fd90.
//
// Solidity: function setMinStake(uint256 _minStake) returns()
func (_ProviderRegistry *ProviderRegistrySession) SetMinStake(_minStake *big.Int) (*types.Transaction, error) {
	return _ProviderRegistry.Contract.SetMinStake(&_ProviderRegistry.TransactOpts, _minStake)
}

// SetMinStake is a paid mutator transaction binding the contract method 0x8c80fd90.
//
// Solidity: function setMinStake(uint256 _minStake) returns()
func (_ProviderRegistry *ProviderRegistryTransactorSession) SetMinStake(_minStake *big.Int) (*types.Transaction, error) {
	return _ProviderRegistry.Contract.SetMinStake(&_ProviderRegistry.TransactOpts, _minStake)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_ProviderRegistry *ProviderRegistryTransactor) TransferOwnership(opts *bind.TransactOpts, newOwner common.Address) (*types.Transaction, error) {
	return _ProviderRegistry.contract.Transact(opts, "transferOwnership", newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_ProviderRegistry *ProviderRegistrySession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _ProviderRegistry.Contract.TransferOwnership(&_ProviderRegistry.TransactOpts, newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_ProviderRegistry *ProviderRegistryTransactorSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _ProviderRegistry.Contract.TransferOwnership(&_ProviderRegistry.TransactOpts, newOwner)
}

// ProviderRegistryDeregisteredIterator is returned from FilterDeregistered and is used to iterate over the raw logs and unpacked data for Deregistered events raised by the ProviderRegistry contract.
type ProviderRegistryDeregisteredIterator struct {
	Event *ProviderRegistryDeregistered // Event containing the contract specifics and raw log

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
func (it *ProviderRegistryDeregisteredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ProviderRegistryDeregistered)
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
		it.Event = new(ProviderRegistryDeregistered)
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
func (it *ProviderRegistryDeregisteredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ProviderRegistryDeregisteredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ProviderRegistryDeregistered represents a Deregistered event raised by the ProviderRegistry contract.
type ProviderRegistryDeregistered struct {
	Provider common.Address
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterDeregistered is a free log retrieval operation binding the contract event 0xafebd0f81ba8c430fcc0c6a6e7a26fd7f868af9c4e4f19db37a0f16502374fd5.
//
// Solidity: event Deregistered(address indexed provider)
func (_ProviderRegistry *ProviderRegistryFilterer) FilterDeregistered(opts *bind.FilterOpts, provider []common.Address) (*ProviderRegistryDeregisteredIterator, error) {

	var providerRule []interface{}
	for _, providerItem := range provider {
		providerRule = append(providerRule, providerItem)
	}

	logs, sub, err := _ProviderRegistry.contract.FilterLogs(opts, "Deregistered", providerRule)
	if err != nil {
		return nil, err
	}
	return &ProviderRegistryDeregisteredIterator{contract: _ProviderRegistry.contract, event: "Deregistered", logs: logs, sub: sub}, nil
}

// WatchDeregistered is a free log subscription operation binding the contract event 0xafebd0f81ba8c430fcc0c6a6e7a26fd7f868af9c4e4f19db37a0f16502374fd5.
//
// Solidity: event Deregistered(address indexed provider)
func (_ProviderRegistry *ProviderRegistryFilterer) WatchDeregistered(opts *bind.WatchOpts, sink chan<- *ProviderRegistryDeregistered, provider []common.Address) (event.Subscription, error) {

	var providerRule []interface{}
	for _, providerItem := range provider {
		providerRule = append(providerRule, providerItem)
	}

	logs, sub, err := _ProviderRegistry.contract.WatchLogs(opts, "Deregistered", providerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ProviderRegistryDeregistered)
				if err := _ProviderRegistry.contract.UnpackLog(event, "Deregistered", log); err != nil {
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

// ParseDeregistered is a log parse operation binding the contract event 0xafebd0f81ba8c430fcc0c6a6e7a26fd7f868af9c4e4f19db37a0f16502374fd5.
//
// Solidity: event Deregistered(address indexed provider)
func (_ProviderRegistry *ProviderRegistryFilterer) ParseDeregistered(log types.Log) (*ProviderRegistryDeregistered, error) {
	event := new(ProviderRegistryDeregistered)
	if err := _ProviderRegistry.contract.UnpackLog(event, "Deregistered", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
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
	Version uint8
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterInitialized is a free log retrieval operation binding the contract event 0x7f26b83ff96e1f2b6a682f133852f6798a09c465da95921460cefb3847402498.
//
// Solidity: event Initialized(uint8 version)
func (_ProviderRegistry *ProviderRegistryFilterer) FilterInitialized(opts *bind.FilterOpts) (*ProviderRegistryInitializedIterator, error) {

	logs, sub, err := _ProviderRegistry.contract.FilterLogs(opts, "Initialized")
	if err != nil {
		return nil, err
	}
	return &ProviderRegistryInitializedIterator{contract: _ProviderRegistry.contract, event: "Initialized", logs: logs, sub: sub}, nil
}

// WatchInitialized is a free log subscription operation binding the contract event 0x7f26b83ff96e1f2b6a682f133852f6798a09c465da95921460cefb3847402498.
//
// Solidity: event Initialized(uint8 version)
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

// ParseInitialized is a log parse operation binding the contract event 0x7f26b83ff96e1f2b6a682f133852f6798a09c465da95921460cefb3847402498.
//
// Solidity: event Initialized(uint8 version)
func (_ProviderRegistry *ProviderRegistryFilterer) ParseInitialized(log types.Log) (*ProviderRegistryInitialized, error) {
	event := new(ProviderRegistryInitialized)
	if err := _ProviderRegistry.contract.UnpackLog(event, "Initialized", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ProviderRegistryMinStakeUpdatedIterator is returned from FilterMinStakeUpdated and is used to iterate over the raw logs and unpacked data for MinStakeUpdated events raised by the ProviderRegistry contract.
type ProviderRegistryMinStakeUpdatedIterator struct {
	Event *ProviderRegistryMinStakeUpdated // Event containing the contract specifics and raw log

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
func (it *ProviderRegistryMinStakeUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ProviderRegistryMinStakeUpdated)
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
		it.Event = new(ProviderRegistryMinStakeUpdated)
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
func (it *ProviderRegistryMinStakeUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ProviderRegistryMinStakeUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ProviderRegistryMinStakeUpdated represents a MinStakeUpdated event raised by the ProviderRegistry contract.
type ProviderRegistryMinStakeUpdated struct {
	NewStake *big.Int
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterMinStakeUpdated is a free log retrieval operation binding the contract event 0x47ab46f2c8d4258304a2f5551c1cbdb6981be49631365d1ba7191288a73f39ef.
//
// Solidity: event MinStakeUpdated(uint256 newStake)
func (_ProviderRegistry *ProviderRegistryFilterer) FilterMinStakeUpdated(opts *bind.FilterOpts) (*ProviderRegistryMinStakeUpdatedIterator, error) {

	logs, sub, err := _ProviderRegistry.contract.FilterLogs(opts, "MinStakeUpdated")
	if err != nil {
		return nil, err
	}
	return &ProviderRegistryMinStakeUpdatedIterator{contract: _ProviderRegistry.contract, event: "MinStakeUpdated", logs: logs, sub: sub}, nil
}

// WatchMinStakeUpdated is a free log subscription operation binding the contract event 0x47ab46f2c8d4258304a2f5551c1cbdb6981be49631365d1ba7191288a73f39ef.
//
// Solidity: event MinStakeUpdated(uint256 newStake)
func (_ProviderRegistry *ProviderRegistryFilterer) WatchMinStakeUpdated(opts *bind.WatchOpts, sink chan<- *ProviderRegistryMinStakeUpdated) (event.Subscription, error) {

	logs, sub, err := _ProviderRegistry.contract.WatchLogs(opts, "MinStakeUpdated")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ProviderRegistryMinStakeUpdated)
				if err := _ProviderRegistry.contract.UnpackLog(event, "MinStakeUpdated", log); err != nil {
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

// ParseMinStakeUpdated is a log parse operation binding the contract event 0x47ab46f2c8d4258304a2f5551c1cbdb6981be49631365d1ba7191288a73f39ef.
//
// Solidity: event MinStakeUpdated(uint256 newStake)
func (_ProviderRegistry *ProviderRegistryFilterer) ParseMinStakeUpdated(log types.Log) (*ProviderRegistryMinStakeUpdated, error) {
	event := new(ProviderRegistryMinStakeUpdated)
	if err := _ProviderRegistry.contract.UnpackLog(event, "MinStakeUpdated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ProviderRegistryOwnershipTransferredIterator is returned from FilterOwnershipTransferred and is used to iterate over the raw logs and unpacked data for OwnershipTransferred events raised by the ProviderRegistry contract.
type ProviderRegistryOwnershipTransferredIterator struct {
	Event *ProviderRegistryOwnershipTransferred // Event containing the contract specifics and raw log

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
func (it *ProviderRegistryOwnershipTransferredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ProviderRegistryOwnershipTransferred)
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
		it.Event = new(ProviderRegistryOwnershipTransferred)
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
func (it *ProviderRegistryOwnershipTransferredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ProviderRegistryOwnershipTransferredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ProviderRegistryOwnershipTransferred represents a OwnershipTransferred event raised by the ProviderRegistry contract.
type ProviderRegistryOwnershipTransferred struct {
	PreviousOwner common.Address
	NewOwner      common.Address
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterOwnershipTransferred is a free log retrieval operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_ProviderRegistry *ProviderRegistryFilterer) FilterOwnershipTransferred(opts *bind.FilterOpts, previousOwner []common.Address, newOwner []common.Address) (*ProviderRegistryOwnershipTransferredIterator, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _ProviderRegistry.contract.FilterLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return &ProviderRegistryOwnershipTransferredIterator{contract: _ProviderRegistry.contract, event: "OwnershipTransferred", logs: logs, sub: sub}, nil
}

// WatchOwnershipTransferred is a free log subscription operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_ProviderRegistry *ProviderRegistryFilterer) WatchOwnershipTransferred(opts *bind.WatchOpts, sink chan<- *ProviderRegistryOwnershipTransferred, previousOwner []common.Address, newOwner []common.Address) (event.Subscription, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _ProviderRegistry.contract.WatchLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ProviderRegistryOwnershipTransferred)
				if err := _ProviderRegistry.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
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

// ParseOwnershipTransferred is a log parse operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_ProviderRegistry *ProviderRegistryFilterer) ParseOwnershipTransferred(log types.Log) (*ProviderRegistryOwnershipTransferred, error) {
	event := new(ProviderRegistryOwnershipTransferred)
	if err := _ProviderRegistry.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ProviderRegistryRegisteredUpdatedIterator is returned from FilterRegisteredUpdated and is used to iterate over the raw logs and unpacked data for RegisteredUpdated events raised by the ProviderRegistry contract.
type ProviderRegistryRegisteredUpdatedIterator struct {
	Event *ProviderRegistryRegisteredUpdated // Event containing the contract specifics and raw log

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
func (it *ProviderRegistryRegisteredUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ProviderRegistryRegisteredUpdated)
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
		it.Event = new(ProviderRegistryRegisteredUpdated)
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
func (it *ProviderRegistryRegisteredUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ProviderRegistryRegisteredUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ProviderRegistryRegisteredUpdated represents a RegisteredUpdated event raised by the ProviderRegistry contract.
type ProviderRegistryRegisteredUpdated struct {
	Provider common.Address
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterRegisteredUpdated is a free log retrieval operation binding the contract event 0x0407973ec6e86eb9f260606351583a2737e1db4d6f44c2414e699bd665ae10dc.
//
// Solidity: event RegisteredUpdated(address indexed provider)
func (_ProviderRegistry *ProviderRegistryFilterer) FilterRegisteredUpdated(opts *bind.FilterOpts, provider []common.Address) (*ProviderRegistryRegisteredUpdatedIterator, error) {

	var providerRule []interface{}
	for _, providerItem := range provider {
		providerRule = append(providerRule, providerItem)
	}

	logs, sub, err := _ProviderRegistry.contract.FilterLogs(opts, "RegisteredUpdated", providerRule)
	if err != nil {
		return nil, err
	}
	return &ProviderRegistryRegisteredUpdatedIterator{contract: _ProviderRegistry.contract, event: "RegisteredUpdated", logs: logs, sub: sub}, nil
}

// WatchRegisteredUpdated is a free log subscription operation binding the contract event 0x0407973ec6e86eb9f260606351583a2737e1db4d6f44c2414e699bd665ae10dc.
//
// Solidity: event RegisteredUpdated(address indexed provider)
func (_ProviderRegistry *ProviderRegistryFilterer) WatchRegisteredUpdated(opts *bind.WatchOpts, sink chan<- *ProviderRegistryRegisteredUpdated, provider []common.Address) (event.Subscription, error) {

	var providerRule []interface{}
	for _, providerItem := range provider {
		providerRule = append(providerRule, providerItem)
	}

	logs, sub, err := _ProviderRegistry.contract.WatchLogs(opts, "RegisteredUpdated", providerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ProviderRegistryRegisteredUpdated)
				if err := _ProviderRegistry.contract.UnpackLog(event, "RegisteredUpdated", log); err != nil {
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

// ParseRegisteredUpdated is a log parse operation binding the contract event 0x0407973ec6e86eb9f260606351583a2737e1db4d6f44c2414e699bd665ae10dc.
//
// Solidity: event RegisteredUpdated(address indexed provider)
func (_ProviderRegistry *ProviderRegistryFilterer) ParseRegisteredUpdated(log types.Log) (*ProviderRegistryRegisteredUpdated, error) {
	event := new(ProviderRegistryRegisteredUpdated)
	if err := _ProviderRegistry.contract.UnpackLog(event, "RegisteredUpdated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
