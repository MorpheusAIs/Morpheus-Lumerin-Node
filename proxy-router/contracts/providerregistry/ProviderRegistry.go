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

// Provider is an auto generated low-level Go binding around an user-defined struct.
type Provider struct {
	Endpoint          string
	Stake             *big.Int
	CreatedAt         *big.Int
	LimitPeriodEnd    *big.Int
	LimitPeriodEarned *big.Int
	IsDeleted         bool
}

// ProviderRegistryMetaData contains all meta data concerning the ProviderRegistry contract.
var ProviderRegistryMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[],\"name\":\"ErrNoStake\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"ErrNoWithdrawableStake\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"ErrProviderNotDeleted\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"KeyExists\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"KeyNotFound\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_user\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"_contractOwner\",\"type\":\"address\"}],\"name\":\"NotContractOwner\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"NotSenderOrOwner\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"ProviderHasActiveBids\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"StakeTooLow\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"ZeroKey\",\"type\":\"error\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"provider\",\"type\":\"address\"}],\"name\":\"ProviderDeregistered\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"newStake\",\"type\":\"uint256\"}],\"name\":\"ProviderMinStakeUpdated\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"provider\",\"type\":\"address\"}],\"name\":\"ProviderRegisteredUpdated\",\"type\":\"event\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"addr\",\"type\":\"address\"}],\"name\":\"providerDeregister\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"addr\",\"type\":\"address\"}],\"name\":\"providerExists\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"providerGetAll\",\"outputs\":[{\"internalType\":\"address[]\",\"name\":\"\",\"type\":\"address[]\"},{\"components\":[{\"internalType\":\"string\",\"name\":\"endpoint\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"stake\",\"type\":\"uint256\"},{\"internalType\":\"uint128\",\"name\":\"createdAt\",\"type\":\"uint128\"},{\"internalType\":\"uint128\",\"name\":\"limitPeriodEnd\",\"type\":\"uint128\"},{\"internalType\":\"uint256\",\"name\":\"limitPeriodEarned\",\"type\":\"uint256\"},{\"internalType\":\"bool\",\"name\":\"isDeleted\",\"type\":\"bool\"}],\"internalType\":\"structProvider[]\",\"name\":\"\",\"type\":\"tuple[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"index\",\"type\":\"uint256\"}],\"name\":\"providerGetByIndex\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"addr\",\"type\":\"address\"},{\"components\":[{\"internalType\":\"string\",\"name\":\"endpoint\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"stake\",\"type\":\"uint256\"},{\"internalType\":\"uint128\",\"name\":\"createdAt\",\"type\":\"uint128\"},{\"internalType\":\"uint128\",\"name\":\"limitPeriodEnd\",\"type\":\"uint128\"},{\"internalType\":\"uint256\",\"name\":\"limitPeriodEarned\",\"type\":\"uint256\"},{\"internalType\":\"bool\",\"name\":\"isDeleted\",\"type\":\"bool\"}],\"internalType\":\"structProvider\",\"name\":\"provider\",\"type\":\"tuple\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"providerGetCount\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"count\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"providerGetIds\",\"outputs\":[{\"internalType\":\"address[]\",\"name\":\"\",\"type\":\"address[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"addr\",\"type\":\"address\"}],\"name\":\"providerMap\",\"outputs\":[{\"components\":[{\"internalType\":\"string\",\"name\":\"endpoint\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"stake\",\"type\":\"uint256\"},{\"internalType\":\"uint128\",\"name\":\"createdAt\",\"type\":\"uint128\"},{\"internalType\":\"uint128\",\"name\":\"limitPeriodEnd\",\"type\":\"uint128\"},{\"internalType\":\"uint256\",\"name\":\"limitPeriodEarned\",\"type\":\"uint256\"},{\"internalType\":\"bool\",\"name\":\"isDeleted\",\"type\":\"bool\"}],\"internalType\":\"structProvider\",\"name\":\"\",\"type\":\"tuple\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"providerMinStake\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"addr\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"addStake\",\"type\":\"uint256\"},{\"internalType\":\"string\",\"name\":\"endpoint\",\"type\":\"string\"}],\"name\":\"providerRegister\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"_minStake\",\"type\":\"uint256\"}],\"name\":\"providerSetMinStake\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"addr\",\"type\":\"address\"}],\"name\":\"providerWithdrawStake\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"index\",\"type\":\"uint256\"}],\"name\":\"providers\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"}]",
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

