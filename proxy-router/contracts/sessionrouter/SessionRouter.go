// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package sessionrouter

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

// Pool is an auto generated low-level Go binding around an user-defined struct.
type Pool struct {
	InitialReward    *big.Int
	RewardDecrease   *big.Int
	PayoutStart      *big.Int
	DecreaseInterval *big.Int
}

// Session is an auto generated low-level Go binding around an user-defined struct.
type Session struct {
	Id                      [32]byte
	User                    common.Address
	Provider                common.Address
	ModelAgentId            [32]byte
	BidID                   [32]byte
	Stake                   *big.Int
	PricePerSecond          *big.Int
	CloseoutReceipt         []byte
	CloseoutType            *big.Int
	ProviderWithdrawnAmount *big.Int
	OpenedAt                *big.Int
	ClosedAt                *big.Int
}

// SessionRouterMetaData contains all meta data concerning the SessionRouter contract.
var SessionRouterMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[],\"name\":\"BidNotFound\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"BidTaken\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"ECDSAInvalidSignature\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"length\",\"type\":\"uint256\"}],\"name\":\"ECDSAInvalidSignatureLength\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"s\",\"type\":\"bytes32\"}],\"name\":\"ECDSAInvalidSignatureS\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_user\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"_contractOwner\",\"type\":\"address\"}],\"name\":\"NotContractOwner\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"NotEnoughWithdrawableBalance\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"NotSenderOrOwner\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"NotUserOrProvider\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"SessionAlreadyClosed\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"SessionNotFound\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"SessionTooShort\",\"type\":\"error\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"providerAddress\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"ProviderClaimed\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"userAddress\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"sessionId\",\"type\":\"bytes32\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"providerId\",\"type\":\"address\"}],\"name\":\"SessionClosed\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"userAddress\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"sessionId\",\"type\":\"bytes32\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"providerId\",\"type\":\"address\"}],\"name\":\"SessionOpened\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"userAddress\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"Staked\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"userAddress\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"Unstaked\",\"type\":\"event\"},{\"inputs\":[],\"name\":\"DAY\",\"outputs\":[{\"internalType\":\"uint32\",\"name\":\"\",\"type\":\"uint32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"MIN_SESSION_DURATION\",\"outputs\":[{\"internalType\":\"uint32\",\"name\":\"\",\"type\":\"uint32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"sessionId\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"amountToWithdraw\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"}],\"name\":\"claimProviderBalance\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"sessionId\",\"type\":\"bytes32\"},{\"internalType\":\"bytes\",\"name\":\"receiptEncoded\",\"type\":\"bytes\"},{\"internalType\":\"bytes\",\"name\":\"signature\",\"type\":\"bytes\"}],\"name\":\"closeSession\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"sessionId\",\"type\":\"bytes32\"}],\"name\":\"deleteHistory\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"timestamp\",\"type\":\"uint256\"}],\"name\":\"getComputeBalance\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"sessionId\",\"type\":\"bytes32\"}],\"name\":\"getProviderClaimableBalance\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"sessionId\",\"type\":\"bytes32\"}],\"name\":\"getSession\",\"outputs\":[{\"components\":[{\"internalType\":\"bytes32\",\"name\":\"id\",\"type\":\"bytes32\"},{\"internalType\":\"address\",\"name\":\"user\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"provider\",\"type\":\"address\"},{\"internalType\":\"bytes32\",\"name\":\"modelAgentId\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"bidID\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"stake\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"pricePerSecond\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"closeoutReceipt\",\"type\":\"bytes\"},{\"internalType\":\"uint256\",\"name\":\"closeoutType\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"providerWithdrawnAmount\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"openedAt\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"closedAt\",\"type\":\"uint256\"}],\"internalType\":\"structSession\",\"name\":\"\",\"type\":\"tuple\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"sessionId\",\"type\":\"bytes32\"}],\"name\":\"getSessionEndTime\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getTodaysBudget\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"signer\",\"type\":\"address\"},{\"internalType\":\"bytes\",\"name\":\"receipt\",\"type\":\"bytes\"},{\"internalType\":\"bytes\",\"name\":\"signature\",\"type\":\"bytes\"}],\"name\":\"isValidReceipt\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"pure\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"bidId\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"_stake\",\"type\":\"uint256\"}],\"name\":\"openSession\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"sessionId\",\"type\":\"bytes32\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"components\":[{\"internalType\":\"uint256\",\"name\":\"initialReward\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"rewardDecrease\",\"type\":\"uint256\"},{\"internalType\":\"uint128\",\"name\":\"payoutStart\",\"type\":\"uint128\"},{\"internalType\":\"uint128\",\"name\":\"decreaseInterval\",\"type\":\"uint128\"}],\"internalType\":\"structPool\",\"name\":\"pool\",\"type\":\"tuple\"}],\"name\":\"setPoolConfig\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"int256\",\"name\":\"delay\",\"type\":\"int256\"}],\"name\":\"setStakeDelay\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"timestamp\",\"type\":\"uint256\"}],\"name\":\"startOfTheDay\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"pure\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"amountToWithdraw\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"}],\"name\":\"withdrawUserStake\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"userAddr\",\"type\":\"address\"}],\"name\":\"withdrawableUserStake\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"avail\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"hold\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"}]",
}

// SessionRouterABI is the input ABI used to generate the binding from.
// Deprecated: Use SessionRouterMetaData.ABI instead.
var SessionRouterABI = SessionRouterMetaData.ABI

// SessionRouter is an auto generated Go binding around an Ethereum contract.
type SessionRouter struct {
	SessionRouterCaller     // Read-only binding to the contract
	SessionRouterTransactor // Write-only binding to the contract
	SessionRouterFilterer   // Log filterer for contract events
}

// SessionRouterCaller is an auto generated read-only Go binding around an Ethereum contract.
type SessionRouterCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// SessionRouterTransactor is an auto generated write-only Go binding around an Ethereum contract.
type SessionRouterTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// SessionRouterFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type SessionRouterFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// SessionRouterSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type SessionRouterSession struct {
	Contract     *SessionRouter    // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// SessionRouterCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type SessionRouterCallerSession struct {
	Contract *SessionRouterCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts        // Call options to use throughout this session
}

// SessionRouterTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type SessionRouterTransactorSession struct {
	Contract     *SessionRouterTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts        // Transaction auth options to use throughout this session
}

// SessionRouterRaw is an auto generated low-level Go binding around an Ethereum contract.
type SessionRouterRaw struct {
	Contract *SessionRouter // Generic contract binding to access the raw methods on
}

// SessionRouterCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type SessionRouterCallerRaw struct {
	Contract *SessionRouterCaller // Generic read-only contract binding to access the raw methods on
}

// SessionRouterTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type SessionRouterTransactorRaw struct {
	Contract *SessionRouterTransactor // Generic write-only contract binding to access the raw methods on
}

// NewSessionRouter creates a new instance of SessionRouter, bound to a specific deployed contract.
func NewSessionRouter(address common.Address, backend bind.ContractBackend) (*SessionRouter, error) {
	contract, err := bindSessionRouter(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &SessionRouter{SessionRouterCaller: SessionRouterCaller{contract: contract}, SessionRouterTransactor: SessionRouterTransactor{contract: contract}, SessionRouterFilterer: SessionRouterFilterer{contract: contract}}, nil
}

// NewSessionRouterCaller creates a new read-only instance of SessionRouter, bound to a specific deployed contract.
func NewSessionRouterCaller(address common.Address, caller bind.ContractCaller) (*SessionRouterCaller, error) {
	contract, err := bindSessionRouter(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &SessionRouterCaller{contract: contract}, nil
}

// NewSessionRouterTransactor creates a new write-only instance of SessionRouter, bound to a specific deployed contract.
func NewSessionRouterTransactor(address common.Address, transactor bind.ContractTransactor) (*SessionRouterTransactor, error) {
	contract, err := bindSessionRouter(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &SessionRouterTransactor{contract: contract}, nil
}

// NewSessionRouterFilterer creates a new log filterer instance of SessionRouter, bound to a specific deployed contract.
func NewSessionRouterFilterer(address common.Address, filterer bind.ContractFilterer) (*SessionRouterFilterer, error) {
	contract, err := bindSessionRouter(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &SessionRouterFilterer{contract: contract}, nil
}

// bindSessionRouter binds a generic wrapper to an already deployed contract.
func bindSessionRouter(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := SessionRouterMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_SessionRouter *SessionRouterRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _SessionRouter.Contract.SessionRouterCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_SessionRouter *SessionRouterRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _SessionRouter.Contract.SessionRouterTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_SessionRouter *SessionRouterRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _SessionRouter.Contract.SessionRouterTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_SessionRouter *SessionRouterCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _SessionRouter.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_SessionRouter *SessionRouterTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _SessionRouter.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_SessionRouter *SessionRouterTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _SessionRouter.Contract.contract.Transact(opts, method, params...)
}

// DAY is a free data retrieval call binding the contract method 0x27cfe856.
//
// Solidity: function DAY() view returns(uint32)
func (_SessionRouter *SessionRouterCaller) DAY(opts *bind.CallOpts) (uint32, error) {
	var out []interface{}
	err := _SessionRouter.contract.Call(opts, &out, "DAY")

	if err != nil {
		return *new(uint32), err
	}

	out0 := *abi.ConvertType(out[0], new(uint32)).(*uint32)

	return out0, err

}

// DAY is a free data retrieval call binding the contract method 0x27cfe856.
//
// Solidity: function DAY() view returns(uint32)
func (_SessionRouter *SessionRouterSession) DAY() (uint32, error) {
	return _SessionRouter.Contract.DAY(&_SessionRouter.CallOpts)
}

// DAY is a free data retrieval call binding the contract method 0x27cfe856.
//
// Solidity: function DAY() view returns(uint32)
func (_SessionRouter *SessionRouterCallerSession) DAY() (uint32, error) {
	return _SessionRouter.Contract.DAY(&_SessionRouter.CallOpts)
}

// MINSESSIONDURATION is a free data retrieval call binding the contract method 0x7d980286.
//
// Solidity: function MIN_SESSION_DURATION() view returns(uint32)
func (_SessionRouter *SessionRouterCaller) MINSESSIONDURATION(opts *bind.CallOpts) (uint32, error) {
	var out []interface{}
	err := _SessionRouter.contract.Call(opts, &out, "MIN_SESSION_DURATION")

	if err != nil {
		return *new(uint32), err
	}

	out0 := *abi.ConvertType(out[0], new(uint32)).(*uint32)

	return out0, err

}

// MINSESSIONDURATION is a free data retrieval call binding the contract method 0x7d980286.
//
// Solidity: function MIN_SESSION_DURATION() view returns(uint32)
func (_SessionRouter *SessionRouterSession) MINSESSIONDURATION() (uint32, error) {
	return _SessionRouter.Contract.MINSESSIONDURATION(&_SessionRouter.CallOpts)
}

// MINSESSIONDURATION is a free data retrieval call binding the contract method 0x7d980286.
//
// Solidity: function MIN_SESSION_DURATION() view returns(uint32)
func (_SessionRouter *SessionRouterCallerSession) MINSESSIONDURATION() (uint32, error) {
	return _SessionRouter.Contract.MINSESSIONDURATION(&_SessionRouter.CallOpts)
}

// GetComputeBalance is a free data retrieval call binding the contract method 0x76738e9e.
//
// Solidity: function getComputeBalance(uint256 timestamp) view returns(uint256)
func (_SessionRouter *SessionRouterCaller) GetComputeBalance(opts *bind.CallOpts, timestamp *big.Int) (*big.Int, error) {
	var out []interface{}
	err := _SessionRouter.contract.Call(opts, &out, "getComputeBalance", timestamp)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetComputeBalance is a free data retrieval call binding the contract method 0x76738e9e.
//
// Solidity: function getComputeBalance(uint256 timestamp) view returns(uint256)
func (_SessionRouter *SessionRouterSession) GetComputeBalance(timestamp *big.Int) (*big.Int, error) {
	return _SessionRouter.Contract.GetComputeBalance(&_SessionRouter.CallOpts, timestamp)
}

// GetComputeBalance is a free data retrieval call binding the contract method 0x76738e9e.
//
// Solidity: function getComputeBalance(uint256 timestamp) view returns(uint256)
func (_SessionRouter *SessionRouterCallerSession) GetComputeBalance(timestamp *big.Int) (*big.Int, error) {
	return _SessionRouter.Contract.GetComputeBalance(&_SessionRouter.CallOpts, timestamp)
}

// GetProviderClaimableBalance is a free data retrieval call binding the contract method 0xa8ca6323.
//
// Solidity: function getProviderClaimableBalance(bytes32 sessionId) view returns(uint256)
func (_SessionRouter *SessionRouterCaller) GetProviderClaimableBalance(opts *bind.CallOpts, sessionId [32]byte) (*big.Int, error) {
	var out []interface{}
	err := _SessionRouter.contract.Call(opts, &out, "getProviderClaimableBalance", sessionId)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetProviderClaimableBalance is a free data retrieval call binding the contract method 0xa8ca6323.
//
// Solidity: function getProviderClaimableBalance(bytes32 sessionId) view returns(uint256)
func (_SessionRouter *SessionRouterSession) GetProviderClaimableBalance(sessionId [32]byte) (*big.Int, error) {
	return _SessionRouter.Contract.GetProviderClaimableBalance(&_SessionRouter.CallOpts, sessionId)
}

// GetProviderClaimableBalance is a free data retrieval call binding the contract method 0xa8ca6323.
//
// Solidity: function getProviderClaimableBalance(bytes32 sessionId) view returns(uint256)
func (_SessionRouter *SessionRouterCallerSession) GetProviderClaimableBalance(sessionId [32]byte) (*big.Int, error) {
	return _SessionRouter.Contract.GetProviderClaimableBalance(&_SessionRouter.CallOpts, sessionId)
}

// GetSession is a free data retrieval call binding the contract method 0x39b240bd.
//
// Solidity: function getSession(bytes32 sessionId) view returns((bytes32,address,address,bytes32,bytes32,uint256,uint256,bytes,uint256,uint256,uint256,uint256))
func (_SessionRouter *SessionRouterCaller) GetSession(opts *bind.CallOpts, sessionId [32]byte) (Session, error) {
	var out []interface{}
	err := _SessionRouter.contract.Call(opts, &out, "getSession", sessionId)

	if err != nil {
		return *new(Session), err
	}

	out0 := *abi.ConvertType(out[0], new(Session)).(*Session)

	return out0, err

}

// GetSession is a free data retrieval call binding the contract method 0x39b240bd.
//
// Solidity: function getSession(bytes32 sessionId) view returns((bytes32,address,address,bytes32,bytes32,uint256,uint256,bytes,uint256,uint256,uint256,uint256))
func (_SessionRouter *SessionRouterSession) GetSession(sessionId [32]byte) (Session, error) {
	return _SessionRouter.Contract.GetSession(&_SessionRouter.CallOpts, sessionId)
}

// GetSession is a free data retrieval call binding the contract method 0x39b240bd.
//
// Solidity: function getSession(bytes32 sessionId) view returns((bytes32,address,address,bytes32,bytes32,uint256,uint256,bytes,uint256,uint256,uint256,uint256))
func (_SessionRouter *SessionRouterCallerSession) GetSession(sessionId [32]byte) (Session, error) {
	return _SessionRouter.Contract.GetSession(&_SessionRouter.CallOpts, sessionId)
}

// GetSessionEndTime is a free data retrieval call binding the contract method 0x3a02141b.
//
// Solidity: function getSessionEndTime(bytes32 sessionId) view returns(uint256)
func (_SessionRouter *SessionRouterCaller) GetSessionEndTime(opts *bind.CallOpts, sessionId [32]byte) (*big.Int, error) {
	var out []interface{}
	err := _SessionRouter.contract.Call(opts, &out, "getSessionEndTime", sessionId)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetSessionEndTime is a free data retrieval call binding the contract method 0x3a02141b.
//
// Solidity: function getSessionEndTime(bytes32 sessionId) view returns(uint256)
func (_SessionRouter *SessionRouterSession) GetSessionEndTime(sessionId [32]byte) (*big.Int, error) {
	return _SessionRouter.Contract.GetSessionEndTime(&_SessionRouter.CallOpts, sessionId)
}

// GetSessionEndTime is a free data retrieval call binding the contract method 0x3a02141b.
//
// Solidity: function getSessionEndTime(bytes32 sessionId) view returns(uint256)
func (_SessionRouter *SessionRouterCallerSession) GetSessionEndTime(sessionId [32]byte) (*big.Int, error) {
	return _SessionRouter.Contract.GetSessionEndTime(&_SessionRouter.CallOpts, sessionId)
}

// GetTodaysBudget is a free data retrieval call binding the contract method 0xa7e7f9a9.
//
// Solidity: function getTodaysBudget() view returns(uint256)
func (_SessionRouter *SessionRouterCaller) GetTodaysBudget(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _SessionRouter.contract.Call(opts, &out, "getTodaysBudget")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetTodaysBudget is a free data retrieval call binding the contract method 0xa7e7f9a9.
//
// Solidity: function getTodaysBudget() view returns(uint256)
func (_SessionRouter *SessionRouterSession) GetTodaysBudget() (*big.Int, error) {
	return _SessionRouter.Contract.GetTodaysBudget(&_SessionRouter.CallOpts)
}

// GetTodaysBudget is a free data retrieval call binding the contract method 0xa7e7f9a9.
//
// Solidity: function getTodaysBudget() view returns(uint256)
func (_SessionRouter *SessionRouterCallerSession) GetTodaysBudget() (*big.Int, error) {
	return _SessionRouter.Contract.GetTodaysBudget(&_SessionRouter.CallOpts)
}

// IsValidReceipt is a free data retrieval call binding the contract method 0x626dd729.
//
// Solidity: function isValidReceipt(address signer, bytes receipt, bytes signature) pure returns(bool)
func (_SessionRouter *SessionRouterCaller) IsValidReceipt(opts *bind.CallOpts, signer common.Address, receipt []byte, signature []byte) (bool, error) {
	var out []interface{}
	err := _SessionRouter.contract.Call(opts, &out, "isValidReceipt", signer, receipt, signature)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsValidReceipt is a free data retrieval call binding the contract method 0x626dd729.
//
// Solidity: function isValidReceipt(address signer, bytes receipt, bytes signature) pure returns(bool)
func (_SessionRouter *SessionRouterSession) IsValidReceipt(signer common.Address, receipt []byte, signature []byte) (bool, error) {
	return _SessionRouter.Contract.IsValidReceipt(&_SessionRouter.CallOpts, signer, receipt, signature)
}

// IsValidReceipt is a free data retrieval call binding the contract method 0x626dd729.
//
// Solidity: function isValidReceipt(address signer, bytes receipt, bytes signature) pure returns(bool)
func (_SessionRouter *SessionRouterCallerSession) IsValidReceipt(signer common.Address, receipt []byte, signature []byte) (bool, error) {
	return _SessionRouter.Contract.IsValidReceipt(&_SessionRouter.CallOpts, signer, receipt, signature)
}

// StartOfTheDay is a free data retrieval call binding the contract method 0xeedd0a72.
//
// Solidity: function startOfTheDay(uint256 timestamp) pure returns(uint256)
func (_SessionRouter *SessionRouterCaller) StartOfTheDay(opts *bind.CallOpts, timestamp *big.Int) (*big.Int, error) {
	var out []interface{}
	err := _SessionRouter.contract.Call(opts, &out, "startOfTheDay", timestamp)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// StartOfTheDay is a free data retrieval call binding the contract method 0xeedd0a72.
//
// Solidity: function startOfTheDay(uint256 timestamp) pure returns(uint256)
func (_SessionRouter *SessionRouterSession) StartOfTheDay(timestamp *big.Int) (*big.Int, error) {
	return _SessionRouter.Contract.StartOfTheDay(&_SessionRouter.CallOpts, timestamp)
}

// StartOfTheDay is a free data retrieval call binding the contract method 0xeedd0a72.
//
// Solidity: function startOfTheDay(uint256 timestamp) pure returns(uint256)
func (_SessionRouter *SessionRouterCallerSession) StartOfTheDay(timestamp *big.Int) (*big.Int, error) {
	return _SessionRouter.Contract.StartOfTheDay(&_SessionRouter.CallOpts, timestamp)
}

// WithdrawableUserStake is a free data retrieval call binding the contract method 0x536f1f82.
//
// Solidity: function withdrawableUserStake(address userAddr) view returns(uint256 avail, uint256 hold)
func (_SessionRouter *SessionRouterCaller) WithdrawableUserStake(opts *bind.CallOpts, userAddr common.Address) (struct {
	Avail *big.Int
	Hold  *big.Int
}, error) {
	var out []interface{}
	err := _SessionRouter.contract.Call(opts, &out, "withdrawableUserStake", userAddr)

	outstruct := new(struct {
		Avail *big.Int
		Hold  *big.Int
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.Avail = *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)
	outstruct.Hold = *abi.ConvertType(out[1], new(*big.Int)).(**big.Int)

	return *outstruct, err

}

// WithdrawableUserStake is a free data retrieval call binding the contract method 0x536f1f82.
//
// Solidity: function withdrawableUserStake(address userAddr) view returns(uint256 avail, uint256 hold)
func (_SessionRouter *SessionRouterSession) WithdrawableUserStake(userAddr common.Address) (struct {
	Avail *big.Int
	Hold  *big.Int
}, error) {
	return _SessionRouter.Contract.WithdrawableUserStake(&_SessionRouter.CallOpts, userAddr)
}

// WithdrawableUserStake is a free data retrieval call binding the contract method 0x536f1f82.
//
// Solidity: function withdrawableUserStake(address userAddr) view returns(uint256 avail, uint256 hold)
func (_SessionRouter *SessionRouterCallerSession) WithdrawableUserStake(userAddr common.Address) (struct {
	Avail *big.Int
	Hold  *big.Int
}, error) {
	return _SessionRouter.Contract.WithdrawableUserStake(&_SessionRouter.CallOpts, userAddr)
}

// ClaimProviderBalance is a paid mutator transaction binding the contract method 0xbab3de02.
//
// Solidity: function claimProviderBalance(bytes32 sessionId, uint256 amountToWithdraw, address to) returns()
func (_SessionRouter *SessionRouterTransactor) ClaimProviderBalance(opts *bind.TransactOpts, sessionId [32]byte, amountToWithdraw *big.Int, to common.Address) (*types.Transaction, error) {
	return _SessionRouter.contract.Transact(opts, "claimProviderBalance", sessionId, amountToWithdraw, to)
}

// ClaimProviderBalance is a paid mutator transaction binding the contract method 0xbab3de02.
//
// Solidity: function claimProviderBalance(bytes32 sessionId, uint256 amountToWithdraw, address to) returns()
func (_SessionRouter *SessionRouterSession) ClaimProviderBalance(sessionId [32]byte, amountToWithdraw *big.Int, to common.Address) (*types.Transaction, error) {
	return _SessionRouter.Contract.ClaimProviderBalance(&_SessionRouter.TransactOpts, sessionId, amountToWithdraw, to)
}

// ClaimProviderBalance is a paid mutator transaction binding the contract method 0xbab3de02.
//
// Solidity: function claimProviderBalance(bytes32 sessionId, uint256 amountToWithdraw, address to) returns()
func (_SessionRouter *SessionRouterTransactorSession) ClaimProviderBalance(sessionId [32]byte, amountToWithdraw *big.Int, to common.Address) (*types.Transaction, error) {
	return _SessionRouter.Contract.ClaimProviderBalance(&_SessionRouter.TransactOpts, sessionId, amountToWithdraw, to)
}

// CloseSession is a paid mutator transaction binding the contract method 0x9775d1ff.
//
// Solidity: function closeSession(bytes32 sessionId, bytes receiptEncoded, bytes signature) returns()
func (_SessionRouter *SessionRouterTransactor) CloseSession(opts *bind.TransactOpts, sessionId [32]byte, receiptEncoded []byte, signature []byte) (*types.Transaction, error) {
	return _SessionRouter.contract.Transact(opts, "closeSession", sessionId, receiptEncoded, signature)
}

// CloseSession is a paid mutator transaction binding the contract method 0x9775d1ff.
//
// Solidity: function closeSession(bytes32 sessionId, bytes receiptEncoded, bytes signature) returns()
func (_SessionRouter *SessionRouterSession) CloseSession(sessionId [32]byte, receiptEncoded []byte, signature []byte) (*types.Transaction, error) {
	return _SessionRouter.Contract.CloseSession(&_SessionRouter.TransactOpts, sessionId, receiptEncoded, signature)
}

// CloseSession is a paid mutator transaction binding the contract method 0x9775d1ff.
//
// Solidity: function closeSession(bytes32 sessionId, bytes receiptEncoded, bytes signature) returns()
func (_SessionRouter *SessionRouterTransactorSession) CloseSession(sessionId [32]byte, receiptEncoded []byte, signature []byte) (*types.Transaction, error) {
	return _SessionRouter.Contract.CloseSession(&_SessionRouter.TransactOpts, sessionId, receiptEncoded, signature)
}

// DeleteHistory is a paid mutator transaction binding the contract method 0xf074ca6b.
//
// Solidity: function deleteHistory(bytes32 sessionId) returns()
func (_SessionRouter *SessionRouterTransactor) DeleteHistory(opts *bind.TransactOpts, sessionId [32]byte) (*types.Transaction, error) {
	return _SessionRouter.contract.Transact(opts, "deleteHistory", sessionId)
}

// DeleteHistory is a paid mutator transaction binding the contract method 0xf074ca6b.
//
// Solidity: function deleteHistory(bytes32 sessionId) returns()
func (_SessionRouter *SessionRouterSession) DeleteHistory(sessionId [32]byte) (*types.Transaction, error) {
	return _SessionRouter.Contract.DeleteHistory(&_SessionRouter.TransactOpts, sessionId)
}

// DeleteHistory is a paid mutator transaction binding the contract method 0xf074ca6b.
//
// Solidity: function deleteHistory(bytes32 sessionId) returns()
func (_SessionRouter *SessionRouterTransactorSession) DeleteHistory(sessionId [32]byte) (*types.Transaction, error) {
	return _SessionRouter.Contract.DeleteHistory(&_SessionRouter.TransactOpts, sessionId)
}

// OpenSession is a paid mutator transaction binding the contract method 0x48c00c90.
//
// Solidity: function openSession(bytes32 bidId, uint256 _stake) returns(bytes32 sessionId)
func (_SessionRouter *SessionRouterTransactor) OpenSession(opts *bind.TransactOpts, bidId [32]byte, _stake *big.Int) (*types.Transaction, error) {
	return _SessionRouter.contract.Transact(opts, "openSession", bidId, _stake)
}

// OpenSession is a paid mutator transaction binding the contract method 0x48c00c90.
//
// Solidity: function openSession(bytes32 bidId, uint256 _stake) returns(bytes32 sessionId)
func (_SessionRouter *SessionRouterSession) OpenSession(bidId [32]byte, _stake *big.Int) (*types.Transaction, error) {
	return _SessionRouter.Contract.OpenSession(&_SessionRouter.TransactOpts, bidId, _stake)
}

// OpenSession is a paid mutator transaction binding the contract method 0x48c00c90.
//
// Solidity: function openSession(bytes32 bidId, uint256 _stake) returns(bytes32 sessionId)
func (_SessionRouter *SessionRouterTransactorSession) OpenSession(bidId [32]byte, _stake *big.Int) (*types.Transaction, error) {
	return _SessionRouter.Contract.OpenSession(&_SessionRouter.TransactOpts, bidId, _stake)
}

// SetPoolConfig is a paid mutator transaction binding the contract method 0x8b1af52a.
//
// Solidity: function setPoolConfig((uint256,uint256,uint128,uint128) pool) returns()
func (_SessionRouter *SessionRouterTransactor) SetPoolConfig(opts *bind.TransactOpts, pool Pool) (*types.Transaction, error) {
	return _SessionRouter.contract.Transact(opts, "setPoolConfig", pool)
}

// SetPoolConfig is a paid mutator transaction binding the contract method 0x8b1af52a.
//
// Solidity: function setPoolConfig((uint256,uint256,uint128,uint128) pool) returns()
func (_SessionRouter *SessionRouterSession) SetPoolConfig(pool Pool) (*types.Transaction, error) {
	return _SessionRouter.Contract.SetPoolConfig(&_SessionRouter.TransactOpts, pool)
}

// SetPoolConfig is a paid mutator transaction binding the contract method 0x8b1af52a.
//
// Solidity: function setPoolConfig((uint256,uint256,uint128,uint128) pool) returns()
func (_SessionRouter *SessionRouterTransactorSession) SetPoolConfig(pool Pool) (*types.Transaction, error) {
	return _SessionRouter.Contract.SetPoolConfig(&_SessionRouter.TransactOpts, pool)
}

// SetStakeDelay is a paid mutator transaction binding the contract method 0x3cadd8bb.
//
// Solidity: function setStakeDelay(int256 delay) returns()
func (_SessionRouter *SessionRouterTransactor) SetStakeDelay(opts *bind.TransactOpts, delay *big.Int) (*types.Transaction, error) {
	return _SessionRouter.contract.Transact(opts, "setStakeDelay", delay)
}

// SetStakeDelay is a paid mutator transaction binding the contract method 0x3cadd8bb.
//
// Solidity: function setStakeDelay(int256 delay) returns()
func (_SessionRouter *SessionRouterSession) SetStakeDelay(delay *big.Int) (*types.Transaction, error) {
	return _SessionRouter.Contract.SetStakeDelay(&_SessionRouter.TransactOpts, delay)
}

// SetStakeDelay is a paid mutator transaction binding the contract method 0x3cadd8bb.
//
// Solidity: function setStakeDelay(int256 delay) returns()
func (_SessionRouter *SessionRouterTransactorSession) SetStakeDelay(delay *big.Int) (*types.Transaction, error) {
	return _SessionRouter.Contract.SetStakeDelay(&_SessionRouter.TransactOpts, delay)
}

// WithdrawUserStake is a paid mutator transaction binding the contract method 0xcd308cb1.
//
// Solidity: function withdrawUserStake(uint256 amountToWithdraw, address to) returns()
func (_SessionRouter *SessionRouterTransactor) WithdrawUserStake(opts *bind.TransactOpts, amountToWithdraw *big.Int, to common.Address) (*types.Transaction, error) {
	return _SessionRouter.contract.Transact(opts, "withdrawUserStake", amountToWithdraw, to)
}

// WithdrawUserStake is a paid mutator transaction binding the contract method 0xcd308cb1.
//
// Solidity: function withdrawUserStake(uint256 amountToWithdraw, address to) returns()
func (_SessionRouter *SessionRouterSession) WithdrawUserStake(amountToWithdraw *big.Int, to common.Address) (*types.Transaction, error) {
	return _SessionRouter.Contract.WithdrawUserStake(&_SessionRouter.TransactOpts, amountToWithdraw, to)
}

// WithdrawUserStake is a paid mutator transaction binding the contract method 0xcd308cb1.
//
// Solidity: function withdrawUserStake(uint256 amountToWithdraw, address to) returns()
func (_SessionRouter *SessionRouterTransactorSession) WithdrawUserStake(amountToWithdraw *big.Int, to common.Address) (*types.Transaction, error) {
	return _SessionRouter.Contract.WithdrawUserStake(&_SessionRouter.TransactOpts, amountToWithdraw, to)
}

// SessionRouterProviderClaimedIterator is returned from FilterProviderClaimed and is used to iterate over the raw logs and unpacked data for ProviderClaimed events raised by the SessionRouter contract.
type SessionRouterProviderClaimedIterator struct {
	Event *SessionRouterProviderClaimed // Event containing the contract specifics and raw log

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
func (it *SessionRouterProviderClaimedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SessionRouterProviderClaimed)
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
		it.Event = new(SessionRouterProviderClaimed)
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
func (it *SessionRouterProviderClaimedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SessionRouterProviderClaimedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SessionRouterProviderClaimed represents a ProviderClaimed event raised by the SessionRouter contract.
type SessionRouterProviderClaimed struct {
	ProviderAddress common.Address
	Amount          *big.Int
	Raw             types.Log // Blockchain specific contextual infos
}

// FilterProviderClaimed is a free log retrieval operation binding the contract event 0x1cd322e3d02eade120b8dceb43a6c1dee437af36e7acd81726c4b54adf5584c2.
//
// Solidity: event ProviderClaimed(address indexed providerAddress, uint256 amount)
func (_SessionRouter *SessionRouterFilterer) FilterProviderClaimed(opts *bind.FilterOpts, providerAddress []common.Address) (*SessionRouterProviderClaimedIterator, error) {

	var providerAddressRule []interface{}
	for _, providerAddressItem := range providerAddress {
		providerAddressRule = append(providerAddressRule, providerAddressItem)
	}

	logs, sub, err := _SessionRouter.contract.FilterLogs(opts, "ProviderClaimed", providerAddressRule)
	if err != nil {
		return nil, err
	}
	return &SessionRouterProviderClaimedIterator{contract: _SessionRouter.contract, event: "ProviderClaimed", logs: logs, sub: sub}, nil
}

// WatchProviderClaimed is a free log subscription operation binding the contract event 0x1cd322e3d02eade120b8dceb43a6c1dee437af36e7acd81726c4b54adf5584c2.
//
// Solidity: event ProviderClaimed(address indexed providerAddress, uint256 amount)
func (_SessionRouter *SessionRouterFilterer) WatchProviderClaimed(opts *bind.WatchOpts, sink chan<- *SessionRouterProviderClaimed, providerAddress []common.Address) (event.Subscription, error) {

	var providerAddressRule []interface{}
	for _, providerAddressItem := range providerAddress {
		providerAddressRule = append(providerAddressRule, providerAddressItem)
	}

	logs, sub, err := _SessionRouter.contract.WatchLogs(opts, "ProviderClaimed", providerAddressRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SessionRouterProviderClaimed)
				if err := _SessionRouter.contract.UnpackLog(event, "ProviderClaimed", log); err != nil {
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

// ParseProviderClaimed is a log parse operation binding the contract event 0x1cd322e3d02eade120b8dceb43a6c1dee437af36e7acd81726c4b54adf5584c2.
//
// Solidity: event ProviderClaimed(address indexed providerAddress, uint256 amount)
func (_SessionRouter *SessionRouterFilterer) ParseProviderClaimed(log types.Log) (*SessionRouterProviderClaimed, error) {
	event := new(SessionRouterProviderClaimed)
	if err := _SessionRouter.contract.UnpackLog(event, "ProviderClaimed", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// SessionRouterSessionClosedIterator is returned from FilterSessionClosed and is used to iterate over the raw logs and unpacked data for SessionClosed events raised by the SessionRouter contract.
type SessionRouterSessionClosedIterator struct {
	Event *SessionRouterSessionClosed // Event containing the contract specifics and raw log

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
func (it *SessionRouterSessionClosedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SessionRouterSessionClosed)
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
		it.Event = new(SessionRouterSessionClosed)
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
func (it *SessionRouterSessionClosedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SessionRouterSessionClosedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SessionRouterSessionClosed represents a SessionClosed event raised by the SessionRouter contract.
type SessionRouterSessionClosed struct {
	UserAddress common.Address
	SessionId   [32]byte
	ProviderId  common.Address
	Raw         types.Log // Blockchain specific contextual infos
}

// FilterSessionClosed is a free log retrieval operation binding the contract event 0x337fbb0a41a596db800dc836595a57815f967185e3596615c646f2455ac3914a.
//
// Solidity: event SessionClosed(address indexed userAddress, bytes32 indexed sessionId, address indexed providerId)
func (_SessionRouter *SessionRouterFilterer) FilterSessionClosed(opts *bind.FilterOpts, userAddress []common.Address, sessionId [][32]byte, providerId []common.Address) (*SessionRouterSessionClosedIterator, error) {

	var userAddressRule []interface{}
	for _, userAddressItem := range userAddress {
		userAddressRule = append(userAddressRule, userAddressItem)
	}
	var sessionIdRule []interface{}
	for _, sessionIdItem := range sessionId {
		sessionIdRule = append(sessionIdRule, sessionIdItem)
	}
	var providerIdRule []interface{}
	for _, providerIdItem := range providerId {
		providerIdRule = append(providerIdRule, providerIdItem)
	}

	logs, sub, err := _SessionRouter.contract.FilterLogs(opts, "SessionClosed", userAddressRule, sessionIdRule, providerIdRule)
	if err != nil {
		return nil, err
	}
	return &SessionRouterSessionClosedIterator{contract: _SessionRouter.contract, event: "SessionClosed", logs: logs, sub: sub}, nil
}

// WatchSessionClosed is a free log subscription operation binding the contract event 0x337fbb0a41a596db800dc836595a57815f967185e3596615c646f2455ac3914a.
//
// Solidity: event SessionClosed(address indexed userAddress, bytes32 indexed sessionId, address indexed providerId)
func (_SessionRouter *SessionRouterFilterer) WatchSessionClosed(opts *bind.WatchOpts, sink chan<- *SessionRouterSessionClosed, userAddress []common.Address, sessionId [][32]byte, providerId []common.Address) (event.Subscription, error) {

	var userAddressRule []interface{}
	for _, userAddressItem := range userAddress {
		userAddressRule = append(userAddressRule, userAddressItem)
	}
	var sessionIdRule []interface{}
	for _, sessionIdItem := range sessionId {
		sessionIdRule = append(sessionIdRule, sessionIdItem)
	}
	var providerIdRule []interface{}
	for _, providerIdItem := range providerId {
		providerIdRule = append(providerIdRule, providerIdItem)
	}

	logs, sub, err := _SessionRouter.contract.WatchLogs(opts, "SessionClosed", userAddressRule, sessionIdRule, providerIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SessionRouterSessionClosed)
				if err := _SessionRouter.contract.UnpackLog(event, "SessionClosed", log); err != nil {
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

// ParseSessionClosed is a log parse operation binding the contract event 0x337fbb0a41a596db800dc836595a57815f967185e3596615c646f2455ac3914a.
//
// Solidity: event SessionClosed(address indexed userAddress, bytes32 indexed sessionId, address indexed providerId)
func (_SessionRouter *SessionRouterFilterer) ParseSessionClosed(log types.Log) (*SessionRouterSessionClosed, error) {
	event := new(SessionRouterSessionClosed)
	if err := _SessionRouter.contract.UnpackLog(event, "SessionClosed", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// SessionRouterSessionOpenedIterator is returned from FilterSessionOpened and is used to iterate over the raw logs and unpacked data for SessionOpened events raised by the SessionRouter contract.
type SessionRouterSessionOpenedIterator struct {
	Event *SessionRouterSessionOpened // Event containing the contract specifics and raw log

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
func (it *SessionRouterSessionOpenedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SessionRouterSessionOpened)
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
		it.Event = new(SessionRouterSessionOpened)
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
func (it *SessionRouterSessionOpenedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SessionRouterSessionOpenedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SessionRouterSessionOpened represents a SessionOpened event raised by the SessionRouter contract.
type SessionRouterSessionOpened struct {
	UserAddress common.Address
	SessionId   [32]byte
	ProviderId  common.Address
	Raw         types.Log // Blockchain specific contextual infos
}

// FilterSessionOpened is a free log retrieval operation binding the contract event 0x2bd7c890baf595977d256a6e784512c873ac58ba612b4895dbb7f784bfbf4839.
//
// Solidity: event SessionOpened(address indexed userAddress, bytes32 indexed sessionId, address indexed providerId)
func (_SessionRouter *SessionRouterFilterer) FilterSessionOpened(opts *bind.FilterOpts, userAddress []common.Address, sessionId [][32]byte, providerId []common.Address) (*SessionRouterSessionOpenedIterator, error) {

	var userAddressRule []interface{}
	for _, userAddressItem := range userAddress {
		userAddressRule = append(userAddressRule, userAddressItem)
	}
	var sessionIdRule []interface{}
	for _, sessionIdItem := range sessionId {
		sessionIdRule = append(sessionIdRule, sessionIdItem)
	}
	var providerIdRule []interface{}
	for _, providerIdItem := range providerId {
		providerIdRule = append(providerIdRule, providerIdItem)
	}

	logs, sub, err := _SessionRouter.contract.FilterLogs(opts, "SessionOpened", userAddressRule, sessionIdRule, providerIdRule)
	if err != nil {
		return nil, err
	}
	return &SessionRouterSessionOpenedIterator{contract: _SessionRouter.contract, event: "SessionOpened", logs: logs, sub: sub}, nil
}

// WatchSessionOpened is a free log subscription operation binding the contract event 0x2bd7c890baf595977d256a6e784512c873ac58ba612b4895dbb7f784bfbf4839.
//
// Solidity: event SessionOpened(address indexed userAddress, bytes32 indexed sessionId, address indexed providerId)
func (_SessionRouter *SessionRouterFilterer) WatchSessionOpened(opts *bind.WatchOpts, sink chan<- *SessionRouterSessionOpened, userAddress []common.Address, sessionId [][32]byte, providerId []common.Address) (event.Subscription, error) {

	var userAddressRule []interface{}
	for _, userAddressItem := range userAddress {
		userAddressRule = append(userAddressRule, userAddressItem)
	}
	var sessionIdRule []interface{}
	for _, sessionIdItem := range sessionId {
		sessionIdRule = append(sessionIdRule, sessionIdItem)
	}
	var providerIdRule []interface{}
	for _, providerIdItem := range providerId {
		providerIdRule = append(providerIdRule, providerIdItem)
	}

	logs, sub, err := _SessionRouter.contract.WatchLogs(opts, "SessionOpened", userAddressRule, sessionIdRule, providerIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SessionRouterSessionOpened)
				if err := _SessionRouter.contract.UnpackLog(event, "SessionOpened", log); err != nil {
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

// ParseSessionOpened is a log parse operation binding the contract event 0x2bd7c890baf595977d256a6e784512c873ac58ba612b4895dbb7f784bfbf4839.
//
// Solidity: event SessionOpened(address indexed userAddress, bytes32 indexed sessionId, address indexed providerId)
func (_SessionRouter *SessionRouterFilterer) ParseSessionOpened(log types.Log) (*SessionRouterSessionOpened, error) {
	event := new(SessionRouterSessionOpened)
	if err := _SessionRouter.contract.UnpackLog(event, "SessionOpened", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// SessionRouterStakedIterator is returned from FilterStaked and is used to iterate over the raw logs and unpacked data for Staked events raised by the SessionRouter contract.
type SessionRouterStakedIterator struct {
	Event *SessionRouterStaked // Event containing the contract specifics and raw log

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
func (it *SessionRouterStakedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SessionRouterStaked)
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
		it.Event = new(SessionRouterStaked)
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
func (it *SessionRouterStakedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SessionRouterStakedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SessionRouterStaked represents a Staked event raised by the SessionRouter contract.
type SessionRouterStaked struct {
	UserAddress common.Address
	Amount      *big.Int
	Raw         types.Log // Blockchain specific contextual infos
}

// FilterStaked is a free log retrieval operation binding the contract event 0x9e71bc8eea02a63969f509818f2dafb9254532904319f9dbda79b67bd34a5f3d.
//
// Solidity: event Staked(address indexed userAddress, uint256 amount)
func (_SessionRouter *SessionRouterFilterer) FilterStaked(opts *bind.FilterOpts, userAddress []common.Address) (*SessionRouterStakedIterator, error) {

	var userAddressRule []interface{}
	for _, userAddressItem := range userAddress {
		userAddressRule = append(userAddressRule, userAddressItem)
	}

	logs, sub, err := _SessionRouter.contract.FilterLogs(opts, "Staked", userAddressRule)
	if err != nil {
		return nil, err
	}
	return &SessionRouterStakedIterator{contract: _SessionRouter.contract, event: "Staked", logs: logs, sub: sub}, nil
}

// WatchStaked is a free log subscription operation binding the contract event 0x9e71bc8eea02a63969f509818f2dafb9254532904319f9dbda79b67bd34a5f3d.
//
// Solidity: event Staked(address indexed userAddress, uint256 amount)
func (_SessionRouter *SessionRouterFilterer) WatchStaked(opts *bind.WatchOpts, sink chan<- *SessionRouterStaked, userAddress []common.Address) (event.Subscription, error) {

	var userAddressRule []interface{}
	for _, userAddressItem := range userAddress {
		userAddressRule = append(userAddressRule, userAddressItem)
	}

	logs, sub, err := _SessionRouter.contract.WatchLogs(opts, "Staked", userAddressRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SessionRouterStaked)
				if err := _SessionRouter.contract.UnpackLog(event, "Staked", log); err != nil {
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
func (_SessionRouter *SessionRouterFilterer) ParseStaked(log types.Log) (*SessionRouterStaked, error) {
	event := new(SessionRouterStaked)
	if err := _SessionRouter.contract.UnpackLog(event, "Staked", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// SessionRouterUnstakedIterator is returned from FilterUnstaked and is used to iterate over the raw logs and unpacked data for Unstaked events raised by the SessionRouter contract.
type SessionRouterUnstakedIterator struct {
	Event *SessionRouterUnstaked // Event containing the contract specifics and raw log

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
func (it *SessionRouterUnstakedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SessionRouterUnstaked)
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
		it.Event = new(SessionRouterUnstaked)
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
func (it *SessionRouterUnstakedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SessionRouterUnstakedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SessionRouterUnstaked represents a Unstaked event raised by the SessionRouter contract.
type SessionRouterUnstaked struct {
	UserAddress common.Address
	Amount      *big.Int
	Raw         types.Log // Blockchain specific contextual infos
}

// FilterUnstaked is a free log retrieval operation binding the contract event 0x0f5bb82176feb1b5e747e28471aa92156a04d9f3ab9f45f28e2d704232b93f75.
//
// Solidity: event Unstaked(address indexed userAddress, uint256 amount)
func (_SessionRouter *SessionRouterFilterer) FilterUnstaked(opts *bind.FilterOpts, userAddress []common.Address) (*SessionRouterUnstakedIterator, error) {

	var userAddressRule []interface{}
	for _, userAddressItem := range userAddress {
		userAddressRule = append(userAddressRule, userAddressItem)
	}

	logs, sub, err := _SessionRouter.contract.FilterLogs(opts, "Unstaked", userAddressRule)
	if err != nil {
		return nil, err
	}
	return &SessionRouterUnstakedIterator{contract: _SessionRouter.contract, event: "Unstaked", logs: logs, sub: sub}, nil
}

// WatchUnstaked is a free log subscription operation binding the contract event 0x0f5bb82176feb1b5e747e28471aa92156a04d9f3ab9f45f28e2d704232b93f75.
//
// Solidity: event Unstaked(address indexed userAddress, uint256 amount)
func (_SessionRouter *SessionRouterFilterer) WatchUnstaked(opts *bind.WatchOpts, sink chan<- *SessionRouterUnstaked, userAddress []common.Address) (event.Subscription, error) {

	var userAddressRule []interface{}
	for _, userAddressItem := range userAddress {
		userAddressRule = append(userAddressRule, userAddressItem)
	}

	logs, sub, err := _SessionRouter.contract.WatchLogs(opts, "Unstaked", userAddressRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SessionRouterUnstaked)
				if err := _SessionRouter.contract.UnpackLog(event, "Unstaked", log); err != nil {
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
func (_SessionRouter *SessionRouterFilterer) ParseUnstaked(log types.Log) (*SessionRouterUnstaked, error) {
	event := new(SessionRouterUnstaked)
	if err := _SessionRouter.contract.UnpackLog(event, "Unstaked", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
