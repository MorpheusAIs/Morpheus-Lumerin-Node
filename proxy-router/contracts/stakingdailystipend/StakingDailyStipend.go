// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package stakingdailystipend

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

// StakingDailyStipendMetaData contains all meta data concerning the StakingDailyStipend contract.
var StakingDailyStipendMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[],\"name\":\"NotEnoughDailyStipend\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"NotEnoughStake\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"NotSenderOrOwner\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"NotSessionRouterOrOwner\",\"type\":\"error\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint8\",\"name\":\"version\",\"type\":\"uint8\"}],\"name\":\"Initialized\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"previousOwner\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"OwnershipTransferred\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"userAddress\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"Staked\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"userAddress\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"Unstaked\",\"type\":\"event\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"userAddress\",\"type\":\"address\"}],\"name\":\"balanceOfDailyStipend\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getComputeBalance\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"userAddress\",\"type\":\"address\"}],\"name\":\"getStakeOnHold\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getTodaysBudget\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"userAddress\",\"type\":\"address\"}],\"name\":\"getTodaysSpend\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_token\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"_tokenAccount\",\"type\":\"address\"}],\"name\":\"initialize\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"owner\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"renounceOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"returnStipend\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"sessionRouter\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"addr\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"stake\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"name\":\"todaysSpend\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"releaseAt\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"token\",\"outputs\":[{\"internalType\":\"contractIERC20\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"tokenAccount\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"from\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"transferDailyStipend\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"transferOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"addr\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"sendToAddr\",\"type\":\"address\"}],\"name\":\"unstake\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"name\":\"userStake\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"userAddress\",\"type\":\"address\"}],\"name\":\"withdrawableStakeBalance\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"}]",
}

// StakingDailyStipendABI is the input ABI used to generate the binding from.
// Deprecated: Use StakingDailyStipendMetaData.ABI instead.
var StakingDailyStipendABI = StakingDailyStipendMetaData.ABI

// StakingDailyStipend is an auto generated Go binding around an Ethereum contract.
type StakingDailyStipend struct {
	StakingDailyStipendCaller     // Read-only binding to the contract
	StakingDailyStipendTransactor // Write-only binding to the contract
	StakingDailyStipendFilterer   // Log filterer for contract events
}

// StakingDailyStipendCaller is an auto generated read-only Go binding around an Ethereum contract.
type StakingDailyStipendCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// StakingDailyStipendTransactor is an auto generated write-only Go binding around an Ethereum contract.
type StakingDailyStipendTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// StakingDailyStipendFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type StakingDailyStipendFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// StakingDailyStipendSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type StakingDailyStipendSession struct {
	Contract     *StakingDailyStipend // Generic contract binding to set the session for
	CallOpts     bind.CallOpts        // Call options to use throughout this session
	TransactOpts bind.TransactOpts    // Transaction auth options to use throughout this session
}

// StakingDailyStipendCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type StakingDailyStipendCallerSession struct {
	Contract *StakingDailyStipendCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts              // Call options to use throughout this session
}

// StakingDailyStipendTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type StakingDailyStipendTransactorSession struct {
	Contract     *StakingDailyStipendTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts              // Transaction auth options to use throughout this session
}

// StakingDailyStipendRaw is an auto generated low-level Go binding around an Ethereum contract.
type StakingDailyStipendRaw struct {
	Contract *StakingDailyStipend // Generic contract binding to access the raw methods on
}

// StakingDailyStipendCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type StakingDailyStipendCallerRaw struct {
	Contract *StakingDailyStipendCaller // Generic read-only contract binding to access the raw methods on
}

// StakingDailyStipendTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type StakingDailyStipendTransactorRaw struct {
	Contract *StakingDailyStipendTransactor // Generic write-only contract binding to access the raw methods on
}