// ProviderExists is a free data retrieval call binding the contract method 0xdfc03505.
//
// Solidity: function providerExists(address addr) view returns(bool)
func (_ProviderRegistry *ProviderRegistryCaller) ProviderExists(opts *bind.CallOpts, addr common.Address) (bool, error) {
	var out []interface{}
	err := _ProviderRegistry.contract.Call(opts, &out, "providerExists", addr)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// ProviderExists is a free data retrieval call binding the contract method 0xdfc03505.
//
// Solidity: function providerExists(address addr) view returns(bool)
func (_ProviderRegistry *ProviderRegistrySession) ProviderExists(addr common.Address) (bool, error) {
	return _ProviderRegistry.Contract.ProviderExists(&_ProviderRegistry.CallOpts, addr)
}

// ProviderExists is a free data retrieval call binding the contract method 0xdfc03505.
//
// Solidity: function providerExists(address addr) view returns(bool)
func (_ProviderRegistry *ProviderRegistryCallerSession) ProviderExists(addr common.Address) (bool, error) {
	return _ProviderRegistry.Contract.ProviderExists(&_ProviderRegistry.CallOpts, addr)
}

// ProviderGetAll is a free data retrieval call binding the contract method 0x86af8fdc.
//
// Solidity: function providerGetAll() view returns(address[], (string,uint256,uint128,uint128,uint256,bool)[])
func (_ProviderRegistry *ProviderRegistryCaller) ProviderGetAll(opts *bind.CallOpts) ([]common.Address, []Provider, error) {
	var out []interface{}
	err := _ProviderRegistry.contract.Call(opts, &out, "providerGetAll")

	if err != nil {
		return *new([]common.Address), *new([]Provider), err
	}

	out0 := *abi.ConvertType(out[0], new([]common.Address)).(*[]common.Address)
	out1 := *abi.ConvertType(out[1], new([]Provider)).(*[]Provider)

	return out0, out1, err

}

// ProviderGetAll is a free data retrieval call binding the contract method 0x86af8fdc.
//
// Solidity: function providerGetAll() view returns(address[], (string,uint256,uint128,uint128,uint256,bool)[])
func (_ProviderRegistry *ProviderRegistrySession) ProviderGetAll() ([]common.Address, []Provider, error) {
	return _ProviderRegistry.Contract.ProviderGetAll(&_ProviderRegistry.CallOpts)
}

// ProviderGetAll is a free data retrieval call binding the contract method 0x86af8fdc.
//
// Solidity: function providerGetAll() view returns(address[], (string,uint256,uint128,uint128,uint256,bool)[])
func (_ProviderRegistry *ProviderRegistryCallerSession) ProviderGetAll() ([]common.Address, []Provider, error) {
	return _ProviderRegistry.Contract.ProviderGetAll(&_ProviderRegistry.CallOpts)
}

// ProviderGetByIndex is a free data retrieval call binding the contract method 0xb8eed333.
//
// Solidity: function providerGetByIndex(uint256 index) view returns(address addr, (string,uint256,uint128,uint128,uint256,bool) provider)
func (_ProviderRegistry *ProviderRegistryCaller) ProviderGetByIndex(opts *bind.CallOpts, index *big.Int) (struct {
	Addr     common.Address
	Provider Provider
}, error) {
	var out []interface{}
	err := _ProviderRegistry.contract.Call(opts, &out, "providerGetByIndex", index)

	outstruct := new(struct {
		Addr     common.Address
		Provider Provider
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.Addr = *abi.ConvertType(out[0], new(common.Address)).(*common.Address)
	outstruct.Provider = *abi.ConvertType(out[1], new(Provider)).(*Provider)

	return *outstruct, err

}

// ProviderGetByIndex is a free data retrieval call binding the contract method 0xb8eed333.
//
// Solidity: function providerGetByIndex(uint256 index) view returns(address addr, (string,uint256,uint128,uint128,uint256,bool) provider)
func (_ProviderRegistry *ProviderRegistrySession) ProviderGetByIndex(index *big.Int) (struct {
	Addr     common.Address
	Provider Provider
}, error) {
	return _ProviderRegistry.Contract.ProviderGetByIndex(&_ProviderRegistry.CallOpts, index)
}

// ProviderGetByIndex is a free data retrieval call binding the contract method 0xb8eed333.
//
// Solidity: function providerGetByIndex(uint256 index) view returns(address addr, (string,uint256,uint128,uint128,uint256,bool) provider)
func (_ProviderRegistry *ProviderRegistryCallerSession) ProviderGetByIndex(index *big.Int) (struct {
	Addr     common.Address
	Provider Provider
}, error) {
	return _ProviderRegistry.Contract.ProviderGetByIndex(&_ProviderRegistry.CallOpts, index)
}

// ProviderGetCount is a free data retrieval call binding the contract method 0x91d2b7eb.
//
// Solidity: function providerGetCount() view returns(uint256 count)
func (_ProviderRegistry *ProviderRegistryCaller) ProviderGetCount(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _ProviderRegistry.contract.Call(opts, &out, "providerGetCount")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// ProviderGetCount is a free data retrieval call binding the contract method 0x91d2b7eb.
//
// Solidity: function providerGetCount() view returns(uint256 count)
func (_ProviderRegistry *ProviderRegistrySession) ProviderGetCount() (*big.Int, error) {
	return _ProviderRegistry.Contract.ProviderGetCount(&_ProviderRegistry.CallOpts)
}

// ProviderGetCount is a free data retrieval call binding the contract method 0x91d2b7eb.
//
// Solidity: function providerGetCount() view returns(uint256 count)
func (_ProviderRegistry *ProviderRegistryCallerSession) ProviderGetCount() (*big.Int, error) {
	return _ProviderRegistry.Contract.ProviderGetCount(&_ProviderRegistry.CallOpts)
}

// ProviderGetIds is a free data retrieval call binding the contract method 0x2e888fe1.
//
// Solidity: function providerGetIds() view returns(address[])
func (_ProviderRegistry *ProviderRegistryCaller) ProviderGetIds(opts *bind.CallOpts) ([]common.Address, error) {
	var out []interface{}
	err := _ProviderRegistry.contract.Call(opts, &out, "providerGetIds")

	if err != nil {
		return *new([]common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new([]common.Address)).(*[]common.Address)

	return out0, err

}

// ProviderGetIds is a free data retrieval call binding the contract method 0x2e888fe1.
//
// Solidity: function providerGetIds() view returns(address[])
func (_ProviderRegistry *ProviderRegistrySession) ProviderGetIds() ([]common.Address, error) {
	return _ProviderRegistry.Contract.ProviderGetIds(&_ProviderRegistry.CallOpts)
}

// ProviderGetIds is a free data retrieval call binding the contract method 0x2e888fe1.
//
// Solidity: function providerGetIds() view returns(address[])
func (_ProviderRegistry *ProviderRegistryCallerSession) ProviderGetIds() ([]common.Address, error) {
	return _ProviderRegistry.Contract.ProviderGetIds(&_ProviderRegistry.CallOpts)
}

// ProviderMap is a free data retrieval call binding the contract method 0xa6c87915.
//
// Solidity: function providerMap(address addr) view returns((string,uint256,uint128,uint128,uint256,bool))
func (_ProviderRegistry *ProviderRegistryCaller) ProviderMap(opts *bind.CallOpts, addr common.Address) (Provider, error) {
	var out []interface{}
	err := _ProviderRegistry.contract.Call(opts, &out, "providerMap", addr)

	if err != nil {
		return *new(Provider), err
	}

	out0 := *abi.ConvertType(out[0], new(Provider)).(*Provider)

	return out0, err

}

// ProviderMap is a free data retrieval call binding the contract method 0xa6c87915.
//
// Solidity: function providerMap(address addr) view returns((string,uint256,uint128,uint128,uint256,bool))
func (_ProviderRegistry *ProviderRegistrySession) ProviderMap(addr common.Address) (Provider, error) {
	return _ProviderRegistry.Contract.ProviderMap(&_ProviderRegistry.CallOpts, addr)
}

// ProviderMap is a free data retrieval call binding the contract method 0xa6c87915.
//
// Solidity: function providerMap(address addr) view returns((string,uint256,uint128,uint128,uint256,bool))
func (_ProviderRegistry *ProviderRegistryCallerSession) ProviderMap(addr common.Address) (Provider, error) {
	return _ProviderRegistry.Contract.ProviderMap(&_ProviderRegistry.CallOpts, addr)
}

// ProviderMinStake is a free data retrieval call binding the contract method 0xbcd5641e.
//
// Solidity: function providerMinStake() view returns(uint256)
func (_ProviderRegistry *ProviderRegistryCaller) ProviderMinStake(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _ProviderRegistry.contract.Call(opts, &out, "providerMinStake")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// ProviderMinStake is a free data retrieval call binding the contract method 0xbcd5641e.
//
// Solidity: function providerMinStake() view returns(uint256)
func (_ProviderRegistry *ProviderRegistrySession) ProviderMinStake() (*big.Int, error) {
	return _ProviderRegistry.Contract.ProviderMinStake(&_ProviderRegistry.CallOpts)
}

// ProviderMinStake is a free data retrieval call binding the contract method 0xbcd5641e.
//
// Solidity: function providerMinStake() view returns(uint256)
func (_ProviderRegistry *ProviderRegistryCallerSession) ProviderMinStake() (*big.Int, error) {
	return _ProviderRegistry.Contract.ProviderMinStake(&_ProviderRegistry.CallOpts)
}

// Providers is a free data retrieval call binding the contract method 0x50f3fc81.
//
// Solidity: function providers(uint256 index) view returns(address)
func (_ProviderRegistry *ProviderRegistryCaller) Providers(opts *bind.CallOpts, index *big.Int) (common.Address, error) {
	var out []interface{}
	err := _ProviderRegistry.contract.Call(opts, &out, "providers", index)

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Providers is a free data retrieval call binding the contract method 0x50f3fc81.
//
// Solidity: function providers(uint256 index) view returns(address)
func (_ProviderRegistry *ProviderRegistrySession) Providers(index *big.Int) (common.Address, error) {
	return _ProviderRegistry.Contract.Providers(&_ProviderRegistry.CallOpts, index)
}

// Providers is a free data retrieval call binding the contract method 0x50f3fc81.
//
// Solidity: function providers(uint256 index) view returns(address)
func (_ProviderRegistry *ProviderRegistryCallerSession) Providers(index *big.Int) (common.Address, error) {
	return _ProviderRegistry.Contract.Providers(&_ProviderRegistry.CallOpts, index)
}

// ProviderDeregister is a paid mutator transaction binding the contract method 0x2ca36c49.
//
// Solidity: function providerDeregister(address addr) returns()
func (_ProviderRegistry *ProviderRegistryTransactor) ProviderDeregister(opts *bind.TransactOpts, addr common.Address) (*types.Transaction, error) {
	return _ProviderRegistry.contract.Transact(opts, "providerDeregister", addr)
}

// ProviderDeregister is a paid mutator transaction binding the contract method 0x2ca36c49.
//
// Solidity: function providerDeregister(address addr) returns()
func (_ProviderRegistry *ProviderRegistrySession) ProviderDeregister(addr common.Address) (*types.Transaction, error) {
	return _ProviderRegistry.Contract.ProviderDeregister(&_ProviderRegistry.TransactOpts, addr)
}

// ProviderDeregister is a paid mutator transaction binding the contract method 0x2ca36c49.
//
// Solidity: function providerDeregister(address addr) returns()
func (_ProviderRegistry *ProviderRegistryTransactorSession) ProviderDeregister(addr common.Address) (*types.Transaction, error) {
	return _ProviderRegistry.Contract.ProviderDeregister(&_ProviderRegistry.TransactOpts, addr)
}

// ProviderRegister is a paid mutator transaction binding the contract method 0x365700cb.
//
// Solidity: function providerRegister(address addr, uint256 addStake, string endpoint) returns()
func (_ProviderRegistry *ProviderRegistryTransactor) ProviderRegister(opts *bind.TransactOpts, addr common.Address, addStake *big.Int, endpoint string) (*types.Transaction, error) {
	return _ProviderRegistry.contract.Transact(opts, "providerRegister", addr, addStake, endpoint)
}

// ProviderRegister is a paid mutator transaction binding the contract method 0x365700cb.
//
// Solidity: function providerRegister(address addr, uint256 addStake, string endpoint) returns()
func (_ProviderRegistry *ProviderRegistrySession) ProviderRegister(addr common.Address, addStake *big.Int, endpoint string) (*types.Transaction, error) {
	return _ProviderRegistry.Contract.ProviderRegister(&_ProviderRegistry.TransactOpts, addr, addStake, endpoint)
}

// ProviderRegister is a paid mutator transaction binding the contract method 0x365700cb.
//
// Solidity: function providerRegister(address addr, uint256 addStake, string endpoint) returns()
func (_ProviderRegistry *ProviderRegistryTransactorSession) ProviderRegister(addr common.Address, addStake *big.Int, endpoint string) (*types.Transaction, error) {
	return _ProviderRegistry.Contract.ProviderRegister(&_ProviderRegistry.TransactOpts, addr, addStake, endpoint)
}

// ProviderSetMinStake is a paid mutator transaction binding the contract method 0x0b7f94d6.
//
// Solidity: function providerSetMinStake(uint256 _minStake) returns()
func (_ProviderRegistry *ProviderRegistryTransactor) ProviderSetMinStake(opts *bind.TransactOpts, _minStake *big.Int) (*types.Transaction, error) {
	return _ProviderRegistry.contract.Transact(opts, "providerSetMinStake", _minStake)
}

// ProviderSetMinStake is a paid mutator transaction binding the contract method 0x0b7f94d6.
//
// Solidity: function providerSetMinStake(uint256 _minStake) returns()
func (_ProviderRegistry *ProviderRegistrySession) ProviderSetMinStake(_minStake *big.Int) (*types.Transaction, error) {
	return _ProviderRegistry.Contract.ProviderSetMinStake(&_ProviderRegistry.TransactOpts, _minStake)
}

// ProviderSetMinStake is a paid mutator transaction binding the contract method 0x0b7f94d6.
//
// Solidity: function providerSetMinStake(uint256 _minStake) returns()
func (_ProviderRegistry *ProviderRegistryTransactorSession) ProviderSetMinStake(_minStake *big.Int) (*types.Transaction, error) {
	return _ProviderRegistry.Contract.ProviderSetMinStake(&_ProviderRegistry.TransactOpts, _minStake)
}

// ProviderWithdrawStake is a paid mutator transaction binding the contract method 0x8209d9ed.
//
// Solidity: function providerWithdrawStake(address addr) returns()
func (_ProviderRegistry *ProviderRegistryTransactor) ProviderWithdrawStake(opts *bind.TransactOpts, addr common.Address) (*types.Transaction, error) {
	return _ProviderRegistry.contract.Transact(opts, "providerWithdrawStake", addr)
}

// ProviderWithdrawStake is a paid mutator transaction binding the contract method 0x8209d9ed.
//
// Solidity: function providerWithdrawStake(address addr) returns()
func (_ProviderRegistry *ProviderRegistrySession) ProviderWithdrawStake(addr common.Address) (*types.Transaction, error) {
	return _ProviderRegistry.Contract.ProviderWithdrawStake(&_ProviderRegistry.TransactOpts, addr)
}

// ProviderWithdrawStake is a paid mutator transaction binding the contract method 0x8209d9ed.
//
// Solidity: function providerWithdrawStake(address addr) returns()
func (_ProviderRegistry *ProviderRegistryTransactorSession) ProviderWithdrawStake(addr common.Address) (*types.Transaction, error) {
	return _ProviderRegistry.Contract.ProviderWithdrawStake(&_ProviderRegistry.TransactOpts, addr)
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