// NewStakingDailyStipend creates a new instance of StakingDailyStipend, bound to a specific deployed contract.
func NewStakingDailyStipend(address common.Address, backend bind.ContractBackend) (*StakingDailyStipend, error) {
	contract, err := bindStakingDailyStipend(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &StakingDailyStipend{StakingDailyStipendCaller: StakingDailyStipendCaller{contract: contract}, StakingDailyStipendTransactor: StakingDailyStipendTransactor{contract: contract}, StakingDailyStipendFilterer: StakingDailyStipendFilterer{contract: contract}}, nil
}

// NewStakingDailyStipendCaller creates a new read-only instance of StakingDailyStipend, bound to a specific deployed contract.
func NewStakingDailyStipendCaller(address common.Address, caller bind.ContractCaller) (*StakingDailyStipendCaller, error) {
	contract, err := bindStakingDailyStipend(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &StakingDailyStipendCaller{contract: contract}, nil
}

// NewStakingDailyStipendTransactor creates a new write-only instance of StakingDailyStipend, bound to a specific deployed contract.
func NewStakingDailyStipendTransactor(address common.Address, transactor bind.ContractTransactor) (*StakingDailyStipendTransactor, error) {
	contract, err := bindStakingDailyStipend(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &StakingDailyStipendTransactor{contract: contract}, nil
}

// NewStakingDailyStipendFilterer creates a new log filterer instance of StakingDailyStipend, bound to a specific deployed contract.
func NewStakingDailyStipendFilterer(address common.Address, filterer bind.ContractFilterer) (*StakingDailyStipendFilterer, error) {
	contract, err := bindStakingDailyStipend(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &StakingDailyStipendFilterer{contract: contract}, nil
}

// bindStakingDailyStipend binds a generic wrapper to an already deployed contract.
func bindStakingDailyStipend(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := StakingDailyStipendMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_StakingDailyStipend *StakingDailyStipendRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _StakingDailyStipend.Contract.StakingDailyStipendCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_StakingDailyStipend *StakingDailyStipendRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _StakingDailyStipend.Contract.StakingDailyStipendTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_StakingDailyStipend *StakingDailyStipendRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _StakingDailyStipend.Contract.StakingDailyStipendTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_StakingDailyStipend *StakingDailyStipendCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _StakingDailyStipend.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_StakingDailyStipend *StakingDailyStipendTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _StakingDailyStipend.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_StakingDailyStipend *StakingDailyStipendTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _StakingDailyStipend.Contract.contract.Transact(opts, method, params...)
}

// BalanceOfDailyStipend is a free data retrieval call binding the contract method 0xf0612b48.
//
// Solidity: function balanceOfDailyStipend(address userAddress) view returns(uint256)
func (_StakingDailyStipend *StakingDailyStipendCaller) BalanceOfDailyStipend(opts *bind.CallOpts, userAddress common.Address) (*big.Int, error) {
	var out []interface{}
	err := _StakingDailyStipend.contract.Call(opts, &out, "balanceOfDailyStipend", userAddress)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// BalanceOfDailyStipend is a free data retrieval call binding the contract method 0xf0612b48.
//
// Solidity: function balanceOfDailyStipend(address userAddress) view returns(uint256)
func (_StakingDailyStipend *StakingDailyStipendSession) BalanceOfDailyStipend(userAddress common.Address) (*big.Int, error) {
	return _StakingDailyStipend.Contract.BalanceOfDailyStipend(&_StakingDailyStipend.CallOpts, userAddress)
}

// BalanceOfDailyStipend is a free data retrieval call binding the contract method 0xf0612b48.
//
// Solidity: function balanceOfDailyStipend(address userAddress) view returns(uint256)
func (_StakingDailyStipend *StakingDailyStipendCallerSession) BalanceOfDailyStipend(userAddress common.Address) (*big.Int, error) {
	return _StakingDailyStipend.Contract.BalanceOfDailyStipend(&_StakingDailyStipend.CallOpts, userAddress)
}

// GetComputeBalance is a free data retrieval call binding the contract method 0x653cdf0c.
//
// Solidity: function getComputeBalance() view returns(uint256)
func (_StakingDailyStipend *StakingDailyStipendCaller) GetComputeBalance(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _StakingDailyStipend.contract.Call(opts, &out, "getComputeBalance")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetComputeBalance is a free data retrieval call binding the contract method 0x653cdf0c.
//
// Solidity: function getComputeBalance() view returns(uint256)
func (_StakingDailyStipend *StakingDailyStipendSession) GetComputeBalance() (*big.Int, error) {
	return _StakingDailyStipend.Contract.GetComputeBalance(&_StakingDailyStipend.CallOpts)
}

// GetComputeBalance is a free data retrieval call binding the contract method 0x653cdf0c.
//
// Solidity: function getComputeBalance() view returns(uint256)
func (_StakingDailyStipend *StakingDailyStipendCallerSession) GetComputeBalance() (*big.Int, error) {
	return _StakingDailyStipend.Contract.GetComputeBalance(&_StakingDailyStipend.CallOpts)
}

// GetStakeOnHold is a free data retrieval call binding the contract method 0x9bc65da5.
//
// Solidity: function getStakeOnHold(address userAddress) view returns(uint256)
func (_StakingDailyStipend *StakingDailyStipendCaller) GetStakeOnHold(opts *bind.CallOpts, userAddress common.Address) (*big.Int, error) {
	var out []interface{}
	err := _StakingDailyStipend.contract.Call(opts, &out, "getStakeOnHold", userAddress)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetStakeOnHold is a free data retrieval call binding the contract method 0x9bc65da5.
//
// Solidity: function getStakeOnHold(address userAddress) view returns(uint256)
func (_StakingDailyStipend *StakingDailyStipendSession) GetStakeOnHold(userAddress common.Address) (*big.Int, error) {
	return _StakingDailyStipend.Contract.GetStakeOnHold(&_StakingDailyStipend.CallOpts, userAddress)
}

// GetStakeOnHold is a free data retrieval call binding the contract method 0x9bc65da5.
//
// Solidity: function getStakeOnHold(address userAddress) view returns(uint256)
func (_StakingDailyStipend *StakingDailyStipendCallerSession) GetStakeOnHold(userAddress common.Address) (*big.Int, error) {
	return _StakingDailyStipend.Contract.GetStakeOnHold(&_StakingDailyStipend.CallOpts, userAddress)
}

// GetTodaysBudget is a free data retrieval call binding the contract method 0xa7e7f9a9.
//
// Solidity: function getTodaysBudget() view returns(uint256)
func (_StakingDailyStipend *StakingDailyStipendCaller) GetTodaysBudget(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _StakingDailyStipend.contract.Call(opts, &out, "getTodaysBudget")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetTodaysBudget is a free data retrieval call binding the contract method 0xa7e7f9a9.
//
// Solidity: function getTodaysBudget() view returns(uint256)
func (_StakingDailyStipend *StakingDailyStipendSession) GetTodaysBudget() (*big.Int, error) {
	return _StakingDailyStipend.Contract.GetTodaysBudget(&_StakingDailyStipend.CallOpts)
}

// GetTodaysBudget is a free data retrieval call binding the contract method 0xa7e7f9a9.
//
// Solidity: function getTodaysBudget() view returns(uint256)
func (_StakingDailyStipend *StakingDailyStipendCallerSession) GetTodaysBudget() (*big.Int, error) {
	return _StakingDailyStipend.Contract.GetTodaysBudget(&_StakingDailyStipend.CallOpts)
}

// GetTodaysSpend is a free data retrieval call binding the contract method 0x02fc4ec8.
//
// Solidity: function getTodaysSpend(address userAddress) view returns(uint256)
func (_StakingDailyStipend *StakingDailyStipendCaller) GetTodaysSpend(opts *bind.CallOpts, userAddress common.Address) (*big.Int, error) {
	var out []interface{}
	err := _StakingDailyStipend.contract.Call(opts, &out, "getTodaysSpend", userAddress)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetTodaysSpend is a free data retrieval call binding the contract method 0x02fc4ec8.
//
// Solidity: function getTodaysSpend(address userAddress) view returns(uint256)
func (_StakingDailyStipend *StakingDailyStipendSession) GetTodaysSpend(userAddress common.Address) (*big.Int, error) {
	return _StakingDailyStipend.Contract.GetTodaysSpend(&_StakingDailyStipend.CallOpts, userAddress)
}

// GetTodaysSpend is a free data retrieval call binding the contract method 0x02fc4ec8.
//
// Solidity: function getTodaysSpend(address userAddress) view returns(uint256)
func (_StakingDailyStipend *StakingDailyStipendCallerSession) GetTodaysSpend(userAddress common.Address) (*big.Int, error) {
	return _StakingDailyStipend.Contract.GetTodaysSpend(&_StakingDailyStipend.CallOpts, userAddress)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_StakingDailyStipend *StakingDailyStipendCaller) Owner(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _StakingDailyStipend.contract.Call(opts, &out, "owner")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_StakingDailyStipend *StakingDailyStipendSession) Owner() (common.Address, error) {
	return _StakingDailyStipend.Contract.Owner(&_StakingDailyStipend.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_StakingDailyStipend *StakingDailyStipendCallerSession) Owner() (common.Address, error) {
	return _StakingDailyStipend.Contract.Owner(&_StakingDailyStipend.CallOpts)
}

// SessionRouter is a free data retrieval call binding the contract method 0x707ba08a.
//
// Solidity: function sessionRouter() view returns(address)
func (_StakingDailyStipend *StakingDailyStipendCaller) SessionRouter(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _StakingDailyStipend.contract.Call(opts, &out, "sessionRouter")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// SessionRouter is a free data retrieval call binding the contract method 0x707ba08a.
//
// Solidity: function sessionRouter() view returns(address)
func (_StakingDailyStipend *StakingDailyStipendSession) SessionRouter() (common.Address, error) {
	return _StakingDailyStipend.Contract.SessionRouter(&_StakingDailyStipend.CallOpts)
}

// SessionRouter is a free data retrieval call binding the contract method 0x707ba08a.
//
// Solidity: function sessionRouter() view returns(address)
func (_StakingDailyStipend *StakingDailyStipendCallerSession) SessionRouter() (common.Address, error) {
	return _StakingDailyStipend.Contract.SessionRouter(&_StakingDailyStipend.CallOpts)
}

// TodaysSpend is a free data retrieval call binding the contract method 0x9b6e4a06.
//
// Solidity: function todaysSpend(address ) view returns(uint256 amount, uint256 releaseAt)
func (_StakingDailyStipend *StakingDailyStipendCaller) TodaysSpend(opts *bind.CallOpts, arg0 common.Address) (struct {
	Amount    *big.Int
	ReleaseAt *big.Int
}, error) {
	var out []interface{}
	err := _StakingDailyStipend.contract.Call(opts, &out, "todaysSpend", arg0)

	outstruct := new(struct {
		Amount    *big.Int
		ReleaseAt *big.Int
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.Amount = *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)
	outstruct.ReleaseAt = *abi.ConvertType(out[1], new(*big.Int)).(**big.Int)

	return *outstruct, err

}

// TodaysSpend is a free data retrieval call binding the contract method 0x9b6e4a06.
//
// Solidity: function todaysSpend(address ) view returns(uint256 amount, uint256 releaseAt)
func (_StakingDailyStipend *StakingDailyStipendSession) TodaysSpend(arg0 common.Address) (struct {
	Amount    *big.Int
	ReleaseAt *big.Int
}, error) {
	return _StakingDailyStipend.Contract.TodaysSpend(&_StakingDailyStipend.CallOpts, arg0)
}

// TodaysSpend is a free data retrieval call binding the contract method 0x9b6e4a06.
//
// Solidity: function todaysSpend(address ) view returns(uint256 amount, uint256 releaseAt)
func (_StakingDailyStipend *StakingDailyStipendCallerSession) TodaysSpend(arg0 common.Address) (struct {
	Amount    *big.Int
	ReleaseAt *big.Int
}, error) {
	return _StakingDailyStipend.Contract.TodaysSpend(&_StakingDailyStipend.CallOpts, arg0)
}

// Token is a free data retrieval call binding the contract method 0xfc0c546a.
//
// Solidity: function token() view returns(address)
func (_StakingDailyStipend *StakingDailyStipendCaller) Token(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _StakingDailyStipend.contract.Call(opts, &out, "token")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Token is a free data retrieval call binding the contract method 0xfc0c546a.
//
// Solidity: function token() view returns(address)
func (_StakingDailyStipend *StakingDailyStipendSession) Token() (common.Address, error) {
	return _StakingDailyStipend.Contract.Token(&_StakingDailyStipend.CallOpts)
}

// Token is a free data retrieval call binding the contract method 0xfc0c546a.
//
// Solidity: function token() view returns(address)
func (_StakingDailyStipend *StakingDailyStipendCallerSession) Token() (common.Address, error) {
	return _StakingDailyStipend.Contract.Token(&_StakingDailyStipend.CallOpts)
}

// TokenAccount is a free data retrieval call binding the contract method 0x30ddebcb.
//
// Solidity: function tokenAccount() view returns(address)
func (_StakingDailyStipend *StakingDailyStipendCaller) TokenAccount(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _StakingDailyStipend.contract.Call(opts, &out, "tokenAccount")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// TokenAccount is a free data retrieval call binding the contract method 0x30ddebcb.
//
// Solidity: function tokenAccount() view returns(address)
func (_StakingDailyStipend *StakingDailyStipendSession) TokenAccount() (common.Address, error) {
	return _StakingDailyStipend.Contract.TokenAccount(&_StakingDailyStipend.CallOpts)
}

// TokenAccount is a free data retrieval call binding the contract method 0x30ddebcb.
//
// Solidity: function tokenAccount() view returns(address)
func (_StakingDailyStipend *StakingDailyStipendCallerSession) TokenAccount() (common.Address, error) {
	return _StakingDailyStipend.Contract.TokenAccount(&_StakingDailyStipend.CallOpts)
}

// UserStake is a free data retrieval call binding the contract method 0x68e5585d.
//
// Solidity: function userStake(address ) view returns(uint256)
func (_StakingDailyStipend *StakingDailyStipendCaller) UserStake(opts *bind.CallOpts, arg0 common.Address) (*big.Int, error) {
	var out []interface{}
	err := _StakingDailyStipend.contract.Call(opts, &out, "userStake", arg0)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// UserStake is a free data retrieval call binding the contract method 0x68e5585d.
//
// Solidity: function userStake(address ) view returns(uint256)
func (_StakingDailyStipend *StakingDailyStipendSession) UserStake(arg0 common.Address) (*big.Int, error) {
	return _StakingDailyStipend.Contract.UserStake(&_StakingDailyStipend.CallOpts, arg0)
}

// UserStake is a free data retrieval call binding the contract method 0x68e5585d.
//
// Solidity: function userStake(address ) view returns(uint256)
func (_StakingDailyStipend *StakingDailyStipendCallerSession) UserStake(arg0 common.Address) (*big.Int, error) {
	return _StakingDailyStipend.Contract.UserStake(&_StakingDailyStipend.CallOpts, arg0)
}

// WithdrawableStakeBalance is a free data retrieval call binding the contract method 0x7594e4d9.
//
// Solidity: function withdrawableStakeBalance(address userAddress) view returns(uint256)
func (_StakingDailyStipend *StakingDailyStipendCaller) WithdrawableStakeBalance(opts *bind.CallOpts, userAddress common.Address) (*big.Int, error) {
	var out []interface{}
	err := _StakingDailyStipend.contract.Call(opts, &out, "withdrawableStakeBalance", userAddress)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// WithdrawableStakeBalance is a free data retrieval call binding the contract method 0x7594e4d9.
//
// Solidity: function withdrawableStakeBalance(address userAddress) view returns(uint256)
func (_StakingDailyStipend *StakingDailyStipendSession) WithdrawableStakeBalance(userAddress common.Address) (*big.Int, error) {
	return _StakingDailyStipend.Contract.WithdrawableStakeBalance(&_StakingDailyStipend.CallOpts, userAddress)
}

// WithdrawableStakeBalance is a free data retrieval call binding the contract method 0x7594e4d9.
//
// Solidity: function withdrawableStakeBalance(address userAddress) view returns(uint256)
func (_StakingDailyStipend *StakingDailyStipendCallerSession) WithdrawableStakeBalance(userAddress common.Address) (*big.Int, error) {
	return _StakingDailyStipend.Contract.WithdrawableStakeBalance(&_StakingDailyStipend.CallOpts, userAddress)
}

// Initialize is a paid mutator transaction binding the contract method 0x485cc955.
//
// Solidity: function initialize(address _token, address _tokenAccount) returns()
func (_StakingDailyStipend *StakingDailyStipendTransactor) Initialize(opts *bind.TransactOpts, _token common.Address, _tokenAccount common.Address) (*types.Transaction, error) {
	return _StakingDailyStipend.contract.Transact(opts, "initialize", _token, _tokenAccount)
}

// Initialize is a paid mutator transaction binding the contract method 0x485cc955.
//
// Solidity: function initialize(address _token, address _tokenAccount) returns()
func (_StakingDailyStipend *StakingDailyStipendSession) Initialize(_token common.Address, _tokenAccount common.Address) (*types.Transaction, error) {
	return _StakingDailyStipend.Contract.Initialize(&_StakingDailyStipend.TransactOpts, _token, _tokenAccount)
}

// Initialize is a paid mutator transaction binding the contract method 0x485cc955.
//
// Solidity: function initialize(address _token, address _tokenAccount) returns()
func (_StakingDailyStipend *StakingDailyStipendTransactorSession) Initialize(_token common.Address, _tokenAccount common.Address) (*types.Transaction, error) {
	return _StakingDailyStipend.Contract.Initialize(&_StakingDailyStipend.TransactOpts, _token, _tokenAccount)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_StakingDailyStipend *StakingDailyStipendTransactor) RenounceOwnership(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _StakingDailyStipend.contract.Transact(opts, "renounceOwnership")
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_StakingDailyStipend *StakingDailyStipendSession) RenounceOwnership() (*types.Transaction, error) {
	return _StakingDailyStipend.Contract.RenounceOwnership(&_StakingDailyStipend.TransactOpts)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_StakingDailyStipend *StakingDailyStipendTransactorSession) RenounceOwnership() (*types.Transaction, error) {
	return _StakingDailyStipend.Contract.RenounceOwnership(&_StakingDailyStipend.TransactOpts)
}

// ReturnStipend is a paid mutator transaction binding the contract method 0x93c87377.
//
// Solidity: function returnStipend(address to, uint256 amount) returns()
func (_StakingDailyStipend *StakingDailyStipendTransactor) ReturnStipend(opts *bind.TransactOpts, to common.Address, amount *big.Int) (*types.Transaction, error) {
	return _StakingDailyStipend.contract.Transact(opts, "returnStipend", to, amount)
}

// ReturnStipend is a paid mutator transaction binding the contract method 0x93c87377.
//
// Solidity: function returnStipend(address to, uint256 amount) returns()
func (_StakingDailyStipend *StakingDailyStipendSession) ReturnStipend(to common.Address, amount *big.Int) (*types.Transaction, error) {
	return _StakingDailyStipend.Contract.ReturnStipend(&_StakingDailyStipend.TransactOpts, to, amount)
}

// ReturnStipend is a paid mutator transaction binding the contract method 0x93c87377.
//
// Solidity: function returnStipend(address to, uint256 amount) returns()
func (_StakingDailyStipend *StakingDailyStipendTransactorSession) ReturnStipend(to common.Address, amount *big.Int) (*types.Transaction, error) {
	return _StakingDailyStipend.Contract.ReturnStipend(&_StakingDailyStipend.TransactOpts, to, amount)
}

// Stake is a paid mutator transaction binding the contract method 0xadc9772e.
//
// Solidity: function stake(address addr, uint256 amount) returns()
func (_StakingDailyStipend *StakingDailyStipendTransactor) Stake(opts *bind.TransactOpts, addr common.Address, amount *big.Int) (*types.Transaction, error) {
	return _StakingDailyStipend.contract.Transact(opts, "stake", addr, amount)
}

// Stake is a paid mutator transaction binding the contract method 0xadc9772e.
//
// Solidity: function stake(address addr, uint256 amount) returns()
func (_StakingDailyStipend *StakingDailyStipendSession) Stake(addr common.Address, amount *big.Int) (*types.Transaction, error) {
	return _StakingDailyStipend.Contract.Stake(&_StakingDailyStipend.TransactOpts, addr, amount)
}

// Stake is a paid mutator transaction binding the contract method 0xadc9772e.
//
// Solidity: function stake(address addr, uint256 amount) returns()
func (_StakingDailyStipend *StakingDailyStipendTransactorSession) Stake(addr common.Address, amount *big.Int) (*types.Transaction, error) {
	return _StakingDailyStipend.Contract.Stake(&_StakingDailyStipend.TransactOpts, addr, amount)
}

// TransferDailyStipend is a paid mutator transaction binding the contract method 0x93a7483a.
//
// Solidity: function transferDailyStipend(address from, address to, uint256 amount) returns()
func (_StakingDailyStipend *StakingDailyStipendTransactor) TransferDailyStipend(opts *bind.TransactOpts, from common.Address, to common.Address, amount *big.Int) (*types.Transaction, error) {
	return _StakingDailyStipend.contract.Transact(opts, "transferDailyStipend", from, to, amount)
}

// TransferDailyStipend is a paid mutator transaction binding the contract method 0x93a7483a.
//
// Solidity: function transferDailyStipend(address from, address to, uint256 amount) returns()
func (_StakingDailyStipend *StakingDailyStipendSession) TransferDailyStipend(from common.Address, to common.Address, amount *big.Int) (*types.Transaction, error) {
	return _StakingDailyStipend.Contract.TransferDailyStipend(&_StakingDailyStipend.TransactOpts, from, to, amount)
}

// TransferDailyStipend is a paid mutator transaction binding the contract method 0x93a7483a.
//
// Solidity: function transferDailyStipend(address from, address to, uint256 amount) returns()
func (_StakingDailyStipend *StakingDailyStipendTransactorSession) TransferDailyStipend(from common.Address, to common.Address, amount *big.Int) (*types.Transaction, error) {
	return _StakingDailyStipend.Contract.TransferDailyStipend(&_StakingDailyStipend.TransactOpts, from, to, amount)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_StakingDailyStipend *StakingDailyStipendTransactor) TransferOwnership(opts *bind.TransactOpts, newOwner common.Address) (*types.Transaction, error) {
	return _StakingDailyStipend.contract.Transact(opts, "transferOwnership", newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_StakingDailyStipend *StakingDailyStipendSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _StakingDailyStipend.Contract.TransferOwnership(&_StakingDailyStipend.TransactOpts, newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_StakingDailyStipend *StakingDailyStipendTransactorSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _StakingDailyStipend.Contract.TransferOwnership(&_StakingDailyStipend.TransactOpts, newOwner)
}

// Unstake is a paid mutator transaction binding the contract method 0x926e31d6.
//
// Solidity: function unstake(address addr, uint256 amount, address sendToAddr) returns()
func (_StakingDailyStipend *StakingDailyStipendTransactor) Unstake(opts *bind.TransactOpts, addr common.Address, amount *big.Int, sendToAddr common.Address) (*types.Transaction, error) {
	return _StakingDailyStipend.contract.Transact(opts, "unstake", addr, amount, sendToAddr)
}

// Unstake is a paid mutator transaction binding the contract method 0x926e31d6.
//
// Solidity: function unstake(address addr, uint256 amount, address sendToAddr) returns()
func (_StakingDailyStipend *StakingDailyStipendSession) Unstake(addr common.Address, amount *big.Int, sendToAddr common.Address) (*types.Transaction, error) {
	return _StakingDailyStipend.Contract.Unstake(&_StakingDailyStipend.TransactOpts, addr, amount, sendToAddr)
}

// Unstake is a paid mutator transaction binding the contract method 0x926e31d6.
//
// Solidity: function unstake(address addr, uint256 amount, address sendToAddr) returns()
func (_StakingDailyStipend *StakingDailyStipendTransactorSession) Unstake(addr common.Address, amount *big.Int, sendToAddr common.Address) (*types.Transaction, error) {
	return _StakingDailyStipend.Contract.Unstake(&_StakingDailyStipend.TransactOpts, addr, amount, sendToAddr)
}

// StakingDailyStipendInitializedIterator is returned from FilterInitialized and is used to iterate over the raw logs and unpacked data for Initialized events raised by the StakingDailyStipend contract.
type StakingDailyStipendInitializedIterator struct {
	Event *StakingDailyStipendInitialized // Event containing the contract specifics and raw log

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
func (it *StakingDailyStipendInitializedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(StakingDailyStipendInitialized)
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
		it.Event = new(StakingDailyStipendInitialized)
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
func (it *StakingDailyStipendInitializedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *StakingDailyStipendInitializedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// StakingDailyStipendInitialized represents a Initialized event raised by the StakingDailyStipend contract.
type StakingDailyStipendInitialized struct {
	Version uint8
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterInitialized is a free log retrieval operation binding the contract event 0x7f26b83ff96e1f2b6a682f133852f6798a09c465da95921460cefb3847402498.
//
// Solidity: event Initialized(uint8 version)
func (_StakingDailyStipend *StakingDailyStipendFilterer) FilterInitialized(opts *bind.FilterOpts) (*StakingDailyStipendInitializedIterator, error) {

	logs, sub, err := _StakingDailyStipend.contract.FilterLogs(opts, "Initialized")
	if err != nil {
		return nil, err
	}
	return &StakingDailyStipendInitializedIterator{contract: _StakingDailyStipend.contract, event: "Initialized", logs: logs, sub: sub}, nil
}

// WatchInitialized is a free log subscription operation binding the contract event 0x7f26b83ff96e1f2b6a682f133852f6798a09c465da95921460cefb3847402498.
//
// Solidity: event Initialized(uint8 version)
func (_StakingDailyStipend *StakingDailyStipendFilterer) WatchInitialized(opts *bind.WatchOpts, sink chan<- *StakingDailyStipendInitialized) (event.Subscription, error) {

	logs, sub, err := _StakingDailyStipend.contract.WatchLogs(opts, "Initialized")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(StakingDailyStipendInitialized)
				if err := _StakingDailyStipend.contract.UnpackLog(event, "Initialized", log); err != nil {
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
func (_StakingDailyStipend *StakingDailyStipendFilterer) ParseInitialized(log types.Log) (*StakingDailyStipendInitialized, error) {
	event := new(StakingDailyStipendInitialized)
	if err := _StakingDailyStipend.contract.UnpackLog(event, "Initialized", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// StakingDailyStipendOwnershipTransferredIterator is returned from FilterOwnershipTransferred and is used to iterate over the raw logs and unpacked data for OwnershipTransferred events raised by the StakingDailyStipend contract.
type StakingDailyStipendOwnershipTransferredIterator struct {
	Event *StakingDailyStipendOwnershipTransferred // Event containing the contract specifics and raw log

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
func (it *StakingDailyStipendOwnershipTransferredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(StakingDailyStipendOwnershipTransferred)
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
		it.Event = new(StakingDailyStipendOwnershipTransferred)
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
func (it *StakingDailyStipendOwnershipTransferredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *StakingDailyStipendOwnershipTransferredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// StakingDailyStipendOwnershipTransferred represents a OwnershipTransferred event raised by the StakingDailyStipend contract.
type StakingDailyStipendOwnershipTransferred struct {
	PreviousOwner common.Address
	NewOwner      common.Address
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterOwnershipTransferred is a free log retrieval operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_StakingDailyStipend *StakingDailyStipendFilterer) FilterOwnershipTransferred(opts *bind.FilterOpts, previousOwner []common.Address, newOwner []common.Address) (*StakingDailyStipendOwnershipTransferredIterator, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _StakingDailyStipend.contract.FilterLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return &StakingDailyStipendOwnershipTransferredIterator{contract: _StakingDailyStipend.contract, event: "OwnershipTransferred", logs: logs, sub: sub}, nil
}

// WatchOwnershipTransferred is a free log subscription operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_StakingDailyStipend *StakingDailyStipendFilterer) WatchOwnershipTransferred(opts *bind.WatchOpts, sink chan<- *StakingDailyStipendOwnershipTransferred, previousOwner []common.Address, newOwner []common.Address) (event.Subscription, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _StakingDailyStipend.contract.WatchLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(StakingDailyStipendOwnershipTransferred)
				if err := _StakingDailyStipend.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
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
func (_StakingDailyStipend *StakingDailyStipendFilterer) ParseOwnershipTransferred(log types.Log) (*StakingDailyStipendOwnershipTransferred, error) {
	event := new(StakingDailyStipendOwnershipTransferred)
	if err := _StakingDailyStipend.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// StakingDailyStipendStakedIterator is returned from FilterStaked and is used to iterate over the raw logs and unpacked data for Staked events raised by the StakingDailyStipend contract.
type StakingDailyStipendStakedIterator struct {
	Event *StakingDailyStipendStaked // Event containing the contract specifics and raw log

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
func (it *StakingDailyStipendStakedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(StakingDailyStipendStaked)
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
		it.Event = new(StakingDailyStipendStaked)
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
func (it *StakingDailyStipendStakedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *StakingDailyStipendStakedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// StakingDailyStipendStaked represents a Staked event raised by the StakingDailyStipend contract.
type StakingDailyStipendStaked struct {
	UserAddress common.Address
	Amount      *big.Int
	Raw         types.Log // Blockchain specific contextual infos
}

// FilterStaked is a free log retrieval operation binding the contract event 0x9e71bc8eea02a63969f509818f2dafb9254532904319f9dbda79b67bd34a5f3d.
//
// Solidity: event Staked(address indexed userAddress, uint256 amount)
func (_StakingDailyStipend *StakingDailyStipendFilterer) FilterStaked(opts *bind.FilterOpts, userAddress []common.Address) (*StakingDailyStipendStakedIterator, error) {

	var userAddressRule []interface{}
	for _, userAddressItem := range userAddress {
		userAddressRule = append(userAddressRule, userAddressItem)
	}

	logs, sub, err := _StakingDailyStipend.contract.FilterLogs(opts, "Staked", userAddressRule)
	if err != nil {
		return nil, err
	}
	return &StakingDailyStipendStakedIterator{contract: _StakingDailyStipend.contract, event: "Staked", logs: logs, sub: sub}, nil
}

// WatchStaked is a free log subscription operation binding the contract event 0x9e71bc8eea02a63969f509818f2dafb9254532904319f9dbda79b67bd34a5f3d.
//
// Solidity: event Staked(address indexed userAddress, uint256 amount)
func (_StakingDailyStipend *StakingDailyStipendFilterer) WatchStaked(opts *bind.WatchOpts, sink chan<- *StakingDailyStipendStaked, userAddress []common.Address) (event.Subscription, error) {

	var userAddressRule []interface{}
	for _, userAddressItem := range userAddress {
		userAddressRule = append(userAddressRule, userAddressItem)
	}

	logs, sub, err := _StakingDailyStipend.contract.WatchLogs(opts, "Staked", userAddressRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(StakingDailyStipendStaked)
				if err := _StakingDailyStipend.contract.UnpackLog(event, "Staked", log); err != nil {
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

// ParseStaked is a log parse operation binding the contract event 0x9e71bc8eea02a63969f509818f2dafb9254532904319f9dbda79b67bd34a5f3d.
//
// Solidity: event Staked(address indexed userAddress, uint256 amount)
func (_StakingDailyStipend *StakingDailyStipendFilterer) ParseStaked(log types.Log) (*StakingDailyStipendStaked, error) {
	event := new(StakingDailyStipendStaked)
	if err := _StakingDailyStipend.contract.UnpackLog(event, "Staked", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// StakingDailyStipendUnstakedIterator is returned from FilterUnstaked and is used to iterate over the raw logs and unpacked data for Unstaked events raised by the StakingDailyStipend contract.
type StakingDailyStipendUnstakedIterator struct {
	Event *StakingDailyStipendUnstaked // Event containing the contract specifics and raw log

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
func (it *StakingDailyStipendUnstakedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(StakingDailyStipendUnstaked)
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
		it.Event = new(StakingDailyStipendUnstaked)
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
func (it *StakingDailyStipendUnstakedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *StakingDailyStipendUnstakedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// StakingDailyStipendUnstaked represents a Unstaked event raised by the StakingDailyStipend contract.
type StakingDailyStipendUnstaked struct {
	UserAddress common.Address
	Amount      *big.Int
	Raw         types.Log // Blockchain specific contextual infos
}

// FilterUnstaked is a free log retrieval operation binding the contract event 0x0f5bb82176feb1b5e747e28471aa92156a04d9f3ab9f45f28e2d704232b93f75.
//
// Solidity: event Unstaked(address indexed userAddress, uint256 amount)
func (_StakingDailyStipend *StakingDailyStipendFilterer) FilterUnstaked(opts *bind.FilterOpts, userAddress []common.Address) (*StakingDailyStipendUnstakedIterator, error) {

	var userAddressRule []interface{}
	for _, userAddressItem := range userAddress {
		userAddressRule = append(userAddressRule, userAddressItem)
	}

	logs, sub, err := _StakingDailyStipend.contract.FilterLogs(opts, "Unstaked", userAddressRule)
	if err != nil {
		return nil, err
	}
	return &StakingDailyStipendUnstakedIterator{contract: _StakingDailyStipend.contract, event: "Unstaked", logs: logs, sub: sub}, nil
}

// WatchUnstaked is a free log subscription operation binding the contract event 0x0f5bb82176feb1b5e747e28471aa92156a04d9f3ab9f45f28e2d704232b93f75.
//
// Solidity: event Unstaked(address indexed userAddress, uint256 amount)
func (_StakingDailyStipend *StakingDailyStipendFilterer) WatchUnstaked(opts *bind.WatchOpts, sink chan<- *StakingDailyStipendUnstaked, userAddress []common.Address) (event.Subscription, error) {

	var userAddressRule []interface{}
	for _, userAddressItem := range userAddress {
		userAddressRule = append(userAddressRule, userAddressItem)
	}

	logs, sub, err := _StakingDailyStipend.contract.WatchLogs(opts, "Unstaked", userAddressRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(StakingDailyStipendUnstaked)
				if err := _StakingDailyStipend.contract.UnpackLog(event, "Unstaked", log); err != nil {
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

// ParseUnstaked is a log parse operation binding the contract event 0x0f5bb82176feb1b5e747e28471aa92156a04d9f3ab9f45f28e2d704232b93f75.
//
// Solidity: event Unstaked(address indexed userAddress, uint256 amount)
func (_StakingDailyStipend *StakingDailyStipendFilterer) ParseUnstaked(log types.Log) (*StakingDailyStipendUnstaked, error) {
	event := new(StakingDailyStipendUnstaked)
	if err := _StakingDailyStipend.contract.UnpackLog(event, "Unstaked", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
